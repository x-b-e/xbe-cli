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

type jobProductionPlanStatusChangesListOptions struct {
	BaseURL                                 string
	Token                                   string
	JSON                                    bool
	NoAuth                                  bool
	Limit                                   int
	Offset                                  int
	Sort                                    string
	JobProductionPlan                       string
	JobProductionPlanCancellationReasonType string
	Status                                  string
	Customer                                string
	StartOn                                 string
	StartOnMin                              string
	StartOnMax                              string
	CreatedAtMin                            string
	CreatedAtMax                            string
	UpdatedAtMin                            string
	UpdatedAtMax                            string
}

type jobProductionPlanStatusChangeRow struct {
	ID                                        string `json:"id"`
	JobProductionPlanID                       string `json:"job_production_plan_id,omitempty"`
	JobProductionPlan                         string `json:"job_production_plan,omitempty"`
	Status                                    string `json:"status,omitempty"`
	Comment                                   string `json:"comment,omitempty"`
	ChangedAt                                 string `json:"changed_at,omitempty"`
	ChangedByID                               string `json:"changed_by_id,omitempty"`
	ChangedBy                                 string `json:"changed_by,omitempty"`
	JobProductionPlanCancellationReasonTypeID string `json:"job_production_plan_cancellation_reason_type_id,omitempty"`
	JobProductionPlanCancellationReasonType   string `json:"job_production_plan_cancellation_reason_type,omitempty"`
}

func newJobProductionPlanStatusChangesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List job production plan status changes",
		Long: `List job production plan status changes.

Output Columns:
  ID                     Status change identifier
  JOB PRODUCTION PLAN    Job production plan name/number
  STATUS                 New status
  CHANGED BY             User who changed the status
  CHANGED AT             When the status changed
  REASON                 Cancellation reason type (if applicable)
  COMMENT                Status change comment (if provided)

Filters:
  --job-production-plan                       Filter by job production plan ID
  --status                                    Filter by status (editing/submitted/rejected/approved/complete/abandoned/cancelled/scrapped)
  --job-production-plan-cancellation-reason-type  Filter by cancellation reason type ID
  --customer                                  Filter by customer ID
  --start-on                                  Filter by job production plan start date (YYYY-MM-DD)
  --start-on-min                              Filter by job production plan start date on/after (YYYY-MM-DD)
  --start-on-max                              Filter by job production plan start date on/before (YYYY-MM-DD)
  --created-at-min                            Filter by created-at on/after (ISO 8601)
  --created-at-max                            Filter by created-at on/before (ISO 8601)
  --updated-at-min                            Filter by updated-at on/after (ISO 8601)
  --updated-at-max                            Filter by updated-at on/before (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List recent status changes
  xbe view job-production-plan-status-changes list

  # Filter by job production plan
  xbe view job-production-plan-status-changes list --job-production-plan 123

  # Filter by status
  xbe view job-production-plan-status-changes list --status approved

  # Output as JSON
  xbe view job-production-plan-status-changes list --json`,
		Args: cobra.NoArgs,
		RunE: runJobProductionPlanStatusChangesList,
	}
	initJobProductionPlanStatusChangesListFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanStatusChangesCmd.AddCommand(newJobProductionPlanStatusChangesListCmd())
}

func initJobProductionPlanStatusChangesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("job-production-plan", "", "Filter by job production plan ID")
	cmd.Flags().String("status", "", "Filter by status")
	cmd.Flags().String("job-production-plan-cancellation-reason-type", "", "Filter by cancellation reason type ID")
	cmd.Flags().String("customer", "", "Filter by customer ID")
	cmd.Flags().String("start-on", "", "Filter by job production plan start date (YYYY-MM-DD)")
	cmd.Flags().String("start-on-min", "", "Filter by job production plan start date on/after (YYYY-MM-DD)")
	cmd.Flags().String("start-on-max", "", "Filter by job production plan start date on/before (YYYY-MM-DD)")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanStatusChangesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseJobProductionPlanStatusChangesListOptions(cmd)
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
	query.Set("fields[job-production-plan-status-changes]", "status,comment,changed-at,created-at,job-production-plan,changed-by,job-production-plan-cancellation-reason-type")
	query.Set("fields[job-production-plans]", "job-number,job-name")
	query.Set("fields[users]", "name")
	query.Set("fields[job-production-plan-cancellation-reason-types]", "name")
	query.Set("include", "job-production-plan,changed-by,job-production-plan-cancellation-reason-type")

	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	} else {
		query.Set("sort", "-changed-at")
	}

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	setFilterIfPresent(query, "filter[job-production-plan]", opts.JobProductionPlan)
	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[job-production-plan-cancellation-reason-type]", opts.JobProductionPlanCancellationReasonType)
	setFilterIfPresent(query, "filter[customer]", opts.Customer)
	setFilterIfPresent(query, "filter[start-on]", opts.StartOn)
	setFilterIfPresent(query, "filter[start-on-min]", opts.StartOnMin)
	setFilterIfPresent(query, "filter[start-on-max]", opts.StartOnMax)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-status-changes", query)
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

	rows := buildJobProductionPlanStatusChangeRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderJobProductionPlanStatusChangesTable(cmd, rows)
}

