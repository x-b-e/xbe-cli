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

type tendersShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type tenderDetails struct {
	ID                                 string `json:"id"`
	Type                               string `json:"type,omitempty"`
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

	JobID      string `json:"job_id,omitempty"`
	JobNumber  string `json:"job_number,omitempty"`
	BuyerID    string `json:"buyer_id,omitempty"`
	BuyerType  string `json:"buyer_type,omitempty"`
	BuyerName  string `json:"buyer,omitempty"`
	SellerID   string `json:"seller_id,omitempty"`
	SellerType string `json:"seller_type,omitempty"`
	SellerName string `json:"seller,omitempty"`

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
}

func newTendersShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show tender details",
		Long: `Show the full details of a tender.

Output Fields:
  ID, type, status, expiration, and note
  Payment and travel settings
  Job, buyer, and seller relationships
  Contact relationships
  Rates, shifts, and service type unit of measures

Arguments:
  <id>    The tender ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a tender
  xbe view tenders show 123

  # JSON output
  xbe view tenders show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTendersShow,
	}
	initTendersShowFlags(cmd)
	return cmd
}

func init() {
	tendersCmd.AddCommand(newTendersShowCmd())
}

func initTendersShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTendersShow(cmd *cobra.Command, args []string) error {
	opts, err := parseTendersShowOptions(cmd)
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
		return fmt.Errorf("tender id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[tenders]", strings.Join([]string{
		"polymorphic-type",
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
	}, ","))
	query.Set("fields[jobs]", "external-job-number")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[truckers]", "company-name")
	query.Set("fields[users]", "name,email-address")
	query.Set("include", "job,buyer,seller,seller-financial-contact,seller-operations-contact,buyer-operations-contact,buyer-financial-contact")

	body, _, err := client.Get(cmd.Context(), "/v1/tenders/"+id, query)
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

	details := buildTenderDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTenderDetails(cmd, details)
}

func parseTendersShowOptions(cmd *cobra.Command) (tendersShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return tendersShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTenderDetails(resp jsonAPISingleResponse) tenderDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := tenderDetails{
		ID:                                 resource.ID,
		Type:                               stringAttr(attrs, "polymorphic-type"),
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
	}

	if rel, ok := resource.Relationships["buyer"]; ok && rel.Data != nil {
		details.BuyerID = rel.Data.ID
		details.BuyerType = rel.Data.Type
	}
	if rel, ok := resource.Relationships["seller"]; ok && rel.Data != nil {
		details.SellerID = rel.Data.ID
		details.SellerType = rel.Data.Type
	}

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	if details.JobID != "" {
		if job, ok := included[resourceKey("jobs", details.JobID)]; ok {
			details.JobNumber = firstNonEmpty(
				stringAttr(job.Attributes, "external-job-number"),
			)
		}
	}

	details.BuyerName = resolveTenderPartyName(included, details.BuyerType, details.BuyerID)
	details.SellerName = resolveTenderPartyName(included, details.SellerType, details.SellerID)

	if details.SellerFinancialContactID != "" {
		if user, ok := included[resourceKey("users", details.SellerFinancialContactID)]; ok {
			details.SellerFinancialContact = firstNonEmpty(
				stringAttr(user.Attributes, "name"),
				stringAttr(user.Attributes, "email-address"),
			)
		}
	}

	if details.SellerOperationsContactID != "" {
		if user, ok := included[resourceKey("users", details.SellerOperationsContactID)]; ok {
			details.SellerOperationsContact = firstNonEmpty(
				stringAttr(user.Attributes, "name"),
				stringAttr(user.Attributes, "email-address"),
			)
		}
	}

	if details.BuyerOperationsContactID != "" {
		if user, ok := included[resourceKey("users", details.BuyerOperationsContactID)]; ok {
			details.BuyerOperationsContact = firstNonEmpty(
				stringAttr(user.Attributes, "name"),
				stringAttr(user.Attributes, "email-address"),
			)
		}
	}

	if details.BuyerFinancialContactID != "" {
		if user, ok := included[resourceKey("users", details.BuyerFinancialContactID)]; ok {
			details.BuyerFinancialContact = firstNonEmpty(
				stringAttr(user.Attributes, "name"),
				stringAttr(user.Attributes, "email-address"),
			)
		}
	}

	return details
}

func renderTenderDetails(cmd *cobra.Command, details tenderDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Type != "" {
		fmt.Fprintf(out, "Type: %s\n", details.Type)
	}
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

	if details.BuyerID != "" || details.BuyerName != "" {
		buyerLabel := formatTenderPartyLabel(details.BuyerType, details.BuyerID, details.BuyerName)
		fmt.Fprintf(out, "Buyer: %s\n", buyerLabel)
	}
	if details.SellerID != "" || details.SellerName != "" {
		sellerLabel := formatTenderPartyLabel(details.SellerType, details.SellerID, details.SellerName)
		fmt.Fprintf(out, "Seller: %s\n", sellerLabel)
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

	return nil
}
