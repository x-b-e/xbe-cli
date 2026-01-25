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

type doBrokerCustomersCreateOptions struct {
	BaseURL                            string
	Token                              string
	JSON                               bool
	Broker                             string
	Customer                           string
	ExternalAccountingBrokerCustomerID string
}

func newDoBrokerCustomersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a broker-customer relationship",
		Long: `Create a broker-customer trading partner relationship.

Required:
  --broker    Broker ID
  --customer  Customer ID

Optional:
  --external-accounting-broker-customer-id  External accounting broker customer ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a broker-customer relationship
  xbe do broker-customers create --broker 123 --customer 456

  # Create with external accounting ID
  xbe do broker-customers create --broker 123 --customer 456 \
    --external-accounting-broker-customer-id "ACCT-42"`,
		Args: cobra.NoArgs,
		RunE: runDoBrokerCustomersCreate,
	}
	initDoBrokerCustomersCreateFlags(cmd)
	return cmd
}

func init() {
	doBrokerCustomersCmd.AddCommand(newDoBrokerCustomersCreateCmd())
}

func initDoBrokerCustomersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("broker", "", "Broker ID")
	cmd.Flags().String("customer", "", "Customer ID")
	cmd.Flags().String("external-accounting-broker-customer-id", "", "External accounting broker customer ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("broker")
	_ = cmd.MarkFlagRequired("customer")
}

func runDoBrokerCustomersCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoBrokerCustomersCreateOptions(cmd)
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

	if opts.Broker == "" {
		err := fmt.Errorf("--broker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.Customer == "" {
		err := fmt.Errorf("--customer is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.ExternalAccountingBrokerCustomerID != "" {
		attributes["external-accounting-broker-customer-id"] = opts.ExternalAccountingBrokerCustomerID
	}

	relationships := map[string]any{
		"organization": map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.Broker,
			},
		},
		"partner": map[string]any{
			"data": map[string]any{
				"type": "customers",
				"id":   opts.Customer,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "broker-customers",
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

	body, _, err := client.Post(cmd.Context(), "/v1/broker-customers", jsonBody)
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

	row := buildBrokerCustomerRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created broker customer %s\n", row.ID)
	return nil
}

func parseDoBrokerCustomersCreateOptions(cmd *cobra.Command) (doBrokerCustomersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	broker, _ := cmd.Flags().GetString("broker")
	customer, _ := cmd.Flags().GetString("customer")
	externalAccountingBrokerCustomerID, _ := cmd.Flags().GetString("external-accounting-broker-customer-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doBrokerCustomersCreateOptions{
		BaseURL:                            baseURL,
		Token:                              token,
		JSON:                               jsonOut,
		Broker:                             broker,
		Customer:                           customer,
		ExternalAccountingBrokerCustomerID: externalAccountingBrokerCustomerID,
	}, nil
}

func buildBrokerCustomerRowFromSingle(resp jsonAPISingleResponse) brokerCustomerRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	row := brokerCustomerRow{
		ID:                                 resp.Data.ID,
		ExternalAccountingBrokerCustomerID: stringAttr(resp.Data.Attributes, "external-accounting-broker-customer-id"),
		TradingPartnerType:                 stringAttr(resp.Data.Attributes, "trading-partner-type"),
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
		row.BrokerName = brokerCustomerNameFromIncluded(rel.Data, included)
	} else if rel, ok := resp.Data.Relationships["organization"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
		row.BrokerName = brokerCustomerNameFromIncluded(rel.Data, included)
	}

	if rel, ok := resp.Data.Relationships["customer"]; ok && rel.Data != nil {
		row.CustomerID = rel.Data.ID
		row.CustomerName = brokerCustomerNameFromIncluded(rel.Data, included)
	} else if rel, ok := resp.Data.Relationships["partner"]; ok && rel.Data != nil {
		row.CustomerID = rel.Data.ID
		row.CustomerName = brokerCustomerNameFromIncluded(rel.Data, included)
	}

	return row
}
