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

type doBrokerInvoicesUpdateOptions struct {
	BaseURL              string
	Token                string
	JSON                 bool
	ID                   string
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

func newDoBrokerInvoicesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a broker invoice",
		Long: `Update a broker invoice.

Optional flags:
  --buyer                  Buyer (Type|ID, e.g., Customer|123)
  --seller                 Seller (Type|ID, e.g., Broker|456)
  --customer               Customer ID (buyer)
  --broker                 Broker ID (seller)
  --time-card-ids          Time card IDs (comma-separated or repeated)
  --invoice-date           Invoice date (YYYY-MM-DD)
  --due-on                 Due date (YYYY-MM-DD)
  --adjustment-amount      Adjustment amount
  --currency-code          Currency code (e.g., USD)
  --notes                  Notes
  --explicit-buyer-name    Explicit buyer name override
  --explicit-buyer-address Explicit buyer address override`,
		Example: `  # Update notes
  xbe do broker-invoices update 123 --notes "Updated"

  # Update invoice dates
  xbe do broker-invoices update 123 --invoice-date 2025-02-01 --due-on 2025-02-28`,
		Args: cobra.ExactArgs(1),
		RunE: runDoBrokerInvoicesUpdate,
	}
	initDoBrokerInvoicesUpdateFlags(cmd)
	return cmd
}

func init() {
	doBrokerInvoicesCmd.AddCommand(newDoBrokerInvoicesUpdateCmd())
}

func initDoBrokerInvoicesUpdateFlags(cmd *cobra.Command) {
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

func runDoBrokerInvoicesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoBrokerInvoicesUpdateOptions(cmd, args)
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
	relationships := map[string]any{}

	if cmd.Flags().Changed("invoice-date") {
		attributes["invoice-date"] = opts.InvoiceDate
	}
	if cmd.Flags().Changed("due-on") {
		attributes["due-on"] = opts.DueOn
	}
	if cmd.Flags().Changed("adjustment-amount") {
		attributes["adjustment-amount"] = opts.AdjustmentAmount
	}
	if cmd.Flags().Changed("currency-code") {
		attributes["currency-code"] = opts.CurrencyCode
	}
	if cmd.Flags().Changed("notes") {
		attributes["notes"] = opts.Notes
	}
	if cmd.Flags().Changed("explicit-buyer-name") {
		attributes["explicit-buyer-name"] = opts.ExplicitBuyerName
	}
	if cmd.Flags().Changed("explicit-buyer-address") {
		attributes["explicit-buyer-address"] = opts.ExplicitBuyerAddress
	}

	buyerType := ""
	buyerID := ""
	if cmd.Flags().Changed("buyer") {
		parsedType, parsedID, err := parseOrganization(opts.Buyer)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		buyerType = parsedType
		buyerID = parsedID
	} else if cmd.Flags().Changed("customer") {
		buyerType = "customers"
		buyerID = strings.TrimSpace(opts.Customer)
	}

	sellerType := ""
	sellerID := ""
	if cmd.Flags().Changed("seller") {
		parsedType, parsedID, err := parseOrganization(opts.Seller)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		sellerType = parsedType
		sellerID = parsedID
	} else if cmd.Flags().Changed("broker") {
		sellerType = "brokers"
		sellerID = strings.TrimSpace(opts.Broker)
	}

	if buyerID != "" {
		relationships["buyer"] = map[string]any{
			"data": map[string]any{
				"type": buyerType,
				"id":   buyerID,
			},
		}
	}
	if sellerID != "" {
		relationships["seller"] = map[string]any{
			"data": map[string]any{
				"type": sellerType,
				"id":   sellerID,
			},
		}
	}

	if cmd.Flags().Changed("time-card-ids") {
		trimmed := make([]string, 0, len(opts.TimeCardIDs))
		for _, id := range opts.TimeCardIDs {
			value := strings.TrimSpace(id)
			if value != "" {
				trimmed = append(trimmed, value)
			}
		}
		dataList := make([]map[string]any, 0, len(trimmed))
		for _, id := range trimmed {
			dataList = append(dataList, map[string]any{
				"type": "time-cards",
				"id":   id,
			})
		}
		relationships["time-cards"] = map[string]any{
			"data": dataList,
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes or relationships to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "broker-invoices",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
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

	body, _, err := client.Patch(cmd.Context(), "/v1/broker-invoices/"+opts.ID, jsonBody)
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

	if opts.JSON {
		row := brokerInvoiceRowFromSingle(resp)
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated broker invoice %s\n", resp.Data.ID)
	return nil
}

func parseDoBrokerInvoicesUpdateOptions(cmd *cobra.Command, args []string) (doBrokerInvoicesUpdateOptions, error) {
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

	return doBrokerInvoicesUpdateOptions{
		BaseURL:              baseURL,
		Token:                token,
		JSON:                 jsonOut,
		ID:                   strings.TrimSpace(args[0]),
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
