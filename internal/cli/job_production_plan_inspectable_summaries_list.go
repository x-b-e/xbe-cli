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

type jobProductionPlanInspectableSummariesListOptions struct {
	BaseURL    string
	Token      string
	JSON       bool
	NoAuth     bool
	Limit      int
	Offset     int
	Sort       string
	Developer  string
	Project    string
	JobNumber  string
	StartOn    string
	StartOnMin string
	StartOnMax string
}

type jobProductionPlanInspectableSummaryRow struct {
	ID                    string `json:"id"`
	JobNumber             string `json:"job_number,omitempty"`
	JobName               string `json:"job_name,omitempty"`
	StartOn               string `json:"start_on,omitempty"`
	StartTime             string `json:"start_time,omitempty"`
	JobSiteName           string `json:"job_site_name,omitempty"`
	Project               string `json:"project,omitempty"`
	CurrentUserCanInspect bool   `json:"current_user_can_inspect"`
}

func newJobProductionPlanInspectableSummariesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List job production plan inspectable summaries",
		Long: `List job production plan inspectable summaries with filtering and pagination.

Output Columns:
  ID        Job production plan ID
  JOB_NUM   Job number
  JOB_NAME  Job name
  START_ON  Start date
  START_AT  Start time
  JOB_SITE  Job site name
  PROJECT   Project name or number
  INSPECT   Current user can inspect (yes/no)

Filters:
  --developer    Filter by developer ID
  --project      Filter by project ID
  --job-number   Filter by job number
  --start-on     Filter by start date (YYYY-MM-DD)
  --start-on-min Filter by minimum start date (YYYY-MM-DD)
  --start-on-max Filter by maximum start date (YYYY-MM-DD)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List inspectable summaries
  xbe view job-production-plan-inspectable-summaries list

  # Filter by developer
  xbe view job-production-plan-inspectable-summaries list --developer 123

  # Filter by date range
  xbe view job-production-plan-inspectable-summaries list --start-on-min 2025-01-01 --start-on-max 2025-01-31

  # Output as JSON
  xbe view job-production-plan-inspectable-summaries list --json`,
		Args: cobra.NoArgs,
		RunE: runJobProductionPlanInspectableSummariesList,
	}
	initJobProductionPlanInspectableSummariesListFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanInspectableSummariesCmd.AddCommand(newJobProductionPlanInspectableSummariesListCmd())
}

func initJobProductionPlanInspectableSummariesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("developer", "", "Filter by developer ID")
	cmd.Flags().String("project", "", "Filter by project ID")
	cmd.Flags().String("job-number", "", "Filter by job number")
	cmd.Flags().String("start-on", "", "Filter by start date (YYYY-MM-DD)")
	cmd.Flags().String("start-on-min", "", "Filter by minimum start date (YYYY-MM-DD)")
	cmd.Flags().String("start-on-max", "", "Filter by maximum start date (YYYY-MM-DD)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanInspectableSummariesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseJobProductionPlanInspectableSummariesListOptions(cmd)
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
	query.Set("fields[job-production-plan-inspectable-summaries]", "job-number,job-name,start-on,start-time,job-site-name,project-name,project-number,current-user-can-inspect")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[developer]", opts.Developer)
	setFilterIfPresent(query, "filter[project]", opts.Project)
	setFilterIfPresent(query, "filter[job-number]", opts.JobNumber)
	setFilterIfPresent(query, "filter[start-on]", opts.StartOn)
	setFilterIfPresent(query, "filter[start-on-min]", opts.StartOnMin)
	setFilterIfPresent(query, "filter[start-on-max]", opts.StartOnMax)

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-inspectable-summaries", query)
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

	rows := buildJobProductionPlanInspectableSummaryRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderJobProductionPlanInspectableSummariesTable(cmd, rows)
}

func parseJobProductionPlanInspectableSummariesListOptions(cmd *cobra.Command) (jobProductionPlanInspectableSummariesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	developer, _ := cmd.Flags().GetString("developer")
	project, _ := cmd.Flags().GetString("project")
	jobNumber, _ := cmd.Flags().GetString("job-number")
	startOn, _ := cmd.Flags().GetString("start-on")
	startOnMin, _ := cmd.Flags().GetString("start-on-min")
	startOnMax, _ := cmd.Flags().GetString("start-on-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanInspectableSummariesListOptions{
		BaseURL:    baseURL,
		Token:      token,
		JSON:       jsonOut,
		NoAuth:     noAuth,
		Limit:      limit,
		Offset:     offset,
		Sort:       sort,
		Developer:  developer,
		Project:    project,
		JobNumber:  jobNumber,
		StartOn:    startOn,
		StartOnMin: startOnMin,
		StartOnMax: startOnMax,
	}, nil
}

func buildJobProductionPlanInspectableSummaryRows(resp jsonAPIResponse) []jobProductionPlanInspectableSummaryRow {
	rows := make([]jobProductionPlanInspectableSummaryRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		projectNumber := stringAttr(attrs, "project-number")
		projectName := stringAttr(attrs, "project-name")
		row := jobProductionPlanInspectableSummaryRow{
			ID:                    resource.ID,
			JobNumber:             stringAttr(attrs, "job-number"),
			JobName:               stringAttr(attrs, "job-name"),
			StartOn:               formatDate(stringAttr(attrs, "start-on")),
			StartTime:             stringAttr(attrs, "start-time"),
			JobSiteName:           stringAttr(attrs, "job-site-name"),
			Project:               firstNonEmpty(projectNumber, projectName),
			CurrentUserCanInspect: boolAttr(attrs, "current-user-can-inspect"),
		}
		rows = append(rows, row)
	}
	return rows
}

func renderJobProductionPlanInspectableSummariesTable(cmd *cobra.Command, rows []jobProductionPlanInspectableSummaryRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No job production plan inspectable summaries found.")
		return nil
	}

	const (
		maxJobName = 28
		maxJobSite = 22
		maxProject = 22
	)

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tJOB_NUM\tJOB_NAME\tSTART_ON\tSTART_AT\tJOB_SITE\tPROJECT\tINSPECT")
	for _, row := range rows {
		inspect := "no"
		if row.CurrentUserCanInspect {
			inspect = "yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.JobNumber,
			truncateString(row.JobName, maxJobName),
			row.StartOn,
			row.StartTime,
			truncateString(row.JobSiteName, maxJobSite),
			truncateString(row.Project, maxProject),
			inspect,
		)
	}
	return writer.Flush()
}
