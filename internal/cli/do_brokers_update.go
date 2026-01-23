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

type doBrokersUpdateOptions struct {
	BaseURL                                               string
	Token                                                 string
	JSON                                                  bool
	Name                                                  string
	Abbreviation                                          string
	DefaultTruckerPaymentTerms                            int
	DefaultCustomerPaymentTerms                           int
	IsTransportOnly                                       string
	DefaultReplyToEmailAddress                            string
	IsActive                                              string
	EnableImplicitTimeCardApproval                        string
	RemitToAddress                                        string
	SendLineupSummariesTo                                 string
	QuickbooksEnabled                                     string
	IsNonDriverPermittedToCheckIn                         string
	IsGeneratingAutomatedShiftFeedback                    string
	IsManagingQualityControlRequirements                  string
	IsManagingDriverVisibility                            string
	SkipMaterialTransactionImageExtraction                string
	HelpText                                              string
	MakeTruckerReportCardVisibleToTruckers                string
	PreferPublicDispatchPhoneNumber                       string
	PublicDispatchPhoneNumberExplicit                     string
	SkipTenderJobScheduleShiftStartingSellerNotifications string
	IsAcceptingOpenDoorIssues                             string
	CanCustomersSeeDriverContactInformation               string
	CanCustomerOperationsSeeDriverContactInformation      string
	ShiftFeedbackReasonNotificationExclusions             string
	DisabledFeedbackTypes                                 string
	QuickbooksEnabledCustomerIds                          string
	EnableEquipmentMovement                               string
	MinDurationOfAutoTruckingIncidentWithDownTime         int
	RequiresCostCodeAllocations                           string
	JobProductionPlanRecapTemplate                        string
	DefaultPredictionSubjectKind                          string
	DefaultPredictionSubjectLeadTimeHours                 int
	ModeledToProjectedConfidenceThreshold                 string
	ModeledToActualConfidenceThreshold                    string
	SlackNtfyChannel                                      string
	SlackNtfyIcon                                         string
	SlackHorizonChannel                                   string
	ActiveEquipmentRentalNotificationDays                 string
	DefaultCustomerRateAgreementTemplate                  string
	DefaultFinancialContact                               string
	DefaultOperationsContact                              string
	DefaultDispatchContact                                string
	Developer                                             string
}

func newDoBrokersUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a broker",
		Long: `Update an existing broker.

Only the fields you specify will be updated. Fields not provided will remain unchanged.

Arguments:
  <id>    The broker ID (required)

Flags (see --help for full list):
  --name                            Update the company name
  --abbreviation                    Update abbreviation
  --default-trucker-payment-terms   Update default trucker payment terms (days)
  --default-customer-payment-terms  Update default customer payment terms (days)
  --is-transport-only               Update transport only status (true/false)
  --default-reply-to-email          Update default reply-to email
  --is-active                       Update active status (true/false, admin only)
  --default-financial-contact       Update default financial contact user ID
  --default-operations-contact      Update default operations contact user ID
  --default-dispatch-contact        Update default dispatch contact user ID`,
		Example: `  # Update the name
  xbe do brokers update 123 --name "New Broker Name"

  # Update abbreviation
  xbe do brokers update 123 --abbreviation "NBN"

  # Update payment terms (30 days)
  xbe do brokers update 123 --default-trucker-payment-terms 30

  # Deactivate a broker
  xbe do brokers update 123 --is-active false

  # Get JSON output
  xbe do brokers update 123 --name "Updated" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoBrokersUpdate,
	}
	initDoBrokersUpdateFlags(cmd)
	return cmd
}

func init() {
	doBrokersCmd.AddCommand(newDoBrokersUpdateCmd())
}

func initDoBrokersUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "New company name")
	cmd.Flags().String("abbreviation", "", "New abbreviation")
	cmd.Flags().Int("default-trucker-payment-terms", 0, "New default trucker payment terms (days)")
	cmd.Flags().Int("default-customer-payment-terms", 0, "New default customer payment terms (days)")
	cmd.Flags().String("is-transport-only", "", "Update transport only status (true/false)")
	cmd.Flags().String("default-reply-to-email", "", "New default reply-to email")
	cmd.Flags().String("is-active", "", "Update active status (true/false, admin only)")
	cmd.Flags().String("enable-implicit-time-card-approval", "", "Enable implicit time card approval (true/false)")
	cmd.Flags().String("remit-to-address", "", "Remit-to address")
	cmd.Flags().String("send-lineup-summaries-to", "", "Send lineup summaries to (email addresses)")
	cmd.Flags().String("quickbooks-enabled", "", "QuickBooks enabled (true/false)")
	cmd.Flags().String("is-non-driver-permitted-to-check-in", "", "Non-driver permitted to check in (true/false)")
	cmd.Flags().String("is-generating-automated-shift-feedback", "", "Generating automated shift feedback (true/false)")
	cmd.Flags().String("is-managing-quality-control-requirements", "", "Managing quality control requirements (true/false)")
	cmd.Flags().String("is-managing-driver-visibility", "", "Managing driver visibility (true/false)")
	cmd.Flags().String("skip-material-transaction-image-extraction", "", "Skip material transaction image extraction (true/false)")
	cmd.Flags().String("help-text", "", "Help text")
	cmd.Flags().String("make-trucker-report-card-visible-to-truckers", "", "Make trucker report card visible to truckers (true/false)")
	cmd.Flags().String("prefer-public-dispatch-phone-number", "", "Prefer public dispatch phone number (true/false)")
	cmd.Flags().String("public-dispatch-phone-number-explicit", "", "Public dispatch phone number")
	cmd.Flags().String("skip-tender-job-schedule-shift-starting-seller-notifications", "", "Skip shift starting seller notifications (true/false)")
	cmd.Flags().String("is-accepting-open-door-issues", "", "Accepting open door issues (true/false)")
	cmd.Flags().String("can-customers-see-driver-contact-information", "", "Customers can see driver contact info (true/false)")
	cmd.Flags().String("can-customer-operations-see-driver-contact-information", "", "Customer operations can see driver contact info (true/false)")
	cmd.Flags().String("shift-feedback-reason-notification-exclusions", "", "Shift feedback reason notification exclusions")
	cmd.Flags().String("disabled-feedback-types", "", "Disabled feedback types")
	cmd.Flags().String("quickbooks-enabled-customer-ids", "", "QuickBooks enabled customer IDs")
	cmd.Flags().String("enable-equipment-movement", "", "Enable equipment movement (true/false)")
	cmd.Flags().Int("min-duration-of-auto-trucking-incident-with-down-time", 0, "Min duration of auto trucking incident with down time (minutes)")
	cmd.Flags().String("requires-cost-code-allocations", "", "Requires cost code allocations (true/false)")
	cmd.Flags().String("job-production-plan-recap-template", "", "Job production plan recap template")
	cmd.Flags().String("default-prediction-subject-kind", "", "Default prediction subject kind")
	cmd.Flags().Int("default-prediction-subject-lead-time-hours", 0, "Default prediction subject lead time hours")
	cmd.Flags().String("modeled-to-projected-confidence-threshold", "", "Modeled to projected confidence threshold")
	cmd.Flags().String("modeled-to-actual-confidence-threshold", "", "Modeled to actual confidence threshold")
	cmd.Flags().String("slack-ntfy-channel", "", "Slack notification channel (admin only)")
	cmd.Flags().String("slack-ntfy-icon", "", "Slack notification icon (admin only)")
	cmd.Flags().String("slack-horizon-channel", "", "Slack horizon channel (admin only)")
	cmd.Flags().String("active-equipment-rental-notification-days", "", "Active equipment rental notification days (JSON array, e.g. \"[1,2,3,4,5]\")")
	cmd.Flags().String("default-customer-rate-agreement-template", "", "Default customer rate agreement template ID")
	cmd.Flags().String("default-financial-contact", "", "Default financial contact user ID")
	cmd.Flags().String("default-operations-contact", "", "Default operations contact user ID")
	cmd.Flags().String("default-dispatch-contact", "", "Default dispatch contact user ID")
	cmd.Flags().String("developer", "", "Developer ID (admin only)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoBrokersUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoBrokersUpdateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Require authentication for write operations
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("broker id is required")
	}

	// Check if any field to update was provided
	truckerTermsChanged := cmd.Flags().Changed("default-trucker-payment-terms")
	customerTermsChanged := cmd.Flags().Changed("default-customer-payment-terms")
	minDurationChanged := cmd.Flags().Changed("min-duration-of-auto-trucking-incident-with-down-time")
	leadTimeHoursChanged := cmd.Flags().Changed("default-prediction-subject-lead-time-hours")
	hasUpdate := opts.Name != "" || opts.Abbreviation != "" ||
		truckerTermsChanged || customerTermsChanged ||
		opts.IsTransportOnly != "" || opts.DefaultReplyToEmailAddress != "" ||
		opts.IsActive != "" || opts.EnableImplicitTimeCardApproval != "" ||
		opts.RemitToAddress != "" || opts.SendLineupSummariesTo != "" ||
		opts.QuickbooksEnabled != "" || opts.IsNonDriverPermittedToCheckIn != "" ||
		opts.IsGeneratingAutomatedShiftFeedback != "" || opts.IsManagingQualityControlRequirements != "" ||
		opts.IsManagingDriverVisibility != "" || opts.SkipMaterialTransactionImageExtraction != "" ||
		opts.HelpText != "" || opts.MakeTruckerReportCardVisibleToTruckers != "" ||
		opts.PreferPublicDispatchPhoneNumber != "" || opts.PublicDispatchPhoneNumberExplicit != "" ||
		opts.SkipTenderJobScheduleShiftStartingSellerNotifications != "" || opts.IsAcceptingOpenDoorIssues != "" ||
		opts.CanCustomersSeeDriverContactInformation != "" || opts.CanCustomerOperationsSeeDriverContactInformation != "" ||
		opts.ShiftFeedbackReasonNotificationExclusions != "" || opts.DisabledFeedbackTypes != "" ||
		opts.QuickbooksEnabledCustomerIds != "" || opts.EnableEquipmentMovement != "" ||
		minDurationChanged || opts.RequiresCostCodeAllocations != "" ||
		opts.JobProductionPlanRecapTemplate != "" || opts.DefaultPredictionSubjectKind != "" ||
		leadTimeHoursChanged || opts.ModeledToProjectedConfidenceThreshold != "" ||
		opts.ModeledToActualConfidenceThreshold != "" || opts.SlackNtfyChannel != "" ||
		opts.SlackNtfyIcon != "" || opts.SlackHorizonChannel != "" ||
		opts.ActiveEquipmentRentalNotificationDays != "" || opts.DefaultCustomerRateAgreementTemplate != "" ||
		opts.DefaultFinancialContact != "" || opts.DefaultOperationsContact != "" ||
		opts.DefaultDispatchContact != "" || opts.Developer != ""

	if !hasUpdate {
		err := fmt.Errorf("at least one field to update is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Build attributes
	attributes := map[string]any{}
	if opts.Name != "" {
		attributes["company-name"] = opts.Name
	}
	if opts.Abbreviation != "" {
		attributes["abbreviation"] = opts.Abbreviation
	}
	if truckerTermsChanged {
		attributes["default-trucker-payment-terms"] = opts.DefaultTruckerPaymentTerms
	}
	if customerTermsChanged {
		attributes["default-customer-payment-terms"] = opts.DefaultCustomerPaymentTerms
	}
	if opts.IsTransportOnly != "" {
		attributes["is-transport-only"] = opts.IsTransportOnly == "true"
	}
	if opts.DefaultReplyToEmailAddress != "" {
		attributes["default-reply-to-email-address"] = opts.DefaultReplyToEmailAddress
	}
	if opts.IsActive != "" {
		attributes["is-active"] = opts.IsActive == "true"
	}
	if opts.EnableImplicitTimeCardApproval != "" {
		attributes["enable-implicit-time-card-approval"] = opts.EnableImplicitTimeCardApproval == "true"
	}
	if opts.RemitToAddress != "" {
		attributes["remit-to-address"] = opts.RemitToAddress
	}
	if opts.SendLineupSummariesTo != "" {
		attributes["send-lineup-summaries-to"] = opts.SendLineupSummariesTo
	}
	if opts.QuickbooksEnabled != "" {
		attributes["quickbooks-enabled"] = opts.QuickbooksEnabled == "true"
	}
	if opts.IsNonDriverPermittedToCheckIn != "" {
		attributes["is-non-driver-permitted-to-check-in"] = opts.IsNonDriverPermittedToCheckIn == "true"
	}
	if opts.IsGeneratingAutomatedShiftFeedback != "" {
		attributes["is-generating-automated-shift-feedback"] = opts.IsGeneratingAutomatedShiftFeedback == "true"
	}
	if opts.IsManagingQualityControlRequirements != "" {
		attributes["is-managing-quality-control-requirements"] = opts.IsManagingQualityControlRequirements == "true"
	}
	if opts.IsManagingDriverVisibility != "" {
		attributes["is-managing-driver-visibility"] = opts.IsManagingDriverVisibility == "true"
	}
	if opts.SkipMaterialTransactionImageExtraction != "" {
		attributes["skip-material-transaction-image-extraction"] = opts.SkipMaterialTransactionImageExtraction == "true"
	}
	if opts.HelpText != "" {
		attributes["help-text"] = opts.HelpText
	}
	if opts.MakeTruckerReportCardVisibleToTruckers != "" {
		attributes["make-trucker-report-card-visible-to-truckers"] = opts.MakeTruckerReportCardVisibleToTruckers == "true"
	}
	if opts.PreferPublicDispatchPhoneNumber != "" {
		attributes["prefer-public-dispatch-phone-number"] = opts.PreferPublicDispatchPhoneNumber == "true"
	}
	if opts.PublicDispatchPhoneNumberExplicit != "" {
		attributes["public-dispatch-phone-number-explicit"] = opts.PublicDispatchPhoneNumberExplicit
	}
	if opts.SkipTenderJobScheduleShiftStartingSellerNotifications != "" {
		attributes["skip-tender-job-schedule-shift-starting-seller-notifications"] = opts.SkipTenderJobScheduleShiftStartingSellerNotifications == "true"
	}
	if opts.IsAcceptingOpenDoorIssues != "" {
		attributes["is-accepting-open-door-issues"] = opts.IsAcceptingOpenDoorIssues == "true"
	}
	if opts.CanCustomersSeeDriverContactInformation != "" {
		attributes["can-customers-see-driver-contact-information"] = opts.CanCustomersSeeDriverContactInformation == "true"
	}
	if opts.CanCustomerOperationsSeeDriverContactInformation != "" {
		attributes["can-customer-operations-see-driver-contact-information"] = opts.CanCustomerOperationsSeeDriverContactInformation == "true"
	}
	if opts.ShiftFeedbackReasonNotificationExclusions != "" {
		attributes["shift-feedback-reason-notification-exclusions"] = opts.ShiftFeedbackReasonNotificationExclusions
	}
	if opts.DisabledFeedbackTypes != "" {
		attributes["disabled-feedback-types"] = opts.DisabledFeedbackTypes
	}
	if opts.QuickbooksEnabledCustomerIds != "" {
		attributes["quickbooks-enabled-customer-ids"] = opts.QuickbooksEnabledCustomerIds
	}
	if opts.EnableEquipmentMovement != "" {
		attributes["enable-equipment-movement"] = opts.EnableEquipmentMovement == "true"
	}
	if minDurationChanged {
		attributes["min-duration-of-auto-trucking-incident-with-down-time"] = opts.MinDurationOfAutoTruckingIncidentWithDownTime
	}
	if opts.RequiresCostCodeAllocations != "" {
		attributes["requires-cost-code-allocations"] = opts.RequiresCostCodeAllocations == "true"
	}
	if opts.JobProductionPlanRecapTemplate != "" {
		attributes["job-production-plan-recap-template"] = opts.JobProductionPlanRecapTemplate
	}
	if opts.DefaultPredictionSubjectKind != "" {
		attributes["default-prediction-subject-kind"] = opts.DefaultPredictionSubjectKind
	}
	if leadTimeHoursChanged {
		attributes["default-prediction-subject-lead-time-hours"] = opts.DefaultPredictionSubjectLeadTimeHours
	}
	if opts.ModeledToProjectedConfidenceThreshold != "" {
		attributes["modeled-to-projected-confidence-threshold"] = opts.ModeledToProjectedConfidenceThreshold
	}
	if opts.ModeledToActualConfidenceThreshold != "" {
		attributes["modeled-to-actual-confidence-threshold"] = opts.ModeledToActualConfidenceThreshold
	}
	if opts.SlackNtfyChannel != "" {
		attributes["slack-ntfy-channel"] = opts.SlackNtfyChannel
	}
	if opts.SlackNtfyIcon != "" {
		attributes["slack-ntfy-icon"] = opts.SlackNtfyIcon
	}
	if opts.SlackHorizonChannel != "" {
		attributes["slack-horizon-channel"] = opts.SlackHorizonChannel
	}
	if opts.ActiveEquipmentRentalNotificationDays != "" {
		var days []int
		if err := json.Unmarshal([]byte(opts.ActiveEquipmentRentalNotificationDays), &days); err != nil {
			return fmt.Errorf("invalid active-equipment-rental-notification-days JSON: %w", err)
		}
		attributes["active-equipment-rental-notification-days"] = days
	}

	// Build relationships
	relationships := map[string]any{}
	if opts.DefaultCustomerRateAgreementTemplate != "" {
		relationships["default-customer-rate-agreement-template"] = map[string]any{
			"data": map[string]string{
				"type": "rate-agreements",
				"id":   opts.DefaultCustomerRateAgreementTemplate,
			},
		}
	}
	if opts.DefaultFinancialContact != "" {
		relationships["default-financial-contact"] = map[string]any{
			"data": map[string]string{
				"type": "users",
				"id":   opts.DefaultFinancialContact,
			},
		}
	}
	if opts.DefaultOperationsContact != "" {
		relationships["default-operations-contact"] = map[string]any{
			"data": map[string]string{
				"type": "users",
				"id":   opts.DefaultOperationsContact,
			},
		}
	}
	if opts.DefaultDispatchContact != "" {
		relationships["default-dispatch-contact"] = map[string]any{
			"data": map[string]string{
				"type": "users",
				"id":   opts.DefaultDispatchContact,
			},
		}
	}
	if opts.Developer != "" {
		relationships["developer"] = map[string]any{
			"data": map[string]string{
				"type": "developers",
				"id":   opts.Developer,
			},
		}
	}

	data := map[string]any{
		"id":         id,
		"type":       "brokers",
		"attributes": attributes,
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/brokers/"+id, jsonBody)
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

	result := map[string]any{
		"id":   resp.Data.ID,
		"name": stringAttr(resp.Data.Attributes, "company-name"),
	}

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), result)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated broker %s (%s)\n", result["id"], result["name"])
	return nil
}

func parseDoBrokersUpdateOptions(cmd *cobra.Command) (doBrokersUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	abbreviation, _ := cmd.Flags().GetString("abbreviation")
	defaultTruckerPaymentTerms, _ := cmd.Flags().GetInt("default-trucker-payment-terms")
	defaultCustomerPaymentTerms, _ := cmd.Flags().GetInt("default-customer-payment-terms")
	isTransportOnly, _ := cmd.Flags().GetString("is-transport-only")
	defaultReplyToEmail, _ := cmd.Flags().GetString("default-reply-to-email")
	isActive, _ := cmd.Flags().GetString("is-active")
	enableImplicitTimeCardApproval, _ := cmd.Flags().GetString("enable-implicit-time-card-approval")
	remitToAddress, _ := cmd.Flags().GetString("remit-to-address")
	sendLineupSummariesTo, _ := cmd.Flags().GetString("send-lineup-summaries-to")
	quickbooksEnabled, _ := cmd.Flags().GetString("quickbooks-enabled")
	isNonDriverPermittedToCheckIn, _ := cmd.Flags().GetString("is-non-driver-permitted-to-check-in")
	isGeneratingAutomatedShiftFeedback, _ := cmd.Flags().GetString("is-generating-automated-shift-feedback")
	isManagingQualityControlRequirements, _ := cmd.Flags().GetString("is-managing-quality-control-requirements")
	isManagingDriverVisibility, _ := cmd.Flags().GetString("is-managing-driver-visibility")
	skipMaterialTransactionImageExtraction, _ := cmd.Flags().GetString("skip-material-transaction-image-extraction")
	helpText, _ := cmd.Flags().GetString("help-text")
	makeTruckerReportCardVisibleToTruckers, _ := cmd.Flags().GetString("make-trucker-report-card-visible-to-truckers")
	preferPublicDispatchPhoneNumber, _ := cmd.Flags().GetString("prefer-public-dispatch-phone-number")
	publicDispatchPhoneNumberExplicit, _ := cmd.Flags().GetString("public-dispatch-phone-number-explicit")
	skipTenderJobScheduleShiftStartingSellerNotifications, _ := cmd.Flags().GetString("skip-tender-job-schedule-shift-starting-seller-notifications")
	isAcceptingOpenDoorIssues, _ := cmd.Flags().GetString("is-accepting-open-door-issues")
	canCustomersSeeDriverContactInformation, _ := cmd.Flags().GetString("can-customers-see-driver-contact-information")
	canCustomerOperationsSeeDriverContactInformation, _ := cmd.Flags().GetString("can-customer-operations-see-driver-contact-information")
	shiftFeedbackReasonNotificationExclusions, _ := cmd.Flags().GetString("shift-feedback-reason-notification-exclusions")
	disabledFeedbackTypes, _ := cmd.Flags().GetString("disabled-feedback-types")
	quickbooksEnabledCustomerIds, _ := cmd.Flags().GetString("quickbooks-enabled-customer-ids")
	enableEquipmentMovement, _ := cmd.Flags().GetString("enable-equipment-movement")
	minDurationOfAutoTruckingIncidentWithDownTime, _ := cmd.Flags().GetInt("min-duration-of-auto-trucking-incident-with-down-time")
	requiresCostCodeAllocations, _ := cmd.Flags().GetString("requires-cost-code-allocations")
	jobProductionPlanRecapTemplate, _ := cmd.Flags().GetString("job-production-plan-recap-template")
	defaultPredictionSubjectKind, _ := cmd.Flags().GetString("default-prediction-subject-kind")
	defaultPredictionSubjectLeadTimeHours, _ := cmd.Flags().GetInt("default-prediction-subject-lead-time-hours")
	modeledToProjectedConfidenceThreshold, _ := cmd.Flags().GetString("modeled-to-projected-confidence-threshold")
	modeledToActualConfidenceThreshold, _ := cmd.Flags().GetString("modeled-to-actual-confidence-threshold")
	slackNtfyChannel, _ := cmd.Flags().GetString("slack-ntfy-channel")
	slackNtfyIcon, _ := cmd.Flags().GetString("slack-ntfy-icon")
	slackHorizonChannel, _ := cmd.Flags().GetString("slack-horizon-channel")
	activeEquipmentRentalNotificationDays, _ := cmd.Flags().GetString("active-equipment-rental-notification-days")
	defaultCustomerRateAgreementTemplate, _ := cmd.Flags().GetString("default-customer-rate-agreement-template")
	defaultFinancialContact, _ := cmd.Flags().GetString("default-financial-contact")
	defaultOperationsContact, _ := cmd.Flags().GetString("default-operations-contact")
	defaultDispatchContact, _ := cmd.Flags().GetString("default-dispatch-contact")
	developer, _ := cmd.Flags().GetString("developer")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doBrokersUpdateOptions{
		BaseURL:                                baseURL,
		Token:                                  token,
		JSON:                                   jsonOut,
		Name:                                   name,
		Abbreviation:                           abbreviation,
		DefaultTruckerPaymentTerms:             defaultTruckerPaymentTerms,
		DefaultCustomerPaymentTerms:            defaultCustomerPaymentTerms,
		IsTransportOnly:                        isTransportOnly,
		DefaultReplyToEmailAddress:             defaultReplyToEmail,
		IsActive:                               isActive,
		EnableImplicitTimeCardApproval:         enableImplicitTimeCardApproval,
		RemitToAddress:                         remitToAddress,
		SendLineupSummariesTo:                  sendLineupSummariesTo,
		QuickbooksEnabled:                      quickbooksEnabled,
		IsNonDriverPermittedToCheckIn:          isNonDriverPermittedToCheckIn,
		IsGeneratingAutomatedShiftFeedback:     isGeneratingAutomatedShiftFeedback,
		IsManagingQualityControlRequirements:   isManagingQualityControlRequirements,
		IsManagingDriverVisibility:             isManagingDriverVisibility,
		SkipMaterialTransactionImageExtraction: skipMaterialTransactionImageExtraction,
		HelpText:                               helpText,
		MakeTruckerReportCardVisibleToTruckers: makeTruckerReportCardVisibleToTruckers,
		PreferPublicDispatchPhoneNumber:        preferPublicDispatchPhoneNumber,
		PublicDispatchPhoneNumberExplicit:      publicDispatchPhoneNumberExplicit,
		SkipTenderJobScheduleShiftStartingSellerNotifications: skipTenderJobScheduleShiftStartingSellerNotifications,
		IsAcceptingOpenDoorIssues:                             isAcceptingOpenDoorIssues,
		CanCustomersSeeDriverContactInformation:               canCustomersSeeDriverContactInformation,
		CanCustomerOperationsSeeDriverContactInformation:      canCustomerOperationsSeeDriverContactInformation,
		ShiftFeedbackReasonNotificationExclusions:             shiftFeedbackReasonNotificationExclusions,
		DisabledFeedbackTypes:                                 disabledFeedbackTypes,
		QuickbooksEnabledCustomerIds:                          quickbooksEnabledCustomerIds,
		EnableEquipmentMovement:                               enableEquipmentMovement,
		MinDurationOfAutoTruckingIncidentWithDownTime:         minDurationOfAutoTruckingIncidentWithDownTime,
		RequiresCostCodeAllocations:                           requiresCostCodeAllocations,
		JobProductionPlanRecapTemplate:                        jobProductionPlanRecapTemplate,
		DefaultPredictionSubjectKind:                          defaultPredictionSubjectKind,
		DefaultPredictionSubjectLeadTimeHours:                 defaultPredictionSubjectLeadTimeHours,
		ModeledToProjectedConfidenceThreshold:                 modeledToProjectedConfidenceThreshold,
		ModeledToActualConfidenceThreshold:                    modeledToActualConfidenceThreshold,
		SlackNtfyChannel:                                      slackNtfyChannel,
		SlackNtfyIcon:                                         slackNtfyIcon,
		SlackHorizonChannel:                                   slackHorizonChannel,
		ActiveEquipmentRentalNotificationDays:                 activeEquipmentRentalNotificationDays,
		DefaultCustomerRateAgreementTemplate:                  defaultCustomerRateAgreementTemplate,
		DefaultFinancialContact:                               defaultFinancialContact,
		DefaultOperationsContact:                              defaultOperationsContact,
		DefaultDispatchContact:                                defaultDispatchContact,
		Developer:                                             developer,
	}, nil
}
