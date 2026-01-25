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

type jobProductionPlanDuplicationWorksListOptions struct {
	BaseURL                     string
	Token                       string
	JSON                        bool
	NoAuth                      bool
	Limit                       int
	Offset                      int
	Sort                        string
	ProcessedAt                 string
	JobProductionPlanTemplateID string
}

func newJobProductionPlanDuplicationWorksListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List job production plan duplication work",
		Long: `List job production plan duplication work items with pagination.

Output Columns:
  ID          Work identifier
  TEMPLATE    Job production plan template ID
  DERIVED     Derived job production plan ID
  JID         Background job ID (if scheduled async)
  START ON    Derived plan start date
  SCHEDULED   When work was scheduled
  PROCESSED   When work finished processing

Filters:
  --processed-at                    Filter by processed-at timestamp (ISO 8601)
  --job-production-plan-template-id Filter by template ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List recent duplication work
  xbe view job-production-plan-duplication-works list --sort -scheduled-at

  # Filter by template
  xbe view job-production-plan-duplication-works list --job-production-plan-template-id 123

  # Filter by processed-at
  xbe view job-production-plan-duplication-works list --processed-at 2025-01-01T00:00:00Z

  # Output as JSON
  xbe view job-production-plan-duplication-works list --json`,
		RunE: runJobProductionPlanDuplicationWorksList,
	}
	initJobProductionPlanDuplicationWorksListFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanDuplicationWorksCmd.AddCommand(newJobProductionPlanDuplicationWorksListCmd())
}

func initJobProductionPlanDuplicationWorksListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort order")
	cmd.Flags().String("processed-at", "", "Filter by processed-at timestamp (ISO 8601)")
	cmd.Flags().String("job-production-plan-template-id", "", "Filter by job production plan template ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanDuplicationWorksList(cmd *cobra.Command, _ []string) error {
	opts, err := parseJobProductionPlanDuplicationWorksListOptions(cmd)
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
	query.Set("fields[job-production-plan-duplication-works]", "jid,scheduled-at,processed-at,start-on,derived-job-production-plan-template-name")

	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}
	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	setFilterIfPresent(query, "filter[processed-at]", opts.ProcessedAt)
	setFilterIfPresent(query, "filter[job-production-plan-template-id]", opts.JobProductionPlanTemplateID)

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-duplication-works", query)
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

	rows := buildJobProductionPlanDuplicationWorkRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderJobProductionPlanDuplicationWorksTable(cmd, rows)
}

func parseJobProductionPlanDuplicationWorksListOptions(cmd *cobra.Command) (jobProductionPlanDuplicationWorksListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	processedAt, _ := cmd.Flags().GetString("processed-at")
	jobProductionPlanTemplateID, _ := cmd.Flags().GetString("job-production-plan-template-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanDuplicationWorksListOptions{
		BaseURL:                     baseURL,
		Token:                       token,
		JSON:                        jsonOut,
		NoAuth:                      noAuth,
		Limit:                       limit,
		Offset:                      offset,
		Sort:                        sort,
		ProcessedAt:                 processedAt,
		JobProductionPlanTemplateID: jobProductionPlanTemplateID,
	}, nil
}

type jobProductionPlanDuplicationWorkRow struct {
	ID                          string `json:"id"`
	JobProductionPlanTemplateID string `json:"job_production_plan_template_id,omitempty"`
	DerivedJobProductionPlanID  string `json:"derived_job_production_plan_id,omitempty"`
	JID                         string `json:"jid,omitempty"`
	StartOn                     string `json:"start_on,omitempty"`
	ScheduledAt                 string `json:"scheduled_at,omitempty"`
	ProcessedAt                 string `json:"processed_at,omitempty"`
}

func buildJobProductionPlanDuplicationWorkRows(resp jsonAPIResponse) []jobProductionPlanDuplicationWorkRow {
	rows := make([]jobProductionPlanDuplicationWorkRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := jobProductionPlanDuplicationWorkRow{
			ID:          resource.ID,
			JID:         stringAttr(attrs, "jid"),
			StartOn:     formatDate(stringAttr(attrs, "start-on")),
			ScheduledAt: formatDateTime(stringAttr(attrs, "scheduled-at")),
			ProcessedAt: formatDateTime(stringAttr(attrs, "processed-at")),
		}

		if rel, ok := resource.Relationships["job-production-plan-template"]; ok && rel.Data != nil {
			row.JobProductionPlanTemplateID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["derived-job-production-plan"]; ok && rel.Data != nil {
			row.DerivedJobProductionPlanID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderJobProductionPlanDuplicationWorksTable(cmd *cobra.Command, rows []jobProductionPlanDuplicationWorkRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No job production plan duplication works found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTEMPLATE\tDERIVED\tJID\tSTART ON\tSCHEDULED\tPROCESSED")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.JobProductionPlanTemplateID,
			row.DerivedJobProductionPlanID,
			row.JID,
			row.StartOn,
			row.ScheduledAt,
			row.ProcessedAt,
		)
	}
	return writer.Flush()
}
