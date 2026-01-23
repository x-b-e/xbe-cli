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

type jobProductionPlanBroadcastMessagesListOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	NoAuth            bool
	Limit             int
	Offset            int
	Sort              string
	JobProductionPlan string
	CreatedBy         string
	IsHidden          string
}

type jobProductionPlanBroadcastMessageRow struct {
	ID                   string `json:"id"`
	JobProductionPlanID  string `json:"job_production_plan_id,omitempty"`
	JobProductionPlan    string `json:"job_production_plan,omitempty"`
	Summary              string `json:"summary,omitempty"`
	Message              string `json:"message,omitempty"`
	IsHidden             bool   `json:"is_hidden"`
	CreatedAt            string `json:"created_at,omitempty"`
	CreatedByID          string `json:"created_by_id,omitempty"`
	CreatedBy            string `json:"created_by,omitempty"`
	CreatedByDisplayName string `json:"-"`
}

func newJobProductionPlanBroadcastMessagesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List job production plan broadcast messages",
		Long: `List job production plan broadcast messages with filtering and pagination.

Output Columns:
  ID                     Broadcast message identifier
  JOB PRODUCTION PLAN    Job production plan name/number
  SUMMARY                Message summary or preview
  HIDDEN                 Whether the message is hidden
  CREATED BY             User who created the message
  CREATED                Created timestamp

Filters:
  --job-production-plan  Filter by job production plan ID (comma-separated for multiple)
  --created-by           Filter by created-by user ID
  --is-hidden            Filter by hidden status (true/false)

By default, hidden messages are excluded (server default).

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List messages for a job production plan
  xbe view job-production-plan-broadcast-messages list --job-production-plan 123

  # Include hidden messages
  xbe view job-production-plan-broadcast-messages list --is-hidden true

  # Filter by creator
  xbe view job-production-plan-broadcast-messages list --created-by 456

  # Output as JSON
  xbe view job-production-plan-broadcast-messages list --json`,
		RunE: runJobProductionPlanBroadcastMessagesList,
	}
	initJobProductionPlanBroadcastMessagesListFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanBroadcastMessagesCmd.AddCommand(newJobProductionPlanBroadcastMessagesListCmd())
}

func initJobProductionPlanBroadcastMessagesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("job-production-plan", "", "Filter by job production plan ID (comma-separated for multiple)")
	cmd.Flags().String("created-by", "", "Filter by created-by user ID")
	cmd.Flags().String("is-hidden", "", "Filter by hidden status (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanBroadcastMessagesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseJobProductionPlanBroadcastMessagesListOptions(cmd)
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
	query.Set("fields[job-production-plan-broadcast-messages]", "summary,message,is-hidden,created-at,job-production-plan,created-by")
	query.Set("fields[job-production-plans]", "job-number,job-name")
	query.Set("fields[users]", "name")
	query.Set("include", "job-production-plan,created-by")
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	} else {
		query.Set("sort", "-created-at")
	}

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	setFilterIfPresent(query, "filter[job_production_plan]", opts.JobProductionPlan)
	setFilterIfPresent(query, "filter[created_by]", opts.CreatedBy)
	setFilterIfPresent(query, "filter[is_hidden]", opts.IsHidden)

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-broadcast-messages", query)
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

	rows := buildJobProductionPlanBroadcastMessageRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderJobProductionPlanBroadcastMessagesTable(cmd, rows)
}

func parseJobProductionPlanBroadcastMessagesListOptions(cmd *cobra.Command) (jobProductionPlanBroadcastMessagesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	createdBy, _ := cmd.Flags().GetString("created-by")
	isHidden, _ := cmd.Flags().GetString("is-hidden")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanBroadcastMessagesListOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		NoAuth:            noAuth,
		Limit:             limit,
		Offset:            offset,
		Sort:              sort,
		JobProductionPlan: jobProductionPlan,
		CreatedBy:         createdBy,
		IsHidden:          isHidden,
	}, nil
}

func buildJobProductionPlanBroadcastMessageRows(resp jsonAPIResponse) []jobProductionPlanBroadcastMessageRow {
	rows := make([]jobProductionPlanBroadcastMessageRow, 0, len(resp.Data))
	included := make(map[string]jsonAPIResource)
	for _, resource := range resp.Included {
		included[resourceKey(resource.Type, resource.ID)] = resource
	}

	for _, resource := range resp.Data {
		rows = append(rows, buildJobProductionPlanBroadcastMessageRow(resource, included))
	}
	return rows
}

func buildJobProductionPlanBroadcastMessageRow(resource jsonAPIResource, included map[string]jsonAPIResource) jobProductionPlanBroadcastMessageRow {
	attrs := resource.Attributes
	row := jobProductionPlanBroadcastMessageRow{
		ID:        resource.ID,
		Summary:   strings.TrimSpace(stringAttr(attrs, "summary")),
		Message:   strings.TrimSpace(stringAttr(attrs, "message")),
		IsHidden:  boolAttr(attrs, "is-hidden"),
		CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
	}

	jppType := ""
	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		row.JobProductionPlanID = rel.Data.ID
		jppType = rel.Data.Type
	}
	createdByType := ""
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		row.CreatedByID = rel.Data.ID
		createdByType = rel.Data.Type
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

	if row.CreatedByID != "" && createdByType != "" {
		if user, ok := included[resourceKey(createdByType, row.CreatedByID)]; ok {
			row.CreatedBy = strings.TrimSpace(stringAttr(user.Attributes, "name"))
		}
	}

	return row
}

func renderJobProductionPlanBroadcastMessagesTable(cmd *cobra.Command, rows []jobProductionPlanBroadcastMessageRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No job production plan broadcast messages found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tJOB PRODUCTION PLAN\tSUMMARY\tHIDDEN\tCREATED BY\tCREATED")

	for _, row := range rows {
		summary := row.Summary
		if summary == "" {
			summary = row.Message
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%t\t%s\t%s\n",
			row.ID,
			truncateString(row.JobProductionPlan, 32),
			truncateString(summary, 48),
			row.IsHidden,
			truncateString(firstNonEmpty(row.CreatedBy, row.CreatedByID), 20),
			truncateString(row.CreatedAt, 20),
		)
	}

	return writer.Flush()
}
