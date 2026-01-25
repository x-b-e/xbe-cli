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

type customerTruckersShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type customerTruckerDetails struct {
	ID           string `json:"id"`
	CustomerID   string `json:"customer_id,omitempty"`
	CustomerName string `json:"customer_name,omitempty"`
	TruckerID    string `json:"trucker_id,omitempty"`
	TruckerName  string `json:"trucker_name,omitempty"`
}

func newCustomerTruckersShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show customer trucker details",
		Long: `Show the full details of a customer trucker link.

Output Fields:
  ID        Customer trucker link identifier
  Customer  Customer name or ID
  Trucker   Trucker name or ID

Arguments:
  <id>  The customer trucker ID (required).`,
		Example: `  # Show customer trucker details
  xbe view customer-truckers show 123

  # Output as JSON
  xbe view customer-truckers show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runCustomerTruckersShow,
	}
	initCustomerTruckersShowFlags(cmd)
	return cmd
}

func init() {
	customerTruckersCmd.AddCommand(newCustomerTruckersShowCmd())
}

func initCustomerTruckersShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCustomerTruckersShow(cmd *cobra.Command, args []string) error {
	opts, err := parseCustomerTruckersShowOptions(cmd)
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
		return fmt.Errorf("customer trucker id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[customer-truckers]", "customer,trucker")
	query.Set("include", "customer,trucker")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[truckers]", "company-name")

	body, _, err := client.Get(cmd.Context(), "/v1/customer-truckers/"+id, query)
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

	details := buildCustomerTruckerDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderCustomerTruckerDetails(cmd, details)
}

func parseCustomerTruckersShowOptions(cmd *cobra.Command) (customerTruckersShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return customerTruckersShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildCustomerTruckerDetails(resp jsonAPISingleResponse) customerTruckerDetails {
	row := customerTruckerRowFromSingle(resp)
	return customerTruckerDetails{
		ID:           row.ID,
		CustomerID:   row.CustomerID,
		CustomerName: row.CustomerName,
		TruckerID:    row.TruckerID,
		TruckerName:  row.TruckerName,
	}
}

func renderCustomerTruckerDetails(cmd *cobra.Command, details customerTruckerDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.CustomerID != "" || details.CustomerName != "" {
		fmt.Fprintf(out, "Customer: %s\n", formatRelated(details.CustomerName, details.CustomerID))
	}
	if details.TruckerID != "" || details.TruckerName != "" {
		fmt.Fprintf(out, "Trucker: %s\n", formatRelated(details.TruckerName, details.TruckerID))
	}

	return nil
}
