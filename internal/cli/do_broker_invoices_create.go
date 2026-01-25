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

type doBrokerInvoicesCreateOptions struct {
	BaseURL              string
	Token                string
	JSON                 bool
	Buyer                string
	Seller               string
	Customer             string
	Broker               string
	TimeCardIDs          []string
	InvoiceDate          string
	DueOn                string
	AdjustmentAmount     string
	CurrencyCode         string
	Notes                string
	ExplicitBuyerName    string
	ExplicitBuyerAddress string
}

func newDoBrokerInvoicesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a broker invoice",
		Long: `Create a broker invoice.

Required flags:
  --customer or --buyer  Customer (Type|ID) to set as buyer
  --broker or --seller   Broker (Type|ID) to set as seller

Optional flags:
  --time-card-ids         Time card IDs (comma-separated or repeated)
  --invoice-date          Invoice date (YYYY-MM-DD)
  --due-on                Due date (YYYY-MM-DD)
  --adjustment-amount     Adjustment amount
  --currency-code         Currency code (e.g., USD)
  --notes                 Notes
  --explicit-buyer-name   Explicit buyer name override
  --explicit-buyer-address Explicit buyer address override`,
		Example: `  # Create a broker invoice from time cards
  xbe do broker-invoices create \
    --customer 123 \
    --broker 456 \
    --time-card-ids 789,790 \
    --invoice-date 2025-01-01 \
    --due-on 2025-01-31 \
    --adjustment-amount 0 \
    --currency-code USD

  # Output as JSON
  xbe do broker-invoices create \
    --customer 123 \
    --broker 456 \
    --invoice-date 2025-01-01 \
    --due-on 2025-01-31 \
    --adjustment-amount 0 \
    --currency-code USD \
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoBrokerInvoicesCreate,
	}
	initDoBrokerInvoicesCreateFlags(cmd)
	return cmd
}

func init() {
	doBrokerInvoicesCmd.AddCommand(newDoBrokerInvoicesCreateCmd())
}

func initDoBrokerInvoicesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("buyer", "", "Buyer (Type|ID, e.g., Customer|123)")
	cmd.Flags().String("seller", "", "Seller (Type|ID, e.g., Broker|456)")
	cmd.Flags().String("customer", "", "Customer ID (buyer)")
	cmd.Flags().String("broker", "", "Broker ID (seller)")
	cmd.Flags().StringSlice("time-card-ids", nil, "Time card IDs (comma-separated or repeated)")
	cmd.Flags().String("invoice-date", "", "Invoice date (YYYY-MM-DD)")
	cmd.Flags().String("due-on", "", "Due date (YYYY-MM-DD)")
	cmd.Flags().String("adjustment-amount", "", "Adjustment amount")
	cmd.Flags().String("currency-code", "", "Currency code (e.g., USD)")
	cmd.Flags().String("notes", "", "Notes")
	cmd.Flags().String("explicit-buyer-name", "", "Explicit buyer name override")
	cmd.Flags().String("explicit-buyer-address", "", "Explicit buyer address override")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoBrokerInvoicesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoBrokerInvoicesCreateOptions(cmd)
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

	buyerType := ""
	buyerID := ""
	if strings.TrimSpace(opts.Buyer) != "" {
		parsedType, parsedID, err := parseOrganization(opts.Buyer)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		buyerType = parsedType
		buyerID = parsedID
	} else if strings.TrimSpace(opts.Customer) != "" {
		buyerType = "customers"
		buyerID = strings.TrimSpace(opts.Customer)
	}

	sellerType := ""
	sellerID := ""
	if strings.TrimSpace(opts.Seller) != "" {
		parsedType, parsedID, err := parseOrganization(opts.Seller)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		sellerType = parsedType
		sellerID = parsedID
	} else if strings.TrimSpace(opts.Broker) != "" {
		sellerType = "brokers"
		sellerID = strings.TrimSpace(opts.Broker)
	}

	if buyerID == "" {
		err := fmt.Errorf("--customer or --buyer is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if sellerID == "" {
		err := fmt.Errorf("--broker or --seller is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	timeCardIDs := make([]string, 0, len(opts.TimeCardIDs))
	for _, id := range opts.TimeCardIDs {
		trimmed := strings.TrimSpace(id)
		if trimmed != "" {
			timeCardIDs = append(timeCardIDs, trimmed)
		}
	}

	attributes := map[string]any{}
	setStringAttrIfPresent(attributes, "invoice-date", strings.TrimSpace(opts.InvoiceDate))
	setStringAttrIfPresent(attributes, "due-on", strings.TrimSpace(opts.DueOn))
	setStringAttrIfPresent(attributes, "adjustment-amount", strings.TrimSpace(opts.AdjustmentAmount))
	setStringAttrIfPresent(attributes, "currency-code", strings.TrimSpace(opts.CurrencyCode))
	setStringAttrIfPresent(attributes, "notes", strings.TrimSpace(opts.Notes))
	setStringAttrIfPresent(attributes, "explicit-buyer-name", strings.TrimSpace(opts.ExplicitBuyerName))
	setStringAttrIfPresent(attributes, "explicit-buyer-address", strings.TrimSpace(opts.ExplicitBuyerAddress))

	relationships := map[string]any{
		"buyer": map[string]any{
			"data": map[string]any{
				"type": buyerType,
				"id":   buyerID,
			},
		},
		"seller": map[string]any{
			"data": map[string]any{
				"type": sellerType,
				"id":   sellerID,
			},
		},
	}

	if len(timeCardIDs) > 0 {
		dataList := make([]map[string]any, 0, len(timeCardIDs))
		for _, id := range timeCardIDs {
			dataList = append(dataList, map[string]any{
				"type": "time-cards",
				"id":   id,
			})
		}
		relationships["time-cards"] = map[string]any{
			"data": dataList,
		}
	}

	data := map[string]any{
		"type":          "broker-invoices",
		"relationships": relationships,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
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

	body, _, err := client.Post(cmd.Context(), "/v1/broker-invoices", jsonBody)
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

	row := brokerInvoiceRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created broker invoice %s\n", row.ID)
	return nil
}

func parseDoBrokerInvoicesCreateOptions(cmd *cobra.Command) (doBrokerInvoicesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	buyer, _ := cmd.Flags().GetString("buyer")
	seller, _ := cmd.Flags().GetString("seller")
	customer, _ := cmd.Flags().GetString("customer")
	broker, _ := cmd.Flags().GetString("broker")
	timeCardIDs, _ := cmd.Flags().GetStringSlice("time-card-ids")
	invoiceDate, _ := cmd.Flags().GetString("invoice-date")
	dueOn, _ := cmd.Flags().GetString("due-on")
	adjustmentAmount, _ := cmd.Flags().GetString("adjustment-amount")
	currencyCode, _ := cmd.Flags().GetString("currency-code")
	notes, _ := cmd.Flags().GetString("notes")
	explicitBuyerName, _ := cmd.Flags().GetString("explicit-buyer-name")
	explicitBuyerAddress, _ := cmd.Flags().GetString("explicit-buyer-address")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doBrokerInvoicesCreateOptions{
		BaseURL:              baseURL,
		Token:                token,
		JSON:                 jsonOut,
		Buyer:                buyer,
		Seller:               seller,
		Customer:             customer,
		Broker:               broker,
		TimeCardIDs:          timeCardIDs,
		InvoiceDate:          invoiceDate,
		DueOn:                dueOn,
		AdjustmentAmount:     adjustmentAmount,
		CurrencyCode:         currencyCode,
		Notes:                notes,
		ExplicitBuyerName:    explicitBuyerName,
		ExplicitBuyerAddress: explicitBuyerAddress,
	}, nil
}
