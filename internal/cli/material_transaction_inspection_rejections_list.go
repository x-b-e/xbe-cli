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

type materialTransactionInspectionRejectionsListOptions struct {
	BaseURL                       string
	Token                         string
	JSON                          bool
	NoAuth                        bool
	Limit                         int
	Offset                        int
	Sort                          string
	MaterialTransactionInspection string
	CreatedAtMin                  string
	CreatedAtMax                  string
	IsCreatedAt                   string
	UpdatedAtMin                  string
	UpdatedAtMax                  string
	IsUpdatedAt                   string
}

type materialTransactionInspectionRejectionRow struct {
	ID                              string `json:"id"`
	MaterialTransactionInspectionID string `json:"material_transaction_inspection_id,omitempty"`
	UnitOfMeasure                   string `json:"unit_of_measure,omitempty"`
	UnitOfMeasureID                 string `json:"unit_of_measure_id,omitempty"`
	Quantity                        string `json:"quantity,omitempty"`
	Note                            string `json:"note,omitempty"`
	RejectedByName                  string `json:"rejected_by_name,omitempty"`
	CreatedAt                       string `json:"created_at,omitempty"`
	UpdatedAt                       string `json:"updated_at,omitempty"`
}

func newMaterialTransactionInspectionRejectionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List material transaction inspection rejections",
		Long: `List material transaction inspection rejections.

Output Columns:
  ID           Rejection identifier
  INSPECTION   Material transaction inspection ID
  QTY          Rejected quantity
  UOM          Unit of measure
  REJECTED BY  Rejected by name (when available)
  NOTE         Rejection note

Filters:
  --material-transaction-inspection  Filter by material transaction inspection ID
  --created-at-min                   Filter by created-at on/after (ISO 8601)
  --created-at-max                   Filter by created-at on/before (ISO 8601)
  --is-created-at                    Filter by presence of created-at (true/false)
  --updated-at-min                   Filter by updated-at on/after (ISO 8601)
  --updated-at-max                   Filter by updated-at on/before (ISO 8601)
  --is-updated-at                    Filter by presence of updated-at (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List inspection rejections
  xbe view material-transaction-inspection-rejections list

  # Filter by inspection
  xbe view material-transaction-inspection-rejections list --material-transaction-inspection 123

  # Filter by created-at range
  xbe view material-transaction-inspection-rejections list \\
    --created-at-min 2026-01-23T00:00:00Z \\
    --created-at-max 2026-01-24T00:00:00Z

  # Output as JSON
  xbe view material-transaction-inspection-rejections list --json`,
		Args: cobra.NoArgs,
		RunE: runMaterialTransactionInspectionRejectionsList,
	}
	initMaterialTransactionInspectionRejectionsListFlags(cmd)
	return cmd
}

func init() {
	materialTransactionInspectionRejectionsCmd.AddCommand(newMaterialTransactionInspectionRejectionsListCmd())
}

func initMaterialTransactionInspectionRejectionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("material-transaction-inspection", "", "Filter by material transaction inspection ID")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by presence of created-at (true/false)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("is-updated-at", "", "Filter by presence of updated-at (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialTransactionInspectionRejectionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMaterialTransactionInspectionRejectionsListOptions(cmd)
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
	query.Set("fields[material-transaction-inspection-rejections]", "quantity,note,rejected-by-name,created-at,updated-at,material-transaction-inspection,unit-of-measure")
	query.Set("fields[unit-of-measures]", "name,abbreviation")
	query.Set("include", "unit-of-measure")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[material_transaction_inspection]", opts.MaterialTransactionInspection)
	setFilterIfPresent(query, "filter[created_at_min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created_at_max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[is_created_at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[updated_at_min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated_at_max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[is_updated_at]", opts.IsUpdatedAt)

	body, _, err := client.Get(cmd.Context(), "/v1/material-transaction-inspection-rejections", query)
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

	rows := buildMaterialTransactionInspectionRejectionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderMaterialTransactionInspectionRejectionsTable(cmd, rows)
}

func parseMaterialTransactionInspectionRejectionsListOptions(cmd *cobra.Command) (materialTransactionInspectionRejectionsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	materialTransactionInspection, _ := cmd.Flags().GetString("material-transaction-inspection")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	isCreatedAt, _ := cmd.Flags().GetString("is-created-at")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	isUpdatedAt, _ := cmd.Flags().GetString("is-updated-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialTransactionInspectionRejectionsListOptions{
		BaseURL:                       baseURL,
		Token:                         token,
		JSON:                          jsonOut,
		NoAuth:                        noAuth,
		Limit:                         limit,
		Offset:                        offset,
		Sort:                          sort,
		MaterialTransactionInspection: materialTransactionInspection,
		CreatedAtMin:                  createdAtMin,
		CreatedAtMax:                  createdAtMax,
		IsCreatedAt:                   isCreatedAt,
		UpdatedAtMin:                  updatedAtMin,
		UpdatedAtMax:                  updatedAtMax,
		IsUpdatedAt:                   isUpdatedAt,
	}, nil
}

func buildMaterialTransactionInspectionRejectionRows(resp jsonAPIResponse) []materialTransactionInspectionRejectionRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]materialTransactionInspectionRejectionRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildMaterialTransactionInspectionRejectionRow(resource, included))
	}

	return rows
}

func buildMaterialTransactionInspectionRejectionRow(resource jsonAPIResource, included map[string]jsonAPIResource) materialTransactionInspectionRejectionRow {
	attrs := resource.Attributes
	row := materialTransactionInspectionRejectionRow{
		ID:             resource.ID,
		Quantity:       strings.TrimSpace(stringAttr(attrs, "quantity")),
		Note:           stringAttr(attrs, "note"),
		RejectedByName: stringAttr(attrs, "rejected-by-name"),
		CreatedAt:      formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:      formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resource.Relationships["material-transaction-inspection"]; ok && rel.Data != nil {
		row.MaterialTransactionInspectionID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["unit-of-measure"]; ok && rel.Data != nil {
		row.UnitOfMeasureID = rel.Data.ID
		if uom, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.UnitOfMeasure = unitOfMeasureLabel(uom.Attributes)
		}
	}

	return row
}

func buildMaterialTransactionInspectionRejectionRowFromSingle(resp jsonAPISingleResponse) materialTransactionInspectionRejectionRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	return buildMaterialTransactionInspectionRejectionRow(resp.Data, included)
}

func renderMaterialTransactionInspectionRejectionsTable(cmd *cobra.Command, rows []materialTransactionInspectionRejectionRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No material transaction inspection rejections found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tINSPECTION\tQTY\tUOM\tREJECTED BY\tNOTE")
	for _, row := range rows {
		uom := firstNonEmpty(row.UnitOfMeasure, row.UnitOfMeasureID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.MaterialTransactionInspectionID,
			row.Quantity,
			truncateString(uom, 8),
			truncateString(row.RejectedByName, 16),
			truncateString(row.Note, 32),
		)
	}

	return writer.Flush()
}
