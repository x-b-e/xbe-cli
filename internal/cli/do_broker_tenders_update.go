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

type doBrokerTendersUpdateOptions struct {
	BaseURL                            string
	Token                              string
	JSON                               bool
	ID                                 string
	Job                                string
	Broker                             string
	Trucker                            string
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
}

func newDoBrokerTendersUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a broker tender",
		Long: `Update a broker tender.

Provide the broker tender ID as an argument and at least one field to update.

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update a broker tender note
  xbe do broker-tenders update 123 --note "Updated note"

  # Update payment terms
  xbe do broker-tenders update 123 --payment-terms 30

  # JSON output
  xbe do broker-tenders update 123 --note "Updated note" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoBrokerTendersUpdate,
	}
	initDoBrokerTendersUpdateFlags(cmd)
	return cmd
}

func init() {
	doBrokerTendersCmd.AddCommand(newDoBrokerTendersUpdateCmd())
}

func initDoBrokerTendersUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job", "", "Job ID")
	cmd.Flags().String("broker", "", "Broker ID (buyer)")
	cmd.Flags().String("trucker", "", "Trucker ID (seller)")
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
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoBrokerTendersUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoBrokerTendersUpdateOptions(cmd, args)
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

	id := strings.TrimSpace(opts.ID)
	if id == "" {
		return fmt.Errorf("broker tender id is required")
	}

	attributes := map[string]any{}
	relationships := map[string]any{}
	hasChanges := false

	if cmd.Flags().Changed("expires-at") {
		attributes["expires-at"] = opts.ExpiresAt
		hasChanges = true
	}
	if cmd.Flags().Changed("note") {
		attributes["note"] = opts.Note
		hasChanges = true
	}
	if cmd.Flags().Changed("is-trucker-shift-rejection-permitted") {
		attributes["is-trucker-shift-rejection-permitted"] = opts.IsTruckerShiftRejectionPermitted == "true"
		hasChanges = true
	}
	if cmd.Flags().Changed("payment-terms") {
		attributes["payment-terms"] = opts.PaymentTerms
		hasChanges = true
	}
	if cmd.Flags().Changed("payment-terms-and-conditions") {
		attributes["payment-terms-and-conditions"] = opts.PaymentTermsAndConditions
		hasChanges = true
	}
	if cmd.Flags().Changed("restrict-to-customer-truckers") {
		attributes["restrict-to-customer-truckers"] = opts.RestrictToCustomerTruckers == "true"
		hasChanges = true
	}
	if cmd.Flags().Changed("maximum-travel-minutes") {
		attributes["maximum-travel-minutes"] = opts.MaximumTravelMinutes
		hasChanges = true
	}
	if cmd.Flags().Changed("billable-travel-minutes-per-travel-mile") {
		attributes["billable-travel-minutes-per-travel-mile"] = opts.BillableTravelMinutesPerTravelMile
		hasChanges = true
	}
	if cmd.Flags().Changed("displays-trips") {
		attributes["displays-trips"] = opts.DisplaysTrips == "true"
		hasChanges = true
	}

	if cmd.Flags().Changed("job") {
		if opts.Job == "" {
			relationships["job"] = map[string]any{"data": nil}
		} else {
			relationships["job"] = map[string]any{
				"data": map[string]any{
					"type": "jobs",
					"id":   opts.Job,
				},
			}
		}
		hasChanges = true
	}

	if cmd.Flags().Changed("broker") {
		if opts.Broker == "" {
			relationships["buyer"] = map[string]any{"data": nil}
		} else {
			relationships["buyer"] = map[string]any{
				"data": map[string]any{
					"type": "brokers",
					"id":   opts.Broker,
				},
			}
		}
		hasChanges = true
	}

	if cmd.Flags().Changed("trucker") {
		if opts.Trucker == "" {
			relationships["seller"] = map[string]any{"data": nil}
		} else {
			relationships["seller"] = map[string]any{
				"data": map[string]any{
					"type": "truckers",
					"id":   opts.Trucker,
				},
			}
		}
		hasChanges = true
	}

	if cmd.Flags().Changed("seller-financial-contact") {
		if opts.SellerFinancialContact == "" {
			relationships["seller-financial-contact"] = map[string]any{"data": nil}
		} else {
			relationships["seller-financial-contact"] = map[string]any{
				"data": map[string]any{
					"type": "users",
					"id":   opts.SellerFinancialContact,
				},
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("seller-operations-contact") {
		if opts.SellerOperationsContact == "" {
			relationships["seller-operations-contact"] = map[string]any{"data": nil}
		} else {
			relationships["seller-operations-contact"] = map[string]any{
				"data": map[string]any{
					"type": "users",
					"id":   opts.SellerOperationsContact,
				},
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("buyer-operations-contact") {
		if opts.BuyerOperationsContact == "" {
			relationships["buyer-operations-contact"] = map[string]any{"data": nil}
		} else {
			relationships["buyer-operations-contact"] = map[string]any{
				"data": map[string]any{
					"type": "users",
					"id":   opts.BuyerOperationsContact,
				},
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("buyer-financial-contact") {
		if opts.BuyerFinancialContact == "" {
			relationships["buyer-financial-contact"] = map[string]any{"data": nil}
		} else {
			relationships["buyer-financial-contact"] = map[string]any{
				"data": map[string]any{
					"type": "users",
					"id":   opts.BuyerFinancialContact,
				},
			}
		}
		hasChanges = true
	}

	if !hasChanges {
		return fmt.Errorf("no fields to update; specify at least one flag")
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "broker-tenders",
			"id":            id,
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

	body, _, err := client.Patch(cmd.Context(), "/v1/broker-tenders/"+id, jsonBody)
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

	row := buildBrokerTenderRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	if row.Status != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Updated broker tender %s (status: %s)\n", row.ID, row.Status)
		return nil
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Updated broker tender %s\n", row.ID)
	return nil
}

func parseDoBrokerTendersUpdateOptions(cmd *cobra.Command, args []string) (doBrokerTendersUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	job, _ := cmd.Flags().GetString("job")
	broker, _ := cmd.Flags().GetString("broker")
	trucker, _ := cmd.Flags().GetString("trucker")
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
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doBrokerTendersUpdateOptions{
		BaseURL:                            baseURL,
		Token:                              token,
		JSON:                               jsonOut,
		ID:                                 args[0],
		Job:                                job,
		Broker:                             broker,
		Trucker:                            trucker,
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
	}, nil
}
