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

type materialTransactionShiftAssignmentsListOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	NoAuth                 bool
	Limit                  int
	Offset                 int
	Sort                   string
	JobProductionPlan      string
	TenderJobScheduleShift string
	Trucker                string
	Broker                 string
	MaterialTransaction    string
	IsProcessed            string
}

type materialTransactionShiftAssignmentRow struct {
	ID                       string   `json:"id"`
	TenderJobScheduleShiftID string   `json:"tender_job_schedule_shift_id,omitempty"`
	JobProductionPlanID      string   `json:"job_production_plan_id,omitempty"`
	JobProductionPlan        string   `json:"job_production_plan,omitempty"`
	TruckerID                string   `json:"trucker_id,omitempty"`
	TruckerName              string   `json:"trucker_name,omitempty"`
	BrokerID                 string   `json:"broker_id,omitempty"`
	BrokerName               string   `json:"broker_name,omitempty"`
	MaterialTransactionIDs   []string `json:"material_transaction_ids,omitempty"`
	MaterialTransactionCount int      `json:"material_transaction_count,omitempty"`
	ProcessedAt              string   `json:"processed_at,omitempty"`
	IsProcessed              bool     `json:"is_processed,omitempty"`
}

func newMaterialTransactionShiftAssignmentsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List material transaction shift assignments",
		Long: `List material transaction shift assignments.

Output Columns:
  ID         Assignment identifier
  SHIFT      Tender job schedule shift ID
  JOB        Job production plan
  TRUCKER    Trucker name
  BROKER     Broker name
  MTXNS      Material transaction count
  PROCESSED  Processed status
  AT         Processed timestamp

Filters:
  --job-production-plan       Filter by job production plan ID
  --tender-job-schedule-shift Filter by tender job schedule shift ID
  --trucker                   Filter by trucker ID
  --broker                    Filter by broker ID
  --material-transaction      Filter by material transaction ID (comma-separated for multiple)
  --is-processed              Filter by processed status (true/false)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List assignments
  xbe view material-transaction-shift-assignments list

  # Filter by shift and processed status
  xbe view material-transaction-shift-assignments list --tender-job-schedule-shift 123 --is-processed true

  # Filter by material transaction
  xbe view material-transaction-shift-assignments list --material-transaction 456

  # Output as JSON
  xbe view material-transaction-shift-assignments list --json`,
		Args: cobra.NoArgs,
		RunE: runMaterialTransactionShiftAssignmentsList,
	}
	initMaterialTransactionShiftAssignmentsListFlags(cmd)
	return cmd
}

func init() {
	materialTransactionShiftAssignmentsCmd.AddCommand(newMaterialTransactionShiftAssignmentsListCmd())
}

func initMaterialTransactionShiftAssignmentsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("job-production-plan", "", "Filter by job production plan ID")
	cmd.Flags().String("tender-job-schedule-shift", "", "Filter by tender job schedule shift ID")
	cmd.Flags().String("trucker", "", "Filter by trucker ID")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("material-transaction", "", "Filter by material transaction ID (comma-separated for multiple)")
	cmd.Flags().String("is-processed", "", "Filter by processed status (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialTransactionShiftAssignmentsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMaterialTransactionShiftAssignmentsListOptions(cmd)
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
	query.Set("fields[material-transaction-shift-assignments]", "material-transaction-ids,tender-job-schedule-shift,job-production-plan,trucker,broker,processed-at,is-processed,created-by")
	query.Set("include", "tender-job-schedule-shift,job-production-plan,trucker,broker,created-by")
	query.Set("fields[job-production-plans]", "job-number,job-name")
	query.Set("fields[truckers]", "company-name")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[users]", "name")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[job-production-plan]", opts.JobProductionPlan)
	setFilterIfPresent(query, "filter[tender-job-schedule-shift]", opts.TenderJobScheduleShift)
	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[material-transactions]", opts.MaterialTransaction)
	setFilterIfPresent(query, "filter[is-processed]", opts.IsProcessed)

	body, _, err := client.Get(cmd.Context(), "/v1/material-transaction-shift-assignments", query)
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

	rows := buildMaterialTransactionShiftAssignmentRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderMaterialTransactionShiftAssignmentsTable(cmd, rows)
}

