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

type projectRevenueItemsListOptions struct {
	BaseURL                   string
	Token                     string
	JSON                      bool
	NoAuth                    bool
	Limit                     int
	Offset                    int
	Sort                      string
	Project                   string
	RevenueClassification     string
	UnitOfMeasure             string
	DeveloperQuantityEstimate string
}

type projectRevenueItemRow struct {
	ID                             string `json:"id"`
	Description                    string `json:"description,omitempty"`
	ExternalDeveloperRevenueItemID string `json:"external_developer_revenue_item_id,omitempty"`
	DeveloperQuantityEstimate      string `json:"developer_quantity_estimate,omitempty"`
	ProjectID                      string `json:"project_id,omitempty"`
	ProjectName                    string `json:"project_name,omitempty"`
	ProjectNumber                  string `json:"project_number,omitempty"`
	RevenueClassificationID        string `json:"revenue_classification_id,omitempty"`
	RevenueClassificationName      string `json:"revenue_classification_name,omitempty"`
	UnitOfMeasureID                string `json:"unit_of_measure_id,omitempty"`
	UnitOfMeasureName              string `json:"unit_of_measure_name,omitempty"`
	UnitOfMeasureAbbreviation      string `json:"unit_of_measure_abbreviation,omitempty"`
}

func newProjectRevenueItemsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project revenue items",
		Long: `List project revenue items with filtering and pagination.

Project revenue items define billable line items for a project and map to
revenue classifications and units of measure.

Output Columns:
  ID             Project revenue item identifier
  PROJECT        Project name/number (or ID)
  CLASSIFICATION Revenue classification (name or ID)
  DESCRIPTION    Revenue item description
  UNIT           Unit of measure
  DEV QTY        Developer quantity estimate
  EXTERNAL ID    External developer revenue item ID

Filters:
  --project                     Filter by project ID
  --revenue-classification      Filter by revenue classification ID
  --unit-of-measure             Filter by unit of measure ID
  --developer-quantity-estimate Filter by developer quantity estimate

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List project revenue items
  xbe view project-revenue-items list

  # Filter by project
  xbe view project-revenue-items list --project 123

  # Filter by revenue classification
  xbe view project-revenue-items list --revenue-classification 456

  # Output as JSON
  xbe view project-revenue-items list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectRevenueItemsList,
	}
	initProjectRevenueItemsListFlags(cmd)
	return cmd
}

func init() {
	projectRevenueItemsCmd.AddCommand(newProjectRevenueItemsListCmd())
}

func initProjectRevenueItemsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("project", "", "Filter by project ID")
	cmd.Flags().String("revenue-classification", "", "Filter by revenue classification ID")
	cmd.Flags().String("unit-of-measure", "", "Filter by unit of measure ID")
	cmd.Flags().String("developer-quantity-estimate", "", "Filter by developer quantity estimate")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectRevenueItemsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectRevenueItemsListOptions(cmd)
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
	query.Set("fields[project-revenue-items]", "description,external-developer-revenue-item-id,developer-quantity-estimate,project,revenue-classification,unit-of-measure")
	query.Set("include", "project,revenue-classification,unit-of-measure")
	query.Set("fields[projects]", "name,number")
	query.Set("fields[project-revenue-classifications]", "name")
	query.Set("fields[unit-of-measures]", "name,abbreviation")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[project]", opts.Project)
	setFilterIfPresent(query, "filter[revenue-classification]", opts.RevenueClassification)
	setFilterIfPresent(query, "filter[unit-of-measure]", opts.UnitOfMeasure)
	setFilterIfPresent(query, "filter[developer-quantity-estimate]", opts.DeveloperQuantityEstimate)

	body, _, err := client.Get(cmd.Context(), "/v1/project-revenue-items", query)
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

	rows := buildProjectRevenueItemRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectRevenueItemsTable(cmd, rows)
}

func parseProjectRevenueItemsListOptions(cmd *cobra.Command) (projectRevenueItemsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	project, _ := cmd.Flags().GetString("project")
	revenueClassification, _ := cmd.Flags().GetString("revenue-classification")
	unitOfMeasure, _ := cmd.Flags().GetString("unit-of-measure")
	developerQuantityEstimate, _ := cmd.Flags().GetString("developer-quantity-estimate")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectRevenueItemsListOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		NoAuth:                    noAuth,
		Limit:                     limit,
		Offset:                    offset,
		Sort:                      sort,
		Project:                   project,
		RevenueClassification:     revenueClassification,
		UnitOfMeasure:             unitOfMeasure,
		DeveloperQuantityEstimate: developerQuantityEstimate,
	}, nil
}

func buildProjectRevenueItemRows(resp jsonAPIResponse) []projectRevenueItemRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]projectRevenueItemRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildProjectRevenueItemRow(resource, included))
	}
	return rows
}

func buildProjectRevenueItemRow(resource jsonAPIResource, included map[string]jsonAPIResource) projectRevenueItemRow {
	attrs := resource.Attributes
	row := projectRevenueItemRow{
		ID:                             resource.ID,
		Description:                    stringAttr(attrs, "description"),
		ExternalDeveloperRevenueItemID: stringAttr(attrs, "external-developer-revenue-item-id"),
		DeveloperQuantityEstimate:      stringAttr(attrs, "developer-quantity-estimate"),
	}

	if rel, ok := resource.Relationships["project"]; ok && rel.Data != nil {
		row.ProjectID = rel.Data.ID
		if included != nil {
			if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.ProjectName = stringAttr(inc.Attributes, "name")
				row.ProjectNumber = stringAttr(inc.Attributes, "number")
			}
		}
	}

	if rel, ok := resource.Relationships["revenue-classification"]; ok && rel.Data != nil {
		row.RevenueClassificationID = rel.Data.ID
		if included != nil {
			if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.RevenueClassificationName = stringAttr(inc.Attributes, "name")
			}
		}
	}

	if rel, ok := resource.Relationships["unit-of-measure"]; ok && rel.Data != nil {
		row.UnitOfMeasureID = rel.Data.ID
		if included != nil {
			if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.UnitOfMeasureName = stringAttr(inc.Attributes, "name")
				row.UnitOfMeasureAbbreviation = stringAttr(inc.Attributes, "abbreviation")
			}
		}
	}

	return row
}

func renderProjectRevenueItemsTable(cmd *cobra.Command, rows []projectRevenueItemRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project revenue items found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPROJECT\tCLASSIFICATION\tDESCRIPTION\tUNIT\tDEV QTY\tEXTERNAL ID")
	for _, row := range rows {
		projectLabel := firstNonEmpty(row.ProjectName, row.ProjectNumber, row.ProjectID)
		classificationLabel := firstNonEmpty(row.RevenueClassificationName, row.RevenueClassificationID)
		unitLabel := firstNonEmpty(row.UnitOfMeasureAbbreviation, row.UnitOfMeasureName, row.UnitOfMeasureID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			projectLabel,
			classificationLabel,
			truncateString(row.Description, 40),
			unitLabel,
			row.DeveloperQuantityEstimate,
			truncateString(row.ExternalDeveloperRevenueItemID, 24),
		)
	}
	return writer.Flush()
}
