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

type timeCardsListOptions struct {
	BaseURL                                string
	Token                                  string
	JSON                                   bool
	NoAuth                                 bool
	Limit                                  int
	Offset                                 int
	Sort                                   string
	Status                                 string
	IsAudited                              string
	WithPayrollCertificationRequirement    string
	WithPayrollCertificationRequirementMet string
	BrokerTender                           string
	BrokerInvoiced                         string
	NotBrokerInvoiced                      string
	TruckerInvoiced                        string
	NotTruckerInvoiced                     string
	ApprovalCount                          string
	ApprovalCountMin                       string
	ApprovalCountMax                       string
	Developer                              string
	OnShortfallBrokerTender                string
	OnShortfallCustomerTender              string
	Customer                               string
	CustomerID                             string
	BusinessUnit                           string
	Trucker                                string
	TruckerID                              string
	Trailer                                string
	TrailerID                              string
	Contractor                             string
	ContractorID                           string
	JobNumber                              string
	Invoice                                string
	TenderJobScheduleShift                 string
	JobScheduleShift                       string
	JobScheduleShiftID                     string
	Broker                                 string
	Driver                                 string
	DriverID                               string
	StartAtMin                             string
	StartAtMax                             string
	IsStartAt                              string
	ShiftDate                              string
	ShiftDateMin                           string
	ShiftDateMax                           string
	ShiftDateBetween                       string
	HasShiftDate                           string
	ApprovedOnMin                          string
	ApprovedOnMax                          string
	TicketNumber                           string
	ApprovableBy                           string
	CostCodes                              string
	GenerateBrokerInvoice                  string
	GenerateTruckerInvoice                 string
	HasTicketReport                        string
	HasSubmitScheduled                     string
}

type timeCardRow struct {
	ID                     string  `json:"id"`
	Status                 string  `json:"status,omitempty"`
	TicketNumber           string  `json:"ticket_number,omitempty"`
	StartAt                string  `json:"start_at,omitempty"`
	EndAt                  string  `json:"end_at,omitempty"`
	TotalHours             float64 `json:"total_hours,omitempty"`
	ApprovalCount          int     `json:"approval_count,omitempty"`
	TruckerID              string  `json:"trucker_id,omitempty"`
	DriverID               string  `json:"driver_id,omitempty"`
	JobID                  string  `json:"job_id,omitempty"`
	JobProductionPlanID    string  `json:"job_production_plan_id,omitempty"`
	TenderJobScheduleShift string  `json:"tender_job_schedule_shift_id,omitempty"`
	JobScheduleShiftID     string  `json:"job_schedule_shift_id,omitempty"`
}

func newTimeCardsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List time cards",
		Long: `List time cards with filtering.

Output Columns:
  ID            Time card identifier
  STATUS        Time card status
  TICKET        Ticket number
  START         Start time
  END           End time
  HOURS         Total hours
  APPROVALS     Approval count
  TRUCKER       Trucker ID
  DRIVER        Driver user ID
  JOB           Job ID
  JPP           Job production plan ID
  SHIFT         Tender job schedule shift ID

Filters:
  --status                                      Filter by status
  --is-audited                                  Filter by audit status (true/false)
  --with-payroll-certification-requirement     Filter by payroll certification requirement (true/false)
  --with-payroll-certification-requirement-met Filter by payroll certification requirement met (true/false)
  --broker-tender                               Filter by broker tender ID
  --broker-invoiced                             Filter by broker invoiced (true/false)
  --not-broker-invoiced                         Filter by not broker invoiced (true/false)
  --trucker-invoiced                            Filter by trucker invoiced (true/false)
  --not-trucker-invoiced                        Filter by not trucker invoiced (true/false)
  --approval-count                              Filter by approval count (exact)
  --approval-count-min                          Filter by minimum approval count
  --approval-count-max                          Filter by maximum approval count
  --developer                                   Filter by developer ID
  --on-shortfall-broker-tender                  Filter by shortfall broker tender (true/false)
  --on-shortfall-customer-tender                Filter by shortfall customer tender (true/false)
  --customer                                    Filter by customer ID
  --customer-id                                 Filter by customer ID (via customer relation)
  --business-unit                               Filter by business unit ID
  --trucker                                     Filter by trucker ID
  --trucker-id                                  Filter by trucker ID (via trucker relation)
  --trailer                                     Filter by trailer ID
  --trailer-id                                  Filter by trailer ID (via trailer relation)
  --contractor                                  Filter by contractor ID
  --contractor-id                               Filter by contractor ID (via contractor relation)
  --job-number                                  Filter by job number
  --invoice                                     Filter by invoice ID
  --tender-job-schedule-shift                   Filter by tender job schedule shift ID
  --job-schedule-shift                          Filter by job schedule shift ID
  --job-schedule-shift-id                       Filter by job schedule shift ID (via relation)
  --broker                                      Filter by broker ID
  --driver                                      Filter by driver user ID
  --driver-id                                   Filter by driver user ID (via relation)
  --start-at-min                                Filter by minimum start time (ISO 8601)
  --start-at-max                                Filter by maximum start time (ISO 8601)
  --is-start-at                                 Filter by presence of start time (true/false)
  --shift-date                                  Filter by shift date (YYYY-MM-DD)
  --shift-date-min                              Filter by minimum shift date (YYYY-MM-DD)
  --shift-date-max                              Filter by maximum shift date (YYYY-MM-DD)
  --shift-date-between                          Filter by shift date range (format: date1|date2)
  --has-shift-date                              Filter by presence of shift date (true/false)
  --approved-on-min                             Filter by minimum approved-on date (YYYY-MM-DD)
  --approved-on-max                             Filter by maximum approved-on date (YYYY-MM-DD)
  --ticket-number                               Filter by ticket number(s) (comma-separated)
  --approvable-by                               Filter by approvable-by user ID
  --cost-codes                                  Filter by cost codes (comma-separated)
  --generate-broker-invoice                     Filter by generate broker invoice (true/false)
  --generate-trucker-invoice                    Filter by generate trucker invoice (true/false)
  --has-ticket-report                           Filter by presence of ticket report (true/false)
  --has-submit-scheduled                        Filter by submit scheduled (true/false)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List time cards
  xbe view time-cards list

  # Filter by status
  xbe view time-cards list --status approved

  # Filter by broker tender
  xbe view time-cards list --broker-tender 123

  # Filter by start time
  xbe view time-cards list --start-at-min 2025-01-01T00:00:00Z --start-at-max 2025-01-31T23:59:59Z

  # Output as JSON
  xbe view time-cards list --json`,
		Args: cobra.NoArgs,
		RunE: runTimeCardsList,
	}
	initTimeCardsListFlags(cmd)
	return cmd
}

func init() {
	timeCardsCmd.AddCommand(newTimeCardsListCmd())
}

func initTimeCardsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("status", "", "Filter by status")
	cmd.Flags().String("is-audited", "", "Filter by audit status (true/false)")
	cmd.Flags().String("with-payroll-certification-requirement", "", "Filter by payroll certification requirement (true/false)")
	cmd.Flags().String("with-payroll-certification-requirement-met", "", "Filter by payroll certification requirement met (true/false)")
	cmd.Flags().String("broker-tender", "", "Filter by broker tender ID")
	cmd.Flags().String("broker-invoiced", "", "Filter by broker invoiced (true/false)")
	cmd.Flags().String("not-broker-invoiced", "", "Filter by not broker invoiced (true/false)")
	cmd.Flags().String("trucker-invoiced", "", "Filter by trucker invoiced (true/false)")
	cmd.Flags().String("not-trucker-invoiced", "", "Filter by not trucker invoiced (true/false)")
	cmd.Flags().String("approval-count", "", "Filter by approval count (exact)")
	cmd.Flags().String("approval-count-min", "", "Filter by minimum approval count")
	cmd.Flags().String("approval-count-max", "", "Filter by maximum approval count")
	cmd.Flags().String("developer", "", "Filter by developer ID")
	cmd.Flags().String("on-shortfall-broker-tender", "", "Filter by shortfall broker tender (true/false)")
	cmd.Flags().String("on-shortfall-customer-tender", "", "Filter by shortfall customer tender (true/false)")
	cmd.Flags().String("customer", "", "Filter by customer ID")
	cmd.Flags().String("customer-id", "", "Filter by customer ID (via customer relation)")
	cmd.Flags().String("business-unit", "", "Filter by business unit ID")
	cmd.Flags().String("trucker", "", "Filter by trucker ID")
	cmd.Flags().String("trucker-id", "", "Filter by trucker ID (via trucker relation)")
	cmd.Flags().String("trailer", "", "Filter by trailer ID")
	cmd.Flags().String("trailer-id", "", "Filter by trailer ID (via trailer relation)")
	cmd.Flags().String("contractor", "", "Filter by contractor ID")
	cmd.Flags().String("contractor-id", "", "Filter by contractor ID (via contractor relation)")
	cmd.Flags().String("job-number", "", "Filter by job number")
	cmd.Flags().String("invoice", "", "Filter by invoice ID")
	cmd.Flags().String("tender-job-schedule-shift", "", "Filter by tender job schedule shift ID")
	cmd.Flags().String("job-schedule-shift", "", "Filter by job schedule shift ID")
	cmd.Flags().String("job-schedule-shift-id", "", "Filter by job schedule shift ID (via relation)")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("driver", "", "Filter by driver user ID")
	cmd.Flags().String("driver-id", "", "Filter by driver user ID (via relation)")
	cmd.Flags().String("start-at-min", "", "Filter by minimum start time (ISO 8601)")
	cmd.Flags().String("start-at-max", "", "Filter by maximum start time (ISO 8601)")
	cmd.Flags().String("is-start-at", "", "Filter by presence of start time (true/false)")
	cmd.Flags().String("shift-date", "", "Filter by shift date (YYYY-MM-DD)")
	cmd.Flags().String("shift-date-min", "", "Filter by minimum shift date (YYYY-MM-DD)")
	cmd.Flags().String("shift-date-max", "", "Filter by maximum shift date (YYYY-MM-DD)")
	cmd.Flags().String("shift-date-between", "", "Filter by shift date range (format: date1|date2)")
	cmd.Flags().String("has-shift-date", "", "Filter by presence of shift date (true/false)")
	cmd.Flags().String("approved-on-min", "", "Filter by minimum approved-on date (YYYY-MM-DD)")
	cmd.Flags().String("approved-on-max", "", "Filter by maximum approved-on date (YYYY-MM-DD)")
	cmd.Flags().String("ticket-number", "", "Filter by ticket number(s) (comma-separated)")
	cmd.Flags().String("approvable-by", "", "Filter by approvable-by user ID")
	cmd.Flags().String("cost-codes", "", "Filter by cost codes (comma-separated)")
	cmd.Flags().String("generate-broker-invoice", "", "Filter by generate broker invoice (true/false)")
	cmd.Flags().String("generate-trucker-invoice", "", "Filter by generate trucker invoice (true/false)")
	cmd.Flags().String("has-ticket-report", "", "Filter by presence of ticket report (true/false)")
	cmd.Flags().String("has-submit-scheduled", "", "Filter by submit scheduled (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTimeCardsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTimeCardsListOptions(cmd)
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
	query.Set("fields[time-cards]", "status,ticket-number,start-at,end-at,total-hours,approval-count,tender-job-schedule-shift,job,job-production-plan,driver,trucker,job-schedule-shift")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[is-audited]", opts.IsAudited)
	setFilterIfPresent(query, "filter[with-payroll-certification-requirement]", opts.WithPayrollCertificationRequirement)
	setFilterIfPresent(query, "filter[with-payroll-certification-requirement-met]", opts.WithPayrollCertificationRequirementMet)
	setFilterIfPresent(query, "filter[broker-tender]", opts.BrokerTender)
	setFilterIfPresent(query, "filter[broker-invoiced]", opts.BrokerInvoiced)
	setFilterIfPresent(query, "filter[not-broker-invoiced]", opts.NotBrokerInvoiced)
	setFilterIfPresent(query, "filter[trucker-invoiced]", opts.TruckerInvoiced)
	setFilterIfPresent(query, "filter[not-trucker-invoiced]", opts.NotTruckerInvoiced)
	setFilterIfPresent(query, "filter[approval-count]", opts.ApprovalCount)
	setFilterIfPresent(query, "filter[approval-count-min]", opts.ApprovalCountMin)
	setFilterIfPresent(query, "filter[approval-count-max]", opts.ApprovalCountMax)
	setFilterIfPresent(query, "filter[developer]", opts.Developer)
	setFilterIfPresent(query, "filter[on-shortfall-broker-tender]", opts.OnShortfallBrokerTender)
	setFilterIfPresent(query, "filter[on-shortfall-customer-tender]", opts.OnShortfallCustomerTender)
	setFilterIfPresent(query, "filter[customer]", opts.Customer)
	setFilterIfPresent(query, "filter[customer-id]", opts.CustomerID)
	setFilterIfPresent(query, "filter[business-unit]", opts.BusinessUnit)
	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)
	setFilterIfPresent(query, "filter[trucker-id]", opts.TruckerID)
	setFilterIfPresent(query, "filter[trailer]", opts.Trailer)
	setFilterIfPresent(query, "filter[trailer-id]", opts.TrailerID)
	setFilterIfPresent(query, "filter[contractor]", opts.Contractor)
	setFilterIfPresent(query, "filter[contractor-id]", opts.ContractorID)
	setFilterIfPresent(query, "filter[job-number]", opts.JobNumber)
	setFilterIfPresent(query, "filter[invoice]", opts.Invoice)
	setFilterIfPresent(query, "filter[tender-job-schedule-shift]", opts.TenderJobScheduleShift)
	setFilterIfPresent(query, "filter[job-schedule-shift]", opts.JobScheduleShift)
	setFilterIfPresent(query, "filter[job-schedule-shift-id]", opts.JobScheduleShiftID)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[driver]", opts.Driver)
	setFilterIfPresent(query, "filter[driver-id]", opts.DriverID)
	setFilterIfPresent(query, "filter[start-at-min]", opts.StartAtMin)
	setFilterIfPresent(query, "filter[start-at-max]", opts.StartAtMax)
	setFilterIfPresent(query, "filter[is-start-at]", opts.IsStartAt)
	setFilterIfPresent(query, "filter[shift-date]", opts.ShiftDate)
	setFilterIfPresent(query, "filter[shift-date-min]", opts.ShiftDateMin)
	setFilterIfPresent(query, "filter[shift-date-max]", opts.ShiftDateMax)
	setFilterIfPresent(query, "filter[has-shift-date]", opts.HasShiftDate)
	setFilterIfPresent(query, "filter[approved-on-min]", opts.ApprovedOnMin)
	setFilterIfPresent(query, "filter[approved-on-max]", opts.ApprovedOnMax)
	setFilterIfPresent(query, "filter[ticket-number]", opts.TicketNumber)
	setFilterIfPresent(query, "filter[approvable-by]", opts.ApprovableBy)
	setFilterIfPresent(query, "filter[cost-codes]", opts.CostCodes)
	setFilterIfPresent(query, "filter[generate-broker-invoice]", opts.GenerateBrokerInvoice)
	setFilterIfPresent(query, "filter[generate-trucker-invoice]", opts.GenerateTruckerInvoice)
	setFilterIfPresent(query, "filter[has-ticket-report]", opts.HasTicketReport)
	setFilterIfPresent(query, "filter[has-submit-scheduled]", opts.HasSubmitScheduled)

	if opts.ShiftDateBetween != "" {
		parts := strings.SplitN(opts.ShiftDateBetween, "|", 2)
		if len(parts) != 2 {
			return fmt.Errorf("--shift-date-between must be in format date1|date2")
		}
		query.Set("filter[shift-date-between]", parts[0]+","+parts[1])
	}

	body, _, err := client.Get(cmd.Context(), "/v1/time-cards", query)
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

	rows := buildTimeCardRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTimeCardsTable(cmd, rows)
}

