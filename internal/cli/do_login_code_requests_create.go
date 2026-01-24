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

type doLoginCodeRequestsCreateOptions struct {
	BaseURL       string
	Token         string
	JSON          bool
	ContactMethod string
	DeviceID      string
}

type loginCodeRequestRow struct {
	ID            string `json:"id"`
	ContactMethod string `json:"contact_method,omitempty"`
	DeviceID      string `json:"device_id,omitempty"`
	Result        string `json:"result,omitempty"`
}

func newDoLoginCodeRequestsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Request a login code",
		Long: `Request a login code.

Login code requests send a one-time login code to the provided email address
or mobile number.

Required flags:
  --contact-method   Email address or mobile number
  --device-id        Device identifier

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Request a login code for an email
  xbe do login-code-requests create --contact-method user@example.com --device-id "xbe-cli"

  # Request a login code for a mobile number
  xbe do login-code-requests create --contact-method "+18155551234" --device-id "xbe-cli"

  # JSON output
  xbe do login-code-requests create --contact-method user@example.com --device-id "xbe-cli" --json`,
		Args: cobra.NoArgs,
		RunE: runDoLoginCodeRequestsCreate,
	}
	initDoLoginCodeRequestsCreateFlags(cmd)
	return cmd
}

func init() {
	doLoginCodeRequestsCmd.AddCommand(newDoLoginCodeRequestsCreateCmd())
}

func initDoLoginCodeRequestsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("contact-method", "", "Email address or mobile number (required)")
	cmd.Flags().String("device-id", "", "Device identifier (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("contact-method")
	cmd.MarkFlagRequired("device-id")
}

func runDoLoginCodeRequestsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoLoginCodeRequestsCreateOptions(cmd)
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

	attributes := map[string]any{
		"contact-method": strings.TrimSpace(opts.ContactMethod),
		"device-id":      strings.TrimSpace(opts.DeviceID),
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "login-code-requests",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/login-code-requests", jsonBody)
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

	row := buildLoginCodeRequestRowFromSingle(resp)
	if row.ContactMethod == "" {
		row.ContactMethod = strings.TrimSpace(opts.ContactMethod)
	}
	if row.DeviceID == "" {
		row.DeviceID = strings.TrimSpace(opts.DeviceID)
	}

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	if row.ContactMethod != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Contact: %s\n", row.ContactMethod)
	}
	if row.DeviceID != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Device: %s\n", row.DeviceID)
	}
	if row.Result != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Result: %s\n", row.Result)
		return nil
	}
	if row.ID != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Created login code request %s\n", row.ID)
		return nil
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Created login code request")
	return nil
}

func parseDoLoginCodeRequestsCreateOptions(cmd *cobra.Command) (doLoginCodeRequestsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	contactMethod, _ := cmd.Flags().GetString("contact-method")
	deviceID, _ := cmd.Flags().GetString("device-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doLoginCodeRequestsCreateOptions{
		BaseURL:       baseURL,
		Token:         token,
		JSON:          jsonOut,
		ContactMethod: contactMethod,
		DeviceID:      deviceID,
	}, nil
}

func buildLoginCodeRequestRowFromSingle(resp jsonAPISingleResponse) loginCodeRequestRow {
	resource := resp.Data
	return loginCodeRequestRow{
		ID:            resource.ID,
		ContactMethod: stringAttr(resource.Attributes, "contact-method"),
		DeviceID:      stringAttr(resource.Attributes, "device-id"),
		Result:        stringAttr(resource.Attributes, "result"),
	}
}
