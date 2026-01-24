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

type tenderJobScheduleShiftsListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Sort    string

	WithDriverAssignmentRefusals                 string
	IsManaged                                    string
	Tender                                       string
	JobScheduleShift                             string
	Job                                          string
	MatchesMaterialPurchaseOrderRelease          string
	DeveloperTruckerCertification                string
	DeveloperTruckerCertificationClass           string
	SellerOperationsContact                      string
	DriverDaySequenceIndex                       string
	TrailerID                                    string
	Trailer                                      string
	TractorID                                    string
	Tractor                                      string
	WithoutSellerOperationsContact               string
	WithSellerOperationsContactAssignedOrDrafted string
	WithTrailerAssignedOrDrafted                 string
	MissingAssignment                            string
	JobProductionPlan                            string
	WithJobProductionPlan                        string
	OnTimeCard                                   string
	IsManagedOrAlive                             string
	ExpectsTimeCards                             string
	TimeCardTicketNumber                         string
	ShiftEndsBefore                              string
	ShiftStartsAfter                             string
	ShiftStartsBefore                            string
	EndAtMin                                     string
	EndAtMax                                     string
	StartDate                                    string
	RelatedTenderStatus                          string
	TenderType                                   string
	TimeCardStatus                               string
	ActiveAsOf                                   string
	Cancelled                                    string
	Trucker                                      string
	Customer                                     string
	Broker                                       string
	SourcedWithTrucker                           string
	AllowsNewTrip                                string
	JobSite                                      string
	Retained                                     string
	TrackedPctMin                                string
	Retainer                                     string
	WithoutReadyToWorkServiceEvent               string
	MissingDriverAssignmentAcknowledgement       string
	TruckerShiftSet                              string
	DefaultTimeCardApprovalProcess               string
	Drivers                                      string
}

type tenderJobScheduleShiftRow struct {
	ID                        string `json:"id"`
	StartAt                   string `json:"start_at,omitempty"`
	TenderID                  string `json:"tender_id,omitempty"`
	TenderType                string `json:"tender_type,omitempty"`
	JobScheduleShiftID        string `json:"job_schedule_shift_id,omitempty"`
	SellerOperationsContactID string `json:"seller_operations_contact_id,omitempty"`
	TrailerID                 string `json:"trailer_id,omitempty"`
	TractorID                 string `json:"tractor_id,omitempty"`
	MaterialTransactionStatus string `json:"material_transaction_status,omitempty"`
}

func newTenderJobScheduleShiftsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tender job schedule shifts",
		Long: `List tender job schedule shifts.

Output Columns:
  ID        Tender job schedule shift ID
  START AT  Scheduled start time
  TENDER    Tender type and ID
  JOB SHIFT Job schedule shift ID
  DRIVER    Seller operations contact (driver) ID
  TRAILER   Trailer ID
  TRACTOR   Tractor ID
  STATUS    Material transaction status

Filters:
  --with-driver-assignment-refusals             Filter shifts with driver assignment refusals (true/false)
  --is-managed                                 Filter by managed flag (true/false)
  --tender                                     Filter by tender ID
  --job-schedule-shift                          Filter by job schedule shift ID
  --job                                        Filter by job ID
  --matches-material-purchase-order-release    Filter by material purchase order release ID
  --developer-trucker-certification            Filter by developer trucker certification ID
  --developer-trucker-certification-classification Filter by certification classification ID
  --seller-operations-contact                  Filter by seller operations contact (driver) ID
  --driver-day-sequence-index                  Filter by driver day sequence index
  --trailer-id                                 Filter by trailer ID (via shift schedule)
  --trailer                                    Filter by trailer ID
  --tractor-id                                 Filter by tractor ID (via shift schedule)
  --tractor                                    Filter by tractor ID
  --without-seller-operations-contact          Filter shifts without seller operations contact (true/false)
  --with-seller-operations-contact-assigned-or-drafted Filter shifts with assigned or drafted driver (true/false)
  --with-trailer-assigned-or-drafted           Filter shifts with assigned or drafted trailer (true/false)
  --missing-assignment                         Filter shifts missing assignment (true/false)
  --job-production-plan                        Filter by job production plan ID
  --with-job-production-plan                   Filter shifts with job production plan (true/false)
  --on-time-card                               Filter shifts on time card (true/false)
  --is-managed-or-alive                        Filter managed or alive (true/false)
  --expects-time-cards                         Filter by expects time cards (true/false)
  --time-card-ticket-number                    Filter by time card ticket number
  --shift-ends-before                          Filter shifts ending before timestamp
  --shift-starts-after                         Filter shifts starting after timestamp
  --shift-starts-before                        Filter shifts starting before timestamp
  --end-at-min                                 Filter shifts ending after timestamp
  --end-at-max                                 Filter shifts ending before timestamp
  --start-date                                 Filter by shift start date (YYYY-MM-DD)
  --related-tender-status                      Filter by related tender status
  --tender-type                                Filter by tender type
  --time-card-status                           Filter by time card status
  --active-as-of                               Filter active shifts as of timestamp
  --cancelled                                  Filter by cancelled flag (true/false)
  --trucker                                    Filter by trucker ID
  --customer                                   Filter by customer ID
  --broker                                     Filter by broker ID
  --sourced-with-trucker                       Filter by sourced-with trucker ID
  --allows-new-trip                            Filter by allows new trip (true/false)
  --job-site                                   Filter by job site ID
  --retained                                   Filter by retained flag (true/false)
  --tracked-pct-min                            Filter by minimum tracked percent
  --retainer                                   Filter by retainer ID
  --without-ready-to-work-service-event        Filter shifts without ready-to-work service event (true/false)
  --missing-driver-assignment-acknowledgement  Filter missing driver assignment acknowledgement (true/false)
  --trucker-shift-set                          Filter by trucker shift set (driver day) ID
  --default-time-card-approval-process         Filter by default time card approval process (admin/field)
  --drivers                                    Filter by driver IDs

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List tender job schedule shifts
  xbe view tender-job-schedule-shifts list

  # Filter by tender
  xbe view tender-job-schedule-shifts list --tender 123

  # Filter by driver
  xbe view tender-job-schedule-shifts list --seller-operations-contact 456

  # Filter by date
  xbe view tender-job-schedule-shifts list --start-date 2025-01-01

  # Output as JSON
  xbe view tender-job-schedule-shifts list --json`,
		Args: cobra.NoArgs,
		RunE: runTenderJobScheduleShiftsList,
	}
	initTenderJobScheduleShiftsListFlags(cmd)
	return cmd
}

func init() {
	tenderJobScheduleShiftsCmd.AddCommand(newTenderJobScheduleShiftsListCmd())
}

func initTenderJobScheduleShiftsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")

	cmd.Flags().String("with-driver-assignment-refusals", "", "Filter shifts with driver assignment refusals (true/false)")
	cmd.Flags().String("is-managed", "", "Filter by managed flag (true/false)")
	cmd.Flags().String("tender", "", "Filter by tender ID")
	cmd.Flags().String("job-schedule-shift", "", "Filter by job schedule shift ID")
	cmd.Flags().String("job", "", "Filter by job ID")
	cmd.Flags().String("matches-material-purchase-order-release", "", "Filter by material purchase order release ID")
	cmd.Flags().String("developer-trucker-certification", "", "Filter by developer trucker certification ID")
	cmd.Flags().String("developer-trucker-certification-classification", "", "Filter by certification classification ID")
	cmd.Flags().String("seller-operations-contact", "", "Filter by seller operations contact (driver) ID")
	cmd.Flags().String("driver-day-sequence-index", "", "Filter by driver day sequence index")
	cmd.Flags().String("trailer-id", "", "Filter by trailer ID (via shift schedule)")
	cmd.Flags().String("trailer", "", "Filter by trailer ID")
	cmd.Flags().String("tractor-id", "", "Filter by tractor ID (via shift schedule)")
	cmd.Flags().String("tractor", "", "Filter by tractor ID")
	cmd.Flags().String("without-seller-operations-contact", "", "Filter shifts without seller operations contact (true/false)")
	cmd.Flags().String("with-seller-operations-contact-assigned-or-drafted", "", "Filter shifts with assigned or drafted driver (true/false)")
	cmd.Flags().String("with-trailer-assigned-or-drafted", "", "Filter shifts with assigned or drafted trailer (true/false)")
	cmd.Flags().String("missing-assignment", "", "Filter shifts missing assignment (true/false)")
	cmd.Flags().String("job-production-plan", "", "Filter by job production plan ID")
	cmd.Flags().String("with-job-production-plan", "", "Filter shifts with job production plan (true/false)")
	cmd.Flags().String("on-time-card", "", "Filter shifts on time card (true/false)")
	cmd.Flags().String("is-managed-or-alive", "", "Filter managed or alive (true/false)")
	cmd.Flags().String("expects-time-cards", "", "Filter by expects time cards (true/false)")
	cmd.Flags().String("time-card-ticket-number", "", "Filter by time card ticket number")
	cmd.Flags().String("shift-ends-before", "", "Filter shifts ending before timestamp")
	cmd.Flags().String("shift-starts-after", "", "Filter shifts starting after timestamp")
	cmd.Flags().String("shift-starts-before", "", "Filter shifts starting before timestamp")
	cmd.Flags().String("end-at-min", "", "Filter shifts ending after timestamp")
	cmd.Flags().String("end-at-max", "", "Filter shifts ending before timestamp")
	cmd.Flags().String("start-date", "", "Filter by shift start date (YYYY-MM-DD)")
	cmd.Flags().String("related-tender-status", "", "Filter by related tender status")
	cmd.Flags().String("tender-type", "", "Filter by tender type")
	cmd.Flags().String("time-card-status", "", "Filter by time card status")
	cmd.Flags().String("active-as-of", "", "Filter active shifts as of timestamp")
	cmd.Flags().String("cancelled", "", "Filter by cancelled flag (true/false)")
	cmd.Flags().String("trucker", "", "Filter by trucker ID")
	cmd.Flags().String("customer", "", "Filter by customer ID")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("sourced-with-trucker", "", "Filter by sourced-with trucker ID")
	cmd.Flags().String("allows-new-trip", "", "Filter by allows new trip (true/false)")
	cmd.Flags().String("job-site", "", "Filter by job site ID")
	cmd.Flags().String("retained", "", "Filter by retained flag (true/false)")
	cmd.Flags().String("tracked-pct-min", "", "Filter by minimum tracked percent")
	cmd.Flags().String("retainer", "", "Filter by retainer ID")
	cmd.Flags().String("without-ready-to-work-service-event", "", "Filter shifts without ready-to-work service event (true/false)")
	cmd.Flags().String("missing-driver-assignment-acknowledgement", "", "Filter missing driver assignment acknowledgement (true/false)")
	cmd.Flags().String("trucker-shift-set", "", "Filter by trucker shift set (driver day) ID")
	cmd.Flags().String("default-time-card-approval-process", "", "Filter by default time card approval process (admin/field)")
	cmd.Flags().String("drivers", "", "Filter by driver IDs")

	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTenderJobScheduleShiftsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTenderJobScheduleShiftsListOptions(cmd)
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
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[with_driver_assignment_refusals]", opts.WithDriverAssignmentRefusals)
	setFilterIfPresent(query, "filter[is_managed]", opts.IsManaged)
	setFilterIfPresent(query, "filter[tender]", opts.Tender)
	setFilterIfPresent(query, "filter[job_schedule_shift]", opts.JobScheduleShift)
	setFilterIfPresent(query, "filter[job]", opts.Job)
	setFilterIfPresent(query, "filter[matches_material_purchase_order_release]", opts.MatchesMaterialPurchaseOrderRelease)
	setFilterIfPresent(query, "filter[developer_trucker_certification]", opts.DeveloperTruckerCertification)
	setFilterIfPresent(query, "filter[developer_trucker_certification_classification]", opts.DeveloperTruckerCertificationClass)
	setFilterIfPresent(query, "filter[seller_operations_contact]", opts.SellerOperationsContact)
	setFilterIfPresent(query, "filter[driver_day_sequence_index]", opts.DriverDaySequenceIndex)
	setFilterIfPresent(query, "filter[trailer_id]", opts.TrailerID)
	setFilterIfPresent(query, "filter[trailer]", opts.Trailer)
	setFilterIfPresent(query, "filter[tractor_id]", opts.TractorID)
	setFilterIfPresent(query, "filter[tractor]", opts.Tractor)
	setFilterIfPresent(query, "filter[without_seller_operations_contact]", opts.WithoutSellerOperationsContact)
	setFilterIfPresent(query, "filter[with_seller_operations_contact_assigned_or_drafted]", opts.WithSellerOperationsContactAssignedOrDrafted)
	setFilterIfPresent(query, "filter[with_trailer_assigned_or_drafted]", opts.WithTrailerAssignedOrDrafted)
	setFilterIfPresent(query, "filter[missing_assignment]", opts.MissingAssignment)
	setFilterIfPresent(query, "filter[job_production_plan]", opts.JobProductionPlan)
	setFilterIfPresent(query, "filter[with_job_production_plan]", opts.WithJobProductionPlan)
	setFilterIfPresent(query, "filter[on_time_card]", opts.OnTimeCard)
	setFilterIfPresent(query, "filter[is_managed_or_alive]", opts.IsManagedOrAlive)
	setFilterIfPresent(query, "filter[expects_time_cards]", opts.ExpectsTimeCards)
	setFilterIfPresent(query, "filter[time_card_ticket_number]", opts.TimeCardTicketNumber)
	setFilterIfPresent(query, "filter[shift_ends_before]", opts.ShiftEndsBefore)
	setFilterIfPresent(query, "filter[shift_starts_after]", opts.ShiftStartsAfter)
	setFilterIfPresent(query, "filter[shift_starts_before]", opts.ShiftStartsBefore)
	setFilterIfPresent(query, "filter[end_at_min]", opts.EndAtMin)
	setFilterIfPresent(query, "filter[end_at_max]", opts.EndAtMax)
	setFilterIfPresent(query, "filter[start_date]", opts.StartDate)
	setFilterIfPresent(query, "filter[related_tender_status]", opts.RelatedTenderStatus)
	setFilterIfPresent(query, "filter[tender_type]", opts.TenderType)
	setFilterIfPresent(query, "filter[time_card_status]", opts.TimeCardStatus)
	setFilterIfPresent(query, "filter[active_as_of]", opts.ActiveAsOf)
	setFilterIfPresent(query, "filter[cancelled]", opts.Cancelled)
	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)
	setFilterIfPresent(query, "filter[customer]", opts.Customer)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[sourced_with_trucker]", opts.SourcedWithTrucker)
	setFilterIfPresent(query, "filter[allows_new_trip]", opts.AllowsNewTrip)
	setFilterIfPresent(query, "filter[job_site]", opts.JobSite)
	setFilterIfPresent(query, "filter[retained]", opts.Retained)
	setFilterIfPresent(query, "filter[tracked_pct_min]", opts.TrackedPctMin)
	setFilterIfPresent(query, "filter[retainer]", opts.Retainer)
	setFilterIfPresent(query, "filter[without_ready_to_work_service_event]", opts.WithoutReadyToWorkServiceEvent)
	setFilterIfPresent(query, "filter[missing_driver_assignment_acknowledgement]", opts.MissingDriverAssignmentAcknowledgement)
	setFilterIfPresent(query, "filter[trucker_shift_set]", opts.TruckerShiftSet)
	setFilterIfPresent(query, "filter[default_time_card_approval_process]", opts.DefaultTimeCardApprovalProcess)
	setFilterIfPresent(query, "filter[drivers]", opts.Drivers)

	body, _, err := client.Get(cmd.Context(), "/v1/tender-job-schedule-shifts", query)
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

	rows := buildTenderJobScheduleShiftRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTenderJobScheduleShiftsTable(cmd, rows)
}

