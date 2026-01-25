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

type doTruckerInvoicesUpdateOptions struct {
	BaseURL              string
	Token                string
	JSON                 bool
	ID                   string
	InvoiceDate          string
	DueOn                string
	AdjustmentAmount     string
	CurrencyCode         string
	Notes                string
	ExplicitBuyerName    string
	ExplicitBuyerAddress string
	BuyerType            string
	BuyerID              string
	SellerType           string
	SellerID             string
	TimeCards            string
}

func newDoTruckerInvoicesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a trucker invoice",
		Long: `Update an existing trucker invoice.

Only the fields you specify will be updated. Fields not provided will remain unchanged.

Arguments:
  <id>    The trucker invoice ID (required)

Flags:
  --invoice-date           Update invoice date (YYYY-MM-DD)
  --due-on                 Update due date (YYYY-MM-DD)
  --adjustment-amount      Update adjustment amount
  --currency-code          Update currency code (USD)
  --notes                  Update notes (use empty string to clear)
  --explicit-buyer-name    Update explicit buyer name
  --explicit-buyer-address Update explicit buyer address
  --buyer-type             Update buyer type (brokers)
  --buyer                  Update buyer ID
  --seller-type            Update seller type (truckers)
  --seller                 Update seller ID
  --time-cards             Update time card IDs (comma-separated, empty to clear)`,
		Example: `  # Update notes
  xbe do trucker-invoices update 123 --notes "Updated notes"

  # Update due date
  xbe do trucker-invoices update 123 --due-on 2025-02-01

  # Update time cards
  xbe do trucker-invoices update 123 --time-cards 111,222`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTruckerInvoicesUpdate,
	}
	initDoTruckerInvoicesUpdateFlags(cmd)
	return cmd
}

func init() {
	doTruckerInvoicesCmd.AddCommand(newDoTruckerInvoicesUpdateCmd())
}

func initDoTruckerInvoicesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("invoice-date", "", "Invoice date (YYYY-MM-DD)")
	cmd.Flags().String("due-on", "", "Due date (YYYY-MM-DD)")
	cmd.Flags().String("adjustment-amount", "", "Adjustment amount")
	cmd.Flags().String("currency-code", "", "Currency code (USD)")
	cmd.Flags().String("notes", "", "Notes")
	cmd.Flags().String("explicit-buyer-name", "", "Explicit buyer name override")
	cmd.Flags().String("explicit-buyer-address", "", "Explicit buyer address override")
	cmd.Flags().String("buyer-type", "", "Buyer type (brokers)")
	cmd.Flags().String("buyer", "", "Buyer ID")
	cmd.Flags().String("seller-type", "", "Seller type (truckers)")
	cmd.Flags().String("seller", "", "Seller ID")
	cmd.Flags().String("time-cards", "", "Time card IDs (comma-separated)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTruckerInvoicesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTruckerInvoicesUpdateOptions(cmd, args)
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
	hasChanges := false

	if cmd.Flags().Changed("invoice-date") {
		attributes["invoice-date"] = opts.InvoiceDate
		hasChanges = true
	}
	if cmd.Flags().Changed("due-on") {
		attributes["due-on"] = opts.DueOn
		hasChanges = true
	}
	if cmd.Flags().Changed("adjustment-amount") {
		attributes["adjustment-amount"] = opts.AdjustmentAmount
		hasChanges = true
	}
	if cmd.Flags().Changed("currency-code") {
		attributes["currency-code"] = opts.CurrencyCode
		hasChanges = true
	}
	if cmd.Flags().Changed("notes") {
		attributes["notes"] = opts.Notes
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-buyer-name") {
		attributes["explicit-buyer-name"] = opts.ExplicitBuyerName
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-buyer-address") {
		attributes["explicit-buyer-address"] = opts.ExplicitBuyerAddress
		hasChanges = true
	}

	if cmd.Flags().Changed("buyer-type") || cmd.Flags().Changed("buyer") {
		if strings.TrimSpace(opts.BuyerType) == "" || strings.TrimSpace(opts.BuyerID) == "" {
			return fmt.Errorf("--buyer-type and --buyer are required together")
		}
		relationships["buyer"] = map[string]any{
			"data": map[string]any{
				"type": opts.BuyerType,
				"id":   opts.BuyerID,
			},
		}
		hasChanges = true
	}

	if cmd.Flags().Changed("seller-type") || cmd.Flags().Changed("seller") {
		if strings.TrimSpace(opts.SellerType) == "" || strings.TrimSpace(opts.SellerID) == "" {
			return fmt.Errorf("--seller-type and --seller are required together")
		}
		relationships["seller"] = map[string]any{
			"data": map[string]any{
				"type": opts.SellerType,
				"id":   opts.SellerID,
			},
		}
		hasChanges = true
	}

	setToManyRelationship := func(flagName, key, resourceType, raw string) {
		if !cmd.Flags().Changed(flagName) {
			return
		}
		if strings.TrimSpace(raw) == "" {
			relationships[key] = map[string]any{"data": []any{}}
			hasChanges = true
			return
		}
		ids := splitCommaList(raw)
		data := make([]map[string]any, 0, len(ids))
		for _, id := range ids {
			data = append(data, map[string]any{
				"type": resourceType,
				"id":   id,
			})
		}
		relationships[key] = map[string]any{"data": data}
		hasChanges = true
	}

	setToManyRelationship("time-cards", "time-cards", "time-cards", opts.TimeCards)

	if !hasChanges {
		return fmt.Errorf("no fields to update")
	}

	data := map[string]any{
		"id":   opts.ID,
		"type": "trucker-invoices",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/trucker-invoices/"+opts.ID, jsonBody)
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

	row := buildTruckerInvoiceRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated trucker invoice %s\n", row.ID)
	return nil
}

func parseDoTruckerInvoicesUpdateOptions(cmd *cobra.Command, args []string) (doTruckerInvoicesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	invoiceDate, _ := cmd.Flags().GetString("invoice-date")
	dueOn, _ := cmd.Flags().GetString("due-on")
	adjustmentAmount, _ := cmd.Flags().GetString("adjustment-amount")
	currencyCode, _ := cmd.Flags().GetString("currency-code")
	notes, _ := cmd.Flags().GetString("notes")
	explicitBuyerName, _ := cmd.Flags().GetString("explicit-buyer-name")
	explicitBuyerAddress, _ := cmd.Flags().GetString("explicit-buyer-address")
	buyerType, _ := cmd.Flags().GetString("buyer-type")
	buyerID, _ := cmd.Flags().GetString("buyer")
	sellerType, _ := cmd.Flags().GetString("seller-type")
	sellerID, _ := cmd.Flags().GetString("seller")
	timeCards, _ := cmd.Flags().GetString("time-cards")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	id := strings.TrimSpace(args[0])
	if id == "" {
		return doTruckerInvoicesUpdateOptions{}, fmt.Errorf("trucker invoice id is required")
	}

	return doTruckerInvoicesUpdateOptions{
		BaseURL:              baseURL,
		Token:                token,
		JSON:                 jsonOut,
		ID:                   id,
		InvoiceDate:          invoiceDate,
		DueOn:                dueOn,
		AdjustmentAmount:     adjustmentAmount,
		CurrencyCode:         currencyCode,
		Notes:                notes,
		ExplicitBuyerName:    explicitBuyerName,
		ExplicitBuyerAddress: explicitBuyerAddress,
		BuyerType:            buyerType,
		BuyerID:              buyerID,
		SellerType:           sellerType,
		SellerID:             sellerID,
		TimeCards:            timeCards,
	}, nil
}
