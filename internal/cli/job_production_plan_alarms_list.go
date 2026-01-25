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

type jobProductionPlanAlarmsListOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	NoAuth              bool
	Limit               int
	Offset              int
	Sort                string
	JobProductionPlanID string
}

type jobProductionPlanAlarmRow struct {
	ID                                 string  `json:"id"`
	JobProductionPlanID                string  `json:"job_production_plan_id,omitempty"`
	Tons                               float64 `json:"tons,omitempty"`
	BaseMaterialTypeFullyQualifiedName string  `json:"base_material_type_fully_qualified_name,omitempty"`
	MaxLatencyMinutes                  int     `json:"max_latency_minutes,omitempty"`
	PlannedAt                          string  `json:"planned_at,omitempty"`
	FulfilledAt                        string  `json:"fulfilled_at,omitempty"`
}

func newJobProductionPlanAlarmsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List job production plan alarms",
		Long: `List job production plan alarms.

Output Columns:
  ID           Alarm identifier
  PLAN         Job production plan ID
  TONS         Tonnage trigger
  BASE TYPE    Base material type fully qualified name
  MAX LATENCY  Max latency (minutes)
  PLANNED AT   Planned fulfillment timestamp
  FULFILLED AT Actual fulfillment timestamp

Filters:
  --job-production-plan  Filter by job production plan ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List alarms
  xbe view job-production-plan-alarms list

  # Filter by job production plan
  xbe view job-production-plan-alarms list --job-production-plan 123

  # Sort by planned time
  xbe view job-production-plan-alarms list --sort planned-at

  # Output as JSON
  xbe view job-production-plan-alarms list --json`,
		Args: cobra.NoArgs,
		RunE: runJobProductionPlanAlarmsList,
	}
	initJobProductionPlanAlarmsListFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanAlarmsCmd.AddCommand(newJobProductionPlanAlarmsListCmd())
}

func initJobProductionPlanAlarmsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("job-production-plan", "", "Filter by job production plan ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanAlarmsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseJobProductionPlanAlarmsListOptions(cmd)
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
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[job-production-plan-alarms]", "tons,base-material-type-fully-qualified-name,max-latency-minutes,planned-at,fulfilled-at")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[job-production-plan]", opts.JobProductionPlanID)

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-alarms", query)
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

	rows := buildJobProductionPlanAlarmRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderJobProductionPlanAlarmsTable(cmd, rows)
}

func parseJobProductionPlanAlarmsListOptions(cmd *cobra.Command) (jobProductionPlanAlarmsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	jobProductionPlanID, _ := cmd.Flags().GetString("job-production-plan")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanAlarmsListOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		NoAuth:              noAuth,
		Limit:               limit,
		Offset:              offset,
		Sort:                sort,
		JobProductionPlanID: jobProductionPlanID,
	}, nil
}

func buildJobProductionPlanAlarmRows(resp jsonAPIResponse) []jobProductionPlanAlarmRow {
	rows := make([]jobProductionPlanAlarmRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := jobProductionPlanAlarmRow{
			ID:                                 resource.ID,
			Tons:                               floatAttr(attrs, "tons"),
			BaseMaterialTypeFullyQualifiedName: stringAttr(attrs, "base-material-type-fully-qualified-name"),
			MaxLatencyMinutes:                  intAttr(attrs, "max-latency-minutes"),
			PlannedAt:                          formatDateTime(stringAttr(attrs, "planned-at")),
			FulfilledAt:                        formatDateTime(stringAttr(attrs, "fulfilled-at")),
		}

		if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
			row.JobProductionPlanID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderJobProductionPlanAlarmsTable(cmd *cobra.Command, rows []jobProductionPlanAlarmRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No job production plan alarms found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPLAN\tTONS\tBASE TYPE\tMAX LATENCY\tPLANNED AT\tFULFILLED AT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.JobProductionPlanID,
			formatTons(row.Tons),
			truncateString(row.BaseMaterialTypeFullyQualifiedName, 28),
			formatMaxLatency(row.MaxLatencyMinutes),
			row.PlannedAt,
			row.FulfilledAt,
		)
	}
	return writer.Flush()
}

func formatTons(value float64) string {
	if value == 0 {
		return ""
	}
	if value == float64(int(value)) {
		return fmt.Sprintf("%.0f", value)
	}
	return fmt.Sprintf("%.2f", value)
}

func formatMaxLatency(value int) string {
	if value == 0 {
		return ""
	}
	return strconv.Itoa(value)
}
