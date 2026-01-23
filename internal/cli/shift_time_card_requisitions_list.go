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

type shiftTimeCardRequisitionsListOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	NoAuth                 bool
	Limit                  int
	Offset                 int
	Sort                   string
	TenderJobScheduleShift string
	TimeCard               string
	Broker                 string
	Trucker                string
	JobProductionPlan      string
	Driver                 string
	CreatedBy              string
	Status                 string
}

type shiftTimeCardRequisitionRow struct {
	ID                       string `json:"id"`
	Status                   string `json:"status,omitempty"`
	IsSubmitted              bool   `json:"is_submitted,omitempty"`
	TenderJobScheduleShiftID string `json:"tender_job_schedule_shift_id,omitempty"`
	TimeCardID               string `json:"time_card_id,omitempty"`
	CreatedByID              string `json:"created_by_id,omitempty"`
}

func newShiftTimeCardRequisitionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List shift time card requisitions",
		Long: `List shift time card requisitions and their statuses.

Output Columns:
  ID          Shift time card requisition ID
  STATUS      Requisition status
  SUBMITTED   Whether the time card was submitted
  SHIFT ID    Tender job schedule shift ID
  TIME CARD   Time card ID (if fulfilled)
  CREATED BY  User who created the requisition

Filters:
  --tender-job-schedule-shift  Filter by tender job schedule shift ID
  --time-card                 Filter by time card ID
  --broker                    Filter by broker ID
  --trucker                   Filter by trucker ID
  --job-production-plan       Filter by job production plan ID
  --driver                    Filter by driver user ID
  --created-by                Filter by created-by user ID
  --status                    Filter by status (open/closed)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List shift time card requisitions
  xbe view shift-time-card-requisitions list

  # Filter by status
  xbe view shift-time-card-requisitions list --status open

  # Filter by broker
  xbe view shift-time-card-requisitions list --broker 123

  # Output as JSON
  xbe view shift-time-card-requisitions list --json`,
		RunE: runShiftTimeCardRequisitionsList,
	}
	initShiftTimeCardRequisitionsListFlags(cmd)
	return cmd
}

func init() {
	shiftTimeCardRequisitionsCmd.AddCommand(newShiftTimeCardRequisitionsListCmd())
}

func initShiftTimeCardRequisitionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("tender-job-schedule-shift", "", "Filter by tender job schedule shift ID")
	cmd.Flags().String("time-card", "", "Filter by time card ID")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("trucker", "", "Filter by trucker ID")
	cmd.Flags().String("job-production-plan", "", "Filter by job production plan ID")
	cmd.Flags().String("driver", "", "Filter by driver user ID")
	cmd.Flags().String("created-by", "", "Filter by created-by user ID")
	cmd.Flags().String("status", "", "Filter by status (open/closed)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runShiftTimeCardRequisitionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseShiftTimeCardRequisitionsListOptions(cmd)
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
	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[tender_job_schedule_shift]", opts.TenderJobScheduleShift)
	setFilterIfPresent(query, "filter[time_card]", opts.TimeCard)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)
	setFilterIfPresent(query, "filter[job_production_plan]", opts.JobProductionPlan)
	setFilterIfPresent(query, "filter[driver]", opts.Driver)
	setFilterIfPresent(query, "filter[created_by]", opts.CreatedBy)
	setFilterIfPresent(query, "filter[status]", opts.Status)

	body, _, err := client.Get(cmd.Context(), "/v1/shift-time-card-requisitions", query)
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

	rows := buildShiftTimeCardRequisitionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderShiftTimeCardRequisitionsTable(cmd, rows)
}

func parseShiftTimeCardRequisitionsListOptions(cmd *cobra.Command) (shiftTimeCardRequisitionsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	tenderJobScheduleShift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	timeCard, _ := cmd.Flags().GetString("time-card")
	broker, _ := cmd.Flags().GetString("broker")
	trucker, _ := cmd.Flags().GetString("trucker")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	driver, _ := cmd.Flags().GetString("driver")
	createdBy, _ := cmd.Flags().GetString("created-by")
	status, _ := cmd.Flags().GetString("status")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return shiftTimeCardRequisitionsListOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		NoAuth:                 noAuth,
		Limit:                  limit,
		Offset:                 offset,
		Sort:                   sort,
		TenderJobScheduleShift: tenderJobScheduleShift,
		TimeCard:               timeCard,
		Broker:                 broker,
		Trucker:                trucker,
		JobProductionPlan:      jobProductionPlan,
		Driver:                 driver,
		CreatedBy:              createdBy,
		Status:                 status,
	}, nil
}

func buildShiftTimeCardRequisitionRows(resp jsonAPIResponse) []shiftTimeCardRequisitionRow {
	rows := make([]shiftTimeCardRequisitionRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := shiftTimeCardRequisitionRow{
			ID:          resource.ID,
			Status:      stringAttr(resource.Attributes, "status"),
			IsSubmitted: boolAttr(resource.Attributes, "is-submitted"),
		}

		if rel, ok := resource.Relationships["tender-job-schedule-shift"]; ok && rel.Data != nil {
			row.TenderJobScheduleShiftID = rel.Data.ID
		}

		if rel, ok := resource.Relationships["time-card"]; ok && rel.Data != nil {
			row.TimeCardID = rel.Data.ID
		}

		if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
			row.CreatedByID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderShiftTimeCardRequisitionsTable(cmd *cobra.Command, rows []shiftTimeCardRequisitionRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No shift time card requisitions found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tSUBMITTED\tSHIFT ID\tTIME CARD\tCREATED BY")
	for _, row := range rows {
		submitted := "no"
		if row.IsSubmitted {
			submitted = "yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Status,
			submitted,
			row.TenderJobScheduleShiftID,
			row.TimeCardID,
			row.CreatedByID,
		)
	}
	return writer.Flush()
}
