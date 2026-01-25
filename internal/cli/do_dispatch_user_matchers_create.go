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

type doDispatchUserMatchersCreateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	PhoneNumber string
}

type dispatchUserMatcherRow struct {
	ID                  string `json:"id"`
	PhoneNumber         string `json:"phone_number,omitempty"`
	DispatchPhoneNumber string `json:"dispatch_phone_number,omitempty"`
}

func newDoDispatchUserMatchersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Match a dispatch user by phone number",
		Long: `Match a dispatch user by phone number.

Required flags:
  --phone-number    Caller phone number to match (required)

Returns the dispatch phone number for the caller.`,
		Example: `  # Match a dispatch user by phone number
  xbe do dispatch-user-matchers create --phone-number "+15551234567"

  # Output as JSON
  xbe do dispatch-user-matchers create --phone-number "+15551234567" --json`,
		Args: cobra.NoArgs,
		RunE: runDoDispatchUserMatchersCreate,
	}
	initDoDispatchUserMatchersCreateFlags(cmd)
	return cmd
}

func init() {
	doDispatchUserMatchersCmd.AddCommand(newDoDispatchUserMatchersCreateCmd())
}

func initDoDispatchUserMatchersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("phone-number", "", "Caller phone number to match (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoDispatchUserMatchersCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoDispatchUserMatchersCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run xbe auth login first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	opts.PhoneNumber = strings.TrimSpace(opts.PhoneNumber)
	if opts.PhoneNumber == "" {
		err := fmt.Errorf("--phone-number is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type": "dispatch-user-matchers",
			"attributes": map[string]any{
				"phone-number": opts.PhoneNumber,
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/dispatch-user-matchers", jsonBody)
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

	row := dispatchUserMatcherRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	if row.DispatchPhoneNumber != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Dispatch phone number: %s\n", row.DispatchPhoneNumber)
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created dispatch user matcher %s\n", row.ID)
	return nil
}

func dispatchUserMatcherRowFromSingle(resp jsonAPISingleResponse) dispatchUserMatcherRow {
	attrs := resp.Data.Attributes
	return dispatchUserMatcherRow{
		ID:                  resp.Data.ID,
		PhoneNumber:         stringAttr(attrs, "phone-number"),
		DispatchPhoneNumber: stringAttr(attrs, "dispatch-phone-number"),
	}
}

func parseDoDispatchUserMatchersCreateOptions(cmd *cobra.Command) (doDispatchUserMatchersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	phoneNumber, _ := cmd.Flags().GetString("phone-number")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doDispatchUserMatchersCreateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		PhoneNumber: phoneNumber,
	}, nil
}
