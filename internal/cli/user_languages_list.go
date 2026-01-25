package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type userLanguagesListOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	NoAuth    bool
	Limit     int
	Offset    int
	User      string
	Language  string
	IsDefault string
}

type userLanguageRow struct {
	ID           string `json:"id"`
	UserID       string `json:"user_id,omitempty"`
	UserName     string `json:"user_name,omitempty"`
	UserEmail    string `json:"user_email,omitempty"`
	LanguageID   string `json:"language_id,omitempty"`
	LanguageName string `json:"language_name,omitempty"`
	LanguageCode string `json:"language_code,omitempty"`
	IsDefault    bool   `json:"is_default"`
}

func newUserLanguagesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List user languages",
		Long: `List user language preferences with filtering and pagination.

User languages associate users with preferred languages and default settings.

Output Columns:
  ID        User language identifier
  USER      User name (falls back to ID)
  LANGUAGE  Language name/code (falls back to ID)
  DEFAULT   Default language flag

Filters:
  --user       Filter by user ID
  --language   Filter by language ID
  --is-default Filter by default status (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --base-url, --token, --no-auth`,
		Example: `  # List user languages
  xbe view user-languages list

  # Filter by user
  xbe view user-languages list --user 123

  # Filter by language
  xbe view user-languages list --language 456

  # Filter by default status
  xbe view user-languages list --is-default true

  # Output as JSON
  xbe view user-languages list --json`,
		Args: cobra.NoArgs,
		RunE: runUserLanguagesList,
	}
	initUserLanguagesListFlags(cmd)
	return cmd
}

func init() {
	userLanguagesCmd.AddCommand(newUserLanguagesListCmd())
}

func initUserLanguagesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("user", "", "Filter by user ID")
	cmd.Flags().String("language", "", "Filter by language ID")
	cmd.Flags().String("is-default", "", "Filter by default status (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runUserLanguagesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseUserLanguagesListOptions(cmd)
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

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[user-languages]", "is-default,user,language")
	query.Set("include", "user,language")
	query.Set("fields[users]", "name,email-address")
	query.Set("fields[languages]", "name,code")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	setFilterIfPresent(query, "filter[user]", opts.User)
	setFilterIfPresent(query, "filter[language]", opts.Language)
	setFilterIfPresent(query, "filter[is-default]", opts.IsDefault)

	body, _, err := client.Get(cmd.Context(), "/v1/user-languages", query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	rows := buildUserLanguageRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderUserLanguagesTable(cmd, rows)
}

func parseUserLanguagesListOptions(cmd *cobra.Command) (userLanguagesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	user, _ := cmd.Flags().GetString("user")
	language, _ := cmd.Flags().GetString("language")
	isDefault, _ := cmd.Flags().GetString("is-default")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return userLanguagesListOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		NoAuth:    noAuth,
		Limit:     limit,
		Offset:    offset,
		User:      user,
		Language:  language,
		IsDefault: isDefault,
	}, nil
}

func buildUserLanguageRows(resp jsonAPIResponse) []userLanguageRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]userLanguageRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := userLanguageRow{
			ID:        resource.ID,
			IsDefault: boolAttr(resource.Attributes, "is-default"),
		}

		if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
			row.UserID = rel.Data.ID
			if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.UserName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
				row.UserEmail = strings.TrimSpace(stringAttr(user.Attributes, "email-address"))
			}
		}

		if rel, ok := resource.Relationships["language"]; ok && rel.Data != nil {
			row.LanguageID = rel.Data.ID
			if language, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.LanguageName = strings.TrimSpace(stringAttr(language.Attributes, "name"))
				row.LanguageCode = strings.TrimSpace(stringAttr(language.Attributes, "code"))
			}
		}

		rows = append(rows, row)
	}
	return rows
}

func renderUserLanguagesTable(cmd *cobra.Command, rows []userLanguageRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No user languages found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tUSER\tLANGUAGE\tDEFAULT")
	for _, row := range rows {
		userDisplay := firstNonEmpty(row.UserName, row.UserID)
		if row.UserName != "" && row.UserID != "" {
			userDisplay = fmt.Sprintf("%s (%s)", row.UserName, row.UserID)
		}

		languageDisplay := firstNonEmpty(row.LanguageName, row.LanguageCode, row.LanguageID)
		if row.LanguageName != "" && row.LanguageCode != "" {
			languageDisplay = fmt.Sprintf("%s (%s)", row.LanguageName, row.LanguageCode)
		}

		defaultLabel := "No"
		if row.IsDefault {
			defaultLabel = "Yes"
		}

		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(userDisplay, 28),
			truncateString(languageDisplay, 28),
			defaultLabel,
		)
	}
	return writer.Flush()
}
