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

type projectRevenueItemQuantityEstimatesListOptions struct {
	BaseURL            string
	Token              string
	JSON               bool
	NoAuth             bool
	Limit              int
	Offset             int
	Sort               string
	ProjectRevenueItem string
	ProjectEstimateSet string
	CreatedBy          string
}

type projectRevenueItemQuantityEstimateRow struct {
	ID                   string         `json:"id"`
	ProjectRevenueItemID string         `json:"project_revenue_item_id,omitempty"`
	ProjectEstimateSetID string         `json:"project_estimate_set_id,omitempty"`
	CreatedByID          string         `json:"created_by_id,omitempty"`
	Description          string         `json:"description,omitempty"`
	Estimate             map[string]any `json:"estimate,omitempty"`
}

func newProjectRevenueItemQuantityEstimatesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project revenue item quantity estimates",
		Long: `List quantity estimates for project revenue items.

Output Columns:
  ID            Quantity estimate identifier
  REVENUE ITEM  Project revenue item ID
  ESTIMATE SET  Project estimate set ID
  CREATED BY    User ID who created the estimate
  DESCRIPTION   Estimate description
  ESTIMATE      Estimate summary

Filters:
  --project-revenue-item  Filter by project revenue item ID
  --project-estimate-set        Filter by project estimate set ID
  --created-by                  Filter by creator user ID

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List quantity estimates
  xbe view project-revenue-item-quantity-estimates list

  # Filter by project estimate set
  xbe view project-revenue-item-quantity-estimates list --project-estimate-set 123

  # Filter by creator
  xbe view project-revenue-item-quantity-estimates list --created-by 456

  # Output as JSON
  xbe view project-revenue-item-quantity-estimates list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectRevenueItemQuantityEstimatesList,
	}
	initProjectRevenueItemQuantityEstimatesListFlags(cmd)
	return cmd
}

func init() {
	projectRevenueItemQuantityEstimatesCmd.AddCommand(newProjectRevenueItemQuantityEstimatesListCmd())
}

func initProjectRevenueItemQuantityEstimatesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("project-revenue-item", "", "Filter by project revenue item ID")
	cmd.Flags().String("project-estimate-set", "", "Filter by project estimate set ID")
	cmd.Flags().String("created-by", "", "Filter by creator user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectRevenueItemQuantityEstimatesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectRevenueItemQuantityEstimatesListOptions(cmd)
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
	query.Set("fields[project-revenue-item-quantity-estimates]", "description,estimate,project-revenue-item,project-estimate-set,created-by")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[project-revenue-item]", opts.ProjectRevenueItem)
	setFilterIfPresent(query, "filter[project-estimate-set]", opts.ProjectEstimateSet)
	setFilterIfPresent(query, "filter[created-by]", opts.CreatedBy)

	body, _, err := client.Get(cmd.Context(), "/v1/project-revenue-item-quantity-estimates", query)
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

	rows := buildProjectRevenueItemQuantityEstimateRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectRevenueItemQuantityEstimatesTable(cmd, rows)
}

func parseProjectRevenueItemQuantityEstimatesListOptions(cmd *cobra.Command) (projectRevenueItemQuantityEstimatesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	projectRevenueItem, _ := cmd.Flags().GetString("project-revenue-item")
	projectEstimateSet, _ := cmd.Flags().GetString("project-estimate-set")
	createdBy, _ := cmd.Flags().GetString("created-by")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectRevenueItemQuantityEstimatesListOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		NoAuth:             noAuth,
		Limit:              limit,
		Offset:             offset,
		Sort:               sort,
		ProjectRevenueItem: projectRevenueItem,
		ProjectEstimateSet: projectEstimateSet,
		CreatedBy:          createdBy,
	}, nil
}

func buildProjectRevenueItemQuantityEstimateRows(resp jsonAPIResponse) []projectRevenueItemQuantityEstimateRow {
	rows := make([]projectRevenueItemQuantityEstimateRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := buildProjectRevenueItemQuantityEstimateRow(resource)
		rows = append(rows, row)
	}
	return rows
}

func projectRevenueItemQuantityEstimateRowFromSingle(resp jsonAPISingleResponse) projectRevenueItemQuantityEstimateRow {
	return buildProjectRevenueItemQuantityEstimateRow(resp.Data)
}

func buildProjectRevenueItemQuantityEstimateRow(resource jsonAPIResource) projectRevenueItemQuantityEstimateRow {
	row := projectRevenueItemQuantityEstimateRow{
		ID:          resource.ID,
		Description: stringAttr(resource.Attributes, "description"),
		Estimate:    estimateAttr(resource.Attributes, "estimate"),
	}

	if rel, ok := resource.Relationships["project-revenue-item"]; ok && rel.Data != nil {
		row.ProjectRevenueItemID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["project-estimate-set"]; ok && rel.Data != nil {
		row.ProjectEstimateSetID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		row.CreatedByID = rel.Data.ID
	}

	return row
}

func renderProjectRevenueItemQuantityEstimatesTable(cmd *cobra.Command, rows []projectRevenueItemQuantityEstimateRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project revenue item quantity estimates found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tREVENUE ITEM\tESTIMATE SET\tCREATED BY\tDESCRIPTION\tESTIMATE")
	for _, row := range rows {
		estimate := estimateSummary(row.Estimate)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.ProjectRevenueItemID, 20),
			truncateString(row.ProjectEstimateSetID, 20),
			truncateString(row.CreatedByID, 20),
			truncateString(row.Description, 24),
			truncateString(estimate, 40),
		)
	}
	return writer.Flush()
}
