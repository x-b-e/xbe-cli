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

type jobProductionPlanTimeCardApproversListOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	NoAuth            bool
	Limit             int
	Offset            int
	Sort              string
	JobProductionPlan string
	User              string
}

type jobProductionPlanTimeCardApproverRow struct {
	ID                  string `json:"id"`
	JobProductionPlanID string `json:"job_production_plan_id,omitempty"`
	JobProductionPlan   string `json:"job_production_plan,omitempty"`
	UserID              string `json:"user_id,omitempty"`
	User                string `json:"user,omitempty"`
	UserEmail           string `json:"user_email,omitempty"`
	UserMobile          string `json:"user_mobile,omitempty"`
}

func newJobProductionPlanTimeCardApproversListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List job production plan time card approvers",
		Long: `List job production plan time card approvers with filtering and pagination.

Output Columns:
  ID       Time card approver identifier
  PLAN     Job production plan (job number/name)
  USER     Approver user name
  EMAIL    Approver email address

Filters:
  --job-production-plan  Filter by job production plan ID
  --user                 Filter by user ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List time card approvers
  xbe view job-production-plan-time-card-approvers list

  # Filter by job production plan
  xbe view job-production-plan-time-card-approvers list --job-production-plan 123

  # Filter by user
  xbe view job-production-plan-time-card-approvers list --user 456

  # JSON output
  xbe view job-production-plan-time-card-approvers list --json`,
		Args: cobra.NoArgs,
		RunE: runJobProductionPlanTimeCardApproversList,
	}
	initJobProductionPlanTimeCardApproversListFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanTimeCardApproversCmd.AddCommand(newJobProductionPlanTimeCardApproversListCmd())
}

func initJobProductionPlanTimeCardApproversListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("job-production-plan", "", "Filter by job production plan ID")
	cmd.Flags().String("user", "", "Filter by user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanTimeCardApproversList(cmd *cobra.Command, _ []string) error {
	opts, err := parseJobProductionPlanTimeCardApproversListOptions(cmd)
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
	query.Set("fields[job-production-plan-time-card-approvers]", "job-production-plan,user")
	query.Set("fields[job-production-plans]", "job-number,job-name")
	query.Set("fields[users]", "name,email-address,mobile-number")
	query.Set("include", "job-production-plan,user")

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
	setFilterIfPresent(query, "filter[user]", opts.User)

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-time-card-approvers", query)
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

	rows := buildJobProductionPlanTimeCardApproverRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderJobProductionPlanTimeCardApproversTable(cmd, rows)
}

func parseJobProductionPlanTimeCardApproversListOptions(cmd *cobra.Command) (jobProductionPlanTimeCardApproversListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	user, _ := cmd.Flags().GetString("user")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanTimeCardApproversListOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		NoAuth:            noAuth,
		Limit:             limit,
		Offset:            offset,
		Sort:              sort,
		JobProductionPlan: jobProductionPlan,
		User:              user,
	}, nil
}

func buildJobProductionPlanTimeCardApproverRows(resp jsonAPIResponse) []jobProductionPlanTimeCardApproverRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]jobProductionPlanTimeCardApproverRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := jobProductionPlanTimeCardApproverRow{
			ID: resource.ID,
		}

		if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
			row.JobProductionPlanID = rel.Data.ID
			if plan, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				jobNumber := stringAttr(plan.Attributes, "job-number")
				jobName := stringAttr(plan.Attributes, "job-name")
				if jobNumber != "" && jobName != "" {
					row.JobProductionPlan = fmt.Sprintf("%s - %s", jobNumber, jobName)
				} else {
					row.JobProductionPlan = firstNonEmpty(jobNumber, jobName)
				}
			}
		}

		if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
			row.UserID = rel.Data.ID
			if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.User = strings.TrimSpace(stringAttr(user.Attributes, "name"))
				row.UserEmail = strings.TrimSpace(stringAttr(user.Attributes, "email-address"))
				row.UserMobile = strings.TrimSpace(stringAttr(user.Attributes, "mobile-number"))
				if row.User == "" {
					row.User = row.UserEmail
				}
			}
		}

		rows = append(rows, row)
	}

	return rows
}

func renderJobProductionPlanTimeCardApproversTable(cmd *cobra.Command, rows []jobProductionPlanTimeCardApproverRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No job production plan time card approvers found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPLAN\tUSER\tEMAIL")
	for _, row := range rows {
		plan := row.JobProductionPlan
		if plan == "" {
			plan = row.JobProductionPlanID
		}
		user := row.User
		if user == "" {
			user = row.UserID
		}

		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(plan, 28),
			truncateString(user, 22),
			truncateString(row.UserEmail, 28),
		)
	}
	return writer.Flush()
}
