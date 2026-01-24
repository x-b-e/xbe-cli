package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
)

type doSamlCodeRedemptionsCreateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	Code    string
}

type samlCodeRedemptionRow struct {
	ID        string `json:"id"`
	Code      string `json:"code,omitempty"`
	AuthToken string `json:"auth_token,omitempty"`
}

func newDoSamlCodeRedemptionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Redeem a SAML login code",
		Long: `Redeem a SAML login code.

SAML code redemptions exchange a SAML login code for an XBE auth token.

Required flags:
  --code   SAML login code

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Redeem a SAML login code
  xbe do saml-code-redemptions create --code "saml_code_value"

  # JSON output
  xbe do saml-code-redemptions create --code "saml_code_value" --json`,
		Args: cobra.NoArgs,
		RunE: runDoSamlCodeRedemptionsCreate,
	}
	initDoSamlCodeRedemptionsCreateFlags(cmd)
	return cmd
}

func init() {
	doSamlCodeRedemptionsCmd.AddCommand(newDoSamlCodeRedemptionsCreateCmd())
}

func initDoSamlCodeRedemptionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("code", "", "SAML login code (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("code")
}

func runDoSamlCodeRedemptionsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoSamlCodeRedemptionsCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"code": strings.TrimSpace(opts.Code),
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "saml-code-redemptions",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/saml-code-redemptions", jsonBody)
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

	row := buildSamlCodeRedemptionRowFromSingle(resp)

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	if row.AuthToken != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Auth token: %s\n", row.AuthToken)
		return nil
	}
	if row.ID != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Redeemed SAML code %s\n", row.ID)
		return nil
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Redeemed SAML code")
	return nil
}

func parseDoSamlCodeRedemptionsCreateOptions(cmd *cobra.Command) (doSamlCodeRedemptionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	code, _ := cmd.Flags().GetString("code")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doSamlCodeRedemptionsCreateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		Code:    code,
	}, nil
}

func buildSamlCodeRedemptionRowFromSingle(resp jsonAPISingleResponse) samlCodeRedemptionRow {
	resource := resp.Data
	return samlCodeRedemptionRow{
		ID:        resource.ID,
		Code:      stringAttr(resource.Attributes, "code"),
		AuthToken: stringAttr(resource.Attributes, "auth-token"),
	}
}
