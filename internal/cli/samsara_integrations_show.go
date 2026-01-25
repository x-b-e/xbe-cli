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

type samsaraIntegrationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type samsaraIntegrationDetails struct {
	ID                  string   `json:"id"`
	IntegrationID       string   `json:"integration_identifier,omitempty"`
	FriendlyName        string   `json:"friendly_name,omitempty"`
	BrokerID            string   `json:"broker_id,omitempty"`
	IntegrationConfigID string   `json:"integration_config_id,omitempty"`
	TruckerIDs          []string `json:"trucker_ids,omitempty"`
}

func newSamsaraIntegrationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show Samsara integration details",
		Long: `Show the full details of a Samsara integration.

Output Fields:
  ID                   Samsara integration identifier
  INTEGRATION ID       Integration identifier from Samsara
  FRIENDLY NAME        Friendly name for the integration
  BROKER ID            Broker ID
  INTEGRATION CONFIG   Integration config ID
  TRUCKER IDS          Linked trucker IDs

Global flags (see xbe --help): --json, --base-url, --token, --no-auth

Arguments:
  <id>    The Samsara integration ID (required). You can find IDs using the list command.`,
		Example: `  # Show a Samsara integration
  xbe view samsara-integrations show 123

  # Get JSON output
  xbe view samsara-integrations show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runSamsaraIntegrationsShow,
	}
	initSamsaraIntegrationsShowFlags(cmd)
	return cmd
}

func init() {
	samsaraIntegrationsCmd.AddCommand(newSamsaraIntegrationsShowCmd())
}

func initSamsaraIntegrationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runSamsaraIntegrationsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseSamsaraIntegrationsShowOptions(cmd)
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
		return fmt.Errorf("samsara integration id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[samsara-integrations]", "integration-identifier,friendly-name,broker,integration-config,truckers")

	body, _, err := client.Get(cmd.Context(), "/v1/samsara-integrations/"+id, query)
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

	details := buildSamsaraIntegrationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderSamsaraIntegrationDetails(cmd, details)
}

func parseSamsaraIntegrationsShowOptions(cmd *cobra.Command) (samsaraIntegrationsShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return samsaraIntegrationsShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return samsaraIntegrationsShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return samsaraIntegrationsShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return samsaraIntegrationsShowOptions{}, err
	}

	return samsaraIntegrationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildSamsaraIntegrationDetails(resp jsonAPISingleResponse) samsaraIntegrationDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := samsaraIntegrationDetails{
		ID:                  resource.ID,
		IntegrationID:       stringAttr(attrs, "integration-identifier"),
		FriendlyName:        stringAttr(attrs, "friendly-name"),
		BrokerID:            relationshipIDFromMap(resource.Relationships, "broker"),
		IntegrationConfigID: relationshipIDFromMap(resource.Relationships, "integration-config"),
		TruckerIDs:          relationshipIDsFromMap(resource.Relationships, "truckers"),
	}

	return details
}

func renderSamsaraIntegrationDetails(cmd *cobra.Command, details samsaraIntegrationDetails) error {
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
	if len(details.TruckerIDs) > 0 {
		fmt.Fprintf(out, "Trucker IDs: %s\n", strings.Join(details.TruckerIDs, ", "))
	}

	return nil
}
