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

type lineupJobScheduleShiftTruckerAssignmentRecommendationsListOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	NoAuth                 bool
	Limit                  int
	Offset                 int
	Sort                   string
	LineupJobScheduleShift string
	CreatedAtMin           string
	CreatedAtMax           string
	UpdatedAtMin           string
	UpdatedAtMax           string
}

type lineupJobScheduleShiftTruckerAssignmentRecommendationRow struct {
	ID                       string `json:"id"`
	LineupJobScheduleShiftID string `json:"lineup_job_schedule_shift_id,omitempty"`
	CandidatesCount          int    `json:"candidates_count"`
}

func newLineupJobScheduleShiftTruckerAssignmentRecommendationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List lineup job schedule shift trucker assignment recommendations",
		Long: `List lineup job schedule shift trucker assignment recommendations.

Output Columns:
  ID             Recommendation identifier
  LINEUP SHIFT   Lineup job schedule shift ID
  CANDIDATES     Number of ranked truckers

Filters:
  --lineup-job-schedule-shift  Filter by lineup job schedule shift ID
  --created-at-min             Filter by created-at on/after (ISO 8601)
  --created-at-max             Filter by created-at on/before (ISO 8601)
  --updated-at-min             Filter by updated-at on/after (ISO 8601)
  --updated-at-max             Filter by updated-at on/before (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List recommendations
  xbe view lineup-job-schedule-shift-trucker-assignment-recommendations list

  # Filter by lineup job schedule shift
  xbe view lineup-job-schedule-shift-trucker-assignment-recommendations list --lineup-job-schedule-shift 123

  # Output as JSON
  xbe view lineup-job-schedule-shift-trucker-assignment-recommendations list --json`,
		Args: cobra.NoArgs,
		RunE: runLineupJobScheduleShiftTruckerAssignmentRecommendationsList,
	}
	initLineupJobScheduleShiftTruckerAssignmentRecommendationsListFlags(cmd)
	return cmd
}

func init() {
	lineupJobScheduleShiftTruckerAssignmentRecommendationsCmd.AddCommand(newLineupJobScheduleShiftTruckerAssignmentRecommendationsListCmd())
}

func initLineupJobScheduleShiftTruckerAssignmentRecommendationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("lineup-job-schedule-shift", "", "Filter by lineup job schedule shift ID")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runLineupJobScheduleShiftTruckerAssignmentRecommendationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseLineupJobScheduleShiftTruckerAssignmentRecommendationsListOptions(cmd)
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
	query.Set("fields[lineup-job-schedule-shift-trucker-assignment-recommendations]", "lineup-job-schedule-shift,candidates")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[lineup-job-schedule-shift]", opts.LineupJobScheduleShift)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/lineup-job-schedule-shift-trucker-assignment-recommendations", query)
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

	rows := buildLineupJobScheduleShiftTruckerAssignmentRecommendationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderLineupJobScheduleShiftTruckerAssignmentRecommendationsTable(cmd, rows)
}

func parseLineupJobScheduleShiftTruckerAssignmentRecommendationsListOptions(cmd *cobra.Command) (lineupJobScheduleShiftTruckerAssignmentRecommendationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	lineupJobScheduleShift, _ := cmd.Flags().GetString("lineup-job-schedule-shift")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return lineupJobScheduleShiftTruckerAssignmentRecommendationsListOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		NoAuth:                 noAuth,
		Limit:                  limit,
		Offset:                 offset,
		Sort:                   sort,
		LineupJobScheduleShift: lineupJobScheduleShift,
		CreatedAtMin:           createdAtMin,
		CreatedAtMax:           createdAtMax,
		UpdatedAtMin:           updatedAtMin,
		UpdatedAtMax:           updatedAtMax,
	}, nil
}

func buildLineupJobScheduleShiftTruckerAssignmentRecommendationRows(resp jsonAPIResponse) []lineupJobScheduleShiftTruckerAssignmentRecommendationRow {
	rows := make([]lineupJobScheduleShiftTruckerAssignmentRecommendationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildLineupJobScheduleShiftTruckerAssignmentRecommendationRow(resource))
	}
	return rows
}

func lineupJobScheduleShiftTruckerAssignmentRecommendationRowFromSingle(resp jsonAPISingleResponse) lineupJobScheduleShiftTruckerAssignmentRecommendationRow {
	return buildLineupJobScheduleShiftTruckerAssignmentRecommendationRow(resp.Data)
}

func buildLineupJobScheduleShiftTruckerAssignmentRecommendationRow(resource jsonAPIResource) lineupJobScheduleShiftTruckerAssignmentRecommendationRow {
	row := lineupJobScheduleShiftTruckerAssignmentRecommendationRow{
		ID:              resource.ID,
		CandidatesCount: candidateCountFromAny(resource.Attributes["candidates"]),
	}

	if rel, ok := resource.Relationships["lineup-job-schedule-shift"]; ok && rel.Data != nil {
		row.LineupJobScheduleShiftID = rel.Data.ID
	}

	return row
}

func renderLineupJobScheduleShiftTruckerAssignmentRecommendationsTable(cmd *cobra.Command, rows []lineupJobScheduleShiftTruckerAssignmentRecommendationRow) error {
	out := cmd.OutOrStdout()
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)

	fmt.Fprintln(w, "ID\tLINEUP SHIFT\tCANDIDATES")
	for _, row := range rows {
		lineupShift := row.LineupJobScheduleShiftID
		if lineupShift == "" {
			lineupShift = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%d\n", row.ID, lineupShift, row.CandidatesCount)
	}

	return w.Flush()
}

func candidateCountFromAny(value any) int {
	if value == nil {
		return 0
	}
	switch typed := value.(type) {
	case []any:
		return len(typed)
	case []map[string]any:
		return len(typed)
	default:
		return 1
	}
}