func parseJobProductionPlanStatusChangesListOptions(cmd *cobra.Command) (jobProductionPlanStatusChangesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	status, _ := cmd.Flags().GetString("status")
	jobProductionPlanCancellationReasonType, _ := cmd.Flags().GetString("job-production-plan-cancellation-reason-type")
	customer, _ := cmd.Flags().GetString("customer")
	startOn, _ := cmd.Flags().GetString("start-on")
	startOnMin, _ := cmd.Flags().GetString("start-on-min")
	startOnMax, _ := cmd.Flags().GetString("start-on-max")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanStatusChangesListOptions{
		BaseURL:                                 baseURL,
		Token:                                   token,
		JSON:                                    jsonOut,
		NoAuth:                                  noAuth,
		Limit:                                   limit,
		Offset:                                  offset,
		Sort:                                    sort,
		JobProductionPlan:                       jobProductionPlan,
		JobProductionPlanCancellationReasonType: jobProductionPlanCancellationReasonType,
		Status:                                  status,
		Customer:                                customer,
		StartOn:                                 startOn,
		StartOnMin:                              startOnMin,
		StartOnMax:                              startOnMax,
		CreatedAtMin:                            createdAtMin,
		CreatedAtMax:                            createdAtMax,
		UpdatedAtMin:                            updatedAtMin,
		UpdatedAtMax:                            updatedAtMax,
	}, nil
}

func buildJobProductionPlanStatusChangeRows(resp jsonAPIResponse) []jobProductionPlanStatusChangeRow {
	rows := make([]jobProductionPlanStatusChangeRow, 0, len(resp.Data))
	included := make(map[string]jsonAPIResource)
	for _, resource := range resp.Included {
		included[resourceKey(resource.Type, resource.ID)] = resource
	}

	for _, resource := range resp.Data {
		rows = append(rows, buildJobProductionPlanStatusChangeRow(resource, included))
	}
	return rows
}

func buildJobProductionPlanStatusChangeRow(resource jsonAPIResource, included map[string]jsonAPIResource) jobProductionPlanStatusChangeRow {
	attrs := resource.Attributes
	row := jobProductionPlanStatusChangeRow{
		ID:        resource.ID,
		Status:    strings.TrimSpace(stringAttr(attrs, "status")),
		Comment:   strings.TrimSpace(stringAttr(attrs, "comment")),
		ChangedAt: formatDateTime(stringAttr(attrs, "changed-at")),
	}

	jppType := ""
	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		row.JobProductionPlanID = rel.Data.ID
		jppType = rel.Data.Type
	}
	changedByType := ""
	if rel, ok := resource.Relationships["changed-by"]; ok && rel.Data != nil {
		row.ChangedByID = rel.Data.ID
		changedByType = rel.Data.Type
	}
	reasonType := ""
	if rel, ok := resource.Relationships["job-production-plan-cancellation-reason-type"]; ok && rel.Data != nil {
		row.JobProductionPlanCancellationReasonTypeID = rel.Data.ID
		reasonType = rel.Data.Type
	}

	if len(included) == 0 {
		return row
	}

	if row.JobProductionPlanID != "" && jppType != "" {
		if jpp, ok := included[resourceKey(jppType, row.JobProductionPlanID)]; ok {
			jobNumber := strings.TrimSpace(stringAttr(jpp.Attributes, "job-number"))
			jobName := strings.TrimSpace(stringAttr(jpp.Attributes, "job-name"))
			if jobNumber != "" && jobName != "" {
				row.JobProductionPlan = fmt.Sprintf("%s - %s", jobNumber, jobName)
			} else {
				row.JobProductionPlan = firstNonEmpty(jobNumber, jobName)
			}
		}
	}

	if row.ChangedByID != "" && changedByType != "" {
		if user, ok := included[resourceKey(changedByType, row.ChangedByID)]; ok {
			row.ChangedBy = strings.TrimSpace(stringAttr(user.Attributes, "name"))
		}
	}

	if row.JobProductionPlanCancellationReasonTypeID != "" && reasonType != "" {
		if reason, ok := included[resourceKey(reasonType, row.JobProductionPlanCancellationReasonTypeID)]; ok {
			row.JobProductionPlanCancellationReasonType = firstNonEmpty(
				strings.TrimSpace(stringAttr(reason.Attributes, "name")),
				strings.TrimSpace(stringAttr(reason.Attributes, "description")),
			)
		}
	}

	return row
}

func buildJobProductionPlanStatusChangeRowFromSingle(resp jsonAPISingleResponse) jobProductionPlanStatusChangeRow {
	included := make(map[string]jsonAPIResource)
	for _, resource := range resp.Included {
		included[resourceKey(resource.Type, resource.ID)] = resource
	}

	return buildJobProductionPlanStatusChangeRow(resp.Data, included)
}

func renderJobProductionPlanStatusChangesTable(cmd *cobra.Command, rows []jobProductionPlanStatusChangeRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No job production plan status changes found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tJOB PRODUCTION PLAN\tSTATUS\tCHANGED BY\tCHANGED AT\tREASON\tCOMMENT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.JobProductionPlan, 32),
			truncateString(row.Status, 12),
			truncateString(firstNonEmpty(row.ChangedBy, row.ChangedByID), 20),
			truncateString(row.ChangedAt, 20),
			truncateString(firstNonEmpty(row.JobProductionPlanCancellationReasonType, row.JobProductionPlanCancellationReasonTypeID), 24),
			truncateString(row.Comment, 40),
		)
	}
	return writer.Flush()
}
