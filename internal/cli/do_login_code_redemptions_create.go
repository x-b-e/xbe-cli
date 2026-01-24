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

type doLoginCodeRedemptionsCreateOptions struct {
	BaseURL  string
	Token    string
	JSON     bool
	Code     string
	DeviceID string
}

type loginCodeRedemptionRow struct {
	ID        string `json:"id"`
	Code      string `json:"code,omitempty"`
	DeviceID  string `json:"device_id,omitempty"`
	AuthToken string `json:"auth_token,omitempty"`
}

func newDoLoginCodeRedemptionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Redeem a login code",
		Long: `Redeem a login code to obtain an auth token.

Required flags:
  --code       Login code to redeem (required)

Optional flags:
  --device-id  Device identifier to bind to the login code`,
		Example: `  # Redeem a login code
  xbe do login-code-redemptions create --code 123456

  # Redeem with a device ID
  xbe do login-code-redemptions create --code 123456 --device-id "device-abc"

  # Output as JSON
  xbe do login-code-redemptions create --code 123456 --json`,
		Args: cobra.NoArgs,
		RunE: runDoLoginCodeRedemptionsCreate,
	}
	initDoLoginCodeRedemptionsCreateFlags(cmd)
	return cmd
}

func init() {
	doLoginCodeRedemptionsCmd.AddCommand(newDoLoginCodeRedemptionsCreateCmd())
}

func initDoLoginCodeRedemptionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("code", "", "Login code to redeem (required)")
	cmd.Flags().String("device-id", "", "Device identifier to bind to the login code")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoLoginCodeRedemptionsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoLoginCodeRedemptionsCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	opts.Code = strings.TrimSpace(opts.Code)
	if opts.Code == "" {
		err := fmt.Errorf("--code is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	opts.DeviceID = strings.TrimSpace(opts.DeviceID)

	attributes := map[string]any{
		"code": opts.Code,
	}
	if opts.DeviceID != "" {
		attributes["device-id"] = opts.DeviceID
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "login-code-redemptions",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/login-code-redemptions", jsonBody)
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

	row := loginCodeRedemptionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	if row.AuthToken != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Auth token: %s\n", row.AuthToken)
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created login code redemption %s\n", row.ID)
	return nil
}

func loginCodeRedemptionRowFromSingle(resp jsonAPISingleResponse) loginCodeRedemptionRow {
	attrs := resp.Data.Attributes
	return loginCodeRedemptionRow{
		ID:        resp.Data.ID,
		Code:      stringAttr(attrs, "code"),
		DeviceID:  firstNonEmpty(stringAttr(attrs, "device-id"), stringAttr(attrs, "device_id")),
		AuthToken: firstNonEmpty(stringAttr(attrs, "auth-token"), stringAttr(attrs, "auth_token")),
	}
}

func parseDoLoginCodeRedemptionsCreateOptions(cmd *cobra.Command) (doLoginCodeRedemptionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	code, _ := cmd.Flags().GetString("code")
	deviceID, _ := cmd.Flags().GetString("device-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doLoginCodeRedemptionsCreateOptions{
		BaseURL:  baseURL,
		Token:    token,
		JSON:     jsonOut,
		Code:     code,
		DeviceID: deviceID,
	}, nil
}
