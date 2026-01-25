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

type projectRevenueItemPriceEstimatesListOptions struct {
	BaseURL            string
	Token              string
	JSON               bool
	NoAuth             bool
	Limit              int
	Offset             int
	Sort               string
	ProjectRevenueItem string
	CreatedBy          string
	ProjectEstimateSet string
}

type projectRevenueItemPriceEstimateRow struct {
	ID                    string `json:"id"`
	ProjectRevenueItemID  string `json:"project_revenue_item_id,omitempty"`
	ProjectEstimateSetID  string `json:"project_estimate_set_id,omitempty"`
	CreatedByID           string `json:"created_by_id,omitempty"`
	Kind                  string `json:"kind,omitempty"`
	PricePerUnitExplicit  string `json:"price_per_unit_explicit,omitempty"`
	CostMultiplier        string `json:"cost_multiplier,omitempty"`
	PricePerUnit          string `json:"price_per_unit,omitempty"`
	PricePerUnitEffective string `json:"price_per_unit_effective,omitempty"`
}

func newProjectRevenueItemPriceEstimatesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project revenue item price estimates",
		Long: `List project revenue item price estimates with filtering and pagination.

Output Columns:
  ID            Price estimate identifier
  REV ITEM      Project revenue item ID
  EST SET       Project estimate set ID
  KIND          Estimate kind (explicit or cost_multiplier)
  PRICE/UNIT    Price per unit (resolved)
  EXPLICIT      Explicit price per unit
  MULTIPLIER    Cost multiplier

Filters:
  --project-revenue-item   Filter by project revenue item ID
  --project-estimate-set   Filter by project estimate set ID
  --created-by             Filter by created-by user ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List project revenue item price estimates
  xbe view project-revenue-item-price-estimates list

  # Filter by project revenue item
  xbe view project-revenue-item-price-estimates list --project-revenue-item 123

  # Filter by estimate set
  xbe view project-revenue-item-price-estimates list --project-estimate-set 456

  # Output as JSON
  xbe view project-revenue-item-price-estimates list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectRevenueItemPriceEstimatesList,
	}
	initProjectRevenueItemPriceEstimatesListFlags(cmd)
	return cmd
}

func init() {
	projectRevenueItemPriceEstimatesCmd.AddCommand(newProjectRevenueItemPriceEstimatesListCmd())
}

func initProjectRevenueItemPriceEstimatesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("project-revenue-item", "", "Filter by project revenue item ID")
	cmd.Flags().String("project-estimate-set", "", "Filter by project estimate set ID")
	cmd.Flags().String("created-by", "", "Filter by created-by user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectRevenueItemPriceEstimatesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectRevenueItemPriceEstimatesListOptions(cmd)
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
	query.Set("fields[project-revenue-item-price-estimates]", "kind,price-per-unit,price-per-unit-explicit,cost-multiplier,price-per-unit-effective,project-revenue-item,project-estimate-set,created-by")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[project-revenue-item]", opts.ProjectRevenueItem)
	setFilterIfPresent(query, "filter[project-estimate-set]", opts.ProjectEstimateSet)
	setFilterIfPresent(query, "filter[created-by]", opts.CreatedBy)

	body, _, err := client.Get(cmd.Context(), "/v1/project-revenue-item-price-estimates", query)
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

	rows := buildProjectRevenueItemPriceEstimateRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectRevenueItemPriceEstimatesTable(cmd, rows)
}

func parseProjectRevenueItemPriceEstimatesListOptions(cmd *cobra.Command) (projectRevenueItemPriceEstimatesListOptions, error) {
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

	return projectRevenueItemPriceEstimatesListOptions{
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

func buildProjectRevenueItemPriceEstimateRows(resp jsonAPIResponse) []projectRevenueItemPriceEstimateRow {
	rows := make([]projectRevenueItemPriceEstimateRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := projectRevenueItemPriceEstimateRow{
			ID:                    resource.ID,
			Kind:                  stringAttr(resource.Attributes, "kind"),
			PricePerUnitExplicit:  stringAttr(resource.Attributes, "price-per-unit-explicit"),
			CostMultiplier:        stringAttr(resource.Attributes, "cost-multiplier"),
			PricePerUnit:          stringAttr(resource.Attributes, "price-per-unit"),
			PricePerUnitEffective: stringAttr(resource.Attributes, "price-per-unit-effective"),
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

		rows = append(rows, row)
	}
	return rows
}

func renderProjectRevenueItemPriceEstimatesTable(cmd *cobra.Command, rows []projectRevenueItemPriceEstimateRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project revenue item price estimates found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tREV ITEM\tEST SET\tKIND\tPRICE/UNIT\tEXPLICIT\tMULTIPLIER")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.ProjectRevenueItemID,
			row.ProjectEstimateSetID,
			row.Kind,
			row.PricePerUnit,
			row.PricePerUnitExplicit,
			row.CostMultiplier,
		)
	}
	return writer.Flush()
}
