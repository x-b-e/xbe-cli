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

type projectPhaseRevenueItemsListOptions struct {
	BaseURL                      string
	Token                        string
	JSON                         bool
	NoAuth                       bool
	Limit                        int
	Offset                       int
	Sort                         string
	ProjectPhase                 string
	ProjectRevenueItem           string
	ProjectRevenueClassification string
	Project                      string
	QuantityStrategy             string
}

type projectPhaseRevenueItemRow struct {
	ID                             string `json:"id"`
	ProjectPhaseID                 string `json:"project_phase_id,omitempty"`
	ProjectRevenueItemID           string `json:"project_revenue_item_id,omitempty"`
	ProjectRevenueClassificationID string `json:"project_revenue_classification_id,omitempty"`
	QuantityStrategy               string `json:"quantity_strategy,omitempty"`
	Note                           string `json:"note,omitempty"`
}

func newProjectPhaseRevenueItemsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project phase revenue items",
		Long: `List project phase revenue items.

Project phase revenue items link project phases to revenue items and define how quantities roll up.

Output Columns:
  ID        Revenue item identifier
  PHASE     Project phase ID
  REV_ITEM  Project revenue item ID
  REV_CLASS Project revenue classification ID
  STRATEGY  Quantity strategy (direct/indirect)
  NOTE      Item note

Filters:
  --project-phase                  Filter by project phase ID
  --project-revenue-item           Filter by project revenue item ID
  --project-revenue-classification Filter by project revenue classification ID
  --project                        Filter by project ID
  --quantity-strategy              Filter by quantity strategy (direct/indirect)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List project phase revenue items
  xbe view project-phase-revenue-items list

  # Filter by project phase
  xbe view project-phase-revenue-items list --project-phase 123

  # Filter by quantity strategy
  xbe view project-phase-revenue-items list --quantity-strategy indirect

  # Output as JSON
  xbe view project-phase-revenue-items list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectPhaseRevenueItemsList,
	}
	initProjectPhaseRevenueItemsListFlags(cmd)
	return cmd
}

func init() {
	projectPhaseRevenueItemsCmd.AddCommand(newProjectPhaseRevenueItemsListCmd())
}

func initProjectPhaseRevenueItemsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("project-phase", "", "Filter by project phase ID")
	cmd.Flags().String("project-revenue-item", "", "Filter by project revenue item ID")
	cmd.Flags().String("project-revenue-classification", "", "Filter by project revenue classification ID")
	cmd.Flags().String("project", "", "Filter by project ID")
	cmd.Flags().String("quantity-strategy", "", "Filter by quantity strategy (direct/indirect)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectPhaseRevenueItemsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectPhaseRevenueItemsListOptions(cmd)
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
	query.Set("fields[project-phase-revenue-items]", "quantity-strategy,note,project-phase,project-revenue-item,project-revenue-classification")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[project-phase]", opts.ProjectPhase)
	setFilterIfPresent(query, "filter[project-revenue-item]", opts.ProjectRevenueItem)
	setFilterIfPresent(query, "filter[project-revenue-classification]", opts.ProjectRevenueClassification)
	setFilterIfPresent(query, "filter[project]", opts.Project)
	setFilterIfPresent(query, "filter[quantity-strategy]", opts.QuantityStrategy)

	body, _, err := client.Get(cmd.Context(), "/v1/project-phase-revenue-items", query)
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

	rows := buildProjectPhaseRevenueItemRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectPhaseRevenueItemsTable(cmd, rows)
}

func parseProjectPhaseRevenueItemsListOptions(cmd *cobra.Command) (projectPhaseRevenueItemsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	projectPhase, _ := cmd.Flags().GetString("project-phase")
	projectRevenueItem, _ := cmd.Flags().GetString("project-revenue-item")
	projectRevenueClassification, _ := cmd.Flags().GetString("project-revenue-classification")
	project, _ := cmd.Flags().GetString("project")
	quantityStrategy, _ := cmd.Flags().GetString("quantity-strategy")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectPhaseRevenueItemsListOptions{
		BaseURL:                      baseURL,
		Token:                        token,
		JSON:                         jsonOut,
		NoAuth:                       noAuth,
		Limit:                        limit,
		Offset:                       offset,
		Sort:                         sort,
		ProjectPhase:                 projectPhase,
		ProjectRevenueItem:           projectRevenueItem,
		ProjectRevenueClassification: projectRevenueClassification,
		Project:                      project,
		QuantityStrategy:             quantityStrategy,
	}, nil
}

func buildProjectPhaseRevenueItemRows(resp jsonAPIResponse) []projectPhaseRevenueItemRow {
	rows := make([]projectPhaseRevenueItemRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := projectPhaseRevenueItemRow{
			ID:                             resource.ID,
			QuantityStrategy:               stringAttr(resource.Attributes, "quantity-strategy"),
			Note:                           stringAttr(resource.Attributes, "note"),
			ProjectPhaseID:                 relationshipIDFromMap(resource.Relationships, "project-phase"),
			ProjectRevenueItemID:           relationshipIDFromMap(resource.Relationships, "project-revenue-item"),
			ProjectRevenueClassificationID: relationshipIDFromMap(resource.Relationships, "project-revenue-classification"),
		}

		rows = append(rows, row)
	}
	return rows
}

func buildProjectPhaseRevenueItemRowFromSingle(resp jsonAPISingleResponse) projectPhaseRevenueItemRow {
	resource := resp.Data
	return projectPhaseRevenueItemRow{
		ID:                             resource.ID,
		QuantityStrategy:               stringAttr(resource.Attributes, "quantity-strategy"),
		Note:                           stringAttr(resource.Attributes, "note"),
		ProjectPhaseID:                 relationshipIDFromMap(resource.Relationships, "project-phase"),
		ProjectRevenueItemID:           relationshipIDFromMap(resource.Relationships, "project-revenue-item"),
		ProjectRevenueClassificationID: relationshipIDFromMap(resource.Relationships, "project-revenue-classification"),
	}
}

func renderProjectPhaseRevenueItemsTable(cmd *cobra.Command, rows []projectPhaseRevenueItemRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project phase revenue items found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPHASE\tREV_ITEM\tREV_CLASS\tSTRATEGY\tNOTE")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.ProjectPhaseID, 15),
			truncateString(row.ProjectRevenueItemID, 15),
			truncateString(row.ProjectRevenueClassificationID, 15),
			row.QuantityStrategy,
			truncateString(row.Note, 30),
		)
	}
	return writer.Flush()
}