func parseTimeCardsListOptions(cmd *cobra.Command) (timeCardsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	status, _ := cmd.Flags().GetString("status")
	isAudited, _ := cmd.Flags().GetString("is-audited")
	withPayrollCertificationRequirement, _ := cmd.Flags().GetString("with-payroll-certification-requirement")
	withPayrollCertificationRequirementMet, _ := cmd.Flags().GetString("with-payroll-certification-requirement-met")
	brokerTender, _ := cmd.Flags().GetString("broker-tender")
	brokerInvoiced, _ := cmd.Flags().GetString("broker-invoiced")
	notBrokerInvoiced, _ := cmd.Flags().GetString("not-broker-invoiced")
	truckerInvoiced, _ := cmd.Flags().GetString("trucker-invoiced")
	notTruckerInvoiced, _ := cmd.Flags().GetString("not-trucker-invoiced")
	approvalCount, _ := cmd.Flags().GetString("approval-count")
	approvalCountMin, _ := cmd.Flags().GetString("approval-count-min")
	approvalCountMax, _ := cmd.Flags().GetString("approval-count-max")
	developer, _ := cmd.Flags().GetString("developer")
	onShortfallBrokerTender, _ := cmd.Flags().GetString("on-shortfall-broker-tender")
	onShortfallCustomerTender, _ := cmd.Flags().GetString("on-shortfall-customer-tender")
	customer, _ := cmd.Flags().GetString("customer")
	customerID, _ := cmd.Flags().GetString("customer-id")
	businessUnit, _ := cmd.Flags().GetString("business-unit")
	trucker, _ := cmd.Flags().GetString("trucker")
	truckerID, _ := cmd.Flags().GetString("trucker-id")
	trailer, _ := cmd.Flags().GetString("trailer")
	trailerID, _ := cmd.Flags().GetString("trailer-id")
	contractor, _ := cmd.Flags().GetString("contractor")
	contractorID, _ := cmd.Flags().GetString("contractor-id")
	jobNumber, _ := cmd.Flags().GetString("job-number")
	invoice, _ := cmd.Flags().GetString("invoice")
	tenderJobScheduleShift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	jobScheduleShift, _ := cmd.Flags().GetString("job-schedule-shift")
	jobScheduleShiftID, _ := cmd.Flags().GetString("job-schedule-shift-id")
	broker, _ := cmd.Flags().GetString("broker")
	driver, _ := cmd.Flags().GetString("driver")
	driverID, _ := cmd.Flags().GetString("driver-id")
	startAtMin, _ := cmd.Flags().GetString("start-at-min")
	startAtMax, _ := cmd.Flags().GetString("start-at-max")
	isStartAt, _ := cmd.Flags().GetString("is-start-at")
	shiftDate, _ := cmd.Flags().GetString("shift-date")
	shiftDateMin, _ := cmd.Flags().GetString("shift-date-min")
	shiftDateMax, _ := cmd.Flags().GetString("shift-date-max")
	shiftDateBetween, _ := cmd.Flags().GetString("shift-date-between")
	hasShiftDate, _ := cmd.Flags().GetString("has-shift-date")
	approvedOnMin, _ := cmd.Flags().GetString("approved-on-min")
	approvedOnMax, _ := cmd.Flags().GetString("approved-on-max")
	ticketNumber, _ := cmd.Flags().GetString("ticket-number")
	approvableBy, _ := cmd.Flags().GetString("approvable-by")
	costCodes, _ := cmd.Flags().GetString("cost-codes")
	generateBrokerInvoice, _ := cmd.Flags().GetString("generate-broker-invoice")
	generateTruckerInvoice, _ := cmd.Flags().GetString("generate-trucker-invoice")
	hasTicketReport, _ := cmd.Flags().GetString("has-ticket-report")
	hasSubmitScheduled, _ := cmd.Flags().GetString("has-submit-scheduled")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return timeCardsListOptions{
		BaseURL:                                baseURL,
		Token:                                  token,
		JSON:                                   jsonOut,
		NoAuth:                                 noAuth,
		Limit:                                  limit,
		Offset:                                 offset,
		Sort:                                   sort,
		Status:                                 status,
		IsAudited:                              isAudited,
		WithPayrollCertificationRequirement:    withPayrollCertificationRequirement,
		WithPayrollCertificationRequirementMet: withPayrollCertificationRequirementMet,
		BrokerTender:                           brokerTender,
		BrokerInvoiced:                         brokerInvoiced,
		NotBrokerInvoiced:                      notBrokerInvoiced,
		TruckerInvoiced:                        truckerInvoiced,
		NotTruckerInvoiced:                     notTruckerInvoiced,
		ApprovalCount:                          approvalCount,
		ApprovalCountMin:                       approvalCountMin,
		ApprovalCountMax:                       approvalCountMax,
		Developer:                              developer,
		OnShortfallBrokerTender:                onShortfallBrokerTender,
		OnShortfallCustomerTender:              onShortfallCustomerTender,
		Customer:                               customer,
		CustomerID:                             customerID,
		BusinessUnit:                           businessUnit,
		Trucker:                                trucker,
		TruckerID:                              truckerID,
		Trailer:                                trailer,
		TrailerID:                              trailerID,
		Contractor:                             contractor,
		ContractorID:                           contractorID,
		JobNumber:                              jobNumber,
		Invoice:                                invoice,
		TenderJobScheduleShift:                 tenderJobScheduleShift,
		JobScheduleShift:                       jobScheduleShift,
		JobScheduleShiftID:                     jobScheduleShiftID,
		Broker:                                 broker,
		Driver:                                 driver,
		DriverID:                               driverID,
		StartAtMin:                             startAtMin,
		StartAtMax:                             startAtMax,
		IsStartAt:                              isStartAt,
		ShiftDate:                              shiftDate,
		ShiftDateMin:                           shiftDateMin,
		ShiftDateMax:                           shiftDateMax,
		ShiftDateBetween:                       shiftDateBetween,
		HasShiftDate:                           hasShiftDate,
		ApprovedOnMin:                          approvedOnMin,
		ApprovedOnMax:                          approvedOnMax,
		TicketNumber:                           ticketNumber,
		ApprovableBy:                           approvableBy,
		CostCodes:                              costCodes,
		GenerateBrokerInvoice:                  generateBrokerInvoice,
		GenerateTruckerInvoice:                 generateTruckerInvoice,
		HasTicketReport:                        hasTicketReport,
		HasSubmitScheduled:                     hasSubmitScheduled,
	}, nil
}

