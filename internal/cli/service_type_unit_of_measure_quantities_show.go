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

type serviceTypeUnitOfMeasureQuantitiesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type serviceTypeUnitOfMeasureQuantityDetails struct {
	ID                                     string `json:"id"`
	Quantity                               string `json:"quantity,omitempty"`
	ExplicitQuantity                       string `json:"explicit_quantity,omitempty"`
	CalculatedQuantity                     string `json:"calculated_quantity,omitempty"`
	UserCanMutate                          bool   `json:"user_can_mutate"`
	QuantifiesAcceptedMaterialTransactions bool   `json:"quantifies_accepted_material_transactions"`
	ServiceTypeUnitOfMeasureID             string `json:"service_type_unit_of_measure_id,omitempty"`
	ServiceTypeUnitOfMeasureName           string `json:"service_type_unit_of_measure_name,omitempty"`
	QuantifiesType                         string `json:"quantifies_type,omitempty"`
	QuantifiesID                           string `json:"quantifies_id,omitempty"`
	TimeCardID                             string `json:"time_card_id,omitempty"`
	MaterialTypeID                         string `json:"material_type_id,omitempty"`
	MaterialTypeName                       string `json:"material_type_name,omitempty"`
	TrailerClassificationID                string `json:"trailer_classification_id,omitempty"`
	TrailerClassificationName              string `json:"trailer_classification_name,omitempty"`

	CalculatedQuantityCustomer string `json:"calculated_quantity_customer,omitempty"`
	QuantityCustomer           string `json:"quantity_customer,omitempty"`
	QuantityBroker             string `json:"quantity_broker,omitempty"`

	CalculatedRateAmountCustomer            string `json:"calculated_rate_amount_customer,omitempty"`
	CalculatedRateBaseAmountCustomer        string `json:"calculated_rate_base_amount_customer,omitempty"`
	CalculatedRateAdjustmentsAmountCustomer string `json:"calculated_rate_adjustments_amount_customer,omitempty"`
	RateAmountCustomer                      string `json:"rate_amount_customer,omitempty"`
	RateAmountBroker                        string `json:"rate_amount_broker,omitempty"`
	RateBaseAmountBroker                    string `json:"rate_base_amount_broker,omitempty"`
	RateAdjustmentsAmountBroker             string `json:"rate_adjustments_amount_broker,omitempty"`

	IsCustomerCostPlus         bool   `json:"is_customer_cost_plus"`
	CostPlusRateAmountCustomer string `json:"cost_plus_rate_amount_customer,omitempty"`

	CalculatedAmountCustomer            string `json:"calculated_amount_customer,omitempty"`
	CalculatedBaseAmountCustomer        string `json:"calculated_base_amount_customer,omitempty"`
	CalculatedAdjustmentsAmountCustomer string `json:"calculated_adjustments_amount_customer,omitempty"`
	AmountCustomer                      string `json:"amount_customer,omitempty"`
	AmountBroker                        string `json:"amount_broker,omitempty"`
	BaseAmountBroker                    string `json:"base_amount_broker,omitempty"`
	AdjustmentsAmountBroker             string `json:"adjustments_amount_broker,omitempty"`

	CalculatedRateStatusCustomer string `json:"calculated_rate_status_customer,omitempty"`
	RateStatusCustomer           string `json:"rate_status_customer,omitempty"`
	RateStatusBroker             string `json:"rate_status_broker,omitempty"`
}

