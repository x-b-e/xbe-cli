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

type jobProductionPlanUnabandonmentsListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Sort    string
}

type jobProductionPlanUnabandonmentRow struct {
	ID                                string `json:"id"`
	JobProductionPlanID               string `json:"job_production_plan_id,omitempty"`
	Comment                           string `json:"comment,omitempty"`
	SuppressStatusChangeNotifications bool   `json:"suppress_status_change_notifications"`
}

func newJobProductionPlanUnabandonmentsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List job production plan unabandonments",
		Long: `List job production plan unabandonments.

Output Columns:
  ID          Unabandonment identifier
  JOB PLAN    Job production plan ID
  SUPPRESS    Suppress status change notifications
  COMMENT     Unabandonment comment (if present)

Filters:
  None

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List unabandonments
  xbe view job-production-plan-unabandonments list

  # Paginate results
  xbe view job-production-plan-unabandonments list --limit 25 --offset 50

  # Output as JSON
  xbe view job-production-plan-unabandonments list --json`,
		Args: cobra.NoArgs,
		RunE: runJobProductionPlanUnabandonmentsList,
	}
	initJobProductionPlanUnabandonmentsListFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanUnabandonmentsCmd.AddCommand(newJobProductionPlanUnabandonmentsListCmd())
}

func initJobProductionPlanUnabandonmentsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanUnabandonmentsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseJobProductionPlanUnabandonmentsListOptions(cmd)
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
	query.Set("fields[job-production-plan-unabandonments]", "comment,suppress-status-change-notifications,job-production-plan")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-unabandonments", query)
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

	rows := buildJobProductionPlanUnabandonmentRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderJobProductionPlanUnabandonmentsTable(cmd, rows)
}

func parseJobProductionPlanUnabandonmentsListOptions(cmd *cobra.Command) (jobProductionPlanUnabandonmentsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanUnabandonmentsListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
	}, nil
}

func buildJobProductionPlanUnabandonmentRows(resp jsonAPIResponse) []jobProductionPlanUnabandonmentRow {
	rows := make([]jobProductionPlanUnabandonmentRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := jobProductionPlanUnabandonmentRow{
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

func buildJobProductionPlanUnabandonmentRowFromSingle(resp jsonAPISingleResponse) jobProductionPlanUnabandonmentRow {
	resource := resp.Data
	attrs := resource.Attributes
	row := jobProductionPlanUnabandonmentRow{
		ID:                                resource.ID,
		Comment:                           stringAttr(attrs, "comment"),
		SuppressStatusChangeNotifications: boolAttr(attrs, "suppress-status-change-notifications"),
	}

	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		row.JobProductionPlanID = rel.Data.ID
	}

	return row
}

func renderJobProductionPlanUnabandonmentsTable(cmd *cobra.Command, rows []jobProductionPlanUnabandonmentRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No job production plan unabandonments found.")
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
