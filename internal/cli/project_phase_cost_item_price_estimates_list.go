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

type projectPhaseCostItemPriceEstimatesListOptions struct {
	BaseURL              string
	Token                string
	JSON                 bool
	NoAuth               bool
	Limit                int
	Offset               int
	Sort                 string
	ProjectPhaseCostItem string
	ProjectEstimateSet   string
	CreatedBy            string
}

type projectPhaseCostItemPriceEstimateRow struct {
	ID                     string `json:"id"`
	ProjectPhaseCostItemID string `json:"project_phase_cost_item_id,omitempty"`
	ProjectEstimateSetID   string `json:"project_estimate_set_id,omitempty"`
	CreatedByID            string `json:"created_by_id,omitempty"`
	Estimate               any    `json:"estimate,omitempty"`
}

func newProjectPhaseCostItemPriceEstimatesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project phase cost item price estimates",
		Long: `List project phase cost item price estimates.

Output Columns:
  ID            Price estimate identifier
  COST_ITEM     Project phase cost item ID
  ESTIMATE_SET  Project estimate set ID
  CREATED_BY    Creator user ID
  ESTIMATE      Estimate summary

Filters:
  --project-phase-cost-item  Filter by project phase cost item ID
  --project-estimate-set     Filter by project estimate set ID
  --created-by               Filter by creator user ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List price estimates
  xbe view project-phase-cost-item-price-estimates list

  # Filter by cost item
  xbe view project-phase-cost-item-price-estimates list --project-phase-cost-item 123

  # Filter by estimate set
  xbe view project-phase-cost-item-price-estimates list --project-estimate-set 456

  # Filter by creator
  xbe view project-phase-cost-item-price-estimates list --created-by 789

  # Output as JSON
  xbe view project-phase-cost-item-price-estimates list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectPhaseCostItemPriceEstimatesList,
	}
	initProjectPhaseCostItemPriceEstimatesListFlags(cmd)
	return cmd
}

func init() {
	projectPhaseCostItemPriceEstimatesCmd.AddCommand(newProjectPhaseCostItemPriceEstimatesListCmd())
}

func initProjectPhaseCostItemPriceEstimatesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("project-phase-cost-item", "", "Filter by project phase cost item ID")
	cmd.Flags().String("project-estimate-set", "", "Filter by project estimate set ID")
	cmd.Flags().String("created-by", "", "Filter by creator user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectPhaseCostItemPriceEstimatesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectPhaseCostItemPriceEstimatesListOptions(cmd)
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
	query.Set("fields[project-phase-cost-item-price-estimates]", "estimate,project-phase-cost-item,project-estimate-set,created-by")
	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[project-phase-cost-item]", opts.ProjectPhaseCostItem)
	setFilterIfPresent(query, "filter[project-estimate-set]", opts.ProjectEstimateSet)
	setFilterIfPresent(query, "filter[created-by]", opts.CreatedBy)

	body, _, err := client.Get(cmd.Context(), "/v1/project-phase-cost-item-price-estimates", query)
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

	rows := buildProjectPhaseCostItemPriceEstimateRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectPhaseCostItemPriceEstimatesTable(cmd, rows)
}

func parseProjectPhaseCostItemPriceEstimatesListOptions(cmd *cobra.Command) (projectPhaseCostItemPriceEstimatesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	projectPhaseCostItem, _ := cmd.Flags().GetString("project-phase-cost-item")
	projectEstimateSet, _ := cmd.Flags().GetString("project-estimate-set")
	createdBy, _ := cmd.Flags().GetString("created-by")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectPhaseCostItemPriceEstimatesListOptions{
		BaseURL:              baseURL,
		Token:                token,
		JSON:                 jsonOut,
		NoAuth:               noAuth,
		Limit:                limit,
		Offset:               offset,
		Sort:                 sort,
		ProjectPhaseCostItem: projectPhaseCostItem,
		ProjectEstimateSet:   projectEstimateSet,
		CreatedBy:            createdBy,
	}, nil
}

func buildProjectPhaseCostItemPriceEstimateRows(resp jsonAPIResponse) []projectPhaseCostItemPriceEstimateRow {
	rows := make([]projectPhaseCostItemPriceEstimateRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := projectPhaseCostItemPriceEstimateRow{
			ID:                     resource.ID,
			Estimate:               resource.Attributes["estimate"],
			ProjectPhaseCostItemID: relationshipIDFromMap(resource.Relationships, "project-phase-cost-item"),
			ProjectEstimateSetID:   relationshipIDFromMap(resource.Relationships, "project-estimate-set"),
			CreatedByID:            relationshipIDFromMap(resource.Relationships, "created-by"),
		}

		rows = append(rows, row)
	}
	return rows
}

func renderProjectPhaseCostItemPriceEstimatesTable(cmd *cobra.Command, rows []projectPhaseCostItemPriceEstimateRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project phase cost item price estimates found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tCOST_ITEM\tESTIMATE_SET\tCREATED_BY\tESTIMATE")
	for _, row := range rows {
		estimate := truncateString(estimateSummary(row.Estimate), 45)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.ProjectPhaseCostItemID,
			row.ProjectEstimateSetID,
			row.CreatedByID,
			estimate,
		)
	}
	return writer.Flush()
}

func estimateSummary(raw any) string {
	if raw == nil {
		return ""
	}

	attrs, ok := raw.(map[string]any)
	if !ok || attrs == nil {
		return fmt.Sprintf("%v", raw)
	}

	className := stringAttr(attrs, "class_name")
	switch className {
	case "NormalDistribution":
		mean := floatAttr(attrs, "mean")
		sd := floatAttr(attrs, "standard_deviation")
		return fmt.Sprintf("Normal(mu=%.2f, sd=%.2f)", mean, sd)
	case "TriangularDistribution":
		min := floatAttr(attrs, "minimum")
		mode := floatAttr(attrs, "mode")
		max := floatAttr(attrs, "maximum")
		return fmt.Sprintf("Triangular(min=%.2f, mode=%.2f, max=%.2f)", min, mode, max)
	default:
		if className != "" {
			return className
		}
	}

	return fmt.Sprintf("%v", raw)
}
