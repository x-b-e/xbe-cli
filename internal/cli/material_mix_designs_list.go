package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type materialMixDesignsListOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	NoAuth           bool
	Limit            int
	Offset           int
	MaterialType     string
	MaterialSupplier string
	Broker           string
	Description      string
	DescriptionLike  string
	StartAtMin       string
	StartAtMax       string
	EndAtMin         string
	EndAtMax         string
	AsOf             string
}

type materialMixDesignRow struct {
	ID                 string `json:"id"`
	Description        string `json:"description,omitempty"`
	Mix                string `json:"mix,omitempty"`
	StartAt            string `json:"start_at,omitempty"`
	EndAt              string `json:"end_at,omitempty"`
	Notes              string `json:"notes,omitempty"`
	MaterialTypeID     string `json:"material_type_id,omitempty"`
	MaterialSupplierID string `json:"material_supplier_id,omitempty"`
	BrokerID           string `json:"broker_id,omitempty"`
}

func newMaterialMixDesignsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List material mix designs",
		Long: `List material mix designs.

Output Columns:
  ID                  Material mix design identifier
  DESCRIPTION         Description
  MIX                 Mix identifier
  MATERIAL TYPE       Material type ID
  SUPPLIER            Material supplier ID

Filters:
  --material-type      Filter by material type ID
  --material-supplier  Filter by material supplier ID
  --broker             Filter by broker ID
  --description        Filter by exact description
  --description-like   Filter by description (partial match)
  --start-at-min       Filter by minimum start time
  --start-at-max       Filter by maximum start time
  --end-at-min         Filter by minimum end time
  --end-at-max         Filter by maximum end time
  --as-of              Filter as of date`,
		Example: `  # List all material mix designs
  xbe view material-mix-designs list

  # Filter by material type
  xbe view material-mix-designs list --material-type 123

  # Filter by description
  xbe view material-mix-designs list --description-like "concrete"

  # Output as JSON
  xbe view material-mix-designs list --json`,
		RunE: runMaterialMixDesignsList,
	}
	initMaterialMixDesignsListFlags(cmd)
	return cmd
}

func init() {
	materialMixDesignsCmd.AddCommand(newMaterialMixDesignsListCmd())
}

func initMaterialMixDesignsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("material-type", "", "Filter by material type ID")
	cmd.Flags().String("material-supplier", "", "Filter by material supplier ID")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("description", "", "Filter by exact description")
	cmd.Flags().String("description-like", "", "Filter by description (partial match)")
	cmd.Flags().String("start-at-min", "", "Filter by minimum start time")
	cmd.Flags().String("start-at-max", "", "Filter by maximum start time")
	cmd.Flags().String("end-at-min", "", "Filter by minimum end time")
	cmd.Flags().String("end-at-max", "", "Filter by maximum end time")
	cmd.Flags().String("as-of", "", "Filter as of date")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialMixDesignsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMaterialMixDesignsListOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run 'xbe auth login' first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("include", "material-type,material-supplier")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	setFilterIfPresent(query, "filter[material_type]", opts.MaterialType)
	setFilterIfPresent(query, "filter[material_supplier]", opts.MaterialSupplier)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[description]", opts.Description)
	setFilterIfPresent(query, "filter[description_like]", opts.DescriptionLike)
	setFilterIfPresent(query, "filter[start_at_min]", opts.StartAtMin)
	setFilterIfPresent(query, "filter[start_at_max]", opts.StartAtMax)
	setFilterIfPresent(query, "filter[end_at_min]", opts.EndAtMin)
	setFilterIfPresent(query, "filter[end_at_max]", opts.EndAtMax)
	setFilterIfPresent(query, "filter[as_of]", opts.AsOf)

	body, _, err := client.Get(cmd.Context(), "/v1/material-mix-designs", query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	handled, err := renderSparseListIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	rows := buildMaterialMixDesignRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderMaterialMixDesignsTable(cmd, rows)
}

func parseMaterialMixDesignsListOptions(cmd *cobra.Command) (materialMixDesignsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	materialType, _ := cmd.Flags().GetString("material-type")
	materialSupplier, _ := cmd.Flags().GetString("material-supplier")
	broker, _ := cmd.Flags().GetString("broker")
	description, _ := cmd.Flags().GetString("description")
	descriptionLike, _ := cmd.Flags().GetString("description-like")
	startAtMin, _ := cmd.Flags().GetString("start-at-min")
	startAtMax, _ := cmd.Flags().GetString("start-at-max")
	endAtMin, _ := cmd.Flags().GetString("end-at-min")
	endAtMax, _ := cmd.Flags().GetString("end-at-max")
	asOf, _ := cmd.Flags().GetString("as-of")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialMixDesignsListOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		NoAuth:           noAuth,
		Limit:            limit,
		Offset:           offset,
		MaterialType:     materialType,
		MaterialSupplier: materialSupplier,
		Broker:           broker,
		Description:      description,
		DescriptionLike:  descriptionLike,
		StartAtMin:       startAtMin,
		StartAtMax:       startAtMax,
		EndAtMin:         endAtMin,
		EndAtMax:         endAtMax,
		AsOf:             asOf,
	}, nil
}

func buildMaterialMixDesignRows(resp jsonAPIResponse) []materialMixDesignRow {
	rows := make([]materialMixDesignRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := materialMixDesignRow{
			ID:          resource.ID,
			Description: stringAttr(resource.Attributes, "description"),
			Mix:         stringAttr(resource.Attributes, "mix"),
			StartAt:     stringAttr(resource.Attributes, "start-at"),
			EndAt:       stringAttr(resource.Attributes, "end-at"),
			Notes:       stringAttr(resource.Attributes, "notes"),
		}

		if rel, ok := resource.Relationships["material-type"]; ok && rel.Data != nil {
			row.MaterialTypeID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["material-supplier"]; ok && rel.Data != nil {
			row.MaterialSupplierID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			row.BrokerID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderMaterialMixDesignsTable(cmd *cobra.Command, rows []materialMixDesignRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No material mix designs found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tDESCRIPTION\tMIX\tMATERIAL TYPE\tSUPPLIER")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Description, 30),
			row.Mix,
			row.MaterialTypeID,
			row.MaterialSupplierID,
		)
	}
	return writer.Flush()
}