func parseTenderJobScheduleShiftsListOptions(cmd *cobra.Command) (tenderJobScheduleShiftsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")

	withDriverAssignmentRefusals, _ := cmd.Flags().GetString("with-driver-assignment-refusals")
	isManaged, _ := cmd.Flags().GetString("is-managed")
	tender, _ := cmd.Flags().GetString("tender")
	jobScheduleShift, _ := cmd.Flags().GetString("job-schedule-shift")
	job, _ := cmd.Flags().GetString("job")
	matchesMaterialPurchaseOrderRelease, _ := cmd.Flags().GetString("matches-material-purchase-order-release")
	developerTruckerCertification, _ := cmd.Flags().GetString("developer-trucker-certification")
	developerTruckerCertificationClass, _ := cmd.Flags().GetString("developer-trucker-certification-classification")
	sellerOperationsContact, _ := cmd.Flags().GetString("seller-operations-contact")
	driverDaySequenceIndex, _ := cmd.Flags().GetString("driver-day-sequence-index")
	trailerID, _ := cmd.Flags().GetString("trailer-id")
	trailer, _ := cmd.Flags().GetString("trailer")
	tractorID, _ := cmd.Flags().GetString("tractor-id")
	tractor, _ := cmd.Flags().GetString("tractor")
	withoutSellerOperationsContact, _ := cmd.Flags().GetString("without-seller-operations-contact")
	withSellerOperationsContactAssignedOrDrafted, _ := cmd.Flags().GetString("with-seller-operations-contact-assigned-or-drafted")
	withTrailerAssignedOrDrafted, _ := cmd.Flags().GetString("with-trailer-assigned-or-drafted")
	missingAssignment, _ := cmd.Flags().GetString("missing-assignment")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	withJobProductionPlan, _ := cmd.Flags().GetString("with-job-production-plan")
	onTimeCard, _ := cmd.Flags().GetString("on-time-card")
	isManagedOrAlive, _ := cmd.Flags().GetString("is-managed-or-alive")
	expectsTimeCards, _ := cmd.Flags().GetString("expects-time-cards")
	timeCardTicketNumber, _ := cmd.Flags().GetString("time-card-ticket-number")
	shiftEndsBefore, _ := cmd.Flags().GetString("shift-ends-before")
	shiftStartsAfter, _ := cmd.Flags().GetString("shift-starts-after")
	shiftStartsBefore, _ := cmd.Flags().GetString("shift-starts-before")
	endAtMin, _ := cmd.Flags().GetString("end-at-min")
	endAtMax, _ := cmd.Flags().GetString("end-at-max")
	startDate, _ := cmd.Flags().GetString("start-date")
	relatedTenderStatus, _ := cmd.Flags().GetString("related-tender-status")
	tenderType, _ := cmd.Flags().GetString("tender-type")
	timeCardStatus, _ := cmd.Flags().GetString("time-card-status")
	activeAsOf, _ := cmd.Flags().GetString("active-as-of")
	cancelled, _ := cmd.Flags().GetString("cancelled")
	trucker, _ := cmd.Flags().GetString("trucker")
	customer, _ := cmd.Flags().GetString("customer")
	broker, _ := cmd.Flags().GetString("broker")
	sourcedWithTrucker, _ := cmd.Flags().GetString("sourced-with-trucker")
	allowsNewTrip, _ := cmd.Flags().GetString("allows-new-trip")
	jobSite, _ := cmd.Flags().GetString("job-site")
	retained, _ := cmd.Flags().GetString("retained")
	trackedPctMin, _ := cmd.Flags().GetString("tracked-pct-min")
	retainer, _ := cmd.Flags().GetString("retainer")
	withoutReadyToWorkServiceEvent, _ := cmd.Flags().GetString("without-ready-to-work-service-event")
	missingDriverAssignmentAcknowledgement, _ := cmd.Flags().GetString("missing-driver-assignment-acknowledgement")
	truckerShiftSet, _ := cmd.Flags().GetString("trucker-shift-set")
	defaultTimeCardApprovalProcess, _ := cmd.Flags().GetString("default-time-card-approval-process")
	drivers, _ := cmd.Flags().GetString("drivers")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return tenderJobScheduleShiftsListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,

		WithDriverAssignmentRefusals:        withDriverAssignmentRefusals,
		IsManaged:                           isManaged,
		Tender:                              tender,
		JobScheduleShift:                    jobScheduleShift,
		Job:                                 job,
		MatchesMaterialPurchaseOrderRelease: matchesMaterialPurchaseOrderRelease,
		DeveloperTruckerCertification:       developerTruckerCertification,
		DeveloperTruckerCertificationClass:  developerTruckerCertificationClass,
		SellerOperationsContact:             sellerOperationsContact,
		DriverDaySequenceIndex:              driverDaySequenceIndex,
		TrailerID:                           trailerID,
		Trailer:                             trailer,
		TractorID:                           tractorID,
		Tractor:                             tractor,
		WithoutSellerOperationsContact:      withoutSellerOperationsContact,
		WithSellerOperationsContactAssignedOrDrafted: withSellerOperationsContactAssignedOrDrafted,
		WithTrailerAssignedOrDrafted:                 withTrailerAssignedOrDrafted,
		MissingAssignment:                            missingAssignment,
		JobProductionPlan:                            jobProductionPlan,
		WithJobProductionPlan:                        withJobProductionPlan,
		OnTimeCard:                                   onTimeCard,
		IsManagedOrAlive:                             isManagedOrAlive,
		ExpectsTimeCards:                             expectsTimeCards,
		TimeCardTicketNumber:                         timeCardTicketNumber,
		ShiftEndsBefore:                              shiftEndsBefore,
		ShiftStartsAfter:                             shiftStartsAfter,
		ShiftStartsBefore:                            shiftStartsBefore,
		EndAtMin:                                     endAtMin,
		EndAtMax:                                     endAtMax,
		StartDate:                                    startDate,
		RelatedTenderStatus:                          relatedTenderStatus,
		TenderType:                                   tenderType,
		TimeCardStatus:                               timeCardStatus,
		ActiveAsOf:                                   activeAsOf,
		Cancelled:                                    cancelled,
		Trucker:                                      trucker,
		Customer:                                     customer,
		Broker:                                       broker,
		SourcedWithTrucker:                           sourcedWithTrucker,
		AllowsNewTrip:                                allowsNewTrip,
		JobSite:                                      jobSite,
		Retained:                                     retained,
		TrackedPctMin:                                trackedPctMin,
		Retainer:                                     retainer,
		WithoutReadyToWorkServiceEvent:               withoutReadyToWorkServiceEvent,
		MissingDriverAssignmentAcknowledgement:       missingDriverAssignmentAcknowledgement,
		TruckerShiftSet:                              truckerShiftSet,
		DefaultTimeCardApprovalProcess:               defaultTimeCardApprovalProcess,
		Drivers:                                      drivers,
	}, nil
}

