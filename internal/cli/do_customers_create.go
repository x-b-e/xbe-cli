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

type doCustomersCreateOptions struct {
	BaseURL string
	Token   string
	JSON    bool

	// Required
	Name   string
	Broker string

	// Contact info
	PhoneNumber string
	FaxNumber   string

	// Address
	CompanyAddress                   string
	CompanyAddressLatitude           string
	CompanyAddressLongitude          string
	CompanyAddressPlaceID            string
	CompanyAddressPlusCode           string
	SkipCompanyAddressGeocoding      string
	IsCompanyAddressFormattedAddress string
	BillToAddress                    string

	// Company info
	CompanyURL                            string
	Notes                                 string
	RequiresUnionDrivers                  string
	IsTruckingCompany                     string
	EstimatedAnnualMaterialTransportSpend string

	// Billing settings
	DefaultPaymentTerms                   int
	GenerateDailyInvoice                  string
	GroupDailyInvoiceByJobSite            string
	AutomaticallyApproveDailyInvoice      string
	BillingPeriodDayCount                 int
	BillingPeriodEndInvoiceOffsetDayCount int
	BillingPeriodStartOn                  string
	SplitBillingPeriodsSpanningMonths     string
	DefaultTimeCardApprovalProcess        string

	// Credit settings
	CreditLimit         string
	CreditType          string
	CreditTypeChangedAt string

	// Operational settings
	IsActive                                  string
	IsControlledByBroker                      string
	IsDeveloper                               string
	Favorite                                  string
	RestrictTendersToCustomerTruckers         string
	RequiresJobProductionPlans                string
	SendLineupSummariesTo                     string
	DefaultAutomaticSubmissionDelayMinutes    int
	DefaultDelayAutomaticSubmissionAfterHours string
	CanManageCrewRequirements                 string
	DefaultIsManagingCrewRequirements         string
	DefaultIsExpectingSafetyMeeting           string
	JobProductionPlanRecapTemplate            string
	JobProductionPlanRecapSummaryTemplate     string
	IsTimeCardStartAtEvidenceRequired         string
	RequiresCostCodeAllocations               string
	EnableNonDefaultContractors               string
	HoldJobProductionPlanApproval             string
	ExcludeFromLineupScenarios                string

	// E-ticketing settings
	IsEticketingEnabled                    string
	IsEticketingRawEnabled                 string
	IsEticketingCycleTimeEnabled           string
	IsMaterialTransactionInspectionEnabled string

	// Crew requirements
	IsExpectingCrewRequirementTimeSheets string
	ExpectingCrewRequirementTimeSheetsOn string

	// Open door issues
	IsAcceptingOpenDoorIssues string

	// Relationships
	Developer                string
	DefaultOperationsContact string
	DefaultFinancialContact  string
	DefaultDispatchContact   string
	DefaultContractor        string
	RateAgreementTemplate    string
}

func newDoCustomersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new customer",
		Long: `Create a new customer.

Required flags:
  --name      The customer company name (required)
  --broker    The broker ID (required)

Optional flags:
  Contact:
    --phone-number                              Phone number
    --fax-number                                Fax number

  Address:
    --company-address                           Company address
    --company-address-latitude                  Latitude
    --company-address-longitude                 Longitude
    --company-address-place-id                  Google Place ID
    --company-address-plus-code                 Plus code
    --skip-company-address-geocoding            Skip geocoding (true/false)
    --is-company-address-formatted-address      Formatted address flag (true/false)
    --bill-to-address                           Bill-to address

  Company info:
    --company-url                               Company website URL
    --notes                                     Notes
    --requires-union-drivers                    Requires union drivers (true/false)
    --is-trucking-company                       Is trucking company (true/false)
    --estimated-annual-material-transport-spend Estimated annual spend

  Billing:
    --default-payment-terms                     Default payment terms (integer)
    --generate-daily-invoice                    Generate daily invoice (true/false)
    --group-daily-invoice-by-job-site           Group daily invoice by job site (true/false)
    --automatically-approve-daily-invoice       Auto-approve daily invoice (true/false)
    --billing-period-day-count                  Billing period day count (integer)
    --billing-period-end-invoice-offset-day-count Billing period end offset (integer)
    --billing-period-start-on                   Billing period start date
    --split-billing-periods-spanning-months     Split billing periods (true/false)
    --default-time-card-approval-process        Default time card approval process

  Credit:
    --credit-limit                              Credit limit
    --credit-type                               Credit type
    --credit-type-changed-at                    Credit type changed timestamp

  Operations:
    --is-active                                 Active status (true/false)
    --is-controlled-by-broker                   Controlled by broker (true/false)
    --is-developer                              Is developer (true/false)
    --favorite                                  Favorite (true/false)
    --restrict-tenders-to-customer-truckers     Restrict tenders (true/false)
    --requires-job-production-plans             Requires job production plans (true/false)
    --send-lineup-summaries-to                  Send lineup summaries to (email)
    --default-automatic-submission-delay-minutes Submission delay minutes (integer)
    --default-delay-automatic-submission-after-hours Delay after hours (true/false)
    --can-manage-crew-requirements              Can manage crew requirements (true/false)
    --default-is-managing-crew-requirements     Default managing crew requirements (true/false)
    --default-is-expecting-safety-meeting       Default expecting safety meeting (true/false)
    --job-production-plan-recap-template        Recap template
    --job-production-plan-recap-summary-template Recap summary template
    --is-time-card-start-at-evidence-required   Time card evidence required (true/false)
    --requires-cost-code-allocations            Requires cost code allocations (true/false)
    --enable-non-default-contractors            Enable non-default contractors (true/false)
    --hold-job-production-plan-approval         Hold job production plan approval (true/false)
    --exclude-from-lineup-scenarios             Exclude from lineup scenarios (true/false)

  E-ticketing:
    --is-eticketing-enabled                     E-ticketing enabled (true/false)
    --is-eticketing-raw-enabled                 E-ticketing raw enabled (true/false)
    --is-eticketing-cycle-time-enabled          E-ticketing cycle time enabled (true/false)
    --is-material-transaction-inspection-enabled Material inspection enabled (true/false)

  Crew requirements:
    --is-expecting-crew-requirement-time-sheets Expecting crew time sheets (true/false)
    --expecting-crew-requirement-time-sheets-on Date expecting time sheets

  Open door:
    --is-accepting-open-door-issues             Accepting open door issues (true/false)

  Relationships:
    --developer                                 Developer ID
    --default-operations-contact                Default operations contact user ID
    --default-financial-contact                 Default financial contact user ID
    --default-dispatch-contact                  Default dispatch contact user ID
    --default-contractor                        Default contractor ID
    --rate-agreement-template                   Rate agreement template ID (create only)`,
		Example: `  # Create a customer
  xbe do customers create --name "ABC Construction" --broker 123

  # Create with contact info
  xbe do customers create --name "XYZ Builders" --broker 123 --phone-number "+15551234567"

  # Create as developer
  xbe do customers create --name "Big Developer" --broker 123 --is-developer true

  # Create with billing settings
  xbe do customers create --name "Customer" --broker 123 --default-payment-terms 30 --generate-daily-invoice true

  # Get JSON output
  xbe do customers create --name "New Customer" --broker 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoCustomersCreate,
	}
	initDoCustomersCreateFlags(cmd)
	return cmd
}

func init() {
	doCustomersCmd.AddCommand(newDoCustomersCreateCmd())
}

func initDoCustomersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")

	// Required
	cmd.Flags().String("name", "", "Company name (required)")
	cmd.Flags().String("broker", "", "Broker ID (required)")

	// Contact info
	cmd.Flags().String("phone-number", "", "Phone number")
	cmd.Flags().String("fax-number", "", "Fax number")

	// Address
	cmd.Flags().String("company-address", "", "Company address")
	cmd.Flags().String("company-address-latitude", "", "Latitude")
	cmd.Flags().String("company-address-longitude", "", "Longitude")
	cmd.Flags().String("company-address-place-id", "", "Google Place ID")
	cmd.Flags().String("company-address-plus-code", "", "Plus code")
	cmd.Flags().String("skip-company-address-geocoding", "", "Skip geocoding (true/false)")
	cmd.Flags().String("is-company-address-formatted-address", "", "Formatted address flag (true/false)")
	cmd.Flags().String("bill-to-address", "", "Bill-to address")

	// Company info
	cmd.Flags().String("company-url", "", "Company website URL")
	cmd.Flags().String("notes", "", "Notes")
	cmd.Flags().String("requires-union-drivers", "", "Requires union drivers (true/false)")
	cmd.Flags().String("is-trucking-company", "", "Is trucking company (true/false)")
	cmd.Flags().String("estimated-annual-material-transport-spend", "", "Estimated annual spend")

	// Billing
	cmd.Flags().Int("default-payment-terms", 0, "Default payment terms")
	cmd.Flags().String("generate-daily-invoice", "", "Generate daily invoice (true/false)")
	cmd.Flags().String("group-daily-invoice-by-job-site", "", "Group daily invoice by job site (true/false)")
	cmd.Flags().String("automatically-approve-daily-invoice", "", "Auto-approve daily invoice (true/false)")
	cmd.Flags().Int("billing-period-day-count", 0, "Billing period day count")
	cmd.Flags().Int("billing-period-end-invoice-offset-day-count", 0, "Billing period end offset")
	cmd.Flags().String("billing-period-start-on", "", "Billing period start date")
	cmd.Flags().String("split-billing-periods-spanning-months", "", "Split billing periods (true/false)")
	cmd.Flags().String("default-time-card-approval-process", "", "Default time card approval process")

	// Credit
	cmd.Flags().String("credit-limit", "", "Credit limit")
	cmd.Flags().String("credit-type", "", "Credit type")
	cmd.Flags().String("credit-type-changed-at", "", "Credit type changed timestamp")

	// Operations
	cmd.Flags().String("is-active", "", "Active status (true/false)")
	cmd.Flags().String("is-controlled-by-broker", "", "Controlled by broker (true/false)")
	cmd.Flags().String("is-developer", "", "Is developer (true/false)")
	cmd.Flags().String("favorite", "", "Favorite (true/false)")
	cmd.Flags().String("restrict-tenders-to-customer-truckers", "", "Restrict tenders (true/false)")
	cmd.Flags().String("requires-job-production-plans", "", "Requires job production plans (true/false)")
	cmd.Flags().String("send-lineup-summaries-to", "", "Send lineup summaries to")
	cmd.Flags().Int("default-automatic-submission-delay-minutes", 0, "Submission delay minutes")
	cmd.Flags().String("default-delay-automatic-submission-after-hours", "", "Delay after hours (true/false)")
	cmd.Flags().String("can-manage-crew-requirements", "", "Can manage crew requirements (true/false)")
	cmd.Flags().String("default-is-managing-crew-requirements", "", "Default managing crew requirements (true/false)")
	cmd.Flags().String("default-is-expecting-safety-meeting", "", "Default expecting safety meeting (true/false)")
	cmd.Flags().String("job-production-plan-recap-template", "", "Recap template")
	cmd.Flags().String("job-production-plan-recap-summary-template", "", "Recap summary template")
	cmd.Flags().String("is-time-card-start-at-evidence-required", "", "Time card evidence required (true/false)")
	cmd.Flags().String("requires-cost-code-allocations", "", "Requires cost code allocations (true/false)")
	cmd.Flags().String("enable-non-default-contractors", "", "Enable non-default contractors (true/false)")
	cmd.Flags().String("hold-job-production-plan-approval", "", "Hold job production plan approval (true/false)")
	cmd.Flags().String("exclude-from-lineup-scenarios", "", "Exclude from lineup scenarios (true/false)")

	// E-ticketing
	cmd.Flags().String("is-eticketing-enabled", "", "E-ticketing enabled (true/false)")
	cmd.Flags().String("is-eticketing-raw-enabled", "", "E-ticketing raw enabled (true/false)")
	cmd.Flags().String("is-eticketing-cycle-time-enabled", "", "E-ticketing cycle time enabled (true/false)")
	cmd.Flags().String("is-material-transaction-inspection-enabled", "", "Material inspection enabled (true/false)")

	// Crew requirements
	cmd.Flags().String("is-expecting-crew-requirement-time-sheets", "", "Expecting crew time sheets (true/false)")
	cmd.Flags().String("expecting-crew-requirement-time-sheets-on", "", "Date expecting time sheets")

	// Open door
	cmd.Flags().String("is-accepting-open-door-issues", "", "Accepting open door issues (true/false)")

	// Relationships
	cmd.Flags().String("developer", "", "Developer ID")
	cmd.Flags().String("default-operations-contact", "", "Default operations contact user ID")
	cmd.Flags().String("default-financial-contact", "", "Default financial contact user ID")
	cmd.Flags().String("default-dispatch-contact", "", "Default dispatch contact user ID")
	cmd.Flags().String("default-contractor", "", "Default contractor ID")
	cmd.Flags().String("rate-agreement-template", "", "Rate agreement template ID")

	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCustomersCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoCustomersCreateOptions(cmd)
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

	// Build attributes
	attributes := map[string]any{
		"company-name": opts.Name,
	}

	// Contact info
	if opts.PhoneNumber != "" {
		attributes["phone-number"] = opts.PhoneNumber
	}
	if opts.FaxNumber != "" {
		attributes["fax-number"] = opts.FaxNumber
	}

	// Address
	if opts.CompanyAddress != "" {
		attributes["company-address"] = opts.CompanyAddress
	}
	if opts.CompanyAddressLatitude != "" {
		attributes["company-address-latitude"] = opts.CompanyAddressLatitude
	}
	if opts.CompanyAddressLongitude != "" {
		attributes["company-address-longitude"] = opts.CompanyAddressLongitude
	}
	if opts.CompanyAddressPlaceID != "" {
		attributes["company-address-place-id"] = opts.CompanyAddressPlaceID
	}
	if opts.CompanyAddressPlusCode != "" {
		attributes["company-address-plus-code"] = opts.CompanyAddressPlusCode
	}
	if opts.SkipCompanyAddressGeocoding != "" {
		attributes["skip-company-address-geocoding"] = opts.SkipCompanyAddressGeocoding == "true"
	}
	if opts.IsCompanyAddressFormattedAddress != "" {
		attributes["is-company-address-formatted-address"] = opts.IsCompanyAddressFormattedAddress == "true"
	}
	if opts.BillToAddress != "" {
		attributes["bill-to-address"] = opts.BillToAddress
	}

	// Company info
	if opts.CompanyURL != "" {
		attributes["company-url"] = opts.CompanyURL
	}
	if opts.Notes != "" {
		attributes["notes"] = opts.Notes
	}
	if opts.RequiresUnionDrivers != "" {
		attributes["requires-union-drivers"] = opts.RequiresUnionDrivers == "true"
	}
	if opts.IsTruckingCompany != "" {
		attributes["is-trucking-company"] = opts.IsTruckingCompany == "true"
	}
	if opts.EstimatedAnnualMaterialTransportSpend != "" {
		attributes["estimated-annual-material-transport-spend"] = opts.EstimatedAnnualMaterialTransportSpend
	}

	// Billing
	if cmd.Flags().Changed("default-payment-terms") {
		attributes["default-payment-terms"] = opts.DefaultPaymentTerms
	}
	if opts.GenerateDailyInvoice != "" {
		attributes["generate-daily-invoice"] = opts.GenerateDailyInvoice == "true"
	}
	if opts.GroupDailyInvoiceByJobSite != "" {
		attributes["group-daily-invoice-by-job-site"] = opts.GroupDailyInvoiceByJobSite == "true"
	}
	if opts.AutomaticallyApproveDailyInvoice != "" {
		attributes["automatically-approve-daily-invoice"] = opts.AutomaticallyApproveDailyInvoice == "true"
	}
	if cmd.Flags().Changed("billing-period-day-count") {
		attributes["billing-period-day-count"] = opts.BillingPeriodDayCount
	}
	if cmd.Flags().Changed("billing-period-end-invoice-offset-day-count") {
		attributes["billing-period-end-invoice-offset-day-count"] = opts.BillingPeriodEndInvoiceOffsetDayCount
	}
	if opts.BillingPeriodStartOn != "" {
		attributes["billing-period-start-on"] = opts.BillingPeriodStartOn
	}
	if opts.SplitBillingPeriodsSpanningMonths != "" {
		attributes["split-billing-periods-spanning-months"] = opts.SplitBillingPeriodsSpanningMonths == "true"
	}
	if opts.DefaultTimeCardApprovalProcess != "" {
		attributes["default-time-card-approval-process"] = opts.DefaultTimeCardApprovalProcess
	}

	// Credit
	if opts.CreditLimit != "" {
		attributes["credit-limit"] = opts.CreditLimit
	}
	if opts.CreditType != "" {
		attributes["credit-type"] = opts.CreditType
	}
	if opts.CreditTypeChangedAt != "" {
		attributes["credit-type-changed-at"] = opts.CreditTypeChangedAt
	}

	// Operations
	if opts.IsActive != "" {
		attributes["is-active"] = opts.IsActive == "true"
	}
	if opts.IsControlledByBroker != "" {
		attributes["is-controlled-by-broker"] = opts.IsControlledByBroker == "true"
	}
	if opts.IsDeveloper != "" {
		attributes["is-developer"] = opts.IsDeveloper == "true"
	}
	if opts.Favorite != "" {
		attributes["favorite"] = opts.Favorite == "true"
	}
	if opts.RestrictTendersToCustomerTruckers != "" {
		attributes["restrict-tenders-to-customer-truckers"] = opts.RestrictTendersToCustomerTruckers == "true"
	}
	if opts.RequiresJobProductionPlans != "" {
		attributes["requires-job-production-plans"] = opts.RequiresJobProductionPlans == "true"
	}
	if opts.SendLineupSummariesTo != "" {
		attributes["send-lineup-summaries-to"] = opts.SendLineupSummariesTo
	}
	if cmd.Flags().Changed("default-automatic-submission-delay-minutes") {
		attributes["default-automatic-submission-delay-minutes"] = opts.DefaultAutomaticSubmissionDelayMinutes
	}
	if opts.DefaultDelayAutomaticSubmissionAfterHours != "" {
		attributes["default-delay-automatic-submission-after-hours"] = opts.DefaultDelayAutomaticSubmissionAfterHours == "true"
	}
	if opts.CanManageCrewRequirements != "" {
		attributes["can-manage-crew-requirements"] = opts.CanManageCrewRequirements == "true"
	}
	if opts.DefaultIsManagingCrewRequirements != "" {
		attributes["default-is-managing-crew-requirements"] = opts.DefaultIsManagingCrewRequirements == "true"
	}
	if opts.DefaultIsExpectingSafetyMeeting != "" {
		attributes["default-is-expecting-safety-meeting"] = opts.DefaultIsExpectingSafetyMeeting == "true"
	}
	if opts.JobProductionPlanRecapTemplate != "" {
		attributes["job-production-plan-recap-template"] = opts.JobProductionPlanRecapTemplate
	}
	if opts.JobProductionPlanRecapSummaryTemplate != "" {
		attributes["job-production-plan-recap-summary-template"] = opts.JobProductionPlanRecapSummaryTemplate
	}
	if opts.IsTimeCardStartAtEvidenceRequired != "" {
		attributes["is-time-card-start-at-evidence-required"] = opts.IsTimeCardStartAtEvidenceRequired == "true"
	}
	if opts.RequiresCostCodeAllocations != "" {
		attributes["requires-cost-code-allocations"] = opts.RequiresCostCodeAllocations == "true"
	}
	if opts.EnableNonDefaultContractors != "" {
		attributes["enable-non-default-contractors"] = opts.EnableNonDefaultContractors == "true"
	}
	if opts.HoldJobProductionPlanApproval != "" {
		attributes["hold-job-production-plan-approval"] = opts.HoldJobProductionPlanApproval == "true"
	}
	if opts.ExcludeFromLineupScenarios != "" {
		attributes["exclude-from-lineup-scenarios"] = opts.ExcludeFromLineupScenarios == "true"
	}

	// E-ticketing
	if opts.IsEticketingEnabled != "" {
		attributes["is-eticketing-enabled"] = opts.IsEticketingEnabled == "true"
	}
	if opts.IsEticketingRawEnabled != "" {
		attributes["is-eticketing-raw-enabled"] = opts.IsEticketingRawEnabled == "true"
	}
	if opts.IsEticketingCycleTimeEnabled != "" {
		attributes["is-eticketing-cycle-time-enabled"] = opts.IsEticketingCycleTimeEnabled == "true"
	}
	if opts.IsMaterialTransactionInspectionEnabled != "" {
		attributes["is-material-transaction-inspection-enabled"] = opts.IsMaterialTransactionInspectionEnabled == "true"
	}

	// Crew requirements
	if opts.IsExpectingCrewRequirementTimeSheets != "" {
		attributes["is-expecting-crew-requirement-time-sheets"] = opts.IsExpectingCrewRequirementTimeSheets == "true"
	}
	if opts.ExpectingCrewRequirementTimeSheetsOn != "" {
		attributes["expecting-crew-requirement-time-sheets-on"] = opts.ExpectingCrewRequirementTimeSheetsOn
	}

	// Open door
	if opts.IsAcceptingOpenDoorIssues != "" {
		attributes["is-accepting-open-door-issues"] = opts.IsAcceptingOpenDoorIssues == "true"
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

	if opts.Developer != "" {
		relationships["developer"] = map[string]any{
			"data": map[string]string{
				"type": "developers",
				"id":   opts.Developer,
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
	if opts.DefaultFinancialContact != "" {
		relationships["default-financial-contact"] = map[string]any{
			"data": map[string]string{
				"type": "users",
				"id":   opts.DefaultFinancialContact,
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
	if opts.DefaultContractor != "" {
		relationships["default-contractor"] = map[string]any{
			"data": map[string]string{
				"type": "contractors",
				"id":   opts.DefaultContractor,
			},
		}
	}
	if opts.RateAgreementTemplate != "" {
		relationships["rate-agreement-template"] = map[string]any{
			"data": map[string]string{
				"type": "rate-agreements",
				"id":   opts.RateAgreementTemplate,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "customers",
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

	body, _, err := client.Post(cmd.Context(), "/v1/customers", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created customer %s (%s)\n", result["id"], result["name"])
	return nil
}

func parseDoCustomersCreateOptions(cmd *cobra.Command) (doCustomersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")

	// Required
	name, _ := cmd.Flags().GetString("name")
	broker, _ := cmd.Flags().GetString("broker")

	// Contact info
	phoneNumber, _ := cmd.Flags().GetString("phone-number")
	faxNumber, _ := cmd.Flags().GetString("fax-number")

	// Address
	companyAddress, _ := cmd.Flags().GetString("company-address")
	companyAddressLatitude, _ := cmd.Flags().GetString("company-address-latitude")
	companyAddressLongitude, _ := cmd.Flags().GetString("company-address-longitude")
	companyAddressPlaceID, _ := cmd.Flags().GetString("company-address-place-id")
	companyAddressPlusCode, _ := cmd.Flags().GetString("company-address-plus-code")
	skipCompanyAddressGeocoding, _ := cmd.Flags().GetString("skip-company-address-geocoding")
	isCompanyAddressFormattedAddress, _ := cmd.Flags().GetString("is-company-address-formatted-address")
	billToAddress, _ := cmd.Flags().GetString("bill-to-address")

	// Company info
	companyURL, _ := cmd.Flags().GetString("company-url")
	notes, _ := cmd.Flags().GetString("notes")
	requiresUnionDrivers, _ := cmd.Flags().GetString("requires-union-drivers")
	isTruckingCompany, _ := cmd.Flags().GetString("is-trucking-company")
	estimatedAnnualMaterialTransportSpend, _ := cmd.Flags().GetString("estimated-annual-material-transport-spend")

	// Billing
	defaultPaymentTerms, _ := cmd.Flags().GetInt("default-payment-terms")
	generateDailyInvoice, _ := cmd.Flags().GetString("generate-daily-invoice")
	groupDailyInvoiceByJobSite, _ := cmd.Flags().GetString("group-daily-invoice-by-job-site")
	automaticallyApproveDailyInvoice, _ := cmd.Flags().GetString("automatically-approve-daily-invoice")
	billingPeriodDayCount, _ := cmd.Flags().GetInt("billing-period-day-count")
	billingPeriodEndInvoiceOffsetDayCount, _ := cmd.Flags().GetInt("billing-period-end-invoice-offset-day-count")
	billingPeriodStartOn, _ := cmd.Flags().GetString("billing-period-start-on")
	splitBillingPeriodsSpanningMonths, _ := cmd.Flags().GetString("split-billing-periods-spanning-months")
	defaultTimeCardApprovalProcess, _ := cmd.Flags().GetString("default-time-card-approval-process")

	// Credit
	creditLimit, _ := cmd.Flags().GetString("credit-limit")
	creditType, _ := cmd.Flags().GetString("credit-type")
	creditTypeChangedAt, _ := cmd.Flags().GetString("credit-type-changed-at")

	// Operations
	isActive, _ := cmd.Flags().GetString("is-active")
	isControlledByBroker, _ := cmd.Flags().GetString("is-controlled-by-broker")
	isDeveloper, _ := cmd.Flags().GetString("is-developer")
	favorite, _ := cmd.Flags().GetString("favorite")
	restrictTendersToCustomerTruckers, _ := cmd.Flags().GetString("restrict-tenders-to-customer-truckers")
	requiresJobProductionPlans, _ := cmd.Flags().GetString("requires-job-production-plans")
	sendLineupSummariesTo, _ := cmd.Flags().GetString("send-lineup-summaries-to")
	defaultAutomaticSubmissionDelayMinutes, _ := cmd.Flags().GetInt("default-automatic-submission-delay-minutes")
	defaultDelayAutomaticSubmissionAfterHours, _ := cmd.Flags().GetString("default-delay-automatic-submission-after-hours")
	canManageCrewRequirements, _ := cmd.Flags().GetString("can-manage-crew-requirements")
	defaultIsManagingCrewRequirements, _ := cmd.Flags().GetString("default-is-managing-crew-requirements")
	defaultIsExpectingSafetyMeeting, _ := cmd.Flags().GetString("default-is-expecting-safety-meeting")
	jobProductionPlanRecapTemplate, _ := cmd.Flags().GetString("job-production-plan-recap-template")
	jobProductionPlanRecapSummaryTemplate, _ := cmd.Flags().GetString("job-production-plan-recap-summary-template")
	isTimeCardStartAtEvidenceRequired, _ := cmd.Flags().GetString("is-time-card-start-at-evidence-required")
	requiresCostCodeAllocations, _ := cmd.Flags().GetString("requires-cost-code-allocations")
	enableNonDefaultContractors, _ := cmd.Flags().GetString("enable-non-default-contractors")
	holdJobProductionPlanApproval, _ := cmd.Flags().GetString("hold-job-production-plan-approval")
	excludeFromLineupScenarios, _ := cmd.Flags().GetString("exclude-from-lineup-scenarios")

	// E-ticketing
	isEticketingEnabled, _ := cmd.Flags().GetString("is-eticketing-enabled")
	isEticketingRawEnabled, _ := cmd.Flags().GetString("is-eticketing-raw-enabled")
	isEticketingCycleTimeEnabled, _ := cmd.Flags().GetString("is-eticketing-cycle-time-enabled")
	isMaterialTransactionInspectionEnabled, _ := cmd.Flags().GetString("is-material-transaction-inspection-enabled")

	// Crew requirements
	isExpectingCrewRequirementTimeSheets, _ := cmd.Flags().GetString("is-expecting-crew-requirement-time-sheets")
	expectingCrewRequirementTimeSheetsOn, _ := cmd.Flags().GetString("expecting-crew-requirement-time-sheets-on")

	// Open door
	isAcceptingOpenDoorIssues, _ := cmd.Flags().GetString("is-accepting-open-door-issues")

	// Relationships
	developer, _ := cmd.Flags().GetString("developer")
	defaultOperationsContact, _ := cmd.Flags().GetString("default-operations-contact")
	defaultFinancialContact, _ := cmd.Flags().GetString("default-financial-contact")
	defaultDispatchContact, _ := cmd.Flags().GetString("default-dispatch-contact")
	defaultContractor, _ := cmd.Flags().GetString("default-contractor")
	rateAgreementTemplate, _ := cmd.Flags().GetString("rate-agreement-template")

	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCustomersCreateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,

		Name:   name,
		Broker: broker,

		PhoneNumber: phoneNumber,
		FaxNumber:   faxNumber,

		CompanyAddress:                   companyAddress,
		CompanyAddressLatitude:           companyAddressLatitude,
		CompanyAddressLongitude:          companyAddressLongitude,
		CompanyAddressPlaceID:            companyAddressPlaceID,
		CompanyAddressPlusCode:           companyAddressPlusCode,
		SkipCompanyAddressGeocoding:      skipCompanyAddressGeocoding,
		IsCompanyAddressFormattedAddress: isCompanyAddressFormattedAddress,
		BillToAddress:                    billToAddress,

		CompanyURL:                            companyURL,
		Notes:                                 notes,
		RequiresUnionDrivers:                  requiresUnionDrivers,
		IsTruckingCompany:                     isTruckingCompany,
		EstimatedAnnualMaterialTransportSpend: estimatedAnnualMaterialTransportSpend,

		DefaultPaymentTerms:                   defaultPaymentTerms,
		GenerateDailyInvoice:                  generateDailyInvoice,
		GroupDailyInvoiceByJobSite:            groupDailyInvoiceByJobSite,
		AutomaticallyApproveDailyInvoice:      automaticallyApproveDailyInvoice,
		BillingPeriodDayCount:                 billingPeriodDayCount,
		BillingPeriodEndInvoiceOffsetDayCount: billingPeriodEndInvoiceOffsetDayCount,
		BillingPeriodStartOn:                  billingPeriodStartOn,
		SplitBillingPeriodsSpanningMonths:     splitBillingPeriodsSpanningMonths,
		DefaultTimeCardApprovalProcess:        defaultTimeCardApprovalProcess,

		CreditLimit:         creditLimit,
		CreditType:          creditType,
		CreditTypeChangedAt: creditTypeChangedAt,

		IsActive:                                  isActive,
		IsControlledByBroker:                      isControlledByBroker,
		IsDeveloper:                               isDeveloper,
		Favorite:                                  favorite,
		RestrictTendersToCustomerTruckers:         restrictTendersToCustomerTruckers,
		RequiresJobProductionPlans:                requiresJobProductionPlans,
		SendLineupSummariesTo:                     sendLineupSummariesTo,
		DefaultAutomaticSubmissionDelayMinutes:    defaultAutomaticSubmissionDelayMinutes,
		DefaultDelayAutomaticSubmissionAfterHours: defaultDelayAutomaticSubmissionAfterHours,
		CanManageCrewRequirements:                 canManageCrewRequirements,
		DefaultIsManagingCrewRequirements:         defaultIsManagingCrewRequirements,
		DefaultIsExpectingSafetyMeeting:           defaultIsExpectingSafetyMeeting,
		JobProductionPlanRecapTemplate:            jobProductionPlanRecapTemplate,
		JobProductionPlanRecapSummaryTemplate:     jobProductionPlanRecapSummaryTemplate,
		IsTimeCardStartAtEvidenceRequired:         isTimeCardStartAtEvidenceRequired,
		RequiresCostCodeAllocations:               requiresCostCodeAllocations,
		EnableNonDefaultContractors:               enableNonDefaultContractors,
		HoldJobProductionPlanApproval:             holdJobProductionPlanApproval,
		ExcludeFromLineupScenarios:                excludeFromLineupScenarios,

		IsEticketingEnabled:                    isEticketingEnabled,
		IsEticketingRawEnabled:                 isEticketingRawEnabled,
		IsEticketingCycleTimeEnabled:           isEticketingCycleTimeEnabled,
		IsMaterialTransactionInspectionEnabled: isMaterialTransactionInspectionEnabled,

		IsExpectingCrewRequirementTimeSheets: isExpectingCrewRequirementTimeSheets,
		ExpectingCrewRequirementTimeSheetsOn: expectingCrewRequirementTimeSheetsOn,

		IsAcceptingOpenDoorIssues: isAcceptingOpenDoorIssues,

		Developer:                developer,
		DefaultOperationsContact: defaultOperationsContact,
		DefaultFinancialContact:  defaultFinancialContact,
		DefaultDispatchContact:   defaultDispatchContact,
		DefaultContractor:        defaultContractor,
		RateAgreementTemplate:    rateAgreementTemplate,
	}, nil
}
