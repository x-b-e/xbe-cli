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

type jobSiteTimesListOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	NoAuth              bool
	Limit               int
	Offset              int
	Sort                string
	JobProductionPlanID string
	UserID              string
	TimeCardID          string
	StartAtMin          string
	StartAtMax          string
	IsStartAt           string
	EndAtMin            string
	EndAtMax            string
	IsEndAt             string
	CreatedAtMin        string
	CreatedAtMax        string
	IsCreatedAt         string
	UpdatedAtMin        string
	UpdatedAtMax        string
	IsUpdatedAt         string
}

type jobSiteTimeRow struct {
	ID                  string  `json:"id"`
	JobProductionPlanID string  `json:"job_production_plan_id,omitempty"`
	UserID              string  `json:"user_id,omitempty"`
	StartAt             string  `json:"start_at,omitempty"`
	EndAt               string  `json:"end_at,omitempty"`
	Hours               float64 `json:"hours,omitempty"`
	Description         string  `json:"description,omitempty"`
}

func newJobSiteTimesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List job site times",
		Long: `List job site times.

Output Columns:
  ID        Job site time identifier
  PLAN      Job production plan ID
  USER      User ID
  START AT  Start timestamp
  END AT    End timestamp
  HOURS     Calculated duration in hours
  DESC      Description

Filters:
  --job-production-plan  Filter by job production plan ID
  --user                 Filter by user ID
  --time-card            Filter by time card ID
  --start-at-min         Filter by start-at on/after (ISO 8601)
  --start-at-max         Filter by start-at on/before (ISO 8601)
  --is-start-at          Filter by presence of start-at (true/false)
  --end-at-min           Filter by end-at on/after (ISO 8601)
  --end-at-max           Filter by end-at on/before (ISO 8601)
  --is-end-at            Filter by presence of end-at (true/false)
  --created-at-min       Filter by created-at on/after (ISO 8601)
  --created-at-max       Filter by created-at on/before (ISO 8601)
  --is-created-at        Filter by presence of created-at (true/false)
  --updated-at-min       Filter by updated-at on/after (ISO 8601)
  --updated-at-max       Filter by updated-at on/before (ISO 8601)
  --is-updated-at        Filter by presence of updated-at (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List job site times
  xbe view job-site-times list

  # Filter by job production plan
  xbe view job-site-times list --job-production-plan 123

  # Filter by user
  xbe view job-site-times list --user 456

  # Filter by time card
  xbe view job-site-times list --time-card 789

  # Filter by start-at range
  xbe view job-site-times list \
    --start-at-min 2026-01-22T00:00:00Z \
    --start-at-max 2026-01-23T00:00:00Z

  # Output as JSON
  xbe view job-site-times list --json`,
		Args: cobra.NoArgs,
		RunE: runJobSiteTimesList,
	}
	initJobSiteTimesListFlags(cmd)
	return cmd
}

func init() {
	jobSiteTimesCmd.AddCommand(newJobSiteTimesListCmd())
}

func initJobSiteTimesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("job-production-plan", "", "Filter by job production plan ID")
	cmd.Flags().String("user", "", "Filter by user ID")
	cmd.Flags().String("time-card", "", "Filter by time card ID")
	cmd.Flags().String("start-at-min", "", "Filter by start-at on/after (ISO 8601)")
	cmd.Flags().String("start-at-max", "", "Filter by start-at on/before (ISO 8601)")
	cmd.Flags().String("is-start-at", "", "Filter by presence of start-at (true/false)")
	cmd.Flags().String("end-at-min", "", "Filter by end-at on/after (ISO 8601)")
	cmd.Flags().String("end-at-max", "", "Filter by end-at on/before (ISO 8601)")
	cmd.Flags().String("is-end-at", "", "Filter by presence of end-at (true/false)")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by presence of created-at (true/false)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("is-updated-at", "", "Filter by presence of updated-at (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobSiteTimesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseJobSiteTimesListOptions(cmd)
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
	query.Set("fields[job-site-times]", "start-at,end-at,description,hours")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[job_production_plan]", opts.JobProductionPlanID)
	setFilterIfPresent(query, "filter[user]", opts.UserID)
	setFilterIfPresent(query, "filter[time_card]", opts.TimeCardID)
	setFilterIfPresent(query, "filter[start_at_min]", opts.StartAtMin)
	setFilterIfPresent(query, "filter[start_at_max]", opts.StartAtMax)
	setFilterIfPresent(query, "filter[is_start_at]", opts.IsStartAt)
	setFilterIfPresent(query, "filter[end_at_min]", opts.EndAtMin)
	setFilterIfPresent(query, "filter[end_at_max]", opts.EndAtMax)
	setFilterIfPresent(query, "filter[is_end_at]", opts.IsEndAt)
	setFilterIfPresent(query, "filter[created_at_min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created_at_max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[is_created_at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[updated_at_min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated_at_max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[is_updated_at]", opts.IsUpdatedAt)

	body, _, err := client.Get(cmd.Context(), "/v1/job-site-times", query)
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

	rows := buildJobSiteTimeRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderJobSiteTimesTable(cmd, rows)
}

