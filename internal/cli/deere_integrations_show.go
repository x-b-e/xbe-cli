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

type deereIntegrationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type deereIntegrationDetails struct {
	ID                  string `json:"id"`
	IntegrationID       string `json:"integration_identifier,omitempty"`
	FriendlyName        string `json:"friendly_name,omitempty"`
	BrokerID            string `json:"broker_id,omitempty"`
	IntegrationConfigID string `json:"integration_config_id,omitempty"`
	OAuthTokenExpiresAt string `json:"oauth_token_expires_at,omitempty"`
	OAuthTokenUpdatedAt string `json:"oauth_token_updated_at,omitempty"`
	IsOAuthLoggedIn     bool   `json:"is_oauth_logged_in,omitempty"`
}

func newDeereIntegrationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show Deere integration details",
		Long: `Show the full details of a Deere integration.

Output Fields:
  ID                   Deere integration identifier
  INTEGRATION ID       Integration identifier from Deere
  FRIENDLY NAME        Friendly name for the integration
  BROKER ID            Broker ID
  INTEGRATION CONFIG   Integration config ID
  OAUTH LOGGED IN      Whether OAuth is connected
  OAUTH UPDATED AT     OAuth token updated timestamp (if available)
  OAUTH EXPIRES AT     OAuth token expiration timestamp (if available)

Arguments:
  <id>    The Deere integration ID (required). You can find IDs using the list command.`,
		Example: `  # Show a Deere integration
  xbe view deere-integrations show 123

  # Get JSON output
  xbe view deere-integrations show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDeereIntegrationsShow,
	}
	initDeereIntegrationsShowFlags(cmd)
	return cmd
}

func init() {
	deereIntegrationsCmd.AddCommand(newDeereIntegrationsShowCmd())
}

func initDeereIntegrationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDeereIntegrationsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseDeereIntegrationsShowOptions(cmd)
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
		return fmt.Errorf("deere integration id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[deere-integrations]", "integration-identifier,friendly-name,broker,integration-config")
	query.Set("meta[deere-integration]", "oauth-token-expires-at,oauth-token-updated-at,is-oauth-logged-in")

	body, _, err := client.Get(cmd.Context(), "/v1/deere-integrations/"+id, query)
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

	details := buildDeereIntegrationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderDeereIntegrationDetails(cmd, details)
}

func parseDeereIntegrationsShowOptions(cmd *cobra.Command) (deereIntegrationsShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return deereIntegrationsShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return deereIntegrationsShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return deereIntegrationsShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return deereIntegrationsShowOptions{}, err
	}

	return deereIntegrationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildDeereIntegrationDetails(resp jsonAPISingleResponse) deereIntegrationDetails {
	resource := resp.Data
	attrs := resource.Attributes
	meta := resource.Meta

	details := deereIntegrationDetails{
		ID:                  resource.ID,
		IntegrationID:       stringAttr(attrs, "integration-identifier"),
		FriendlyName:        stringAttr(attrs, "friendly-name"),
		BrokerID:            relationshipIDFromMap(resource.Relationships, "broker"),
		IntegrationConfigID: relationshipIDFromMap(resource.Relationships, "integration-config"),
	}

	if meta != nil {
		details.OAuthTokenExpiresAt = stringAttr(meta, "oauth_token_expires_at")
		if details.OAuthTokenExpiresAt == "" {
			details.OAuthTokenExpiresAt = stringAttr(meta, "oauth-token-expires-at")
		}
		details.OAuthTokenUpdatedAt = stringAttr(meta, "oauth_token_updated_at")
		if details.OAuthTokenUpdatedAt == "" {
			details.OAuthTokenUpdatedAt = stringAttr(meta, "oauth-token-updated-at")
		}
		details.IsOAuthLoggedIn = boolAttr(meta, "is_oauth_logged_in")
		if !details.IsOAuthLoggedIn {
			details.IsOAuthLoggedIn = boolAttr(meta, "is-oauth-logged-in")
		}
	}

	return details
}

func renderDeereIntegrationDetails(cmd *cobra.Command, details deereIntegrationDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.IntegrationID != "" {
		fmt.Fprintf(out, "Integration ID: %s\n", details.IntegrationID)
	}
	if details.FriendlyName != "" {
		fmt.Fprintf(out, "Friendly Name: %s\n", details.FriendlyName)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
	}
	if details.IntegrationConfigID != "" {
		fmt.Fprintf(out, "Integration Config ID: %s\n", details.IntegrationConfigID)
	}

	fmt.Fprintf(out, "OAuth Logged In: %s\n", formatYesNo(details.IsOAuthLoggedIn))
	if details.OAuthTokenUpdatedAt != "" {
		fmt.Fprintf(out, "OAuth Updated At: %s\n", details.OAuthTokenUpdatedAt)
	}
	if details.OAuthTokenExpiresAt != "" {
		fmt.Fprintf(out, "OAuth Expires At: %s\n", details.OAuthTokenExpiresAt)
	}

	return nil
}
