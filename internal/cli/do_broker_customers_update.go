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

type doBrokerCustomersUpdateOptions struct {
	BaseURL                            string
	Token                              string
	JSON                               bool
	ID                                 string
	ExternalAccountingBrokerCustomerID string
}

func newDoBrokerCustomersUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a broker-customer relationship",
		Long: `Update a broker-customer relationship.

At least one field is required.

Updatable fields:
  --external-accounting-broker-customer-id  External accounting broker customer ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update external accounting ID
  xbe do broker-customers update 123 --external-accounting-broker-customer-id "ACCT-42"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoBrokerCustomersUpdate,
	}
	initDoBrokerCustomersUpdateFlags(cmd)
	return cmd
}

func init() {
	doBrokerCustomersCmd.AddCommand(newDoBrokerCustomersUpdateCmd())
}

func initDoBrokerCustomersUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("external-accounting-broker-customer-id", "", "External accounting broker customer ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoBrokerCustomersUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoBrokerCustomersUpdateOptions(cmd, args)
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
	if opts.ExternalAccountingBrokerCustomerID != "" {
		attributes["external-accounting-broker-customer-id"] = opts.ExternalAccountingBrokerCustomerID
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("at least one field is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "broker-customers",
			"id":         opts.ID,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/broker-customers/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated broker customer %s\n", row.ID)
	return nil
}

func parseDoBrokerCustomersUpdateOptions(cmd *cobra.Command, args []string) (doBrokerCustomersUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	externalAccountingBrokerCustomerID, _ := cmd.Flags().GetString("external-accounting-broker-customer-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doBrokerCustomersUpdateOptions{
		BaseURL:                            baseURL,
		Token:                              token,
		JSON:                               jsonOut,
		ID:                                 args[0],
		ExternalAccountingBrokerCustomerID: externalAccountingBrokerCustomerID,
	}, nil
}
