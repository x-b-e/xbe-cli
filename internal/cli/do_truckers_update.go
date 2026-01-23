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

type doTruckersUpdateOptions struct {
	BaseURL string
	Token   string
	JSON    bool

	// Basic info
	Name           string
	CompanyAddress string

	// Address
	CompanyAddressPlaceID       string
	CompanyAddressPlusCode      string
	SkipCompanyAddressGeocoding string

	// Contact info
	PhoneNumber string
	FaxNumber   string

	// Company info
	HasUnionDrivers          string
	EstimatedTrailerCapacity string
	Notes                    string
	ReferralSource           string
	ColorHex                 string

	// Financial settings
	TaxIdentifier        string
	DefaultPaymentTerms  int
	BillingRequirement   string
	GenerateDailyInvoice string

	// Payment address
	RemitToAddress            string
	PaymentAddressLineOne     string
	PaymentAddressLineTwo     string
	PaymentAddressCity        string
	PaymentAddressStateCode   string
	PaymentAddressPostalCode  string
	PaymentAddressCountryCode string

	// Notification settings
	NotifyDefaultFinancialContactOfTimeCardPreApprovals string
	NotifyDefaultFinancialContactOfTimeCardRejections   string

	// Time sheet settings
	IsExpectingTruckerShiftSetTimeSheets        string
	ExpectingTruckerShiftSetTimeSheetsOn        string
	TimeSheetSubmissionTerms                    string
	IsTimeCardCreatingTimeSheetLineItemExplicit string

	// Shift settings
	ManageDriverAssignmentAcknowledgement string
	DefaultPreTripMinutes                 int
	DefaultPostTripMinutes                int
	AreShiftsExpectingTimeCards           string

	// Status
	IsActive                  string
	IsControlledByBroker      string
	Favorite                  string
	IsAcceptingOpenDoorIssues string

	// Validation skips
	SkipReasonableDefaultOperationsContactValidation string
	SkipReasonableDefaultTrailerValidation           string

	// Relationships (update only)
	DefaultOperationsContact string
	DefaultFinancialContact  string
	DefaultTrailer           string
}

func newDoTruckersUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a trucker",
		Long: `Update an existing trucker.

Only the fields you specify will be updated. Fields not provided will remain unchanged.

Arguments:
  <id>    The trucker ID (required)

Flags:
  Basic:
    --name                              Update company name
    --company-address                   Update company address

  Address:
    --company-address-place-id          Update Google Place ID
    --company-address-plus-code         Update plus code
    --skip-company-address-geocoding    Update skip geocoding (true/false)

  Contact:
    --phone-number                      Update phone number
    --fax-number                        Update fax number

  Company info:
    --has-union-drivers                 Update has union drivers (true/false)
    --estimated-trailer-capacity        Update estimated trailer capacity
    --notes                             Update notes
    --referral-source                   Update referral source
    --color-hex                         Update color hex code for UI

  Financial:
    --tax-identifier                    Update tax identifier (EIN/SSN)
    --default-payment-terms             Update default payment terms (integer)
    --billing-requirement               Update billing requirement
    --generate-daily-invoice            Update generate daily invoice (true/false)

  Payment address:
    --remit-to-address                  Update remit-to address
    --payment-address-line-one          Update payment address line 1
    --payment-address-line-two          Update payment address line 2
    --payment-address-city              Update payment address city
    --payment-address-state-code        Update payment address state code
    --payment-address-postal-code       Update payment address postal code
    --payment-address-country-code      Update payment address country code

  Notifications:
    --notify-default-financial-contact-of-time-card-pre-approvals Update notify of pre-approvals (true/false)
    --notify-default-financial-contact-of-time-card-rejections    Update notify of rejections (true/false)

  Time sheets:
    --is-expecting-trucker-shift-set-time-sheets    Update expecting time sheets (true/false)
    --expecting-trucker-shift-set-time-sheets-on    Update date expecting time sheets
    --time-sheet-submission-terms                   Update time sheet submission terms
    --is-time-card-creating-time-sheet-line-item-explicit Update create line items (true/false)

  Shifts:
    --manage-driver-assignment-acknowledgement      Update manage driver acknowledgement
    --default-pre-trip-minutes                      Update default pre-trip minutes (integer)
    --default-post-trip-minutes                     Update default post-trip minutes (integer)
    --are-shifts-expecting-time-cards               Update expecting time cards (true/false)

  Status:
    --is-active                         Update active status (true/false)
    --is-controlled-by-broker           Update controlled by broker (true/false)
    --favorite                          Update favorite (true/false)
    --is-accepting-open-door-issues     Update accepting open door issues (true/false)

  Relationships:
    --default-operations-contact        Update default operations contact user ID
    --default-financial-contact         Update default financial contact user ID
    --default-trailer                   Update default trailer ID`,
		Example: `  # Update the name
  xbe do truckers update 123 --name "New Company Name"

  # Update contact info
  xbe do truckers update 123 --phone-number "+15559876543"

  # Deactivate a trucker
  xbe do truckers update 123 --is-active false

  # Update financial settings
  xbe do truckers update 123 --default-payment-terms 45 --tax-identifier "98-7654321"

  # Update default contacts
  xbe do truckers update 123 --default-operations-contact 456 --default-financial-contact 789

  # Get JSON output
  xbe do truckers update 123 --name "Updated" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTruckersUpdate,
	}
	initDoTruckersUpdateFlags(cmd)
	return cmd
}

func init() {
	doTruckersCmd.AddCommand(newDoTruckersUpdateCmd())
}

func initDoTruckersUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")

	// Basic info
	cmd.Flags().String("name", "", "New company name")
	cmd.Flags().String("company-address", "", "New company address")

	// Address
	cmd.Flags().String("company-address-place-id", "", "New Google Place ID")
	cmd.Flags().String("company-address-plus-code", "", "New plus code")
	cmd.Flags().String("skip-company-address-geocoding", "", "Skip geocoding (true/false)")

	// Contact
	cmd.Flags().String("phone-number", "", "New phone number")
	cmd.Flags().String("fax-number", "", "New fax number")

	// Company info
	cmd.Flags().String("has-union-drivers", "", "Has union drivers (true/false)")
	cmd.Flags().String("estimated-trailer-capacity", "", "New estimated trailer capacity")
	cmd.Flags().String("notes", "", "New notes")
	cmd.Flags().String("referral-source", "", "New referral source")
	cmd.Flags().String("color-hex", "", "New color hex code")

	// Financial
	cmd.Flags().String("tax-identifier", "", "New tax identifier")
	cmd.Flags().Int("default-payment-terms", 0, "New default payment terms")
	cmd.Flags().String("billing-requirement", "", "New billing requirement")
	cmd.Flags().String("generate-daily-invoice", "", "Generate daily invoice (true/false)")

	// Payment address
	cmd.Flags().String("remit-to-address", "", "New remit-to address")
	cmd.Flags().String("payment-address-line-one", "", "New payment address line 1")
	cmd.Flags().String("payment-address-line-two", "", "New payment address line 2")
	cmd.Flags().String("payment-address-city", "", "New payment address city")
	cmd.Flags().String("payment-address-state-code", "", "New payment address state code")
	cmd.Flags().String("payment-address-postal-code", "", "New payment address postal code")
	cmd.Flags().String("payment-address-country-code", "", "New payment address country code")

	// Notifications
	cmd.Flags().String("notify-default-financial-contact-of-time-card-pre-approvals", "", "Notify of pre-approvals (true/false)")
	cmd.Flags().String("notify-default-financial-contact-of-time-card-rejections", "", "Notify of rejections (true/false)")

	// Time sheets
	cmd.Flags().String("is-expecting-trucker-shift-set-time-sheets", "", "Expecting time sheets (true/false)")
	cmd.Flags().String("expecting-trucker-shift-set-time-sheets-on", "", "New date expecting time sheets")
	cmd.Flags().String("time-sheet-submission-terms", "", "New time sheet submission terms")
	cmd.Flags().String("is-time-card-creating-time-sheet-line-item-explicit", "", "Create line items (true/false)")

	// Shifts
	cmd.Flags().String("manage-driver-assignment-acknowledgement", "", "New manage driver acknowledgement")
	cmd.Flags().Int("default-pre-trip-minutes", 0, "New default pre-trip minutes")
	cmd.Flags().Int("default-post-trip-minutes", 0, "New default post-trip minutes")
	cmd.Flags().String("are-shifts-expecting-time-cards", "", "Expecting time cards (true/false)")

	// Status
	cmd.Flags().String("is-active", "", "Active status (true/false)")
	cmd.Flags().String("is-controlled-by-broker", "", "Controlled by broker (true/false)")
	cmd.Flags().String("favorite", "", "Favorite (true/false)")
	cmd.Flags().String("is-accepting-open-door-issues", "", "Accepting open door issues (true/false)")

	// Validation skips
	cmd.Flags().String("skip-reasonable-default-operations-contact-validation", "", "Skip validation (true/false)")
	cmd.Flags().String("skip-reasonable-default-trailer-validation", "", "Skip validation (true/false)")

	// Relationships
	cmd.Flags().String("default-operations-contact", "", "Default operations contact user ID")
	cmd.Flags().String("default-financial-contact", "", "Default financial contact user ID")
	cmd.Flags().String("default-trailer", "", "Default trailer ID")

	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTruckersUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTruckersUpdateOptions(cmd)
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
		return fmt.Errorf("trucker id is required")
	}

	// Build attributes - check if any field is provided
	attributes := map[string]any{}
	hasChanges := false

	// Basic info
	if opts.Name != "" {
		attributes["company-name"] = opts.Name
		hasChanges = true
	}
	if opts.CompanyAddress != "" {
		attributes["company-address"] = opts.CompanyAddress
		hasChanges = true
	}

	// Address
	if opts.CompanyAddressPlaceID != "" {
		attributes["company-address-place-id"] = opts.CompanyAddressPlaceID
		hasChanges = true
	}
	if opts.CompanyAddressPlusCode != "" {
		attributes["company-address-plus-code"] = opts.CompanyAddressPlusCode
		hasChanges = true
	}
	if opts.SkipCompanyAddressGeocoding != "" {
		attributes["skip-company-address-geocoding"] = opts.SkipCompanyAddressGeocoding == "true"
		hasChanges = true
	}

	// Contact
	if opts.PhoneNumber != "" {
		attributes["phone-number"] = opts.PhoneNumber
		hasChanges = true
	}
	if opts.FaxNumber != "" {
		attributes["fax-number"] = opts.FaxNumber
		hasChanges = true
	}

	// Company info
	if opts.HasUnionDrivers != "" {
		attributes["has-union-drivers"] = opts.HasUnionDrivers == "true"
		hasChanges = true
	}
	if opts.EstimatedTrailerCapacity != "" {
		attributes["estimated-trailer-capacity"] = opts.EstimatedTrailerCapacity
		hasChanges = true
	}
	if opts.Notes != "" {
		attributes["notes"] = opts.Notes
		hasChanges = true
	}
	if opts.ReferralSource != "" {
		attributes["referral-source"] = opts.ReferralSource
		hasChanges = true
	}
	if opts.ColorHex != "" {
		attributes["color-hex"] = opts.ColorHex
		hasChanges = true
	}

	// Financial
	if opts.TaxIdentifier != "" {
		attributes["tax-identifier"] = opts.TaxIdentifier
		hasChanges = true
	}
	if cmd.Flags().Changed("default-payment-terms") {
		attributes["default-payment-terms"] = opts.DefaultPaymentTerms
		hasChanges = true
	}
	if opts.BillingRequirement != "" {
		attributes["billing-requirement"] = opts.BillingRequirement
		hasChanges = true
	}
	if opts.GenerateDailyInvoice != "" {
		attributes["generate-daily-invoice"] = opts.GenerateDailyInvoice == "true"
		hasChanges = true
	}

	// Payment address
	if opts.RemitToAddress != "" {
		attributes["remit-to-address"] = opts.RemitToAddress
		hasChanges = true
	}
	if opts.PaymentAddressLineOne != "" {
		attributes["payment-address-line-one"] = opts.PaymentAddressLineOne
		hasChanges = true
	}
	if opts.PaymentAddressLineTwo != "" {
		attributes["payment-address-line-two"] = opts.PaymentAddressLineTwo
		hasChanges = true
	}
	if opts.PaymentAddressCity != "" {
		attributes["payment-address-city"] = opts.PaymentAddressCity
		hasChanges = true
	}
	if opts.PaymentAddressStateCode != "" {
		attributes["payment-address-state-code"] = opts.PaymentAddressStateCode
		hasChanges = true
	}
	if opts.PaymentAddressPostalCode != "" {
		attributes["payment-address-postal-code"] = opts.PaymentAddressPostalCode
		hasChanges = true
	}
	if opts.PaymentAddressCountryCode != "" {
		attributes["payment-address-country-code"] = opts.PaymentAddressCountryCode
		hasChanges = true
	}

	// Notifications
	if opts.NotifyDefaultFinancialContactOfTimeCardPreApprovals != "" {
		attributes["notify-default-financial-contact-of-time-card-pre-approvals"] = opts.NotifyDefaultFinancialContactOfTimeCardPreApprovals == "true"
		hasChanges = true
	}
	if opts.NotifyDefaultFinancialContactOfTimeCardRejections != "" {
		attributes["notify-default-financial-contact-of-time-card-rejections"] = opts.NotifyDefaultFinancialContactOfTimeCardRejections == "true"
		hasChanges = true
	}

	// Time sheets
	if opts.IsExpectingTruckerShiftSetTimeSheets != "" {
		attributes["is-expecting-trucker-shift-set-time-sheets"] = opts.IsExpectingTruckerShiftSetTimeSheets == "true"
		hasChanges = true
	}
	if opts.ExpectingTruckerShiftSetTimeSheetsOn != "" {
		attributes["expecting-trucker-shift-set-time-sheets-on"] = opts.ExpectingTruckerShiftSetTimeSheetsOn
		hasChanges = true
	}
	if opts.TimeSheetSubmissionTerms != "" {
		attributes["time-sheet-submission-terms"] = opts.TimeSheetSubmissionTerms
		hasChanges = true
	}
	if opts.IsTimeCardCreatingTimeSheetLineItemExplicit != "" {
		attributes["is-time-card-creating-time-sheet-line-item-explicit"] = opts.IsTimeCardCreatingTimeSheetLineItemExplicit == "true"
		hasChanges = true
	}

	// Shifts
	if opts.ManageDriverAssignmentAcknowledgement != "" {
		attributes["manage-driver-assignment-acknowledgement"] = opts.ManageDriverAssignmentAcknowledgement
		hasChanges = true
	}
	if cmd.Flags().Changed("default-pre-trip-minutes") {
		attributes["default-pre-trip-minutes"] = opts.DefaultPreTripMinutes
		hasChanges = true
	}
	if cmd.Flags().Changed("default-post-trip-minutes") {
		attributes["default-post-trip-minutes"] = opts.DefaultPostTripMinutes
		hasChanges = true
	}
	if opts.AreShiftsExpectingTimeCards != "" {
		attributes["are-shifts-expecting-time-cards"] = opts.AreShiftsExpectingTimeCards == "true"
		hasChanges = true
	}

	// Status
	if opts.IsActive != "" {
		attributes["is-active"] = opts.IsActive == "true"
		hasChanges = true
	}
	if opts.IsControlledByBroker != "" {
		attributes["is-controlled-by-broker"] = opts.IsControlledByBroker == "true"
		hasChanges = true
	}
	if opts.Favorite != "" {
		attributes["favorite"] = opts.Favorite == "true"
		hasChanges = true
	}
	if opts.IsAcceptingOpenDoorIssues != "" {
		attributes["is-accepting-open-door-issues"] = opts.IsAcceptingOpenDoorIssues == "true"
		hasChanges = true
	}

	// Validation skips
	if opts.SkipReasonableDefaultOperationsContactValidation != "" {
		attributes["skip-reasonable-default-operations-contact-validation"] = opts.SkipReasonableDefaultOperationsContactValidation == "true"
		hasChanges = true
	}
	if opts.SkipReasonableDefaultTrailerValidation != "" {
		attributes["skip-reasonable-default-trailer-validation"] = opts.SkipReasonableDefaultTrailerValidation == "true"
		hasChanges = true
	}

	// Build relationships
	relationships := map[string]any{}

	if opts.DefaultOperationsContact != "" {
		relationships["default-operations-contact"] = map[string]any{
			"data": map[string]string{
				"type": "users",
				"id":   opts.DefaultOperationsContact,
			},
		}
		hasChanges = true
	}
	if opts.DefaultFinancialContact != "" {
		relationships["default-financial-contact"] = map[string]any{
			"data": map[string]string{
				"type": "users",
				"id":   opts.DefaultFinancialContact,
			},
		}
		hasChanges = true
	}
	if opts.DefaultTrailer != "" {
		relationships["default-trailer"] = map[string]any{
			"data": map[string]string{
				"type": "trailers",
				"id":   opts.DefaultTrailer,
			},
		}
		hasChanges = true
	}

	if !hasChanges {
		err := fmt.Errorf("at least one field to update is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"id":         id,
			"type":       "truckers",
			"attributes": attributes,
		},
	}

	if len(relationships) > 0 {
		requestBody["data"].(map[string]any)["relationships"] = relationships
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/truckers/"+id, jsonBody)
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

	row := truckerRow{
		ID:   resp.Data.ID,
		Name: stringAttr(resp.Data.Attributes, "company-name"),
	}

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated trucker %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoTruckersUpdateOptions(cmd *cobra.Command) (doTruckersUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")

	// Basic info
	name, _ := cmd.Flags().GetString("name")
	companyAddress, _ := cmd.Flags().GetString("company-address")

	// Address
	companyAddressPlaceID, _ := cmd.Flags().GetString("company-address-place-id")
	companyAddressPlusCode, _ := cmd.Flags().GetString("company-address-plus-code")
	skipCompanyAddressGeocoding, _ := cmd.Flags().GetString("skip-company-address-geocoding")

	// Contact
	phoneNumber, _ := cmd.Flags().GetString("phone-number")
	faxNumber, _ := cmd.Flags().GetString("fax-number")

	// Company info
	hasUnionDrivers, _ := cmd.Flags().GetString("has-union-drivers")
	estimatedTrailerCapacity, _ := cmd.Flags().GetString("estimated-trailer-capacity")
	notes, _ := cmd.Flags().GetString("notes")
	referralSource, _ := cmd.Flags().GetString("referral-source")
	colorHex, _ := cmd.Flags().GetString("color-hex")

	// Financial
	taxIdentifier, _ := cmd.Flags().GetString("tax-identifier")
	defaultPaymentTerms, _ := cmd.Flags().GetInt("default-payment-terms")
	billingRequirement, _ := cmd.Flags().GetString("billing-requirement")
	generateDailyInvoice, _ := cmd.Flags().GetString("generate-daily-invoice")

	// Payment address
	remitToAddress, _ := cmd.Flags().GetString("remit-to-address")
	paymentAddressLineOne, _ := cmd.Flags().GetString("payment-address-line-one")
	paymentAddressLineTwo, _ := cmd.Flags().GetString("payment-address-line-two")
	paymentAddressCity, _ := cmd.Flags().GetString("payment-address-city")
	paymentAddressStateCode, _ := cmd.Flags().GetString("payment-address-state-code")
	paymentAddressPostalCode, _ := cmd.Flags().GetString("payment-address-postal-code")
	paymentAddressCountryCode, _ := cmd.Flags().GetString("payment-address-country-code")

	// Notifications
	notifyDefaultFinancialContactOfTimeCardPreApprovals, _ := cmd.Flags().GetString("notify-default-financial-contact-of-time-card-pre-approvals")
	notifyDefaultFinancialContactOfTimeCardRejections, _ := cmd.Flags().GetString("notify-default-financial-contact-of-time-card-rejections")

	// Time sheets
	isExpectingTruckerShiftSetTimeSheets, _ := cmd.Flags().GetString("is-expecting-trucker-shift-set-time-sheets")
	expectingTruckerShiftSetTimeSheetsOn, _ := cmd.Flags().GetString("expecting-trucker-shift-set-time-sheets-on")
	timeSheetSubmissionTerms, _ := cmd.Flags().GetString("time-sheet-submission-terms")
	isTimeCardCreatingTimeSheetLineItemExplicit, _ := cmd.Flags().GetString("is-time-card-creating-time-sheet-line-item-explicit")

	// Shifts
	manageDriverAssignmentAcknowledgement, _ := cmd.Flags().GetString("manage-driver-assignment-acknowledgement")
	defaultPreTripMinutes, _ := cmd.Flags().GetInt("default-pre-trip-minutes")
	defaultPostTripMinutes, _ := cmd.Flags().GetInt("default-post-trip-minutes")
	areShiftsExpectingTimeCards, _ := cmd.Flags().GetString("are-shifts-expecting-time-cards")

	// Status
	isActive, _ := cmd.Flags().GetString("is-active")
	isControlledByBroker, _ := cmd.Flags().GetString("is-controlled-by-broker")
	favorite, _ := cmd.Flags().GetString("favorite")
	isAcceptingOpenDoorIssues, _ := cmd.Flags().GetString("is-accepting-open-door-issues")

	// Validation skips
	skipReasonableDefaultOperationsContactValidation, _ := cmd.Flags().GetString("skip-reasonable-default-operations-contact-validation")
	skipReasonableDefaultTrailerValidation, _ := cmd.Flags().GetString("skip-reasonable-default-trailer-validation")

	// Relationships
	defaultOperationsContact, _ := cmd.Flags().GetString("default-operations-contact")
	defaultFinancialContact, _ := cmd.Flags().GetString("default-financial-contact")
	defaultTrailer, _ := cmd.Flags().GetString("default-trailer")

	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTruckersUpdateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,

		Name:           name,
		CompanyAddress: companyAddress,

		CompanyAddressPlaceID:       companyAddressPlaceID,
		CompanyAddressPlusCode:      companyAddressPlusCode,
		SkipCompanyAddressGeocoding: skipCompanyAddressGeocoding,

		PhoneNumber: phoneNumber,
		FaxNumber:   faxNumber,

		HasUnionDrivers:          hasUnionDrivers,
		EstimatedTrailerCapacity: estimatedTrailerCapacity,
		Notes:                    notes,
		ReferralSource:           referralSource,
		ColorHex:                 colorHex,

		TaxIdentifier:        taxIdentifier,
		DefaultPaymentTerms:  defaultPaymentTerms,
		BillingRequirement:   billingRequirement,
		GenerateDailyInvoice: generateDailyInvoice,

		RemitToAddress:            remitToAddress,
		PaymentAddressLineOne:     paymentAddressLineOne,
		PaymentAddressLineTwo:     paymentAddressLineTwo,
		PaymentAddressCity:        paymentAddressCity,
		PaymentAddressStateCode:   paymentAddressStateCode,
		PaymentAddressPostalCode:  paymentAddressPostalCode,
		PaymentAddressCountryCode: paymentAddressCountryCode,

		NotifyDefaultFinancialContactOfTimeCardPreApprovals: notifyDefaultFinancialContactOfTimeCardPreApprovals,
		NotifyDefaultFinancialContactOfTimeCardRejections:   notifyDefaultFinancialContactOfTimeCardRejections,

		IsExpectingTruckerShiftSetTimeSheets:        isExpectingTruckerShiftSetTimeSheets,
		ExpectingTruckerShiftSetTimeSheetsOn:        expectingTruckerShiftSetTimeSheetsOn,
		TimeSheetSubmissionTerms:                    timeSheetSubmissionTerms,
		IsTimeCardCreatingTimeSheetLineItemExplicit: isTimeCardCreatingTimeSheetLineItemExplicit,

		ManageDriverAssignmentAcknowledgement: manageDriverAssignmentAcknowledgement,
		DefaultPreTripMinutes:                 defaultPreTripMinutes,
		DefaultPostTripMinutes:                defaultPostTripMinutes,
		AreShiftsExpectingTimeCards:           areShiftsExpectingTimeCards,

		IsActive:                  isActive,
		IsControlledByBroker:      isControlledByBroker,
		Favorite:                  favorite,
		IsAcceptingOpenDoorIssues: isAcceptingOpenDoorIssues,

		SkipReasonableDefaultOperationsContactValidation: skipReasonableDefaultOperationsContactValidation,
		SkipReasonableDefaultTrailerValidation:           skipReasonableDefaultTrailerValidation,

		DefaultOperationsContact: defaultOperationsContact,
		DefaultFinancialContact:  defaultFinancialContact,
		DefaultTrailer:           defaultTrailer,
	}, nil
}