func newServiceTypeUnitOfMeasureQuantitiesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show service type unit of measure quantity details",
		Long: `Show the full details of a service type unit of measure quantity.

Output Fields:
  ID, quantities, and relationship details
  Customer/broker calculated amounts and rates (when visible)

Arguments:
  <id>  The service type unit of measure quantity ID (required).`,
		Example: `  # Show a service type unit of measure quantity
  xbe view service-type-unit-of-measure-quantities show 123

  # Output as JSON
  xbe view service-type-unit-of-measure-quantities show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runServiceTypeUnitOfMeasureQuantitiesShow,
	}
	initServiceTypeUnitOfMeasureQuantitiesShowFlags(cmd)
	return cmd
}

func init() {
	serviceTypeUnitOfMeasureQuantitiesCmd.AddCommand(newServiceTypeUnitOfMeasureQuantitiesShowCmd())
}

func initServiceTypeUnitOfMeasureQuantitiesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runServiceTypeUnitOfMeasureQuantitiesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseServiceTypeUnitOfMeasureQuantitiesShowOptions(cmd)
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
		return fmt.Errorf("service type unit of measure quantity id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[service-type-unit-of-measure-quantities]", strings.Join([]string{
		"quantity",
		"explicit-quantity",
		"calculated-quantity",
		"user-can-mutate",
		"quantifies-accepted-material-transactions",
		"calculated-quantity-customer",
		"quantity-customer",
		"quantity-broker",
		"calculated-rate-amount-customer",
		"calculated-rate-base-amount-customer",
		"calculated-rate-adjustments-amount-customer",
		"rate-amount-customer",
		"rate-amount-broker",
		"rate-base-amount-broker",
		"rate-adjustments-amount-broker",
		"is-customer-cost-plus",
		"cost-plus-rate-amount-customer",
		"calculated-amount-customer",
		"calculated-base-amount-customer",
		"calculated-adjustments-amount-customer",
		"amount-customer",
		"amount-broker",
		"base-amount-broker",
		"adjustments-amount-broker",
		"calculated-rate-status-customer",
		"rate-status-customer",
		"rate-status-broker",
		"service-type-unit-of-measure",
		"quantifies",
		"time-card",
		"material-type",
		"trailer-classification",
	}, ","))
	query.Set("include", "service-type-unit-of-measure,material-type,trailer-classification")
	query.Set("fields[service-type-unit-of-measures]", "name")
	query.Set("fields[material-types]", "name,display-name")
	query.Set("fields[trailer-classifications]", "name,abbreviation")

	body, _, err := client.Get(cmd.Context(), "/v1/service-type-unit-of-measure-quantities/"+id, query)
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

	details := buildServiceTypeUnitOfMeasureQuantityDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderServiceTypeUnitOfMeasureQuantityDetails(cmd, details)
}

func parseServiceTypeUnitOfMeasureQuantitiesShowOptions(cmd *cobra.Command) (serviceTypeUnitOfMeasureQuantitiesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return serviceTypeUnitOfMeasureQuantitiesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildServiceTypeUnitOfMeasureQuantityDetails(resp jsonAPISingleResponse) serviceTypeUnitOfMeasureQuantityDetails {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	resource := resp.Data
	attrs := resource.Attributes
	details := serviceTypeUnitOfMeasureQuantityDetails{
		ID:                                      resource.ID,
		Quantity:                                stringAttr(attrs, "quantity"),
		ExplicitQuantity:                        stringAttr(attrs, "explicit-quantity"),
		CalculatedQuantity:                      stringAttr(attrs, "calculated-quantity"),
		UserCanMutate:                           boolAttr(attrs, "user-can-mutate"),
		QuantifiesAcceptedMaterialTransactions:  boolAttr(attrs, "quantifies-accepted-material-transactions"),
		CalculatedQuantityCustomer:              stringAttr(attrs, "calculated-quantity-customer"),
		QuantityCustomer:                        stringAttr(attrs, "quantity-customer"),
		QuantityBroker:                          stringAttr(attrs, "quantity-broker"),
		CalculatedRateAmountCustomer:            stringAttr(attrs, "calculated-rate-amount-customer"),
		CalculatedRateBaseAmountCustomer:        stringAttr(attrs, "calculated-rate-base-amount-customer"),
		CalculatedRateAdjustmentsAmountCustomer: stringAttr(attrs, "calculated-rate-adjustments-amount-customer"),
		RateAmountCustomer:                      stringAttr(attrs, "rate-amount-customer"),
		RateAmountBroker:                        stringAttr(attrs, "rate-amount-broker"),
		RateBaseAmountBroker:                    stringAttr(attrs, "rate-base-amount-broker"),
		RateAdjustmentsAmountBroker:             stringAttr(attrs, "rate-adjustments-amount-broker"),
		IsCustomerCostPlus:                      boolAttr(attrs, "is-customer-cost-plus"),
		CostPlusRateAmountCustomer:              stringAttr(attrs, "cost-plus-rate-amount-customer"),
		CalculatedAmountCustomer:                stringAttr(attrs, "calculated-amount-customer"),
		CalculatedBaseAmountCustomer:            stringAttr(attrs, "calculated-base-amount-customer"),
		CalculatedAdjustmentsAmountCustomer:     stringAttr(attrs, "calculated-adjustments-amount-customer"),
		AmountCustomer:                          stringAttr(attrs, "amount-customer"),
		AmountBroker:                            stringAttr(attrs, "amount-broker"),
		BaseAmountBroker:                        stringAttr(attrs, "base-amount-broker"),
		AdjustmentsAmountBroker:                 stringAttr(attrs, "adjustments-amount-broker"),
		CalculatedRateStatusCustomer:            stringAttr(attrs, "calculated-rate-status-customer"),
		RateStatusCustomer:                      stringAttr(attrs, "rate-status-customer"),
		RateStatusBroker:                        stringAttr(attrs, "rate-status-broker"),
	}

	if rel, ok := resource.Relationships["service-type-unit-of-measure"]; ok && rel.Data != nil {
		details.ServiceTypeUnitOfMeasureID = rel.Data.ID
		if stuom, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.ServiceTypeUnitOfMeasureName = strings.TrimSpace(stringAttr(stuom.Attributes, "name"))
		}
	}

	if rel, ok := resource.Relationships["quantifies"]; ok && rel.Data != nil {
		details.QuantifiesType = rel.Data.Type
		details.QuantifiesID = rel.Data.ID
	}

	if rel, ok := resource.Relationships["time-card"]; ok && rel.Data != nil {
		details.TimeCardID = rel.Data.ID
	}

	if rel, ok := resource.Relationships["material-type"]; ok && rel.Data != nil {
		details.MaterialTypeID = rel.Data.ID
		if materialType, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.MaterialTypeName = materialTypeLabel(materialType.Attributes)
		}
	}

	if rel, ok := resource.Relationships["trailer-classification"]; ok && rel.Data != nil {
		details.TrailerClassificationID = rel.Data.ID
		if trailer, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.TrailerClassificationName = trailerClassificationLabel(trailer.Attributes)
		}
	}

	return details
}

func renderServiceTypeUnitOfMeasureQuantityDetails(cmd *cobra.Command, details serviceTypeUnitOfMeasureQuantityDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Quantity != "" {
		fmt.Fprintf(out, "Quantity: %s\n", details.Quantity)
	}
	if details.ExplicitQuantity != "" {
		fmt.Fprintf(out, "Explicit Quantity: %s\n", details.ExplicitQuantity)
	}
	if details.CalculatedQuantity != "" {
		fmt.Fprintf(out, "Calculated Quantity: %s\n", details.CalculatedQuantity)
	}

	if details.ServiceTypeUnitOfMeasureID != "" {
		label := details.ServiceTypeUnitOfMeasureID
		if details.ServiceTypeUnitOfMeasureName != "" {
			label = fmt.Sprintf("%s (%s)", details.ServiceTypeUnitOfMeasureName, details.ServiceTypeUnitOfMeasureID)
		}
		fmt.Fprintf(out, "Service Type UOM: %s\n", label)
	}

	if details.QuantifiesType != "" && details.QuantifiesID != "" {
		fmt.Fprintf(out, "Quantifies: %s/%s\n", details.QuantifiesType, details.QuantifiesID)
	}
	if details.TimeCardID != "" {
		fmt.Fprintf(out, "Time Card: %s\n", details.TimeCardID)
	}
	if details.MaterialTypeID != "" {
		label := details.MaterialTypeID
		if details.MaterialTypeName != "" {
			label = fmt.Sprintf("%s (%s)", details.MaterialTypeName, details.MaterialTypeID)
		}
		fmt.Fprintf(out, "Material Type: %s\n", label)
	}
	if details.TrailerClassificationID != "" {
		label := details.TrailerClassificationID
		if details.TrailerClassificationName != "" {
			label = fmt.Sprintf("%s (%s)", details.TrailerClassificationName, details.TrailerClassificationID)
		}
		fmt.Fprintf(out, "Trailer Classification: %s\n", label)
	}

	fmt.Fprintf(out, "User Can Mutate: %t\n", details.UserCanMutate)
	fmt.Fprintf(out, "Quantifies Accepted Material Transactions: %t\n", details.QuantifiesAcceptedMaterialTransactions)

	if details.QuantityCustomer != "" {
		fmt.Fprintf(out, "Customer Quantity: %s\n", details.QuantityCustomer)
	}
	if details.CalculatedQuantityCustomer != "" {
		fmt.Fprintf(out, "Customer Calculated Quantity: %s\n", details.CalculatedQuantityCustomer)
	}
	if details.QuantityBroker != "" {
		fmt.Fprintf(out, "Broker Quantity: %s\n", details.QuantityBroker)
	}

	if details.IsCustomerCostPlus {
		fmt.Fprintf(out, "Customer Cost Plus: %t\n", details.IsCustomerCostPlus)
	}
	if details.CostPlusRateAmountCustomer != "" {
		fmt.Fprintf(out, "Customer Cost Plus Rate Amount: %s\n", details.CostPlusRateAmountCustomer)
	}

	if details.RateAmountCustomer != "" {
		fmt.Fprintf(out, "Customer Rate Amount: %s\n", details.RateAmountCustomer)
	}
	if details.CalculatedRateAmountCustomer != "" {
		fmt.Fprintf(out, "Customer Calculated Rate Amount: %s\n", details.CalculatedRateAmountCustomer)
	}
	if details.CalculatedRateBaseAmountCustomer != "" {
		fmt.Fprintf(out, "Customer Calculated Rate Base Amount: %s\n", details.CalculatedRateBaseAmountCustomer)
	}
	if details.CalculatedRateAdjustmentsAmountCustomer != "" {
		fmt.Fprintf(out, "Customer Calculated Rate Adjustments Amount: %s\n", details.CalculatedRateAdjustmentsAmountCustomer)
	}
	if details.RateAmountBroker != "" {
		fmt.Fprintf(out, "Broker Rate Amount: %s\n", details.RateAmountBroker)
	}
	if details.RateBaseAmountBroker != "" {
		fmt.Fprintf(out, "Broker Rate Base Amount: %s\n", details.RateBaseAmountBroker)
	}
	if details.RateAdjustmentsAmountBroker != "" {
		fmt.Fprintf(out, "Broker Rate Adjustments Amount: %s\n", details.RateAdjustmentsAmountBroker)
	}

	if details.AmountCustomer != "" {
		fmt.Fprintf(out, "Customer Amount: %s\n", details.AmountCustomer)
	}
	if details.CalculatedAmountCustomer != "" {
		fmt.Fprintf(out, "Customer Calculated Amount: %s\n", details.CalculatedAmountCustomer)
	}
	if details.CalculatedBaseAmountCustomer != "" {
		fmt.Fprintf(out, "Customer Calculated Base Amount: %s\n", details.CalculatedBaseAmountCustomer)
	}
	if details.CalculatedAdjustmentsAmountCustomer != "" {
		fmt.Fprintf(out, "Customer Calculated Adjustments Amount: %s\n", details.CalculatedAdjustmentsAmountCustomer)
	}

	if details.AmountBroker != "" {
		fmt.Fprintf(out, "Broker Amount: %s\n", details.AmountBroker)
	}
	if details.BaseAmountBroker != "" {
		fmt.Fprintf(out, "Broker Base Amount: %s\n", details.BaseAmountBroker)
	}
	if details.AdjustmentsAmountBroker != "" {
		fmt.Fprintf(out, "Broker Adjustments Amount: %s\n", details.AdjustmentsAmountBroker)
	}

	if details.CalculatedRateStatusCustomer != "" {
		fmt.Fprintf(out, "Customer Calculated Rate Status: %s\n", details.CalculatedRateStatusCustomer)
	}
	if details.RateStatusCustomer != "" {
		fmt.Fprintf(out, "Customer Rate Status: %s\n", details.RateStatusCustomer)
	}
	if details.RateStatusBroker != "" {
		fmt.Fprintf(out, "Broker Rate Status: %s\n", details.RateStatusBroker)
	}

	return nil
}
