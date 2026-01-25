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

type doCustomerTendersCreateOptions struct {
	BaseURL                            string
	Token                              string
	JSON                               bool
	Job                                string
	Customer                           string
	Broker                             string
	ExpiresAt                          string
	Note                               string
	IsTruckerShiftRejectionPermitted   string
	PaymentTerms                       string
	PaymentTermsAndConditions          string
	RestrictToCustomerTruckers         string
	MaximumTravelMinutes               string
	BillableTravelMinutesPerTravelMile string
	DisplaysTrips                      string
	SellerFinancialContact             string
	SellerOperationsContact            string
	BuyerOperationsContact             string
	BuyerFinancialContact              string
	CertificationRequirements          string
}

func newDoCustomerTendersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a customer tender",
		Long: `Create a customer tender.

Required flags:
  --job        Job ID
  --customer   Customer ID (buyer)
  --broker     Broker ID (seller)

Optional fields:
  --expires-at                         Expiration time (RFC3339)
  --note                               Note for the tender
  --is-trucker-shift-rejection-permitted  Allow trucker shift rejection (true/false)
  --payment-terms                      Payment terms (integer)
  --payment-terms-and-conditions       Payment terms and conditions
  --restrict-to-customer-truckers      Restrict to customer truckers (true/false)
  --maximum-travel-minutes             Maximum travel minutes
  --billable-travel-minutes-per-travel-mile Billable travel minutes per travel mile
  --displays-trips                     Display trips (true/false)
  --seller-financial-contact           Seller financial contact user ID
  --seller-operations-contact          Seller operations contact user ID
  --buyer-operations-contact           Buyer operations contact user ID
  --buyer-financial-contact            Buyer financial contact user ID
  --certification-requirements         Certification requirement IDs (comma-separated)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a customer tender
  xbe do customer-tenders create --job 123 --customer 456 --broker 789

  # Create with payment terms and note
  xbe do customer-tenders create --job 123 --customer 456 --broker 789 --payment-terms 30 --note "Dispatch ASAP"

  # JSON output
  xbe do customer-tenders create --job 123 --customer 456 --broker 789 --json`,
		RunE: runDoCustomerTendersCreate,
	}
	initDoCustomerTendersCreateFlags(cmd)
	return cmd
}

func init() {
	doCustomerTendersCmd.AddCommand(newDoCustomerTendersCreateCmd())
}

func initDoCustomerTendersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job", "", "Job ID (required)")
	cmd.Flags().String("customer", "", "Customer ID (required)")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("expires-at", "", "Expiration time (RFC3339)")
	cmd.Flags().String("note", "", "Note for the tender")
	cmd.Flags().String("is-trucker-shift-rejection-permitted", "", "Allow trucker shift rejection (true/false)")
	cmd.Flags().String("payment-terms", "", "Payment terms (integer)")
	cmd.Flags().String("payment-terms-and-conditions", "", "Payment terms and conditions")
	cmd.Flags().String("restrict-to-customer-truckers", "", "Restrict to customer truckers (true/false)")
	cmd.Flags().String("maximum-travel-minutes", "", "Maximum travel minutes")
	cmd.Flags().String("billable-travel-minutes-per-travel-mile", "", "Billable travel minutes per travel mile")
	cmd.Flags().String("displays-trips", "", "Display trips (true/false)")
	cmd.Flags().String("seller-financial-contact", "", "Seller financial contact user ID")
	cmd.Flags().String("seller-operations-contact", "", "Seller operations contact user ID")
	cmd.Flags().String("buyer-operations-contact", "", "Buyer operations contact user ID")
	cmd.Flags().String("buyer-financial-contact", "", "Buyer financial contact user ID")
	cmd.Flags().String("certification-requirements", "", "Certification requirement IDs (comma-separated)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("job")
	cmd.MarkFlagRequired("customer")
	cmd.MarkFlagRequired("broker")
}

func runDoCustomerTendersCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoCustomerTendersCreateOptions(cmd)
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
	if opts.ExpiresAt != "" {
		attributes["expires-at"] = opts.ExpiresAt
	}
	if opts.Note != "" {
		attributes["note"] = opts.Note
	}
	if opts.IsTruckerShiftRejectionPermitted != "" {
		attributes["is-trucker-shift-rejection-permitted"] = opts.IsTruckerShiftRejectionPermitted == "true"
	}
	if opts.PaymentTerms != "" {
		attributes["payment-terms"] = opts.PaymentTerms
	}
	if opts.PaymentTermsAndConditions != "" {
		attributes["payment-terms-and-conditions"] = opts.PaymentTermsAndConditions
	}
	if opts.RestrictToCustomerTruckers != "" {
		attributes["restrict-to-customer-truckers"] = opts.RestrictToCustomerTruckers == "true"
	}
	if opts.MaximumTravelMinutes != "" {
		attributes["maximum-travel-minutes"] = opts.MaximumTravelMinutes
	}
	if opts.BillableTravelMinutesPerTravelMile != "" {
		attributes["billable-travel-minutes-per-travel-mile"] = opts.BillableTravelMinutesPerTravelMile
	}
	if opts.DisplaysTrips != "" {
		attributes["displays-trips"] = opts.DisplaysTrips == "true"
	}

	relationships := map[string]any{
		"job": map[string]any{
			"data": map[string]any{
				"type": "jobs",
				"id":   opts.Job,
			},
		},
		"buyer": map[string]any{
			"data": map[string]any{
				"type": "customers",
				"id":   opts.Customer,
			},
		},
		"seller": map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.Broker,
			},
		},
	}

	if opts.SellerFinancialContact != "" {
		relationships["seller-financial-contact"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.SellerFinancialContact,
			},
		}
	}
	if opts.SellerOperationsContact != "" {
		relationships["seller-operations-contact"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.SellerOperationsContact,
			},
		}
	}
	if opts.BuyerOperationsContact != "" {
		relationships["buyer-operations-contact"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.BuyerOperationsContact,
			},
		}
	}
	if opts.BuyerFinancialContact != "" {
		relationships["buyer-financial-contact"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.BuyerFinancialContact,
			},
		}
	}

	if opts.CertificationRequirements != "" {
		ids := splitCommaList(opts.CertificationRequirements)
		if len(ids) > 0 {
			data := make([]map[string]any, 0, len(ids))
			for _, id := range ids {
				data = append(data, map[string]any{
					"type": "certification-requirements",
					"id":   id,
				})
			}
			relationships["certification-requirements"] = map[string]any{"data": data}
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "customer-tenders",
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

	body, _, err := client.Post(cmd.Context(), "/v1/customer-tenders", jsonBody)
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

	row := buildCustomerTenderRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	if row.Status != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Created customer tender %s (status: %s)\n", row.ID, row.Status)
		return nil
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Created customer tender %s\n", row.ID)
	return nil
}

func parseDoCustomerTendersCreateOptions(cmd *cobra.Command) (doCustomerTendersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	job, _ := cmd.Flags().GetString("job")
	customer, _ := cmd.Flags().GetString("customer")
	broker, _ := cmd.Flags().GetString("broker")
	expiresAt, _ := cmd.Flags().GetString("expires-at")
	note, _ := cmd.Flags().GetString("note")
	isTruckerShiftRejectionPermitted, _ := cmd.Flags().GetString("is-trucker-shift-rejection-permitted")
	paymentTerms, _ := cmd.Flags().GetString("payment-terms")
	paymentTermsAndConditions, _ := cmd.Flags().GetString("payment-terms-and-conditions")
	restrictToCustomerTruckers, _ := cmd.Flags().GetString("restrict-to-customer-truckers")
	maximumTravelMinutes, _ := cmd.Flags().GetString("maximum-travel-minutes")
	billableTravelMinutesPerTravelMile, _ := cmd.Flags().GetString("billable-travel-minutes-per-travel-mile")
	displaysTrips, _ := cmd.Flags().GetString("displays-trips")
	sellerFinancialContact, _ := cmd.Flags().GetString("seller-financial-contact")
	sellerOperationsContact, _ := cmd.Flags().GetString("seller-operations-contact")
	buyerOperationsContact, _ := cmd.Flags().GetString("buyer-operations-contact")
	buyerFinancialContact, _ := cmd.Flags().GetString("buyer-financial-contact")
	certificationRequirements, _ := cmd.Flags().GetString("certification-requirements")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCustomerTendersCreateOptions{
		BaseURL:                            baseURL,
		Token:                              token,
		JSON:                               jsonOut,
		Job:                                job,
		Customer:                           customer,
		Broker:                             broker,
		ExpiresAt:                          expiresAt,
		Note:                               note,
		IsTruckerShiftRejectionPermitted:   isTruckerShiftRejectionPermitted,
		PaymentTerms:                       paymentTerms,
		PaymentTermsAndConditions:          paymentTermsAndConditions,
		RestrictToCustomerTruckers:         restrictToCustomerTruckers,
		MaximumTravelMinutes:               maximumTravelMinutes,
		BillableTravelMinutesPerTravelMile: billableTravelMinutesPerTravelMile,
		DisplaysTrips:                      displaysTrips,
		SellerFinancialContact:             sellerFinancialContact,
		SellerOperationsContact:            sellerOperationsContact,
		BuyerOperationsContact:             buyerOperationsContact,
		BuyerFinancialContact:              buyerFinancialContact,
		CertificationRequirements:          certificationRequirements,
	}, nil
}
