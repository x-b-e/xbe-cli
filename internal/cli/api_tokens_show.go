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

type apiTokensShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type apiTokenDetails struct {
	ID             string `json:"id"`
	Name           string `json:"name,omitempty"`
	UserID         string `json:"user_id,omitempty"`
	UserName       string `json:"user_name,omitempty"`
	UserEmail      string `json:"user_email,omitempty"`
	CreatedByID    string `json:"created_by_id,omitempty"`
	CreatedByName  string `json:"created_by_name,omitempty"`
	CreatedByEmail string `json:"created_by_email,omitempty"`
	ExpiresAt      string `json:"expires_at,omitempty"`
	RevokedAt      string `json:"revoked_at,omitempty"`
	LastUsedAt     string `json:"last_used_at,omitempty"`
	Token          string `json:"token,omitempty"`
}

func newApiTokensShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show API token details",
		Long: `Show the full details of a specific API token.

API tokens are issued to a user and can be revoked or set to expire. The raw
token value is only returned at creation time and is not available in show.

Output Fields:
  ID             Token identifier
  Name           Token label (if set)
  User           User name, email, and ID
  Created By     Creator name, email, and ID (if present)
  Expires At     Expiration timestamp
  Revoked At     Revocation timestamp
  Last Used At   Last usage timestamp

Arguments:
  <id>    The API token ID (required). You can find IDs using the list command.`,
		Example: `  # View an API token by ID
  xbe view api-tokens show 123

  # Get JSON output
  xbe view api-tokens show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runApiTokensShow,
	}
	initApiTokensShowFlags(cmd)
	return cmd
}

func init() {
	apiTokensCmd.AddCommand(newApiTokensShowCmd())
}

func initApiTokensShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runApiTokensShow(cmd *cobra.Command, args []string) error {
	opts, err := parseApiTokensShowOptions(cmd)
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
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run xbe auth login first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("api token id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[api-tokens]", "name,expires-at,revoked-at,last-used-at,user,created-by,token")
	query.Set("fields[users]", "name,email-address")
	query.Set("include", "user,created-by")

	body, _, err := client.Get(cmd.Context(), "/v1/api-tokens/"+id, query)
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

	details := buildApiTokenDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderApiTokenDetails(cmd, details)
}

func parseApiTokensShowOptions(cmd *cobra.Command) (apiTokensShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return apiTokensShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildApiTokenDetails(resp jsonAPISingleResponse) apiTokenDetails {
	attrs := resp.Data.Attributes

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := apiTokenDetails{
		ID:         resp.Data.ID,
		Name:       strings.TrimSpace(stringAttr(attrs, "name")),
		ExpiresAt:  formatDateTime(stringAttr(attrs, "expires-at")),
		RevokedAt:  formatDateTime(stringAttr(attrs, "revoked-at")),
		LastUsedAt: formatDateTime(stringAttr(attrs, "last-used-at")),
		Token:      stringAttr(attrs, "token"),
	}

	if rel, ok := resp.Data.Relationships["user"]; ok && rel.Data != nil {
		details.UserID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.UserName = stringAttr(user.Attributes, "name")
			details.UserEmail = stringAttr(user.Attributes, "email-address")
		}
	}

	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.CreatedByName = stringAttr(user.Attributes, "name")
			details.CreatedByEmail = stringAttr(user.Attributes, "email-address")
		}
	}

	return details
}

func renderApiTokenDetails(cmd *cobra.Command, details apiTokenDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Name != "" {
		fmt.Fprintf(out, "Name: %s\n", details.Name)
	}

	if details.UserID != "" {
		userLabel := details.UserID
		if details.UserName != "" && details.UserEmail != "" {
			userLabel = fmt.Sprintf("%s <%s> (%s)", details.UserName, details.UserEmail, details.UserID)
		} else if details.UserName != "" {
			userLabel = fmt.Sprintf("%s (%s)", details.UserName, details.UserID)
		}
		fmt.Fprintf(out, "User: %s\n", userLabel)
	}

	if details.CreatedByID != "" {
		creatorLabel := details.CreatedByID
		if details.CreatedByName != "" && details.CreatedByEmail != "" {
			creatorLabel = fmt.Sprintf("%s <%s> (%s)", details.CreatedByName, details.CreatedByEmail, details.CreatedByID)
		} else if details.CreatedByName != "" {
			creatorLabel = fmt.Sprintf("%s (%s)", details.CreatedByName, details.CreatedByID)
		}
		fmt.Fprintf(out, "Created By: %s\n", creatorLabel)
	}

	if details.ExpiresAt != "" {
		fmt.Fprintf(out, "Expires At: %s\n", details.ExpiresAt)
	}
	if details.RevokedAt != "" {
		fmt.Fprintf(out, "Revoked At: %s\n", details.RevokedAt)
	}
	if details.LastUsedAt != "" {
		fmt.Fprintf(out, "Last Used At: %s\n", details.LastUsedAt)
	}
	if details.Token != "" {
		fmt.Fprintf(out, "Token: %s\n", details.Token)
	}

	return nil
}
