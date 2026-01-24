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

type doUserSearchesCreateOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	ContactMethod     string
	ContactValue      string
	OnlyAdminOrMember string
}

func newDoUserSearchesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Run a user search",
		Long: `Run a user search by contact method and value.

Required flags:
  --contact-method        Contact method (email_address or mobile_number)
  --contact-value         Contact value to search

Optional flags:
  --only-admin-or-member  Restrict matches to admins or members (true/false)`,
		Example: `  # Search by email address
  xbe do user-searches create --contact-method email_address --contact-value "user@example.com"

  # Search by mobile number
  xbe do user-searches create --contact-method mobile_number --contact-value "+15551234567"

  # Restrict matches to admins or members
  xbe do user-searches create --contact-method email_address --contact-value "user@example.com" --only-admin-or-member true

  # Output as JSON
  xbe do user-searches create --contact-method email_address --contact-value "user@example.com" --json`,
		Args: cobra.NoArgs,
		RunE: runDoUserSearchesCreate,
	}
	initDoUserSearchesCreateFlags(cmd)
	return cmd
}

func init() {
	doUserSearchesCmd.AddCommand(newDoUserSearchesCreateCmd())
}

func initDoUserSearchesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("contact-method", "", "Contact method (email_address or mobile_number)")
	cmd.Flags().String("contact-value", "", "Contact value to search")
	cmd.Flags().String("only-admin-or-member", "", "Restrict matches to admins or members (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoUserSearchesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoUserSearchesCreateOptions(cmd)
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

	opts.ContactMethod = strings.TrimSpace(opts.ContactMethod)
	if opts.ContactMethod == "" {
		err := fmt.Errorf("--contact-method is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	opts.ContactValue = strings.TrimSpace(opts.ContactValue)
	if opts.ContactValue == "" {
		err := fmt.Errorf("--contact-value is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	switch opts.ContactMethod {
	case "email_address", "mobile_number":
		// valid
	default:
		err := fmt.Errorf("--contact-method must be one of: email_address, mobile_number")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"contact-method": opts.ContactMethod,
		"contact-value":  opts.ContactValue,
	}

	if opts.OnlyAdminOrMember != "" {
		attributes["only-admin-or-member"] = opts.OnlyAdminOrMember == "true"
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "user-searches",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/user-searches", jsonBody)
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

	row := userSearchRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	if row.MatchingUserID != "" {
		label := userSearchMatchingLabel(row)
		fmt.Fprintf(cmd.OutOrStdout(), "Matching user: %s\n", label)
		return nil
	}

	fmt.Fprintln(cmd.OutOrStdout(), "No matching user found.")
	return nil
}

func parseDoUserSearchesCreateOptions(cmd *cobra.Command) (doUserSearchesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	contactMethod, _ := cmd.Flags().GetString("contact-method")
	contactValue, _ := cmd.Flags().GetString("contact-value")
	onlyAdminOrMember, _ := cmd.Flags().GetString("only-admin-or-member")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doUserSearchesCreateOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		ContactMethod:     contactMethod,
		ContactValue:      contactValue,
		OnlyAdminOrMember: onlyAdminOrMember,
	}, nil
}
