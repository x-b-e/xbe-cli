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

type materialTypeUnavailabilitiesListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	MaterialType string
	StartAtMin   string
	StartAtMax   string
	EndAtMin     string
	EndAtMax     string
}

type materialTypeUnavailabilityRow struct {
	ID             string `json:"id"`
	MaterialTypeID string `json:"material_type_id,omitempty"`
	MaterialType   string `json:"material_type,omitempty"`
	StartAt        string `json:"start_at,omitempty"`
	EndAt          string `json:"end_at,omitempty"`
	Description    string `json:"description,omitempty"`
}

func newMaterialTypeUnavailabilitiesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List material type unavailabilities",
		Long: `List material type unavailabilities.

Output Columns:
  ID            Material type unavailability identifier
  MATERIAL TYPE Material type name (or ID)
  START AT      Start timestamp
  END AT        End timestamp
  DESCRIPTION   Description

Filters:
  --material-type  Filter by material type ID
  --start-at-min   Filter by minimum start time
  --start-at-max   Filter by maximum start time
  --end-at-min     Filter by minimum end time
  --end-at-max     Filter by maximum end time

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List material type unavailabilities
  xbe view material-type-unavailabilities list

  # Filter by material type
  xbe view material-type-unavailabilities list --material-type 123

  # Filter by time window
  xbe view material-type-unavailabilities list --start-at-min 2025-01-01T00:00:00Z --end-at-max 2025-02-01T00:00:00Z

  # Output as JSON
  xbe view material-type-unavailabilities list --json`,
		Args: cobra.NoArgs,
		RunE: runMaterialTypeUnavailabilitiesList,
	}
	initMaterialTypeUnavailabilitiesListFlags(cmd)
	return cmd
}

func init() {
	materialTypeUnavailabilitiesCmd.AddCommand(newMaterialTypeUnavailabilitiesListCmd())
}

func initMaterialTypeUnavailabilitiesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("material-type", "", "Filter by material type ID")
	cmd.Flags().String("start-at-min", "", "Filter by minimum start time")
	cmd.Flags().String("start-at-max", "", "Filter by maximum start time")
	cmd.Flags().String("end-at-min", "", "Filter by minimum end time")
	cmd.Flags().String("end-at-max", "", "Filter by maximum end time")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialTypeUnavailabilitiesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMaterialTypeUnavailabilitiesListOptions(cmd)
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
	query.Set("fields[material-type-unavailabilities]", "start-at,end-at,description,material-type")
	query.Set("fields[material-types]", "name,display-name")
	query.Set("include", "material-type")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[material_type]", opts.MaterialType)
	setFilterIfPresent(query, "filter[start_at_min]", opts.StartAtMin)
	setFilterIfPresent(query, "filter[start_at_max]", opts.StartAtMax)
	setFilterIfPresent(query, "filter[end_at_min]", opts.EndAtMin)
	setFilterIfPresent(query, "filter[end_at_max]", opts.EndAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/material-type-unavailabilities", query)
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

	rows := buildMaterialTypeUnavailabilityRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderMaterialTypeUnavailabilitiesTable(cmd, rows)
}

func parseMaterialTypeUnavailabilitiesListOptions(cmd *cobra.Command) (materialTypeUnavailabilitiesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	materialType, _ := cmd.Flags().GetString("material-type")
	startAtMin, _ := cmd.Flags().GetString("start-at-min")
	startAtMax, _ := cmd.Flags().GetString("start-at-max")
	endAtMin, _ := cmd.Flags().GetString("end-at-min")
	endAtMax, _ := cmd.Flags().GetString("end-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialTypeUnavailabilitiesListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		MaterialType: materialType,
		StartAtMin:   startAtMin,
		StartAtMax:   startAtMax,
		EndAtMin:     endAtMin,
		EndAtMax:     endAtMax,
	}, nil
}

func buildMaterialTypeUnavailabilityRows(resp jsonAPIResponse) []materialTypeUnavailabilityRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]materialTypeUnavailabilityRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := materialTypeUnavailabilityRow{
			ID:          resource.ID,
			StartAt:     formatDateTime(stringAttr(attrs, "start-at")),
			EndAt:       formatDateTime(stringAttr(attrs, "end-at")),
			Description: stringAttr(attrs, "description"),
		}

		if rel, ok := resource.Relationships["material-type"]; ok && rel.Data != nil {
			row.MaterialTypeID = rel.Data.ID
			if materialType, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.MaterialType = firstNonEmpty(
					stringAttr(materialType.Attributes, "display-name"),
					stringAttr(materialType.Attributes, "name"),
				)
			}
		}

		rows = append(rows, row)
	}

	return rows
}

func renderMaterialTypeUnavailabilitiesTable(cmd *cobra.Command, rows []materialTypeUnavailabilityRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No material type unavailabilities found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tMATERIAL TYPE\tSTART AT\tEND AT\tDESCRIPTION")
	for _, row := range rows {
		materialType := row.MaterialType
		if materialType == "" {
			materialType = row.MaterialTypeID
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(materialType, 32),
			row.StartAt,
			row.EndAt,
			truncateString(row.Description, 40),
		)
	}
	writer.Flush()
	return nil
}
