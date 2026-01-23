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

type doTruckersCreateOptions struct {
	BaseURL string
	Token   string
	JSON    bool

	// Required
	Name           string
	Broker         string
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
}

func newDoTruckersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new trucker",
		Long: `Create a new trucker.

Required flags:
  --name             The trucker company name (required)
  --broker           The broker ID (required)
  --company-address  The company address (required)

Optional flags:
  Address:
    --company-address-place-id          Google Place ID
    --company-address-plus-code         Plus code
    --skip-company-address-geocoding    Skip geocoding (true/false)

  Contact:
    --phone-number                      Phone number
    --fax-number                        Fax number

  Company info:
    --has-union-drivers                 Has union drivers (true/false)
    --estimated-trailer-capacity        Estimated trailer capacity
    --notes                             Notes
    --referral-source                   Referral source
    --color-hex                         Color hex code for UI

  Financial:
    --tax-identifier                    Tax identifier (EIN/SSN)
    --default-payment-terms             Default payment terms (integer)
    --billing-requirement               Billing requirement
    --generate-daily-invoice            Generate daily invoice (true/false)

  Payment address:
    --payment-address-line-one          Payment address line 1
    --payment-address-line-two          Payment address line 2
    --payment-address-city              Payment address city
    --payment-address-state-code        Payment address state code
    --payment-address-postal-code       Payment address postal code
    --payment-address-country-code      Payment address country code

  Notifications:
    --notify-default-financial-contact-of-time-card-pre-approvals Notify of pre-approvals (true/false)
    --notify-default-financial-contact-of-time-card-rejections    Notify of rejections (true/false)

  Time sheets:
    --is-expecting-trucker-shift-set-time-sheets    Expecting time sheets (true/false)
    --expecting-trucker-shift-set-time-sheets-on    Date expecting time sheets
    --time-sheet-submission-terms                   Time sheet submission terms
    --is-time-card-creating-time-sheet-line-item-explicit Create line items (true/false)

  Shifts:
    --manage-driver-assignment-acknowledgement      Manage driver acknowledgement
    --default-pre-trip-minutes                      Default pre-trip minutes (integer)
    --default-post-trip-minutes                     Default post-trip minutes (integer)
    --are-shifts-expecting-time-cards               Expecting time cards (true/false)

  Status:
    --is-active                         Active status (true/false)
    --is-controlled-by-broker           Controlled by broker (true/false)
    --favorite                          Favorite (true/false)
    --is-accepting-open-door-issues     Accepting open door issues (true/false)`,
		Example: `  # Create a trucker
  xbe do truckers create --name "ABC Trucking" --broker 123 --company-address "123 Main St"

  # Create with contact info
  xbe do truckers create --name "XYZ Transport" --broker 123 --company-address "456 Oak Ave" --phone-number "+15551234567"

  # Create with tax ID and payment terms
  xbe do truckers create --name "Local Haulers" --broker 123 --company-address "789 Elm Rd" --tax-identifier "12-3456789" --default-payment-terms 30

  # Get JSON output
  xbe do truckers create --name "New Trucker" --broker 123 --company-address "100 First St" --json`,
		Args: cobra.NoArgs,
		RunE: runDoTruckersCreate,
	}
	initDoTruckersCreateFlags(cmd)
	return cmd
}

func init() {
	doTruckersCmd.AddCommand(newDoTruckersCreateCmd())
}

func initDoTruckersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")

	// Required
	cmd.Flags().String("name", "", "Company name (required)")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("company-address", "", "Company address (required)")

	// Address
	cmd.Flags().String("company-address-place-id", "", "Google Place ID")
	cmd.Flags().String("company-address-plus-code", "", "Plus code")
	cmd.Flags().String("skip-company-address-geocoding", "", "Skip geocoding (true/false)")

	// Contact
	cmd.Flags().String("phone-number", "", "Phone number")
	cmd.Flags().String("fax-number", "", "Fax number")

	// Company info
	cmd.Flags().String("has-union-drivers", "", "Has union drivers (true/false)")
	cmd.Flags().String("estimated-trailer-capacity", "", "Estimated trailer capacity")
	cmd.Flags().String("notes", "", "Notes")
	cmd.Flags().String("referral-source", "", "Referral source")
	cmd.Flags().String("color-hex", "", "Color hex code")

	// Financial
	cmd.Flags().String("tax-identifier", "", "Tax identifier (EIN/SSN)")
	cmd.Flags().Int("default-payment-terms", 0, "Default payment terms")
	cmd.Flags().String("billing-requirement", "", "Billing requirement")
	cmd.Flags().String("generate-daily-invoice", "", "Generate daily invoice (true/false)")

	// Payment address
	cmd.Flags().String("payment-address-line-one", "", "Payment address line 1")
	cmd.Flags().String("payment-address-line-two", "", "Payment address line 2")
	cmd.Flags().String("payment-address-city", "", "Payment address city")
	cmd.Flags().String("payment-address-state-code", "", "Payment address state code")
	cmd.Flags().String("payment-address-postal-code", "", "Payment address postal code")
	cmd.Flags().String("payment-address-country-code", "", "Payment address country code")

	// Notifications
	cmd.Flags().String("notify-default-financial-contact-of-time-card-pre-approvals", "", "Notify of pre-approvals (true/false)")
	cmd.Flags().String("notify-default-financial-contact-of-time-card-rejections", "", "Notify of rejections (true/false)")

	// Time sheets
	cmd.Flags().String("is-expecting-trucker-shift-set-time-sheets", "", "Expecting time sheets (true/false)")
	cmd.Flags().String("expecting-trucker-shift-set-time-sheets-on", "", "Date expecting time sheets")
	cmd.Flags().String("time-sheet-submission-terms", "", "Time sheet submission terms")
	cmd.Flags().String("is-time-card-creating-time-sheet-line-item-explicit", "", "Create line items (true/false)")

	// Shifts
	cmd.Flags().String("manage-driver-assignment-acknowledgement", "", "Manage driver acknowledgement")
	cmd.Flags().Int("default-pre-trip-minutes", 0, "Default pre-trip minutes")
	cmd.Flags().Int("default-post-trip-minutes", 0, "Default post-trip minutes")
	cmd.Flags().String("are-shifts-expecting-time-cards", "", "Expecting time cards (true/false)")

	// Status
	cmd.Flags().String("is-active", "", "Active status (true/false)")
	cmd.Flags().String("is-controlled-by-broker", "", "Controlled by broker (true/false)")
	cmd.Flags().String("favorite", "", "Favorite (true/false)")
	cmd.Flags().String("is-accepting-open-door-issues", "", "Accepting open door issues (true/false)")

	// Validation skips
	cmd.Flags().String("skip-reasonable-default-operations-contact-validation", "", "Skip validation (true/false)")
	cmd.Flags().String("skip-reasonable-default-trailer-validation", "", "Skip validation (true/false)")

	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTruckersCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTruckersCreateOptions(cmd)
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

	// Require name
	if opts.Name == "" {
		err := fmt.Errorf("--name is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Require broker
	if opts.Broker == "" {
		err := fmt.Errorf("--broker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Require company-address
	if opts.CompanyAddress == "" {
		err := fmt.Errorf("--company-address is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Build attributes
	attributes := map[string]any{
		"company-name":    opts.Name,
		"company-address": opts.CompanyAddress,
	}

	// Address
	if opts.CompanyAddressPlaceID != "" {
		attributes["company-address-place-id"] = opts.CompanyAddressPlaceID
	}
	if opts.CompanyAddressPlusCode != "" {
		attributes["company-address-plus-code"] = opts.CompanyAddressPlusCode
	}
	if opts.SkipCompanyAddressGeocoding != "" {
		attributes["skip-company-address-geocoding"] = opts.SkipCompanyAddressGeocoding == "true"
	}

	// Contact
	if opts.PhoneNumber != "" {
		attributes["phone-number"] = opts.PhoneNumber
	}
	if opts.FaxNumber != "" {
		attributes["fax-number"] = opts.FaxNumber
	}

	// Company info
	if opts.HasUnionDrivers != "" {
		attributes["has-union-drivers"] = opts.HasUnionDrivers == "true"
	}
	if opts.EstimatedTrailerCapacity != "" {
		attributes["estimated-trailer-capacity"] = opts.EstimatedTrailerCapacity
	}
	if opts.Notes != "" {
		attributes["notes"] = opts.Notes
	}
	if opts.ReferralSource != "" {
		attributes["referral-source"] = opts.ReferralSource
	}
	if opts.ColorHex != "" {
		attributes["color-hex"] = opts.ColorHex
	}

	// Financial
	if opts.TaxIdentifier != "" {
		attributes["tax-identifier"] = opts.TaxIdentifier
	}
	if cmd.Flags().Changed("default-payment-terms") {
		attributes["default-payment-terms"] = opts.DefaultPaymentTerms
	}
	if opts.BillingRequirement != "" {
		attributes["billing-requirement"] = opts.BillingRequirement
	}
	if opts.GenerateDailyInvoice != "" {
		attributes["generate-daily-invoice"] = opts.GenerateDailyInvoice == "true"
	}

	// Payment address
	if opts.PaymentAddressLineOne != "" {
		attributes["payment-address-line-one"] = opts.PaymentAddressLineOne
	}
	if opts.PaymentAddressLineTwo != "" {
		attributes["payment-address-line-two"] = opts.PaymentAddressLineTwo
	}
	if opts.PaymentAddressCity != "" {
		attributes["payment-address-city"] = opts.PaymentAddressCity
	}
	if opts.PaymentAddressStateCode != "" {
		attributes["payment-address-state-code"] = opts.PaymentAddressStateCode
	}
	if opts.PaymentAddressPostalCode != "" {
		attributes["payment-address-postal-code"] = opts.PaymentAddressPostalCode
	}
	if opts.PaymentAddressCountryCode != "" {
		attributes["payment-address-country-code"] = opts.PaymentAddressCountryCode
	}

	// Notifications
	if opts.NotifyDefaultFinancialContactOfTimeCardPreApprovals != "" {
		attributes["notify-default-financial-contact-of-time-card-pre-approvals"] = opts.NotifyDefaultFinancialContactOfTimeCardPreApprovals == "true"
	}
	if opts.NotifyDefaultFinancialContactOfTimeCardRejections != "" {
		attributes["notify-default-financial-contact-of-time-card-rejections"] = opts.NotifyDefaultFinancialContactOfTimeCardRejections == "true"
	}

	// Time sheets
	if opts.IsExpectingTruckerShiftSetTimeSheets != "" {
		attributes["is-expecting-trucker-shift-set-time-sheets"] = opts.IsExpectingTruckerShiftSetTimeSheets == "true"
	}
	if opts.ExpectingTruckerShiftSetTimeSheetsOn != "" {
		attributes["expecting-trucker-shift-set-time-sheets-on"] = opts.ExpectingTruckerShiftSetTimeSheetsOn
	}
	if opts.TimeSheetSubmissionTerms != "" {
		attributes["time-sheet-submission-terms"] = opts.TimeSheetSubmissionTerms
	}
	if opts.IsTimeCardCreatingTimeSheetLineItemExplicit != "" {
		attributes["is-time-card-creating-time-sheet-line-item-explicit"] = opts.IsTimeCardCreatingTimeSheetLineItemExplicit == "true"
	}

	// Shifts
	if opts.ManageDriverAssignmentAcknowledgement != "" {
		attributes["manage-driver-assignment-acknowledgement"] = opts.ManageDriverAssignmentAcknowledgement
	}
	if cmd.Flags().Changed("default-pre-trip-minutes") {
		attributes["default-pre-trip-minutes"] = opts.DefaultPreTripMinutes
	}
	if cmd.Flags().Changed("default-post-trip-minutes") {
		attributes["default-post-trip-minutes"] = opts.DefaultPostTripMinutes
	}
	if opts.AreShiftsExpectingTimeCards != "" {
		attributes["are-shifts-expecting-time-cards"] = opts.AreShiftsExpectingTimeCards == "true"
	}

	// Status
	if opts.IsActive != "" {
		attributes["is-active"] = opts.IsActive == "true"
	}
	if opts.IsControlledByBroker != "" {
		attributes["is-controlled-by-broker"] = opts.IsControlledByBroker == "true"
	}
	if opts.Favorite != "" {
		attributes["favorite"] = opts.Favorite == "true"
	}
	if opts.IsAcceptingOpenDoorIssues != "" {
		attributes["is-accepting-open-door-issues"] = opts.IsAcceptingOpenDoorIssues == "true"
	}

	// Validation skips
	if opts.SkipReasonableDefaultOperationsContactValidation != "" {
		attributes["skip-reasonable-default-operations-contact-validation"] = opts.SkipReasonableDefaultOperationsContactValidation == "true"
	}
	if opts.SkipReasonableDefaultTrailerValidation != "" {
		attributes["skip-reasonable-default-trailer-validation"] = opts.SkipReasonableDefaultTrailerValidation == "true"
	}

	// Build relationships
	relationships := map[string]any{
		"broker": map[string]any{
			"data": map[string]string{
				"type": "brokers",
				"id":   opts.Broker,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "truckers",
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

	body, _, err := client.Post(cmd.Context(), "/v1/truckers", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created trucker %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoTruckersCreateOptions(cmd *cobra.Command) (doTruckersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")

	// Required
	name, _ := cmd.Flags().GetString("name")
	broker, _ := cmd.Flags().GetString("broker")
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

	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTruckersCreateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,

		Name:           name,
		Broker:         broker,
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
	}, nil
}
