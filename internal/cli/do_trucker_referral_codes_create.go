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

type doTruckerReferralCodesCreateOptions struct {
	BaseURL  string
	Token    string
	JSON     bool
	BrokerID string
	Code     string
	Value    string
}

func newDoTruckerReferralCodesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a trucker referral code",
		Long: `Create a trucker referral code.

Required flags:
  --broker  Broker ID (required)
  --code    Referral code (required)

Optional flags:
  --value   Referral value (numeric)

Note: Codes are normalized to uppercase and whitespace is removed.`,
		Example: `  # Create a trucker referral code
  xbe do trucker-referral-codes create --broker 123 --code "REF-123"

  # Create with a value
  xbe do trucker-referral-codes create --broker 123 --code "REF-123" --value 50`,
		Args: cobra.NoArgs,
		RunE: runDoTruckerReferralCodesCreate,
	}
	initDoTruckerReferralCodesCreateFlags(cmd)
	return cmd
}

func init() {
	doTruckerReferralCodesCmd.AddCommand(newDoTruckerReferralCodesCreateCmd())
}

func initDoTruckerReferralCodesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("code", "", "Referral code (required)")
	cmd.Flags().String("value", "", "Referral value (numeric)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTruckerReferralCodesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTruckerReferralCodesCreateOptions(cmd)
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

	if strings.TrimSpace(opts.BrokerID) == "" {
		err := fmt.Errorf("--broker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.Code) == "" {
		err := fmt.Errorf("--code is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"code": opts.Code,
	}
	if strings.TrimSpace(opts.Value) != "" {
		attributes["value"] = opts.Value
	}

	relationships := map[string]any{
		"broker": map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.BrokerID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "trucker-referral-codes",
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

	body, _, err := client.Post(cmd.Context(), "/v1/trucker-referral-codes", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created trucker referral code %s\n", row.ID)
	return nil
}

func parseDoTruckerReferralCodesCreateOptions(cmd *cobra.Command) (doTruckerReferralCodesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	brokerID, _ := cmd.Flags().GetString("broker")
	code, _ := cmd.Flags().GetString("code")
	value, _ := cmd.Flags().GetString("value")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTruckerReferralCodesCreateOptions{
		BaseURL:  baseURL,
		Token:    token,
		JSON:     jsonOut,
		BrokerID: brokerID,
		Code:     code,
		Value:    value,
	}, nil
}
