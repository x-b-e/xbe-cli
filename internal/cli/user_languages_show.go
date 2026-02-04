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

type userLanguagesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type userLanguageDetails struct {
	ID           string `json:"id"`
	IsDefault    bool   `json:"is_default"`
	UserID       string `json:"user_id,omitempty"`
	UserName     string `json:"user_name,omitempty"`
	UserEmail    string `json:"user_email,omitempty"`
	LanguageID   string `json:"language_id,omitempty"`
	LanguageName string `json:"language_name,omitempty"`
	LanguageCode string `json:"language_code,omitempty"`
}

func newUserLanguagesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show user language details",
		Long: `Show the full details of a user language preference.

Output Fields:
  ID           User language identifier
  Is Default   Default language flag
  User         User ID, name, and email (if available)
  Language     Language ID, name, and code (if available)

Arguments:
  <id>    User language ID (required)

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a user language
  xbe view user-languages show 123

  # JSON output
  xbe view user-languages show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runUserLanguagesShow,
	}
	initUserLanguagesShowFlags(cmd)
	return cmd
}

func init() {
	userLanguagesCmd.AddCommand(newUserLanguagesShowCmd())
}

func initUserLanguagesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runUserLanguagesShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseUserLanguagesShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("user language id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("include", "user,language")
	query.Set("fields[user-languages]", "is-default,user,language")
	query.Set("fields[users]", "name,email-address")
	query.Set("fields[languages]", "name,code")

	body, _, err := client.Get(cmd.Context(), "/v1/user-languages/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildUserLanguageDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderUserLanguageDetails(cmd, details)
}

func parseUserLanguagesShowOptions(cmd *cobra.Command) (userLanguagesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return userLanguagesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildUserLanguageDetails(resp jsonAPISingleResponse) userLanguageDetails {
	attrs := resp.Data.Attributes
	details := userLanguageDetails{
		ID:        resp.Data.ID,
		IsDefault: boolAttr(attrs, "is-default"),
	}

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	if rel, ok := resp.Data.Relationships["user"]; ok && rel.Data != nil {
		details.UserID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.UserName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
			details.UserEmail = strings.TrimSpace(stringAttr(user.Attributes, "email-address"))
		}
	}

	if rel, ok := resp.Data.Relationships["language"]; ok && rel.Data != nil {
		details.LanguageID = rel.Data.ID
		if language, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.LanguageName = strings.TrimSpace(stringAttr(language.Attributes, "name"))
			details.LanguageCode = strings.TrimSpace(stringAttr(language.Attributes, "code"))
		}
	}

	return details
}

func renderUserLanguageDetails(cmd *cobra.Command, details userLanguageDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	fmt.Fprintf(out, "Is Default: %s\n", formatBool(details.IsDefault))
	fmt.Fprintln(out, "")

	fmt.Fprintln(out, "User:")
	if details.UserID != "" {
		fmt.Fprintf(out, "  ID: %s\n", details.UserID)
	}
	if details.UserName != "" {
		fmt.Fprintf(out, "  Name: %s\n", details.UserName)
	}
	if details.UserEmail != "" {
		fmt.Fprintf(out, "  Email: %s\n", details.UserEmail)
	}
	fmt.Fprintln(out, "")

	fmt.Fprintln(out, "Language:")
	if details.LanguageID != "" {
		fmt.Fprintf(out, "  ID: %s\n", details.LanguageID)
	}
	if details.LanguageName != "" {
		fmt.Fprintf(out, "  Name: %s\n", details.LanguageName)
	}
	if details.LanguageCode != "" {
		fmt.Fprintf(out, "  Code: %s\n", details.LanguageCode)
	}

	return nil
}
