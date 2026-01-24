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

type doCustomerVendorsUpdateOptions struct {
	BaseURL                            string
	Token                              string
	JSON                               bool
	ID                                 string
	ExternalAccountingCustomerVendorID string
}

func newDoCustomerVendorsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a customer-vendor relationship",
		Long: `Update a customer-vendor relationship.

At least one field is required.

Updatable fields:
  --external-accounting-customer-vendor-id  External accounting customer vendor ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update external accounting ID
  xbe do customer-vendors update 123 --external-accounting-customer-vendor-id "ACCT-42"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoCustomerVendorsUpdate,
	}
	initDoCustomerVendorsUpdateFlags(cmd)
	return cmd
}

func init() {
	doCustomerVendorsCmd.AddCommand(newDoCustomerVendorsUpdateCmd())
}

func initDoCustomerVendorsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("external-accounting-customer-vendor-id", "", "External accounting customer vendor ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCustomerVendorsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoCustomerVendorsUpdateOptions(cmd, args)
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
	if opts.ExternalAccountingCustomerVendorID != "" {
		attributes["external-accounting-customer-vendor-id"] = opts.ExternalAccountingCustomerVendorID
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("at least one field is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "customer-vendors",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/customer-vendors/"+opts.ID, jsonBody)
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

	row := buildCustomerVendorRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated customer vendor %s\n", row.ID)
	return nil
}

func parseDoCustomerVendorsUpdateOptions(cmd *cobra.Command, args []string) (doCustomerVendorsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	externalAccountingCustomerVendorID, _ := cmd.Flags().GetString("external-accounting-customer-vendor-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCustomerVendorsUpdateOptions{
		BaseURL:                            baseURL,
		Token:                              token,
		JSON:                               jsonOut,
		ID:                                 args[0],
		ExternalAccountingCustomerVendorID: externalAccountingCustomerVendorID,
	}, nil
}