func parseJobSiteTimesListOptions(cmd *cobra.Command) (jobSiteTimesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	jobProductionPlanID, _ := cmd.Flags().GetString("job-production-plan")
	userID, _ := cmd.Flags().GetString("user")
	timeCardID, _ := cmd.Flags().GetString("time-card")
	startAtMin, _ := cmd.Flags().GetString("start-at-min")
	startAtMax, _ := cmd.Flags().GetString("start-at-max")
	isStartAt, _ := cmd.Flags().GetString("is-start-at")
	endAtMin, _ := cmd.Flags().GetString("end-at-min")
	endAtMax, _ := cmd.Flags().GetString("end-at-max")
	isEndAt, _ := cmd.Flags().GetString("is-end-at")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	isCreatedAt, _ := cmd.Flags().GetString("is-created-at")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	isUpdatedAt, _ := cmd.Flags().GetString("is-updated-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobSiteTimesListOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		NoAuth:              noAuth,
		Limit:               limit,
		Offset:              offset,
		Sort:                sort,
		JobProductionPlanID: jobProductionPlanID,
		UserID:              userID,
		TimeCardID:          timeCardID,
		StartAtMin:          startAtMin,
		StartAtMax:          startAtMax,
		IsStartAt:           isStartAt,
		EndAtMin:            endAtMin,
		EndAtMax:            endAtMax,
		IsEndAt:             isEndAt,
		CreatedAtMin:        createdAtMin,
		CreatedAtMax:        createdAtMax,
		IsCreatedAt:         isCreatedAt,
		UpdatedAtMin:        updatedAtMin,
		UpdatedAtMax:        updatedAtMax,
		IsUpdatedAt:         isUpdatedAt,
	}, nil
}

func buildJobSiteTimeRows(resp jsonAPIResponse) []jobSiteTimeRow {
	rows := make([]jobSiteTimeRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, jobSiteTimeRowFromResource(resource))
	}
	return rows
}

func jobSiteTimeRowFromResource(resource jsonAPIResource) jobSiteTimeRow {
	attrs := resource.Attributes
	row := jobSiteTimeRow{
		ID:          resource.ID,
		StartAt:     formatDateTime(stringAttr(attrs, "start-at")),
		EndAt:       formatDateTime(stringAttr(attrs, "end-at")),
		Hours:       floatAttr(attrs, "hours"),
		Description: stringAttr(attrs, "description"),
	}

	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		row.JobProductionPlanID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
		row.UserID = rel.Data.ID
	}

	return row
}

func buildJobSiteTimeRowFromSingle(resp jsonAPISingleResponse) jobSiteTimeRow {
	return jobSiteTimeRowFromResource(resp.Data)
}

func renderJobSiteTimesTable(cmd *cobra.Command, rows []jobSiteTimeRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No job site times found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPLAN\tUSER\tSTART AT\tEND AT\tHOURS\tDESC")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.JobProductionPlanID,
			row.UserID,
			truncateString(row.StartAt, 19),
			truncateString(row.EndAt, 19),
			formatHours(row.Hours),
			truncateString(row.Description, 24),
		)
	}
	return writer.Flush()
}

func formatHours(value float64) string {
	if value == 0 {
		return ""
	}
	if value == float64(int(value)) {
		return fmt.Sprintf("%.0f", value)
	}
	return fmt.Sprintf("%.2f", value)
}