func buildTenderJobScheduleShiftRows(resp jsonAPIResponse) []tenderJobScheduleShiftRow {
	rows := make([]tenderJobScheduleShiftRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := tenderJobScheduleShiftRow{
			ID:                        resource.ID,
			StartAt:                   formatDateTime(stringAttr(attrs, "start-at")),
			TenderType:                stringAttr(attrs, "tender-type"),
			MaterialTransactionStatus: stringAttr(attrs, "material-transaction-status"),
			TenderID:                  relationshipIDFromMap(resource.Relationships, "tender"),
			JobScheduleShiftID:        relationshipIDFromMap(resource.Relationships, "job-schedule-shift"),
			SellerOperationsContactID: relationshipIDFromMap(resource.Relationships, "seller-operations-contact"),
			TrailerID:                 relationshipIDFromMap(resource.Relationships, "trailer"),
			TractorID:                 relationshipIDFromMap(resource.Relationships, "tractor"),
		}

		if row.TenderType == "" {
			if rel, ok := resource.Relationships["tender"]; ok && rel.Data != nil {
				row.TenderType = rel.Data.Type
			}
		}

		rows = append(rows, row)
	}
	return rows
}

func buildTenderJobScheduleShiftRowFromSingle(resp jsonAPISingleResponse) tenderJobScheduleShiftRow {
	rows := buildTenderJobScheduleShiftRows(jsonAPIResponse{Data: []jsonAPIResource{resp.Data}})
	if len(rows) == 0 {
		return tenderJobScheduleShiftRow{}
	}
	return rows[0]
}

func renderTenderJobScheduleShiftsTable(cmd *cobra.Command, rows []tenderJobScheduleShiftRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No tender job schedule shifts found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTART AT\tTENDER\tJOB SHIFT\tDRIVER\tTRAILER\tTRACTOR\tSTATUS")
	for _, row := range rows {
		tenderRef := ""
		if row.TenderID != "" && row.TenderType != "" {
			tenderRef = fmt.Sprintf("%s:%s", row.TenderType, row.TenderID)
		} else if row.TenderID != "" {
			tenderRef = row.TenderID
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.StartAt,
			tenderRef,
			row.JobScheduleShiftID,
			row.SellerOperationsContactID,
			row.TrailerID,
			row.TractorID,
			row.MaterialTransactionStatus,
		)
	}
	return writer.Flush()
}
