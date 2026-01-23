package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doTimeCardsCreateOptions struct {
	BaseURL                                     string
	Token                                       string
	JSON                                        bool
	BrokerTender                                string
	TenderJobScheduleShift                      string
	TicketNumber                                string
	StartAt                                     string
	EndAt                                       string
	DownMinutes                                 string
	SkipAutoSubmissionUponMaterialTxnAcceptance string
	SubmittedTravelMinutes                      string
	ExplicitIsTrailerRequiredForApproval        string
	ExplicitEnforceTicketNumberUniqueness       string
	IsTimeCardCreatingTimeSheetLineItemExplicit string
	ExplicitIsInvoiceableWhenApproved           string
	EnforceTicketNumberUniqueness               string
	ApprovalProcess                             string
	GenerateBrokerInvoice                       string
	GenerateTruckerInvoice                      string
	SubmittedCustomerTravelMinutes              string
}

func newDoTimeCardsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a time card",
		Long: `Create a time card.

Required flags:
  --broker-tender             Broker tender ID (required)
  --tender-job-schedule-shift Tender job schedule shift ID (required)

Optional flags:
  --ticket-number                                   Ticket number
  --start-at                                        Start time (ISO 8601)
  --end-at                                          End time (ISO 8601)
  --down-minutes                                    Down minutes
  --submitted-travel-minutes                        Submitted travel minutes
  --submitted-customer-travel-minutes               Submitted customer travel minutes
  --skip-auto-submission-upon-material-transaction-acceptance Skip auto submission (true/false)
  --explicit-is-trailer-required-for-approval       Explicit trailer required (true/false)
  --explicit-enforce-ticket-number-uniqueness       Explicit ticket number uniqueness (true/false)
  --enforce-ticket-number-uniqueness                Enforce ticket number uniqueness (true/false)
  --is-time-card-creating-time-sheet-line-item-explicit Create time sheet line item (true/false)
  --explicit-is-invoiceable-when-approved           Explicit invoiceable when approved (true/false)
  --approval-process                                Approval process
  --generate-broker-invoice                          Generate broker invoice (true/false)
  --generate-trucker-invoice                         Generate trucker invoice (true/false)`,
		Example: `  # Create a time card
  xbe do time-cards create \
    --broker-tender 123 \
    --tender-job-schedule-shift 456 \
    --ticket-number "TC-001" \
    --start-at 2025-01-01T07:00:00Z \
    --end-at 2025-01-01T15:00:00Z`,
		Args: cobra.NoArgs,
		RunE: runDoTimeCardsCreate,
	}
	initDoTimeCardsCreateFlags(cmd)
	return cmd
}

func init() {
	doTimeCardsCmd.AddCommand(newDoTimeCardsCreateCmd())
}

func initDoTimeCardsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("broker-tender", "", "Broker tender ID (required)")
	cmd.Flags().String("tender-job-schedule-shift", "", "Tender job schedule shift ID (required)")
	cmd.Flags().String("ticket-number", "", "Ticket number")
	cmd.Flags().String("start-at", "", "Start time (ISO 8601)")
	cmd.Flags().String("end-at", "", "End time (ISO 8601)")
	cmd.Flags().String("down-minutes", "", "Down minutes")
	cmd.Flags().String("submitted-travel-minutes", "", "Submitted travel minutes")
	cmd.Flags().String("submitted-customer-travel-minutes", "", "Submitted customer travel minutes")
	cmd.Flags().String("skip-auto-submission-upon-material-transaction-acceptance", "", "Skip auto submission upon material transaction acceptance (true/false)")
	cmd.Flags().String("explicit-is-trailer-required-for-approval", "", "Explicit trailer required (true/false)")
	cmd.Flags().String("explicit-enforce-ticket-number-uniqueness", "", "Explicit ticket number uniqueness (true/false)")
	cmd.Flags().String("enforce-ticket-number-uniqueness", "", "Enforce ticket number uniqueness (true/false)")
	cmd.Flags().String("is-time-card-creating-time-sheet-line-item-explicit", "", "Create time sheet line item (true/false)")
	cmd.Flags().String("explicit-is-invoiceable-when-approved", "", "Explicit invoiceable when approved (true/false)")
	cmd.Flags().String("approval-process", "", "Approval process")
	cmd.Flags().String("generate-broker-invoice", "", "Generate broker invoice (true/false)")
	cmd.Flags().String("generate-trucker-invoice", "", "Generate trucker invoice (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTimeCardsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTimeCardsCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
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

	if opts.BrokerTender == "" {
		err := fmt.Errorf("--broker-tender is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.TenderJobScheduleShift == "" {
		err := fmt.Errorf("--tender-job-schedule-shift is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.TicketNumber != "" {
		attributes["ticket-number"] = opts.TicketNumber
	}
	if opts.StartAt != "" {
		attributes["start-at"] = opts.StartAt
	}
	if opts.EndAt != "" {
		attributes["end-at"] = opts.EndAt
	}
	if opts.DownMinutes != "" {
		attributes["down-minutes"] = opts.DownMinutes
	}
	if opts.SubmittedTravelMinutes != "" {
		attributes["submitted-travel-minutes"] = opts.SubmittedTravelMinutes
	}
	if opts.SubmittedCustomerTravelMinutes != "" {
		attributes["submitted-customer-travel-minutes"] = opts.SubmittedCustomerTravelMinutes
	}
	if opts.SkipAutoSubmissionUponMaterialTxnAcceptance != "" {
		attributes["skip-auto-submission-upon-material-transaction-acceptance"] = opts.SkipAutoSubmissionUponMaterialTxnAcceptance == "true"
	}
	if opts.ExplicitIsTrailerRequiredForApproval != "" {
		attributes["explicit-is-trailer-required-for-approval"] = opts.ExplicitIsTrailerRequiredForApproval == "true"
	}
	if opts.ExplicitEnforceTicketNumberUniqueness != "" {
		attributes["explicit-enforce-ticket-number-uniqueness"] = opts.ExplicitEnforceTicketNumberUniqueness == "true"
	}
	if opts.EnforceTicketNumberUniqueness != "" {
		attributes["enforce-ticket-number-uniqueness"] = opts.EnforceTicketNumberUniqueness == "true"
	}
	if opts.IsTimeCardCreatingTimeSheetLineItemExplicit != "" {
		attributes["is-time-card-creating-time-sheet-line-item-explicit"] = opts.IsTimeCardCreatingTimeSheetLineItemExplicit == "true"
	}
	if opts.ExplicitIsInvoiceableWhenApproved != "" {
		attributes["explicit-is-invoiceable-when-approved"] = opts.ExplicitIsInvoiceableWhenApproved == "true"
	}
	if opts.ApprovalProcess != "" {
		attributes["approval-process"] = opts.ApprovalProcess
	}
	if opts.GenerateBrokerInvoice != "" {
		attributes["generate-broker-invoice"] = opts.GenerateBrokerInvoice == "true"
	}
	if opts.GenerateTruckerInvoice != "" {
		attributes["generate-trucker-invoice"] = opts.GenerateTruckerInvoice == "true"
	}

	relationships := map[string]any{
		"broker-tender": map[string]any{
			"data": map[string]any{
				"type": "broker-tenders",
				"id":   opts.BrokerTender,
			},
		},
		"tender-job-schedule-shift": map[string]any{
			"data": map[string]any{
				"type": "tender-job-schedule-shifts",
				"id":   opts.TenderJobScheduleShift,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "time-cards",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/time-cards", jsonBody)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPISingleResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	row := timeCardRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created time card %s\n", row.ID)
	return nil
}

func parseDoTimeCardsCreateOptions(cmd *cobra.Command) (doTimeCardsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	brokerTender, _ := cmd.Flags().GetString("broker-tender")
	tenderJobScheduleShift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	ticketNumber, _ := cmd.Flags().GetString("ticket-number")
	startAt, _ := cmd.Flags().GetString("start-at")
	endAt, _ := cmd.Flags().GetString("end-at")
	downMinutes, _ := cmd.Flags().GetString("down-minutes")
	submittedTravelMinutes, _ := cmd.Flags().GetString("submitted-travel-minutes")
	submittedCustomerTravelMinutes, _ := cmd.Flags().GetString("submitted-customer-travel-minutes")
	skipAutoSubmission, _ := cmd.Flags().GetString("skip-auto-submission-upon-material-transaction-acceptance")
	explicitIsTrailerRequiredForApproval, _ := cmd.Flags().GetString("explicit-is-trailer-required-for-approval")
	explicitEnforceTicketNumberUniqueness, _ := cmd.Flags().GetString("explicit-enforce-ticket-number-uniqueness")
	enforceTicketNumberUniqueness, _ := cmd.Flags().GetString("enforce-ticket-number-uniqueness")
	isTimeCardCreatingTimeSheetLineItemExplicit, _ := cmd.Flags().GetString("is-time-card-creating-time-sheet-line-item-explicit")
	explicitIsInvoiceableWhenApproved, _ := cmd.Flags().GetString("explicit-is-invoiceable-when-approved")
	approvalProcess, _ := cmd.Flags().GetString("approval-process")
	generateBrokerInvoice, _ := cmd.Flags().GetString("generate-broker-invoice")
	generateTruckerInvoice, _ := cmd.Flags().GetString("generate-trucker-invoice")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTimeCardsCreateOptions{
		BaseURL:                        baseURL,
		Token:                          token,
		JSON:                           jsonOut,
		BrokerTender:                   brokerTender,
		TenderJobScheduleShift:         tenderJobScheduleShift,
		TicketNumber:                   ticketNumber,
		StartAt:                        startAt,
		EndAt:                          endAt,
		DownMinutes:                    downMinutes,
		SubmittedTravelMinutes:         submittedTravelMinutes,
		SubmittedCustomerTravelMinutes: submittedCustomerTravelMinutes,
		SkipAutoSubmissionUponMaterialTxnAcceptance: skipAutoSubmission,
		ExplicitIsTrailerRequiredForApproval:        explicitIsTrailerRequiredForApproval,
		ExplicitEnforceTicketNumberUniqueness:       explicitEnforceTicketNumberUniqueness,
		EnforceTicketNumberUniqueness:               enforceTicketNumberUniqueness,
		IsTimeCardCreatingTimeSheetLineItemExplicit: isTimeCardCreatingTimeSheetLineItemExplicit,
		ExplicitIsInvoiceableWhenApproved:           explicitIsInvoiceableWhenApproved,
		ApprovalProcess:                             approvalProcess,
		GenerateBrokerInvoice:                       generateBrokerInvoice,
		GenerateTruckerInvoice:                      generateTruckerInvoice,
	}, nil
}

func timeCardRowFromSingle(resp jsonAPISingleResponse) timeCardRow {
	resource := resp.Data
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
	if rel, ok := resource.Relationships["driver"]; ok && rel.Data != nil {
		row.DriverID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["trucker"]; ok && rel.Data != nil {
		row.TruckerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["job-schedule-shift"]; ok && rel.Data != nil {
		row.JobScheduleShiftID = rel.Data.ID
	}

	return row
}
