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

type projectTransportPlanDriverAssignmentRecommendationsListOptions struct {
	BaseURL                    string
	Token                      string
	JSON                       bool
	NoAuth                     bool
	Limit                      int
	Offset                     int
	Sort                       string
	ProjectTransportPlanDriver string
}

type projectTransportPlanDriverAssignmentRecommendationRow struct {
	ID                           string `json:"id"`
	ProjectTransportPlanDriverID string `json:"project_transport_plan_driver_id,omitempty"`
	CandidateCount               int    `json:"candidate_count,omitempty"`
}

func newProjectTransportPlanDriverAssignmentRecommendationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project transport plan driver assignment recommendations",
		Long: `List project transport plan driver assignment recommendations.

Output Columns:
  ID          Recommendation identifier
  DRIVER      Project transport plan driver ID
  CANDIDATES  Candidate driver count

Filters:
  --project-transport-plan-driver  Filter by project transport plan driver ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List recommendations
  xbe view project-transport-plan-driver-assignment-recommendations list

  # Filter by project transport plan driver
  xbe view project-transport-plan-driver-assignment-recommendations list --project-transport-plan-driver 123

  # JSON output
  xbe view project-transport-plan-driver-assignment-recommendations list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectTransportPlanDriverAssignmentRecommendationsList,
	}
	initProjectTransportPlanDriverAssignmentRecommendationsListFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanDriverAssignmentRecommendationsCmd.AddCommand(newProjectTransportPlanDriverAssignmentRecommendationsListCmd())
}

func initProjectTransportPlanDriverAssignmentRecommendationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("project-transport-plan-driver", "", "Filter by project transport plan driver ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanDriverAssignmentRecommendationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectTransportPlanDriverAssignmentRecommendationsListOptions(cmd)
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
	query.Set("fields[project-transport-plan-driver-assignment-recommendations]", "candidates,project-transport-plan-driver")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[project-transport-plan-driver]", opts.ProjectTransportPlanDriver)

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-driver-assignment-recommendations", query)
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

	rows := buildProjectTransportPlanDriverAssignmentRecommendationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectTransportPlanDriverAssignmentRecommendationsTable(cmd, rows)
}

func parseProjectTransportPlanDriverAssignmentRecommendationsListOptions(cmd *cobra.Command) (projectTransportPlanDriverAssignmentRecommendationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	projectTransportPlanDriver, _ := cmd.Flags().GetString("project-transport-plan-driver")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanDriverAssignmentRecommendationsListOptions{
		BaseURL:                    baseURL,
		Token:                      token,
		JSON:                       jsonOut,
		NoAuth:                     noAuth,
		Limit:                      limit,
		Offset:                     offset,
		Sort:                       sort,
		ProjectTransportPlanDriver: projectTransportPlanDriver,
	}, nil
}

func buildProjectTransportPlanDriverAssignmentRecommendationRows(resp jsonAPIResponse) []projectTransportPlanDriverAssignmentRecommendationRow {
	rows := make([]projectTransportPlanDriverAssignmentRecommendationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := projectTransportPlanDriverAssignmentRecommendationRow{
			ID:             resource.ID,
			CandidateCount: countRecommendationCandidates(resource.Attributes),
		}

		if rel, ok := resource.Relationships["project-transport-plan-driver"]; ok && rel.Data != nil {
			row.ProjectTransportPlanDriverID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func countRecommendationCandidates(attrs map[string]any) int {
	if attrs == nil {
		return 0
	}
	value, ok := attrs["candidates"]
	if !ok || value == nil {
		return 0
	}
	switch typed := value.(type) {
	case []any:
		return len(typed)
	case []string:
		return len(typed)
	default:
		return 0
	}
}

func renderProjectTransportPlanDriverAssignmentRecommendationsTable(cmd *cobra.Command, rows []projectTransportPlanDriverAssignmentRecommendationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project transport plan driver assignment recommendations found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tDRIVER\tCANDIDATES")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%d\n",
			row.ID,
			row.ProjectTransportPlanDriverID,
			row.CandidateCount,
		)
	}
	return writer.Flush()
}
