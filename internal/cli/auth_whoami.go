package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type whoamiOptions struct {
	BaseURL string
	Token   string
	JSON    bool
}

type whoamiResult struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Mobile  string `json:"mobile,omitempty"`
	IsAdmin bool   `json:"is_admin"`
}

var authWhoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show the current authenticated user",
	Long: `Show the current authenticated user.

Verifies your authentication by fetching your user profile from the API.
This confirms that your token is valid and shows who you are logged in as.

Output Fields:
  ID       Your unique user identifier
  Name     Your display name
  Email    Your email address
  Admin    Whether you have admin privileges`,
	Example: `  # Check who you're logged in as
  xbe auth whoami

  # Get output as JSON
  xbe auth whoami --json

  # Check against a specific environment
  xbe auth whoami --base-url https://staging.x-b-e.com`,
	RunE: runAuthWhoami,
}

func init() {
	authCmd.AddCommand(authWhoamiCmd)

	authWhoamiCmd.Flags().Bool("json", false, "Output JSON")
	authWhoamiCmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	authWhoamiCmd.Flags().String("token", "", "API token (optional)")
}

func runAuthWhoami(cmd *cobra.Command, _ []string) error {
	opts, err := parseWhoamiOptions(cmd)
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

	if opts.Token == "" {
		err := errors.New("not authenticated: use 'xbe auth login' to authenticate")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[users]", "name,email-address,mobile-number,is-admin")

	body, _, err := client.Get(cmd.Context(), "/v1/users/me", query)
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

	result := buildWhoamiResult(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), result)
	}

	return renderWhoami(cmd, result)
}

func parseWhoamiOptions(cmd *cobra.Command) (whoamiOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return whoamiOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return whoamiOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return whoamiOptions{}, err
	}

	return whoamiOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
	}, nil
}

func buildWhoamiResult(resp jsonAPISingleResponse) whoamiResult {
	attrs := resp.Data.Attributes
	return whoamiResult{
		ID:      resp.Data.ID,
		Name:    strings.TrimSpace(stringAttr(attrs, "name")),
		Email:   strings.TrimSpace(stringAttr(attrs, "email-address")),
		Mobile:  strings.TrimSpace(stringAttr(attrs, "mobile-number")),
		IsAdmin: boolAttr(attrs, "is-admin"),
	}
}

func renderWhoami(cmd *cobra.Command, result whoamiResult) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "Logged in as %s\n", result.Name)
	fmt.Fprintf(out, "  ID:    %s\n", result.ID)
	fmt.Fprintf(out, "  Email: %s\n", result.Email)
	if result.Mobile != "" {
		fmt.Fprintf(out, "  Mobile: %s\n", result.Mobile)
	}
	if result.IsAdmin {
		fmt.Fprintf(out, "  Admin: yes\n")
	}

	return nil
}
