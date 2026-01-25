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

type rawMaterialTransactionSalesCustomersShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type rawMaterialTransactionSalesCustomerDetails struct {
	ID                 string `json:"id"`
	RawSalesCustomerID string `json:"raw_sales_customer_id,omitempty"`
	CustomerID         string `json:"customer_id,omitempty"`
	CustomerName       string `json:"customer_name,omitempty"`
	BrokerID           string `json:"broker_id,omitempty"`
	BrokerName         string `json:"broker_name,omitempty"`
}

func newRawMaterialTransactionSalesCustomersShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show raw material transaction sales customer details",
		Long: `Show the full details of a raw material transaction sales customer.

Output Fields:
  ID               Raw material transaction sales customer identifier
  Raw Sales ID     Raw sales customer identifier
  Customer         Customer name or ID
  Broker           Broker name or ID

Arguments:
  <id>  The raw material transaction sales customer ID (required).`,
		Example: `  # Show raw material transaction sales customer details
  xbe view raw-material-transaction-sales-customers show 123

  # Output as JSON
  xbe view raw-material-transaction-sales-customers show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runRawMaterialTransactionSalesCustomersShow,
	}
	initRawMaterialTransactionSalesCustomersShowFlags(cmd)
	return cmd
}

func init() {
	rawMaterialTransactionSalesCustomersCmd.AddCommand(newRawMaterialTransactionSalesCustomersShowCmd())
}

func initRawMaterialTransactionSalesCustomersShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runRawMaterialTransactionSalesCustomersShow(cmd *cobra.Command, args []string) error {
	opts, err := parseRawMaterialTransactionSalesCustomersShowOptions(cmd)
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
		return fmt.Errorf("raw material transaction sales customer id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[raw-material-transaction-sales-customers]", "raw-sales-customer-id,customer,broker")
	query.Set("include", "customer,broker")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[brokers]", "company-name")

	body, _, err := client.Get(cmd.Context(), "/v1/raw-material-transaction-sales-customers/"+id, query)
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

	details := buildRawMaterialTransactionSalesCustomerDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderRawMaterialTransactionSalesCustomerDetails(cmd, details)
}

func parseRawMaterialTransactionSalesCustomersShowOptions(cmd *cobra.Command) (rawMaterialTransactionSalesCustomersShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return rawMaterialTransactionSalesCustomersShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildRawMaterialTransactionSalesCustomerDetails(resp jsonAPISingleResponse) rawMaterialTransactionSalesCustomerDetails {
	row := rawMaterialTransactionSalesCustomerRowFromSingle(resp)
	return rawMaterialTransactionSalesCustomerDetails{
		ID:                 row.ID,
		RawSalesCustomerID: row.RawSalesCustomerID,
		CustomerID:         row.CustomerID,
		CustomerName:       row.CustomerName,
		BrokerID:           row.BrokerID,
		BrokerName:         row.BrokerName,
	}
}

func renderRawMaterialTransactionSalesCustomerDetails(cmd *cobra.Command, details rawMaterialTransactionSalesCustomerDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.RawSalesCustomerID != "" {
		fmt.Fprintf(out, "Raw Sales Customer ID: %s\n", details.RawSalesCustomerID)
	}
	if details.CustomerID != "" || details.CustomerName != "" {
		fmt.Fprintf(out, "Customer: %s\n", formatRelated(details.CustomerName, details.CustomerID))
	}
	if details.BrokerID != "" || details.BrokerName != "" {
		fmt.Fprintf(out, "Broker: %s\n", formatRelated(details.BrokerName, details.BrokerID))
	}

	return nil
}
