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

type jobProductionPlanRejectionsListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Sort    string
}

type jobProductionPlanRejectionRow struct {
	ID                                string `json:"id"`
	JobProductionPlanID               string `json:"job_production_plan_id,omitempty"`
	Comment                           string `json:"comment,omitempty"`
	SuppressStatusChangeNotifications bool   `json:"suppress_status_change_notifications"`
}

func newJobProductionPlanRejectionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List job production plan rejections",
		Long: `List job production plan rejections.

Output Columns:
  ID          Rejection identifier
  JOB PLAN    Job production plan ID
  SUPPRESS    Suppress status change notifications
  COMMENT     Rejection comment (if present)

Filters:
  None

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List rejections
  xbe view job-production-plan-rejections list

  # Paginate results
  xbe view job-production-plan-rejections list --limit 25 --offset 50

  # Output as JSON
  xbe view job-production-plan-rejections list --json`,
		Args: cobra.NoArgs,
		RunE: runJobProductionPlanRejectionsList,
	}
	initJobProductionPlanRejectionsListFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanRejectionsCmd.AddCommand(newJobProductionPlanRejectionsListCmd())
}

func initJobProductionPlanRejectionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanRejectionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseJobProductionPlanRejectionsListOptions(cmd)
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
	query.Set("fields[job-production-plan-rejections]", "comment,suppress-status-change-notifications,job-production-plan")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-rejections", query)
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

	rows := buildJobProductionPlanRejectionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderJobProductionPlanRejectionsTable(cmd, rows)
}

func parseJobProductionPlanRejectionsListOptions(cmd *cobra.Command) (jobProductionPlanRejectionsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanRejectionsListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
	}, nil
}

func buildJobProductionPlanRejectionRows(resp jsonAPIResponse) []jobProductionPlanRejectionRow {
	rows := make([]jobProductionPlanRejectionRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := jobProductionPlanRejectionRow{
			ID:                                resource.ID,
			Comment:                           stringAttr(attrs, "comment"),
			SuppressStatusChangeNotifications: boolAttr(attrs, "suppress-status-change-notifications"),
		}

		if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
			row.JobProductionPlanID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func buildJobProductionPlanRejectionRowFromSingle(resp jsonAPISingleResponse) jobProductionPlanRejectionRow {
	resource := resp.Data
	attrs := resource.Attributes
	row := jobProductionPlanRejectionRow{
		ID:                                resource.ID,
		Comment:                           stringAttr(attrs, "comment"),
		SuppressStatusChangeNotifications: boolAttr(attrs, "suppress-status-change-notifications"),
	}

	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		row.JobProductionPlanID = rel.Data.ID
	}

	return row
}

func renderJobProductionPlanRejectionsTable(cmd *cobra.Command, rows []jobProductionPlanRejectionRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No job production plan rejections found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tJOB PLAN\tSUPPRESS\tCOMMENT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%t\t%s\n",
			row.ID,
			row.JobProductionPlanID,
			row.SuppressStatusChangeNotifications,
			truncateString(row.Comment, 50),
		)
	}
	return writer.Flush()
}
