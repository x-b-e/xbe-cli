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

type projectTransportPlanTrailerAssignmentRecommendationsListOptions struct {
	BaseURL                     string
	Token                       string
	JSON                        bool
	NoAuth                      bool
	Limit                       int
	Offset                      int
	Sort                        string
	ProjectTransportPlanTrailer string
}

type projectTransportPlanTrailerAssignmentRecommendationRow struct {
	ID                            string `json:"id"`
	ProjectTransportPlanTrailerID string `json:"project_transport_plan_trailer_id,omitempty"`
	CandidateCount                int    `json:"candidate_count,omitempty"`
}

func newProjectTransportPlanTrailerAssignmentRecommendationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project transport plan trailer assignment recommendations",
		Long: `List project transport plan trailer assignment recommendations.

Output Columns:
  ID          Recommendation identifier
  TRAILER     Project transport plan trailer ID
  CANDIDATES  Candidate trailer count

Filters:
  --project-transport-plan-trailer  Filter by project transport plan trailer ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List recommendations
  xbe view project-transport-plan-trailer-assignment-recommendations list

  # Filter by project transport plan trailer
  xbe view project-transport-plan-trailer-assignment-recommendations list --project-transport-plan-trailer 123

  # JSON output
  xbe view project-transport-plan-trailer-assignment-recommendations list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectTransportPlanTrailerAssignmentRecommendationsList,
	}
	initProjectTransportPlanTrailerAssignmentRecommendationsListFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanTrailerAssignmentRecommendationsCmd.AddCommand(newProjectTransportPlanTrailerAssignmentRecommendationsListCmd())
}

func initProjectTransportPlanTrailerAssignmentRecommendationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("project-transport-plan-trailer", "", "Filter by project transport plan trailer ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanTrailerAssignmentRecommendationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectTransportPlanTrailerAssignmentRecommendationsListOptions(cmd)
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
	query.Set("fields[project-transport-plan-trailer-assignment-recommendations]", "candidates,project-transport-plan-trailer")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[project-transport-plan-trailer]", opts.ProjectTransportPlanTrailer)

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-trailer-assignment-recommendations", query)
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

	rows := buildProjectTransportPlanTrailerAssignmentRecommendationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectTransportPlanTrailerAssignmentRecommendationsTable(cmd, rows)
}

func parseProjectTransportPlanTrailerAssignmentRecommendationsListOptions(cmd *cobra.Command) (projectTransportPlanTrailerAssignmentRecommendationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	projectTransportPlanTrailer, _ := cmd.Flags().GetString("project-transport-plan-trailer")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanTrailerAssignmentRecommendationsListOptions{
		BaseURL:                     baseURL,
		Token:                       token,
		JSON:                        jsonOut,
		NoAuth:                      noAuth,
		Limit:                       limit,
		Offset:                      offset,
		Sort:                        sort,
		ProjectTransportPlanTrailer: projectTransportPlanTrailer,
	}, nil
}

func buildProjectTransportPlanTrailerAssignmentRecommendationRows(resp jsonAPIResponse) []projectTransportPlanTrailerAssignmentRecommendationRow {
	rows := make([]projectTransportPlanTrailerAssignmentRecommendationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := projectTransportPlanTrailerAssignmentRecommendationRow{
			ID:             resource.ID,
			CandidateCount: countRecommendationCandidates(resource.Attributes),
		}

		if rel, ok := resource.Relationships["project-transport-plan-trailer"]; ok && rel.Data != nil {
			row.ProjectTransportPlanTrailerID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderProjectTransportPlanTrailerAssignmentRecommendationsTable(cmd *cobra.Command, rows []projectTransportPlanTrailerAssignmentRecommendationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project transport plan trailer assignment recommendations found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTRAILER\tCANDIDATES")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%d\n",
			row.ID,
			row.ProjectTransportPlanTrailerID,
			row.CandidateCount,
		)
	}
	return writer.Flush()
}
