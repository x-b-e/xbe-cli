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

type jobProductionPlanScheduleChangeWorksListOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	NoAuth            bool
	Limit             int
	Offset            int
	Sort              string
	JobProductionPlan string
	CreatedBy         string
	TimeKind          string
	Broker            string
	Customer          string
	Project           string
}

type jobProductionPlanScheduleChangeWorkRow struct {
	ID                  string `json:"id"`
	JobProductionPlanID string `json:"job_production_plan_id,omitempty"`
	CreatedByID         string `json:"created_by_id,omitempty"`
	TimeKind            string `json:"time_kind,omitempty"`
	OffsetSeconds       string `json:"offset_seconds,omitempty"`
	ScheduledAt         string `json:"scheduled_at,omitempty"`
	ProcessedAt         string `json:"processed_at,omitempty"`
}

func newJobProductionPlanScheduleChangeWorksListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List job production plan schedule change works",
		Long: `List job production plan schedule change works with filtering and pagination.

Output Columns:
  ID              Work identifier
  JOB_PLAN        Job production plan ID
  TIME_KIND       Time kind (both, material_site, job_site)
  OFFSET_SECONDS  Offset in seconds
  SCHEDULED_AT    Scheduled timestamp
  PROCESSED_AT    Processed timestamp
  CREATED_BY      Creator user ID

Filters:
  --job-production-plan  Filter by job production plan ID
  --created-by           Filter by creator user ID
  --time-kind            Filter by time kind (both, material_site, job_site)
  --broker               Filter by broker ID
  --customer             Filter by customer ID
  --project              Filter by project ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List schedule change works
  xbe view job-production-plan-schedule-change-works list

  # Filter by job production plan
  xbe view job-production-plan-schedule-change-works list --job-production-plan 123

  # Filter by time kind
  xbe view job-production-plan-schedule-change-works list --time-kind both

  # Output as JSON
  xbe view job-production-plan-schedule-change-works list --json`,
		Args: cobra.NoArgs,
		RunE: runJobProductionPlanScheduleChangeWorksList,
	}
	initJobProductionPlanScheduleChangeWorksListFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanScheduleChangeWorksCmd.AddCommand(newJobProductionPlanScheduleChangeWorksListCmd())
}

func initJobProductionPlanScheduleChangeWorksListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("job-production-plan", "", "Filter by job production plan ID")
	cmd.Flags().String("created-by", "", "Filter by creator user ID")
	cmd.Flags().String("time-kind", "", "Filter by time kind (both, material_site, job_site)")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("customer", "", "Filter by customer ID")
	cmd.Flags().String("project", "", "Filter by project ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanScheduleChangeWorksList(cmd *cobra.Command, _ []string) error {
	opts, err := parseJobProductionPlanScheduleChangeWorksListOptions(cmd)
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
	query.Set("fields[job-production-plan-schedule-change-works]", "offset-seconds,time-kind,scheduled-at,processed-at,job-production-plan,created-by")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[job-production-plan]", opts.JobProductionPlan)
	setFilterIfPresent(query, "filter[created-by]", opts.CreatedBy)
	setFilterIfPresent(query, "filter[time-kind]", opts.TimeKind)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[customer]", opts.Customer)
	setFilterIfPresent(query, "filter[project]", opts.Project)

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-schedule-change-works", query)
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

	rows := buildJobProductionPlanScheduleChangeWorkRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderJobProductionPlanScheduleChangeWorksTable(cmd, rows)
}

func parseJobProductionPlanScheduleChangeWorksListOptions(cmd *cobra.Command) (jobProductionPlanScheduleChangeWorksListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	createdBy, _ := cmd.Flags().GetString("created-by")
	timeKind, _ := cmd.Flags().GetString("time-kind")
	broker, _ := cmd.Flags().GetString("broker")
	customer, _ := cmd.Flags().GetString("customer")
	project, _ := cmd.Flags().GetString("project")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanScheduleChangeWorksListOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		NoAuth:            noAuth,
		Limit:             limit,
		Offset:            offset,
		Sort:              sort,
		JobProductionPlan: jobProductionPlan,
		CreatedBy:         createdBy,
		TimeKind:          timeKind,
		Broker:            broker,
		Customer:          customer,
		Project:           project,
	}, nil
}

func buildJobProductionPlanScheduleChangeWorkRows(resp jsonAPIResponse) []jobProductionPlanScheduleChangeWorkRow {
	rows := make([]jobProductionPlanScheduleChangeWorkRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := jobProductionPlanScheduleChangeWorkRow{
			ID:            resource.ID,
			TimeKind:      stringAttr(attrs, "time-kind"),
			OffsetSeconds: stringAttr(attrs, "offset-seconds"),
			ScheduledAt:   formatDateTime(stringAttr(attrs, "scheduled-at")),
			ProcessedAt:   formatDateTime(stringAttr(attrs, "processed-at")),
		}

		if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
			row.JobProductionPlanID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
			row.CreatedByID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderJobProductionPlanScheduleChangeWorksTable(cmd *cobra.Command, rows []jobProductionPlanScheduleChangeWorkRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No job production plan schedule change works found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tJOB_PLAN\tTIME_KIND\tOFFSET_SECONDS\tSCHEDULED_AT\tPROCESSED_AT\tCREATED_BY")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.JobProductionPlanID,
			row.TimeKind,
			row.OffsetSeconds,
			row.ScheduledAt,
			row.ProcessedAt,
			row.CreatedByID,
		)
	}
	return writer.Flush()
}
