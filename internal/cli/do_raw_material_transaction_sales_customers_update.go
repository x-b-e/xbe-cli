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

type doRawMaterialTransactionSalesCustomersUpdateOptions struct {
	BaseURL            string
	Token              string
	JSON               bool
	ID                 string
	RawSalesCustomerID string
	Customer           string
}

func newDoRawMaterialTransactionSalesCustomersUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a raw material transaction sales customer",
		Long: `Update an existing raw material transaction sales customer.

Provide at least one field to update. Fields not provided remain unchanged.

Arguments:
  <id>  The raw material transaction sales customer ID (required)

Flags:
  --raw-sales-customer-id  Raw sales customer identifier
  --customer               Customer ID`,
		Example: `  # Update the raw sales customer identifier
  xbe do raw-material-transaction-sales-customers update 123 --raw-sales-customer-id RAW-456

  # Update the customer
  xbe do raw-material-transaction-sales-customers update 123 --customer 789`,
		Args: cobra.ExactArgs(1),
		RunE: runDoRawMaterialTransactionSalesCustomersUpdate,
	}
	initDoRawMaterialTransactionSalesCustomersUpdateFlags(cmd)
	return cmd
}

func init() {
	doRawMaterialTransactionSalesCustomersCmd.AddCommand(newDoRawMaterialTransactionSalesCustomersUpdateCmd())
}

func initDoRawMaterialTransactionSalesCustomersUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("raw-sales-customer-id", "", "Raw sales customer identifier")
	cmd.Flags().String("customer", "", "Customer ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoRawMaterialTransactionSalesCustomersUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoRawMaterialTransactionSalesCustomersUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("raw-sales-customer-id") {
		if strings.TrimSpace(opts.RawSalesCustomerID) == "" {
			err := fmt.Errorf("--raw-sales-customer-id cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["raw-sales-customer-id"] = opts.RawSalesCustomerID
	}

	relationships := map[string]any{}
	if cmd.Flags().Changed("customer") {
		if strings.TrimSpace(opts.Customer) == "" {
			err := fmt.Errorf("--customer cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["customer"] = map[string]any{
			"data": map[string]any{
				"type": "customers",
				"id":   opts.Customer,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("at least one field must be specified for update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestData := map[string]any{
		"type": "raw-material-transaction-sales-customers",
		"id":   opts.ID,
	}

	if len(attributes) > 0 {
		requestData["attributes"] = attributes
	}
	if len(relationships) > 0 {
		requestData["relationships"] = relationships
	}

	requestBody := map[string]any{
		"data": requestData,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/raw-material-transaction-sales-customers/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated raw material transaction sales customer %s\n", row.ID)
	return nil
}

func parseDoRawMaterialTransactionSalesCustomersUpdateOptions(cmd *cobra.Command, args []string) (doRawMaterialTransactionSalesCustomersUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	rawSalesCustomerID, _ := cmd.Flags().GetString("raw-sales-customer-id")
	customer, _ := cmd.Flags().GetString("customer")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doRawMaterialTransactionSalesCustomersUpdateOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		ID:                 args[0],
		RawSalesCustomerID: rawSalesCustomerID,
		Customer:           customer,
	}, nil
}
