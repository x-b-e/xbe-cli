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

type doCustomerTruckersCreateOptions struct {
	BaseURL  string
	Token    string
	JSON     bool
	Customer string
	Trucker  string
}

func newDoCustomerTruckersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a customer trucker link",
		Long: `Create a customer trucker link.

Required flags:
  --customer  Customer ID (required)
  --trucker   Trucker ID (required)

Note: The customer and trucker must belong to the same broker.`,
		Example: `  # Link a customer to a trucker
  xbe do customer-truckers create --customer 123 --trucker 456

  # Output as JSON
  xbe do customer-truckers create --customer 123 --trucker 456 --json`,
		Args: cobra.NoArgs,
		RunE: runDoCustomerTruckersCreate,
	}
	initDoCustomerTruckersCreateFlags(cmd)
	return cmd
}

func init() {
	doCustomerTruckersCmd.AddCommand(newDoCustomerTruckersCreateCmd())
}

func initDoCustomerTruckersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("customer", "", "Customer ID (required)")
	cmd.Flags().String("trucker", "", "Trucker ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCustomerTruckersCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoCustomerTruckersCreateOptions(cmd)
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

	if opts.Customer == "" {
		err := fmt.Errorf("--customer is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.Trucker == "" {
		err := fmt.Errorf("--trucker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"customer": map[string]any{
			"data": map[string]any{
				"type": "customers",
				"id":   opts.Customer,
			},
		},
		"trucker": map[string]any{
			"data": map[string]any{
				"type": "truckers",
				"id":   opts.Trucker,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "customer-truckers",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/customer-truckers", jsonBody)
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

	row := customerTruckerRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created customer trucker %s\n", row.ID)
	return nil
}

func parseDoCustomerTruckersCreateOptions(cmd *cobra.Command) (doCustomerTruckersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	customer, _ := cmd.Flags().GetString("customer")
	trucker, _ := cmd.Flags().GetString("trucker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCustomerTruckersCreateOptions{
		BaseURL:  baseURL,
		Token:    token,
		JSON:     jsonOut,
		Customer: customer,
		Trucker:  trucker,
	}, nil
}
