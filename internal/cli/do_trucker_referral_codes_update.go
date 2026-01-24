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

type doTruckerReferralCodesUpdateOptions struct {
	BaseURL  string
	Token    string
	JSON     bool
	ID       string
	BrokerID string
	Code     string
	Value    string
}

func newDoTruckerReferralCodesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a trucker referral code",
		Long: `Update a trucker referral code.

Arguments:
  <id>  The trucker referral code ID (required)

Optional flags:
  --code    Referral code
  --value   Referral value (numeric)
  --broker  Broker ID

Note: Codes are normalized to uppercase and whitespace is removed.`,
		Example: `  # Update a trucker referral code
  xbe do trucker-referral-codes update 123 --code "REF-456"

  # Update referral value
  xbe do trucker-referral-codes update 123 --value 75`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTruckerReferralCodesUpdate,
	}
	initDoTruckerReferralCodesUpdateFlags(cmd)
	return cmd
}

func init() {
	doTruckerReferralCodesCmd.AddCommand(newDoTruckerReferralCodesUpdateCmd())
}

func initDoTruckerReferralCodesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("code", "", "Referral code")
	cmd.Flags().String("value", "", "Referral value (numeric)")
	cmd.Flags().String("broker", "", "Broker ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTruckerReferralCodesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTruckerReferralCodesUpdateOptions(cmd, args)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	if strings.TrimSpace(opts.ID) == "" {
		return fmt.Errorf("trucker referral code id is required")
	}

	attributes := map[string]any{}
	relationships := map[string]any{}

	if cmd.Flags().Changed("code") {
		if strings.TrimSpace(opts.Code) == "" {
			return fmt.Errorf("--code cannot be empty")
		}
		attributes["code"] = opts.Code
	}
	if cmd.Flags().Changed("value") {
		if strings.TrimSpace(opts.Value) == "" {
			return fmt.Errorf("--value cannot be empty")
		}
		attributes["value"] = opts.Value
	}
	if cmd.Flags().Changed("broker") {
		if strings.TrimSpace(opts.BrokerID) == "" {
			return fmt.Errorf("--broker cannot be empty")
		}
		relationships["broker"] = map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.BrokerID,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "trucker-referral-codes",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{"data": data}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/trucker-referral-codes/"+opts.ID, jsonBody)
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

	row := truckerReferralCodeRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated trucker referral code %s\n", row.ID)
	return nil
}

func parseDoTruckerReferralCodesUpdateOptions(cmd *cobra.Command, args []string) (doTruckerReferralCodesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	code, _ := cmd.Flags().GetString("code")
	value, _ := cmd.Flags().GetString("value")
	brokerID, _ := cmd.Flags().GetString("broker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTruckerReferralCodesUpdateOptions{
		BaseURL:  baseURL,
		Token:    token,
		JSON:     jsonOut,
		ID:       args[0],
		BrokerID: brokerID,
		Code:     code,
		Value:    value,
	}, nil
}
