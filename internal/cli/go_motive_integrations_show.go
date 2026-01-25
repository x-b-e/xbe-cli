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

type goMotiveIntegrationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type goMotiveIntegrationDetails struct {
	ID                    string   `json:"id"`
	IntegrationIdentifier string   `json:"integration_identifier,omitempty"`
	FriendlyName          string   `json:"friendly_name,omitempty"`
	BrokerID              string   `json:"broker_id,omitempty"`
	BrokerName            string   `json:"broker_name,omitempty"`
	IntegrationConfigID   string   `json:"integration_config_id,omitempty"`
	IntegrationConfigName string   `json:"integration_config_name,omitempty"`
	TruckerIDs            []string `json:"trucker_ids,omitempty"`
	OAuthTokenExpiresAt   string   `json:"oauth_token_expires_at,omitempty"`
	OAuthTokenUpdatedAt   string   `json:"oauth_token_updated_at,omitempty"`
	IsOAuthLoggedIn       bool     `json:"is_oauth_logged_in"`
}

func newGoMotiveIntegrationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show GoMotive integration details",
		Long: `Show the full details of a GoMotive integration.

Output Fields:
  ID                   GoMotive integration identifier
  Friendly Name        Friendly name
  Integration ID       GoMotive integration identifier
  Broker               Broker name or ID
  Integration Config   Integration config name or ID
  OAuth Logged In      OAuth connection status
  OAuth Token Expires  OAuth token expiration timestamp
  OAuth Token Updated  OAuth token updated timestamp
  Truckers             Related trucker IDs

Arguments:
  <id>  GoMotive integration ID (required). Find IDs using the list command.`,
		Example: `  # Show GoMotive integration details
  xbe view go-motive-integrations show 123

  # Output as JSON
  xbe view go-motive-integrations show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runGoMotiveIntegrationsShow,
	}
	initGoMotiveIntegrationsShowFlags(cmd)
	return cmd
}

func init() {
	goMotiveIntegrationsCmd.AddCommand(newGoMotiveIntegrationsShowCmd())
}

func initGoMotiveIntegrationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runGoMotiveIntegrationsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseGoMotiveIntegrationsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run 'xbe auth login' first.")
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("go-motive integration id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[go-motive-integrations]", "integration-identifier,friendly-name,broker,integration-config,truckers")
	query.Set("include", "broker,integration-config")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[integration-configs]", "friendly-name")
	query.Set("meta[go-motive-integration]", "oauth_token_expires_at,oauth_token_updated_at,is_oauth_logged_in")

	body, _, err := client.Get(cmd.Context(), "/v1/go-motive-integrations/"+id, query)
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

	details := buildGoMotiveIntegrationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderGoMotiveIntegrationDetails(cmd, details)
}

func parseGoMotiveIntegrationsShowOptions(cmd *cobra.Command) (goMotiveIntegrationsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return goMotiveIntegrationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildGoMotiveIntegrationDetails(resp jsonAPISingleResponse) goMotiveIntegrationDetails {
	row := goMotiveIntegrationRowFromSingle(resp)
	resource := resp.Data

	details := goMotiveIntegrationDetails{
		ID:                    row.ID,
		IntegrationIdentifier: row.IntegrationIdentifier,
		FriendlyName:          row.FriendlyName,
		BrokerID:              row.BrokerID,
		BrokerName:            row.BrokerName,
		IntegrationConfigID:   row.IntegrationConfigID,
		IntegrationConfigName: row.IntegrationConfigName,
		OAuthTokenExpiresAt:   formatDateTime(stringAttr(resource.Meta, "oauth_token_expires_at")),
		OAuthTokenUpdatedAt:   formatDateTime(stringAttr(resource.Meta, "oauth_token_updated_at")),
		IsOAuthLoggedIn:       boolAttr(resource.Meta, "is_oauth_logged_in"),
	}

	if rel, ok := resource.Relationships["truckers"]; ok {
		ids := relationshipIDs(rel)
		if len(ids) > 0 {
			details.TruckerIDs = make([]string, 0, len(ids))
			for _, id := range ids {
				details.TruckerIDs = append(details.TruckerIDs, id.ID)
			}
		}
	}

	return details
}

func renderGoMotiveIntegrationDetails(cmd *cobra.Command, details goMotiveIntegrationDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.FriendlyName != "" {
		fmt.Fprintf(out, "Friendly Name: %s\n", details.FriendlyName)
	}
	if details.IntegrationIdentifier != "" {
		fmt.Fprintf(out, "Integration ID: %s\n", details.IntegrationIdentifier)
	}
	if details.BrokerID != "" || details.BrokerName != "" {
		fmt.Fprintf(out, "Broker: %s\n", formatRelated(details.BrokerName, details.BrokerID))
	}
	if details.IntegrationConfigID != "" || details.IntegrationConfigName != "" {
		fmt.Fprintf(out, "Integration Config: %s\n", formatRelated(details.IntegrationConfigName, details.IntegrationConfigID))
	}
	fmt.Fprintf(out, "OAuth Logged In: %s\n", formatBool(details.IsOAuthLoggedIn))
	if details.OAuthTokenExpiresAt != "" {
		fmt.Fprintf(out, "OAuth Token Expires: %s\n", details.OAuthTokenExpiresAt)
	}
	if details.OAuthTokenUpdatedAt != "" {
		fmt.Fprintf(out, "OAuth Token Updated: %s\n", details.OAuthTokenUpdatedAt)
	}
	if len(details.TruckerIDs) > 0 {
		fmt.Fprintf(out, "Truckers: %s\n", strings.Join(details.TruckerIDs, ", "))
	}
	return nil
}
