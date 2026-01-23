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

type doTimeCardsUpdateOptions struct {
	BaseURL                                     string
	Token                                       string
	JSON                                        bool
	ID                                          string
	TicketNumber                                string
	StartAt                                     string
	EndAt                                       string
	DownMinutes                                 string
	SubmittedTravelMinutes                      string
	SubmittedCustomerTravelMinutes              string
	SkipAutoSubmissionUponMaterialTxnAcceptance string
	ExplicitIsTrailerRequiredForApproval        string
	ExplicitEnforceTicketNumberUniqueness       string
	EnforceTicketNumberUniqueness               string
	IsTimeCardCreatingTimeSheetLineItemExplicit string
	ExplicitIsInvoiceableWhenApproved           string
	ApprovalProcess                             string
	GenerateBrokerInvoice                       string
	GenerateTruckerInvoice                      string
	ResetSubmitAt                               bool
}

func newDoTimeCardsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a time card",
		Long: `Update a time card.

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
  --generate-trucker-invoice                         Generate trucker invoice (true/false)
  --reset-submit-at                                 Reset submit timestamp`,
		Example: `  # Update ticket number
  xbe do time-cards update 123 --ticket-number \"TC-002\"

  # Reset submit timestamp
  xbe do time-cards update 123 --reset-submit-at`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTimeCardsUpdate,
	}
	initDoTimeCardsUpdateFlags(cmd)
	return cmd
}

func init() {
	doTimeCardsCmd.AddCommand(newDoTimeCardsUpdateCmd())
}

func initDoTimeCardsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
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
	cmd.Flags().Bool("reset-submit-at", false, "Reset submit timestamp")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTimeCardsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTimeCardsUpdateOptions(cmd, args)
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

	attributes := map[string]any{}
	if cmd.Flags().Changed("ticket-number") {
		attributes["ticket-number"] = opts.TicketNumber
	}
	if cmd.Flags().Changed("start-at") {
		attributes["start-at"] = opts.StartAt
	}
	if cmd.Flags().Changed("end-at") {
		attributes["end-at"] = opts.EndAt
	}
	if cmd.Flags().Changed("down-minutes") {
		attributes["down-minutes"] = opts.DownMinutes
	}
	if cmd.Flags().Changed("submitted-travel-minutes") {
		attributes["submitted-travel-minutes"] = opts.SubmittedTravelMinutes
	}
	if cmd.Flags().Changed("submitted-customer-travel-minutes") {
		attributes["submitted-customer-travel-minutes"] = opts.SubmittedCustomerTravelMinutes
	}
	if cmd.Flags().Changed("skip-auto-submission-upon-material-transaction-acceptance") {
		attributes["skip-auto-submission-upon-material-transaction-acceptance"] = opts.SkipAutoSubmissionUponMaterialTxnAcceptance == "true"
	}
	if cmd.Flags().Changed("explicit-is-trailer-required-for-approval") {
		attributes["explicit-is-trailer-required-for-approval"] = opts.ExplicitIsTrailerRequiredForApproval == "true"
	}
	if cmd.Flags().Changed("explicit-enforce-ticket-number-uniqueness") {
		attributes["explicit-enforce-ticket-number-uniqueness"] = opts.ExplicitEnforceTicketNumberUniqueness == "true"
	}
	if cmd.Flags().Changed("enforce-ticket-number-uniqueness") {
		attributes["enforce-ticket-number-uniqueness"] = opts.EnforceTicketNumberUniqueness == "true"
	}
	if cmd.Flags().Changed("is-time-card-creating-time-sheet-line-item-explicit") {
		attributes["is-time-card-creating-time-sheet-line-item-explicit"] = opts.IsTimeCardCreatingTimeSheetLineItemExplicit == "true"
	}
	if cmd.Flags().Changed("explicit-is-invoiceable-when-approved") {
		attributes["explicit-is-invoiceable-when-approved"] = opts.ExplicitIsInvoiceableWhenApproved == "true"
	}
	if cmd.Flags().Changed("approval-process") {
		attributes["approval-process"] = opts.ApprovalProcess
	}
	if cmd.Flags().Changed("generate-broker-invoice") {
		attributes["generate-broker-invoice"] = opts.GenerateBrokerInvoice == "true"
	}
	if cmd.Flags().Changed("generate-trucker-invoice") {
		attributes["generate-trucker-invoice"] = opts.GenerateTruckerInvoice == "true"
	}
	if cmd.Flags().Changed("reset-submit-at") {
		attributes["reset-submit-at"] = opts.ResetSubmitAt
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "time-cards",
			"id":         opts.ID,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/time-cards/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated time card %s\n", row.ID)
	return nil
}

func parseDoTimeCardsUpdateOptions(cmd *cobra.Command, args []string) (doTimeCardsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
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
	resetSubmitAt, _ := cmd.Flags().GetBool("reset-submit-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTimeCardsUpdateOptions{
		BaseURL:                        baseURL,
		Token:                          token,
		JSON:                           jsonOut,
		ID:                             args[0],
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
		ResetSubmitAt:                               resetSubmitAt,
	}, nil
}
