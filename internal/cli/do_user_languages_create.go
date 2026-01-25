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

type doUserLanguagesCreateOptions struct {
	BaseURL    string
	Token      string
	JSON       bool
	UserID     string
	LanguageID string
	IsDefault  string
}

func newDoUserLanguagesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a user language",
		Long: `Create a user language preference.

Required flags:
  --user      User ID (required)
  --language  Language ID (required)

Optional flags:
  --is-default  Default language flag (true/false)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a user language
  xbe do user-languages create --user 123 --language 456

  # Create and mark default
  xbe do user-languages create --user 123 --language 456 --is-default true`,
		Args: cobra.NoArgs,
		RunE: runDoUserLanguagesCreate,
	}
	initDoUserLanguagesCreateFlags(cmd)
	return cmd
}

func init() {
	doUserLanguagesCmd.AddCommand(newDoUserLanguagesCreateCmd())
}

func initDoUserLanguagesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("user", "", "User ID (required)")
	cmd.Flags().String("language", "", "Language ID (required)")
	cmd.Flags().String("is-default", "", "Default language flag (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoUserLanguagesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoUserLanguagesCreateOptions(cmd)
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

	if opts.UserID == "" {
		err := fmt.Errorf("--user is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.LanguageID == "" {
		err := fmt.Errorf("--language is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	setBoolAttrIfPresent(attributes, "is-default", opts.IsDefault)

	relationships := map[string]any{
		"user": map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.UserID,
			},
		},
		"language": map[string]any{
			"data": map[string]any{
				"type": "languages",
				"id":   opts.LanguageID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "user-languages",
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

	body, _, err := client.Post(cmd.Context(), "/v1/user-languages", jsonBody)
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

	row := buildUserLanguageRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created user language %s\n", row.ID)
	return nil
}

func parseDoUserLanguagesCreateOptions(cmd *cobra.Command) (doUserLanguagesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	userID, _ := cmd.Flags().GetString("user")
	languageID, _ := cmd.Flags().GetString("language")
	isDefault, _ := cmd.Flags().GetString("is-default")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doUserLanguagesCreateOptions{
		BaseURL:    baseURL,
		Token:      token,
		JSON:       jsonOut,
		UserID:     userID,
		LanguageID: languageID,
		IsDefault:  isDefault,
	}, nil
}

func buildUserLanguageRowFromSingle(resp jsonAPISingleResponse) userLanguageRow {
	attrs := resp.Data.Attributes

	row := userLanguageRow{
		ID:        resp.Data.ID,
		IsDefault: boolAttr(attrs, "is-default"),
	}

	if rel, ok := resp.Data.Relationships["user"]; ok && rel.Data != nil {
		row.UserID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["language"]; ok && rel.Data != nil {
		row.LanguageID = rel.Data.ID
	}

	return row
}
