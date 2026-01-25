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

type doBrokerVendorsUpdateOptions struct {
	BaseURL                          string
	Token                            string
	JSON                             bool
	ID                               string
	ExternalAccountingBrokerVendorID string
}

func newDoBrokerVendorsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a broker-vendor relationship",
		Long: `Update a broker-vendor relationship.

At least one field is required.

Updatable fields:
  --external-accounting-broker-vendor-id  External accounting broker vendor ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update external accounting ID
  xbe do broker-vendors update 123 --external-accounting-broker-vendor-id "ACCT-42"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoBrokerVendorsUpdate,
	}
	initDoBrokerVendorsUpdateFlags(cmd)
	return cmd
}

func init() {
	doBrokerVendorsCmd.AddCommand(newDoBrokerVendorsUpdateCmd())
}

func initDoBrokerVendorsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("external-accounting-broker-vendor-id", "", "External accounting broker vendor ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoBrokerVendorsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoBrokerVendorsUpdateOptions(cmd, args)
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
	if opts.ExternalAccountingBrokerVendorID != "" {
		attributes["external-accounting-broker-vendor-id"] = opts.ExternalAccountingBrokerVendorID
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("at least one field is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "broker-vendors",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/broker-vendors/"+opts.ID, jsonBody)
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

	row := buildBrokerVendorRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated broker vendor %s\n", row.ID)
	return nil
}

func parseDoBrokerVendorsUpdateOptions(cmd *cobra.Command, args []string) (doBrokerVendorsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	externalAccountingBrokerVendorID, _ := cmd.Flags().GetString("external-accounting-broker-vendor-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doBrokerVendorsUpdateOptions{
		BaseURL:                          baseURL,
		Token:                            token,
		JSON:                             jsonOut,
		ID:                               args[0],
		ExternalAccountingBrokerVendorID: externalAccountingBrokerVendorID,
	}, nil
}
