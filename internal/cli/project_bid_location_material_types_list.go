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

type projectBidLocationMaterialTypesListOptions struct {
	BaseURL            string
	Token              string
	JSON               bool
	NoAuth             bool
	Limit              int
	Offset             int
	ProjectBidLocation string
	MaterialType       string
}

type projectBidLocationMaterialTypeRow struct {
	ID                        string `json:"id"`
	Quantity                  string `json:"quantity,omitempty"`
	Notes                     string `json:"notes,omitempty"`
	ProjectBidLocationID      string `json:"project_bid_location_id,omitempty"`
	ProjectBidLocationName    string `json:"project_bid_location_name,omitempty"`
	MaterialTypeID            string `json:"material_type_id,omitempty"`
	MaterialTypeName          string `json:"material_type_name,omitempty"`
	UnitOfMeasureID           string `json:"unit_of_measure_id,omitempty"`
	UnitOfMeasureName         string `json:"unit_of_measure_name,omitempty"`
	UnitOfMeasureAbbreviation string `json:"unit_of_measure_abbreviation,omitempty"`
}

func newProjectBidLocationMaterialTypesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project bid location material types",
		Long: `List project bid location material types with filtering and pagination.

Project bid location material types define planned quantities and notes for a
material type at a specific project bid location.

Output Columns:
  ID                     Project bid location material type identifier
  PROJECT BID LOCATION   Project bid location name (or ID)
  MATERIAL TYPE          Material type name (or ID)
  UNIT                   Unit of measure (abbreviation or name)
  QUANTITY               Planned quantity

Filters:
  --project-bid-location  Filter by project bid location ID
  --material-type         Filter by material type ID`,
		Example: `  # List all project bid location material types
  xbe view project-bid-location-material-types list

  # Filter by project bid location
  xbe view project-bid-location-material-types list --project-bid-location 123

  # Filter by material type
  xbe view project-bid-location-material-types list --material-type 456

  # Output as JSON
  xbe view project-bid-location-material-types list --json`,
		RunE: runProjectBidLocationMaterialTypesList,
	}
	initProjectBidLocationMaterialTypesListFlags(cmd)
	return cmd
}

func init() {
	projectBidLocationMaterialTypesCmd.AddCommand(newProjectBidLocationMaterialTypesListCmd())
}

func initProjectBidLocationMaterialTypesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("project-bid-location", "", "Filter by project bid location ID")
	cmd.Flags().String("material-type", "", "Filter by material type ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectBidLocationMaterialTypesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectBidLocationMaterialTypesListOptions(cmd)
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
	query.Set("fields[project-bid-location-material-types]", "quantity,notes,project-bid-location,material-type,unit-of-measure")
	query.Set("include", "project-bid-location,material-type,unit-of-measure")
	query.Set("fields[project-bid-locations]", "name")
	query.Set("fields[material-types]", "name")
	query.Set("fields[unit-of-measures]", "name,abbreviation")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[project-bid-location]", opts.ProjectBidLocation)
	setFilterIfPresent(query, "filter[material-type]", opts.MaterialType)

	body, _, err := client.Get(cmd.Context(), "/v1/project-bid-location-material-types", query)
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

	rows := buildProjectBidLocationMaterialTypeRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectBidLocationMaterialTypesTable(cmd, rows)
}

func parseProjectBidLocationMaterialTypesListOptions(cmd *cobra.Command) (projectBidLocationMaterialTypesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	projectBidLocation, _ := cmd.Flags().GetString("project-bid-location")
	materialType, _ := cmd.Flags().GetString("material-type")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectBidLocationMaterialTypesListOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		NoAuth:             noAuth,
		Limit:              limit,
		Offset:             offset,
		ProjectBidLocation: projectBidLocation,
		MaterialType:       materialType,
	}, nil
}

func buildProjectBidLocationMaterialTypeRows(resp jsonAPIResponse) []projectBidLocationMaterialTypeRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]projectBidLocationMaterialTypeRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := projectBidLocationMaterialTypeRow{
			ID:       resource.ID,
			Quantity: stringAttr(resource.Attributes, "quantity"),
			Notes:    stringAttr(resource.Attributes, "notes"),
		}

		if rel, ok := resource.Relationships["project-bid-location"]; ok && rel.Data != nil {
			row.ProjectBidLocationID = rel.Data.ID
			if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.ProjectBidLocationName = stringAttr(inc.Attributes, "name")
			}
		}

		if rel, ok := resource.Relationships["material-type"]; ok && rel.Data != nil {
			row.MaterialTypeID = rel.Data.ID
			if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.MaterialTypeName = stringAttr(inc.Attributes, "name")
			}
		}

		if rel, ok := resource.Relationships["unit-of-measure"]; ok && rel.Data != nil {
			row.UnitOfMeasureID = rel.Data.ID
			if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.UnitOfMeasureName = stringAttr(inc.Attributes, "name")
				row.UnitOfMeasureAbbreviation = stringAttr(inc.Attributes, "abbreviation")
			}
		}

		rows = append(rows, row)
	}

	return rows
}

func renderProjectBidLocationMaterialTypesTable(cmd *cobra.Command, rows []projectBidLocationMaterialTypeRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project bid location material types found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPROJECT BID LOCATION\tMATERIAL TYPE\tUNIT\tQUANTITY")
	for _, row := range rows {
		projectBidLocationLabel := firstNonEmpty(row.ProjectBidLocationName, row.ProjectBidLocationID)
		materialTypeLabel := firstNonEmpty(row.MaterialTypeName, row.MaterialTypeID)
		unitLabel := firstNonEmpty(row.UnitOfMeasureAbbreviation, row.UnitOfMeasureName, row.UnitOfMeasureID)

		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(projectBidLocationLabel, 26),
			truncateString(materialTypeLabel, 22),
			truncateString(unitLabel, 12),
			row.Quantity,
		)
	}
	writer.Flush()
	return nil
}
