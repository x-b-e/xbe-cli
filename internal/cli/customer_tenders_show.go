package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type customerTendersShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type customerTenderDetails struct {
	ID                                 string `json:"id"`
	Status                             string `json:"status,omitempty"`
	ExpiresAt                          string `json:"expires_at,omitempty"`
	Note                               string `json:"note,omitempty"`
	PaymentTerms                       string `json:"payment_terms,omitempty"`
	PaymentTermsAndConditions          string `json:"payment_terms_and_conditions,omitempty"`
	RestrictToCustomerTruckers         bool   `json:"restrict_to_customer_truckers"`
	MaximumTravelMinutes               string `json:"maximum_travel_minutes,omitempty"`
	BillableTravelMinutesPerTravelMile string `json:"billable_travel_minutes_per_travel_mile,omitempty"`
	DisplaysTrips                      bool   `json:"displays_trips"`
	IsTruckerShiftRejectionPermitted   bool   `json:"is_trucker_shift_rejection_permitted"`
	HasBuyerComplianceErrors           bool   `json:"has_buyer_compliance_errors"`
	HasSellerComplianceErrors          bool   `json:"has_seller_compliance_errors"`
	IsTimeCardStartAtEvidenceRequired  bool   `json:"is_time_card_start_at_evidence_required"`
	IsManaged                          bool   `json:"is_managed"`

	JobID        string `json:"job_id,omitempty"`
	JobNumber    string `json:"job_number,omitempty"`
	BuyerID      string `json:"buyer_id,omitempty"`
	BuyerType    string `json:"buyer_type,omitempty"`
	SellerID     string `json:"seller_id,omitempty"`
	SellerType   string `json:"seller_type,omitempty"`
	CustomerID   string `json:"customer_id,omitempty"`
	CustomerName string `json:"customer,omitempty"`
	BrokerID     string `json:"broker_id,omitempty"`
	BrokerName   string `json:"broker,omitempty"`

	SellerFinancialContactID  string `json:"seller_financial_contact_id,omitempty"`
	SellerFinancialContact    string `json:"seller_financial_contact,omitempty"`
	SellerOperationsContactID string `json:"seller_operations_contact_id,omitempty"`
	SellerOperationsContact   string `json:"seller_operations_contact,omitempty"`
	BuyerOperationsContactID  string `json:"buyer_operations_contact_id,omitempty"`
	BuyerOperationsContact    string `json:"buyer_operations_contact,omitempty"`
	BuyerFinancialContactID   string `json:"buyer_financial_contact_id,omitempty"`
	BuyerFinancialContact     string `json:"buyer_financial_contact,omitempty"`

	RateIDs                       []string `json:"rate_ids,omitempty"`
	ShiftSetTimeCardConstraintIDs []string `json:"shift_set_time_card_constraint_ids,omitempty"`
	TenderStatusChangeIDs         []string `json:"tender_status_change_ids,omitempty"`
	TenderJobScheduleShiftIDs     []string `json:"tender_job_schedule_shift_ids,omitempty"`
	ServiceTypeUnitOfMeasureIDs   []string `json:"service_type_unit_of_measure_ids,omitempty"`
	CertificationRequirementIDs   []string `json:"certification_requirement_ids,omitempty"`
	ExternalIdentificationIDs     []string `json:"external_identification_ids,omitempty"`
}

func newCustomerTendersShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show customer tender details",
		Long: `Show the full details of a customer tender.

Output Fields:
  ID, status, expiration, and note
  Payment and travel settings
  Job, customer, and broker relationships
  Contact relationships
  Rates, shifts, and certification requirements

Arguments:
  <id>    The customer tender ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a customer tender
  xbe view customer-tenders show 123

  # JSON output
  xbe view customer-tenders show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runCustomerTendersShow,
	}
	initCustomerTendersShowFlags(cmd)
	return cmd
}

func init() {
	customerTendersCmd.AddCommand(newCustomerTendersShowCmd())
}

func initCustomerTendersShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCustomerTendersShow(cmd *cobra.Command, args []string) error {
	opts, err := parseCustomerTendersShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("customer tender id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[customer-tenders]", strings.Join([]string{
		"status",
		"expires-at",
		"note",
		"is-trucker-shift-rejection-permitted",
		"payment-terms",
		"payment-terms-and-conditions",
		"restrict-to-customer-truckers",
		"maximum-travel-minutes",
		"billable-travel-minutes-per-travel-mile",
		"displays-trips",
		"has-buyer-compliance-errors",
		"has-seller-compliance-errors",
		"is-time-card-start-at-evidence-required",
		"is-managed",
		"job",
		"buyer",
		"seller",
		"seller-financial-contact",
		"seller-operations-contact",
		"buyer-operations-contact",
		"buyer-financial-contact",
		"rates",
		"shift-set-time-card-constraints",
		"tender-status-changes",
		"tender-job-schedule-shifts",
		"service-type-unit-of-measures",
		"certification-requirements",
		"external-identifications",
	}, ","))
	query.Set("fields[jobs]", "job-number,job-name")
	query.Set("fields[customers]", "company-name,name")
	query.Set("fields[brokers]", "company-name,name")
	query.Set("fields[users]", "full-name,email-address,name")
	query.Set("include", "job,buyer,seller,seller-financial-contact,seller-operations-contact,buyer-operations-contact,buyer-financial-contact")

	body, _, err := client.Get(cmd.Context(), "/v1/customer-tenders/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildCustomerTenderDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderCustomerTenderDetails(cmd, details)
}

func parseCustomerTendersShowOptions(cmd *cobra.Command) (customerTendersShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return customerTendersShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildCustomerTenderDetails(resp jsonAPISingleResponse) customerTenderDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := customerTenderDetails{
		ID:                                 resource.ID,
		Status:                             stringAttr(attrs, "status"),
		ExpiresAt:                          formatDateTime(stringAttr(attrs, "expires-at")),
		Note:                               stringAttr(attrs, "note"),
		PaymentTerms:                       stringAttr(attrs, "payment-terms"),
		PaymentTermsAndConditions:          stringAttr(attrs, "payment-terms-and-conditions"),
		RestrictToCustomerTruckers:         boolAttr(attrs, "restrict-to-customer-truckers"),
		MaximumTravelMinutes:               stringAttr(attrs, "maximum-travel-minutes"),
		BillableTravelMinutesPerTravelMile: stringAttr(attrs, "billable-travel-minutes-per-travel-mile"),
		DisplaysTrips:                      boolAttr(attrs, "displays-trips"),
		IsTruckerShiftRejectionPermitted:   boolAttr(attrs, "is-trucker-shift-rejection-permitted"),
		HasBuyerComplianceErrors:           boolAttr(attrs, "has-buyer-compliance-errors"),
		HasSellerComplianceErrors:          boolAttr(attrs, "has-seller-compliance-errors"),
		IsTimeCardStartAtEvidenceRequired:  boolAttr(attrs, "is-time-card-start-at-evidence-required"),
		IsManaged:                          boolAttr(attrs, "is-managed"),
		JobID:                              relationshipIDFromMap(resource.Relationships, "job"),
		SellerFinancialContactID:           relationshipIDFromMap(resource.Relationships, "seller-financial-contact"),
		SellerOperationsContactID:          relationshipIDFromMap(resource.Relationships, "seller-operations-contact"),
		BuyerOperationsContactID:           relationshipIDFromMap(resource.Relationships, "buyer-operations-contact"),
		BuyerFinancialContactID:            relationshipIDFromMap(resource.Relationships, "buyer-financial-contact"),
		RateIDs:                            relationshipIDsFromMap(resource.Relationships, "rates"),
		ShiftSetTimeCardConstraintIDs:      relationshipIDsFromMap(resource.Relationships, "shift-set-time-card-constraints"),
		TenderStatusChangeIDs:              relationshipIDsFromMap(resource.Relationships, "tender-status-changes"),
		TenderJobScheduleShiftIDs:          relationshipIDsFromMap(resource.Relationships, "tender-job-schedule-shifts"),
		ServiceTypeUnitOfMeasureIDs:        relationshipIDsFromMap(resource.Relationships, "service-type-unit-of-measures"),
		CertificationRequirementIDs:        relationshipIDsFromMap(resource.Relationships, "certification-requirements"),
		ExternalIdentificationIDs:          relationshipIDsFromMap(resource.Relationships, "external-identifications"),
	}

	if rel, ok := resource.Relationships["buyer"]; ok && rel.Data != nil {
		details.BuyerID = rel.Data.ID
		details.BuyerType = rel.Data.Type
		if rel.Data.Type == "customers" {
			details.CustomerID = rel.Data.ID
		}
	}
	if rel, ok := resource.Relationships["seller"]; ok && rel.Data != nil {
		details.SellerID = rel.Data.ID
		details.SellerType = rel.Data.Type
		if rel.Data.Type == "brokers" {
			details.BrokerID = rel.Data.ID
		}
	}

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	if details.JobID != "" {
		if job, ok := included[resourceKey("jobs", details.JobID)]; ok {
			details.JobNumber = firstNonEmpty(
				stringAttr(job.Attributes, "job-number"),
				stringAttr(job.Attributes, "job-name"),
			)
		}
	}

	if details.CustomerID != "" {
		if customer, ok := included[resourceKey("customers", details.CustomerID)]; ok {
			details.CustomerName = firstNonEmpty(
				stringAttr(customer.Attributes, "company-name"),
				stringAttr(customer.Attributes, "name"),
			)
		}
	}

	if details.BrokerID != "" {
		if broker, ok := included[resourceKey("brokers", details.BrokerID)]; ok {
			details.BrokerName = firstNonEmpty(
				stringAttr(broker.Attributes, "company-name"),
				stringAttr(broker.Attributes, "name"),
			)
		}
	}

	if details.SellerFinancialContactID != "" {
		if user, ok := included[resourceKey("users", details.SellerFinancialContactID)]; ok {
			details.SellerFinancialContact = firstNonEmpty(
				stringAttr(user.Attributes, "full-name"),
				stringAttr(user.Attributes, "name"),
				stringAttr(user.Attributes, "email-address"),
			)
		}
	}

	if details.SellerOperationsContactID != "" {
		if user, ok := included[resourceKey("users", details.SellerOperationsContactID)]; ok {
			details.SellerOperationsContact = firstNonEmpty(
				stringAttr(user.Attributes, "full-name"),
				stringAttr(user.Attributes, "name"),
				stringAttr(user.Attributes, "email-address"),
			)
		}
	}

	if details.BuyerOperationsContactID != "" {
		if user, ok := included[resourceKey("users", details.BuyerOperationsContactID)]; ok {
			details.BuyerOperationsContact = firstNonEmpty(
				stringAttr(user.Attributes, "full-name"),
				stringAttr(user.Attributes, "name"),
				stringAttr(user.Attributes, "email-address"),
			)
		}
	}

	if details.BuyerFinancialContactID != "" {
		if user, ok := included[resourceKey("users", details.BuyerFinancialContactID)]; ok {
			details.BuyerFinancialContact = firstNonEmpty(
				stringAttr(user.Attributes, "full-name"),
				stringAttr(user.Attributes, "name"),
				stringAttr(user.Attributes, "email-address"),
			)
		}
	}

	return details
}

func renderCustomerTenderDetails(cmd *cobra.Command, details customerTenderDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.ExpiresAt != "" {
		fmt.Fprintf(out, "Expires At: %s\n", details.ExpiresAt)
	}
	if details.Note != "" {
		fmt.Fprintf(out, "Note: %s\n", details.Note)
	}
	if details.PaymentTerms != "" {
		fmt.Fprintf(out, "Payment Terms: %s\n", details.PaymentTerms)
	}
	if details.PaymentTermsAndConditions != "" {
		fmt.Fprintf(out, "Payment Terms and Conditions: %s\n", details.PaymentTermsAndConditions)
	}
	fmt.Fprintf(out, "Restrict To Customer Truckers: %t\n", details.RestrictToCustomerTruckers)
	if details.MaximumTravelMinutes != "" {
		fmt.Fprintf(out, "Maximum Travel Minutes: %s\n", details.MaximumTravelMinutes)
	}
	if details.BillableTravelMinutesPerTravelMile != "" {
		fmt.Fprintf(out, "Billable Travel Minutes Per Travel Mile: %s\n", details.BillableTravelMinutesPerTravelMile)
	}
	fmt.Fprintf(out, "Displays Trips: %t\n", details.DisplaysTrips)
	fmt.Fprintf(out, "Is Trucker Shift Rejection Permitted: %t\n", details.IsTruckerShiftRejectionPermitted)
	fmt.Fprintf(out, "Has Buyer Compliance Errors: %t\n", details.HasBuyerComplianceErrors)
	fmt.Fprintf(out, "Has Seller Compliance Errors: %t\n", details.HasSellerComplianceErrors)
	fmt.Fprintf(out, "Is Time Card Start At Evidence Required: %t\n", details.IsTimeCardStartAtEvidenceRequired)
	fmt.Fprintf(out, "Is Managed: %t\n", details.IsManaged)

	if details.JobID != "" || details.JobNumber != "" {
		label := firstNonEmpty(details.JobNumber, details.JobID)
		if details.JobNumber != "" && details.JobID != "" {
			label = fmt.Sprintf("%s (%s)", details.JobNumber, details.JobID)
		}
		fmt.Fprintf(out, "Job: %s\n", label)
	}

	if details.CustomerID != "" || details.CustomerName != "" {
		name := details.CustomerName
		if name == "" {
			name = details.CustomerID
			fmt.Fprintf(out, "Customer: %s\n", name)
		} else if details.CustomerID != "" {
			fmt.Fprintf(out, "Customer: %s (%s)\n", name, details.CustomerID)
		} else {
			fmt.Fprintf(out, "Customer: %s\n", name)
		}
	}

	if details.BrokerID != "" || details.BrokerName != "" {
		name := details.BrokerName
		if name == "" {
			name = details.BrokerID
			fmt.Fprintf(out, "Broker: %s\n", name)
		} else if details.BrokerID != "" {
			fmt.Fprintf(out, "Broker: %s (%s)\n", name, details.BrokerID)
		} else {
			fmt.Fprintf(out, "Broker: %s\n", name)
		}
	}

	if details.BuyerID != "" {
		fmt.Fprintf(out, "Buyer: %s %s\n", details.BuyerType, details.BuyerID)
	}
	if details.SellerID != "" {
		fmt.Fprintf(out, "Seller: %s %s\n", details.SellerType, details.SellerID)
	}

	if details.SellerOperationsContactID != "" || details.SellerOperationsContact != "" {
		label := firstNonEmpty(details.SellerOperationsContact, details.SellerOperationsContactID)
		if details.SellerOperationsContact != "" && details.SellerOperationsContactID != "" {
			label = fmt.Sprintf("%s (%s)", details.SellerOperationsContact, details.SellerOperationsContactID)
		}
		fmt.Fprintf(out, "Seller Operations Contact: %s\n", label)
	}
	if details.SellerFinancialContactID != "" || details.SellerFinancialContact != "" {
		label := firstNonEmpty(details.SellerFinancialContact, details.SellerFinancialContactID)
		if details.SellerFinancialContact != "" && details.SellerFinancialContactID != "" {
			label = fmt.Sprintf("%s (%s)", details.SellerFinancialContact, details.SellerFinancialContactID)
		}
		fmt.Fprintf(out, "Seller Financial Contact: %s\n", label)
	}
	if details.BuyerOperationsContactID != "" || details.BuyerOperationsContact != "" {
		label := firstNonEmpty(details.BuyerOperationsContact, details.BuyerOperationsContactID)
		if details.BuyerOperationsContact != "" && details.BuyerOperationsContactID != "" {
			label = fmt.Sprintf("%s (%s)", details.BuyerOperationsContact, details.BuyerOperationsContactID)
		}
		fmt.Fprintf(out, "Buyer Operations Contact: %s\n", label)
	}
	if details.BuyerFinancialContactID != "" || details.BuyerFinancialContact != "" {
		label := firstNonEmpty(details.BuyerFinancialContact, details.BuyerFinancialContactID)
		if details.BuyerFinancialContact != "" && details.BuyerFinancialContactID != "" {
			label = fmt.Sprintf("%s (%s)", details.BuyerFinancialContact, details.BuyerFinancialContactID)
		}
		fmt.Fprintf(out, "Buyer Financial Contact: %s\n", label)
	}

	if len(details.RateIDs) > 0 {
		fmt.Fprintf(out, "Rate IDs: %s\n", strings.Join(details.RateIDs, ", "))
	}
	if len(details.ShiftSetTimeCardConstraintIDs) > 0 {
		fmt.Fprintf(out, "Shift Set Time Card Constraint IDs: %s\n", strings.Join(details.ShiftSetTimeCardConstraintIDs, ", "))
	}
	if len(details.TenderStatusChangeIDs) > 0 {
		fmt.Fprintf(out, "Tender Status Change IDs: %s\n", strings.Join(details.TenderStatusChangeIDs, ", "))
	}
	if len(details.TenderJobScheduleShiftIDs) > 0 {
		fmt.Fprintf(out, "Tender Job Schedule Shift IDs: %s\n", strings.Join(details.TenderJobScheduleShiftIDs, ", "))
	}
	if len(details.ServiceTypeUnitOfMeasureIDs) > 0 {
		fmt.Fprintf(out, "Service Type Unit Of Measure IDs: %s\n", strings.Join(details.ServiceTypeUnitOfMeasureIDs, ", "))
	}
	if len(details.CertificationRequirementIDs) > 0 {
		fmt.Fprintf(out, "Certification Requirement IDs: %s\n", strings.Join(details.CertificationRequirementIDs, ", "))
	}
	if len(details.ExternalIdentificationIDs) > 0 {
		fmt.Fprintf(out, "External Identification IDs: %s\n", strings.Join(details.ExternalIdentificationIDs, ", "))
	}

	return nil
}
