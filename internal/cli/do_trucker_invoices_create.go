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

type doTruckerInvoicesCreateOptions struct {
	BaseURL              string
	Token                string
	JSON                 bool
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

func newDoTruckerInvoicesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a trucker invoice",
		Long: `Create a trucker invoice.

Required flags:
  --buyer-type         Buyer type (brokers)
  --buyer              Buyer ID
  --seller-type        Seller type (truckers)
  --seller             Seller ID
  --invoice-date       Invoice date (YYYY-MM-DD)
  --due-on             Due date (YYYY-MM-DD)
  --adjustment-amount  Adjustment amount
  --currency-code      Currency code (USD)

Optional flags:
  --notes                 Notes
  --explicit-buyer-name   Explicit buyer name override
  --explicit-buyer-address Explicit buyer address override
  --time-cards            Time card IDs (comma-separated)`,
		Example: `  # Create a trucker invoice
  xbe do trucker-invoices create \\
    --buyer-type brokers --buyer 123 \\
    --seller-type truckers --seller 456 \\
    --invoice-date 2025-01-01 --due-on 2025-01-10 \\
    --adjustment-amount 0.00 --currency-code USD

  # Create with notes and time cards
  xbe do trucker-invoices create \\
    --buyer-type brokers --buyer 123 \\
    --seller-type truckers --seller 456 \\
    --invoice-date 2025-01-01 --due-on 2025-01-10 \\
    --adjustment-amount 10.00 --currency-code USD \\
    --notes \"Manual adjustment\" \\
    --time-cards 111,222

  # Get JSON output
  xbe do trucker-invoices create --buyer-type brokers --buyer 123 --seller-type truckers --seller 456 \\
    --invoice-date 2025-01-01 --due-on 2025-01-10 --adjustment-amount 0.00 --currency-code USD --json`,
		Args: cobra.NoArgs,
		RunE: runDoTruckerInvoicesCreate,
	}
	initDoTruckerInvoicesCreateFlags(cmd)
	return cmd
}

func init() {
	doTruckerInvoicesCmd.AddCommand(newDoTruckerInvoicesCreateCmd())
}

func initDoTruckerInvoicesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("buyer-type", "", "Buyer type (brokers)")
	cmd.Flags().String("buyer", "", "Buyer ID")
	cmd.Flags().String("seller-type", "", "Seller type (truckers)")
	cmd.Flags().String("seller", "", "Seller ID")
	cmd.Flags().String("invoice-date", "", "Invoice date (YYYY-MM-DD)")
	cmd.Flags().String("due-on", "", "Due date (YYYY-MM-DD)")
	cmd.Flags().String("adjustment-amount", "", "Adjustment amount")
	cmd.Flags().String("currency-code", "", "Currency code (USD)")
	cmd.Flags().String("notes", "", "Notes")
	cmd.Flags().String("explicit-buyer-name", "", "Explicit buyer name override")
	cmd.Flags().String("explicit-buyer-address", "", "Explicit buyer address override")
	cmd.Flags().String("time-cards", "", "Time card IDs (comma-separated)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("buyer-type")
	cmd.MarkFlagRequired("buyer")
	cmd.MarkFlagRequired("seller-type")
	cmd.MarkFlagRequired("seller")
	cmd.MarkFlagRequired("invoice-date")
	cmd.MarkFlagRequired("due-on")
	cmd.MarkFlagRequired("adjustment-amount")
	cmd.MarkFlagRequired("currency-code")
}

func runDoTruckerInvoicesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTruckerInvoicesCreateOptions(cmd)
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

	if strings.TrimSpace(opts.BuyerType) == "" {
		err := fmt.Errorf("--buyer-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.BuyerID) == "" {
		err := fmt.Errorf("--buyer is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.SellerType) == "" {
		err := fmt.Errorf("--seller-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.SellerID) == "" {
		err := fmt.Errorf("--seller is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.InvoiceDate) == "" {
		err := fmt.Errorf("--invoice-date is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.DueOn) == "" {
		err := fmt.Errorf("--due-on is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.AdjustmentAmount) == "" {
		err := fmt.Errorf("--adjustment-amount is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.CurrencyCode) == "" {
		err := fmt.Errorf("--currency-code is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"invoice-date":      opts.InvoiceDate,
		"due-on":            opts.DueOn,
		"adjustment-amount": opts.AdjustmentAmount,
		"currency-code":     opts.CurrencyCode,
	}
	if opts.Notes != "" {
		attributes["notes"] = opts.Notes
	}
	if opts.ExplicitBuyerName != "" {
		attributes["explicit-buyer-name"] = opts.ExplicitBuyerName
	}
	if opts.ExplicitBuyerAddress != "" {
		attributes["explicit-buyer-address"] = opts.ExplicitBuyerAddress
	}

	relationships := map[string]any{
		"buyer": map[string]any{
			"data": map[string]any{
				"type": opts.BuyerType,
				"id":   opts.BuyerID,
			},
		},
		"seller": map[string]any{
			"data": map[string]any{
				"type": opts.SellerType,
				"id":   opts.SellerID,
			},
		},
	}

	if strings.TrimSpace(opts.TimeCards) != "" {
		ids := splitCommaList(opts.TimeCards)
		if len(ids) > 0 {
			data := make([]map[string]any, 0, len(ids))
			for _, id := range ids {
				data = append(data, map[string]any{
					"type": "time-cards",
					"id":   id,
				})
			}
			relationships["time-cards"] = map[string]any{"data": data}
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "trucker-invoices",
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

	body, _, err := client.Post(cmd.Context(), "/v1/trucker-invoices", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created trucker invoice %s\n", row.ID)
	return nil
}

func parseDoTruckerInvoicesCreateOptions(cmd *cobra.Command) (doTruckerInvoicesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	buyerType, _ := cmd.Flags().GetString("buyer-type")
	buyerID, _ := cmd.Flags().GetString("buyer")
	sellerType, _ := cmd.Flags().GetString("seller-type")
	sellerID, _ := cmd.Flags().GetString("seller")
	invoiceDate, _ := cmd.Flags().GetString("invoice-date")
	dueOn, _ := cmd.Flags().GetString("due-on")
	adjustmentAmount, _ := cmd.Flags().GetString("adjustment-amount")
	currencyCode, _ := cmd.Flags().GetString("currency-code")
	notes, _ := cmd.Flags().GetString("notes")
	explicitBuyerName, _ := cmd.Flags().GetString("explicit-buyer-name")
	explicitBuyerAddress, _ := cmd.Flags().GetString("explicit-buyer-address")
	timeCards, _ := cmd.Flags().GetString("time-cards")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTruckerInvoicesCreateOptions{
		BaseURL:              baseURL,
		Token:                token,
		JSON:                 jsonOut,
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