func buildTimeCardRows(resp jsonAPIResponse) []timeCardRow {
	rows := make([]timeCardRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := timeCardRow{
			ID:            resource.ID,
			Status:        stringAttr(attrs, "status"),
			TicketNumber:  stringAttr(attrs, "ticket-number"),
			StartAt:       formatDateTime(stringAttr(attrs, "start-at")),
			EndAt:         formatDateTime(stringAttr(attrs, "end-at")),
			TotalHours:    floatAttr(attrs, "total-hours"),
			ApprovalCount: int(floatAttr(attrs, "approval-count")),
		}

		if rel, ok := resource.Relationships["tender-job-schedule-shift"]; ok && rel.Data != nil {
			row.TenderJobScheduleShift = rel.Data.ID
		}
		if rel, ok := resource.Relationships["job"]; ok && rel.Data != nil {
			row.JobID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
			row.JobProductionPlanID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["driver"]; ok && rel.Data != nil {
			row.DriverID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["trucker"]; ok && rel.Data != nil {
			row.TruckerID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["job-schedule-shift"]; ok && rel.Data != nil {
			row.JobScheduleShiftID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderTimeCardsTable(cmd *cobra.Command, rows []timeCardRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No time cards found.")
		return nil
	}

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tSTATUS\tTICKET\tSTART\tEND\tHOURS\tAPPROVALS\tTRUCKER\tDRIVER\tJOB\tJPP\tSHIFT")
	for _, row := range rows {
		hours := ""
		if row.TotalHours != 0 {
			hours = fmt.Sprintf("%.2f", row.TotalHours)
		}
		approvals := ""
		if row.ApprovalCount != 0 {
			approvals = strconv.Itoa(row.ApprovalCount)
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Status,
			row.TicketNumber,
			row.StartAt,
			row.EndAt,
			hours,
			approvals,
			row.TruckerID,
			row.DriverID,
			row.JobID,
			row.JobProductionPlanID,
			row.TenderJobScheduleShift,
		)
	}
	return w.Flush()
}
