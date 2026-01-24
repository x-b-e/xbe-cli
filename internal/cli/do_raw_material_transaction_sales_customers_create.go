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

type doRawMaterialTransactionSalesCustomersCreateOptions struct {
	BaseURL            string
	Token              string
	JSON               bool
	RawSalesCustomerID string
	Customer           string
}

func newDoRawMaterialTransactionSalesCustomersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a raw material transaction sales customer",
		Long: `Create a raw material transaction sales customer.

Required flags:
  --raw-sales-customer-id  Raw sales customer identifier (required)
  --customer               Customer ID (required)

Note: The broker is derived from the customer and must match existing broker rules.`,
		Example: `  # Create a raw material transaction sales customer
  xbe do raw-material-transaction-sales-customers create \\
    --raw-sales-customer-id RAW-123 \\
    --customer 456

  # Output as JSON
  xbe do raw-material-transaction-sales-customers create --raw-sales-customer-id RAW-123 --customer 456 --json`,
		Args: cobra.NoArgs,
		RunE: runDoRawMaterialTransactionSalesCustomersCreate,
	}
	initDoRawMaterialTransactionSalesCustomersCreateFlags(cmd)
	return cmd
}

func init() {
	doRawMaterialTransactionSalesCustomersCmd.AddCommand(newDoRawMaterialTransactionSalesCustomersCreateCmd())
}

func initDoRawMaterialTransactionSalesCustomersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("raw-sales-customer-id", "", "Raw sales customer identifier (required)")
	cmd.Flags().String("customer", "", "Customer ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("raw-sales-customer-id")
	_ = cmd.MarkFlagRequired("customer")
}

func runDoRawMaterialTransactionSalesCustomersCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoRawMaterialTransactionSalesCustomersCreateOptions(cmd)
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

	if strings.TrimSpace(opts.RawSalesCustomerID) == "" {
		err := fmt.Errorf("--raw-sales-customer-id is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.Customer) == "" {
		err := fmt.Errorf("--customer is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"raw-sales-customer-id": opts.RawSalesCustomerID,
	}

	relationships := map[string]any{
		"customer": map[string]any{
			"data": map[string]any{
				"type": "customers",
				"id":   opts.Customer,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "raw-material-transaction-sales-customers",
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

	body, _, err := client.Post(cmd.Context(), "/v1/raw-material-transaction-sales-customers", jsonBody)
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

	row := rawMaterialTransactionSalesCustomerRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created raw material transaction sales customer %s\n", row.ID)
	return nil
}

func parseDoRawMaterialTransactionSalesCustomersCreateOptions(cmd *cobra.Command) (doRawMaterialTransactionSalesCustomersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	rawSalesCustomerID, _ := cmd.Flags().GetString("raw-sales-customer-id")
	customer, _ := cmd.Flags().GetString("customer")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doRawMaterialTransactionSalesCustomersCreateOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		RawSalesCustomerID: rawSalesCustomerID,
		Customer:           customer,
	}, nil
}
