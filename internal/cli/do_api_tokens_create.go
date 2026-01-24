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

type doApiTokensCreateOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	UserID    string
	Name      string
	ExpiresAt string
}

func newDoApiTokensCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new API token",
		Long: `Create a new API token.

Required flags:
  --user          User ID (required)

Optional flags:
  --name          Token label
  --expires-at    Expiration timestamp (RFC3339)

The token value is only returned at creation time.`,
		Example: `  # Create an API token
  xbe do api-tokens create --user 123

  # Create with a name
  xbe do api-tokens create --user 123 --name "CI token"

  # Create with expiration
  xbe do api-tokens create --user 123 --expires-at 2026-12-31T00:00:00Z`,
		Args: cobra.NoArgs,
		RunE: runDoApiTokensCreate,
	}
	initDoApiTokensCreateFlags(cmd)
	return cmd
}

func init() {
	doApiTokensCmd.AddCommand(newDoApiTokensCreateCmd())
}

func initDoApiTokensCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("user", "", "User ID (required)")
	cmd.Flags().String("name", "", "Token label")
	cmd.Flags().String("expires-at", "", "Expiration timestamp (RFC3339)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoApiTokensCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoApiTokensCreateOptions(cmd)
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

	if opts.UserID == "" {
		err := fmt.Errorf("--user is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if strings.TrimSpace(opts.Name) != "" {
		attributes["name"] = opts.Name
	}
	if strings.TrimSpace(opts.ExpiresAt) != "" {
		attributes["expires-at"] = opts.ExpiresAt
	}

	relationships := map[string]any{
		"user": map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.UserID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "api-tokens",
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

	body, _, err := client.Post(cmd.Context(), "/v1/api-tokens", jsonBody)
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

	row := buildApiTokenRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created API token %s\n", row.ID)
	if row.Token != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Token: %s\n", row.Token)
		fmt.Fprintln(cmd.OutOrStdout(), "Store this token now; it will not be shown again.")
	}
	return nil
}

func parseDoApiTokensCreateOptions(cmd *cobra.Command) (doApiTokensCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	userID, _ := cmd.Flags().GetString("user")
	name, _ := cmd.Flags().GetString("name")
	expiresAt, _ := cmd.Flags().GetString("expires-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doApiTokensCreateOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		UserID:    userID,
		Name:      name,
		ExpiresAt: expiresAt,
	}, nil
}

func buildApiTokenRowFromSingle(resp jsonAPISingleResponse) apiTokenRow {
	attrs := resp.Data.Attributes

	row := apiTokenRow{
		ID:         resp.Data.ID,
		Name:       strings.TrimSpace(stringAttr(attrs, "name")),
		ExpiresAt:  formatDateTime(stringAttr(attrs, "expires-at")),
		RevokedAt:  formatDateTime(stringAttr(attrs, "revoked-at")),
		LastUsedAt: formatDateTime(stringAttr(attrs, "last-used-at")),
		Token:      stringAttr(attrs, "token"),
	}

	if rel, ok := resp.Data.Relationships["user"]; ok && rel.Data != nil {
		row.UserID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		row.CreatedBy = rel.Data.ID
	}

	return row
}