func parseMaterialTransactionShiftAssignmentsListOptions(cmd *cobra.Command) (materialTransactionShiftAssignmentsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	tenderJobScheduleShift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	trucker, _ := cmd.Flags().GetString("trucker")
	broker, _ := cmd.Flags().GetString("broker")
	materialTransaction, _ := cmd.Flags().GetString("material-transaction")
	isProcessed, _ := cmd.Flags().GetString("is-processed")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialTransactionShiftAssignmentsListOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		NoAuth:                 noAuth,
		Limit:                  limit,
		Offset:                 offset,
		Sort:                   sort,
		JobProductionPlan:      jobProductionPlan,
		TenderJobScheduleShift: tenderJobScheduleShift,
		Trucker:                trucker,
		Broker:                 broker,
		MaterialTransaction:    materialTransaction,
		IsProcessed:            isProcessed,
	}, nil
}

func buildMaterialTransactionShiftAssignmentRows(resp jsonAPIResponse) []materialTransactionShiftAssignmentRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]materialTransactionShiftAssignmentRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildMaterialTransactionShiftAssignmentRow(resource, included))
	}

	return rows
}

func materialTransactionShiftAssignmentRowFromSingle(resp jsonAPISingleResponse) materialTransactionShiftAssignmentRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	return buildMaterialTransactionShiftAssignmentRow(resp.Data, included)
}

func buildMaterialTransactionShiftAssignmentRow(resource jsonAPIResource, included map[string]jsonAPIResource) materialTransactionShiftAssignmentRow {
	attrs := resource.Attributes
	mtxnIDs := stringSliceAttr(attrs, "material-transaction-ids")

	row := materialTransactionShiftAssignmentRow{
		ID:                       resource.ID,
		MaterialTransactionIDs:   mtxnIDs,
		MaterialTransactionCount: len(mtxnIDs),
		ProcessedAt:              formatDateTime(stringAttr(attrs, "processed-at")),
		IsProcessed:              boolAttr(attrs, "is-processed"),
	}

	if rel, ok := resource.Relationships["tender-job-schedule-shift"]; ok && rel.Data != nil {
		row.TenderJobScheduleShiftID = rel.Data.ID
	}

	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		row.JobProductionPlanID = rel.Data.ID
		if jpp, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.JobProductionPlan = firstNonEmpty(
				stringAttr(jpp.Attributes, "job-number"),
				stringAttr(jpp.Attributes, "job-name"),
			)
		}
	}

	if rel, ok := resource.Relationships["trucker"]; ok && rel.Data != nil {
		row.TruckerID = rel.Data.ID
		if trucker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.TruckerName = strings.TrimSpace(stringAttr(trucker.Attributes, "company-name"))
		}
	}

	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
		if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.BrokerName = strings.TrimSpace(stringAttr(broker.Attributes, "company-name"))
		}
	}

	return row
}

func renderMaterialTransactionShiftAssignmentsTable(cmd *cobra.Command, rows []materialTransactionShiftAssignmentRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No material transaction shift assignments found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSHIFT\tJOB\tTRUCKER\tBROKER\tMTXNS\tPROCESSED\tAT")
	for _, row := range rows {
		jobLabel := firstNonEmpty(row.JobProductionPlan, row.JobProductionPlanID)
		truckerLabel := firstNonEmpty(row.TruckerName, row.TruckerID)
		brokerLabel := firstNonEmpty(row.BrokerName, row.BrokerID)

		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%d\t%t\t%s\n",
			row.ID,
			row.TenderJobScheduleShiftID,
			truncateString(jobLabel, 24),
			truncateString(truckerLabel, 24),
			truncateString(brokerLabel, 24),
			row.MaterialTransactionCount,
			row.IsProcessed,
			row.ProcessedAt,
		)
	}
	return writer.Flush()
}
