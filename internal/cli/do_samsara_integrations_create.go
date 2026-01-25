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

type doSamsaraIntegrationsCreateOptions struct {
	BaseURL               string
	Token                 string
	JSON                  bool
	IntegrationIdentifier string
	FriendlyName          string
	Broker                string
	IntegrationConfig     string
}

func newDoSamsaraIntegrationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Samsara integration",
		Long: `Create a Samsara integration.

Required flags:
  --integration-identifier   Integration identifier from Samsara (required)
  --friendly-name            Friendly name for the integration (required)
  --broker                   Broker ID (required)
  --integration-config       Integration config ID (required)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a Samsara integration
  xbe do samsara-integrations create \\
    --integration-identifier "samsara-123" \\
    --friendly-name "Main Samsara Account" \\
    --broker 456 \\
    --integration-config 789

  # Get JSON output
  xbe do samsara-integrations create --integration-identifier "samsara-123" --friendly-name "Main Samsara Account" --broker 456 --integration-config 789 --json`,
		Args: cobra.NoArgs,
		RunE: runDoSamsaraIntegrationsCreate,
	}
	initDoSamsaraIntegrationsCreateFlags(cmd)
	return cmd
}

func init() {
	doSamsaraIntegrationsCmd.AddCommand(newDoSamsaraIntegrationsCreateCmd())
}

func initDoSamsaraIntegrationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("integration-identifier", "", "Integration identifier from Samsara (required)")
	cmd.Flags().String("friendly-name", "", "Friendly name for the integration (required)")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("integration-config", "", "Integration config ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoSamsaraIntegrationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoSamsaraIntegrationsCreateOptions(cmd)
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

	if opts.IntegrationIdentifier == "" {
		err := fmt.Errorf("--integration-identifier is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.FriendlyName == "" {
		err := fmt.Errorf("--friendly-name is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.Broker == "" {
		err := fmt.Errorf("--broker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.IntegrationConfig == "" {
		err := fmt.Errorf("--integration-config is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"integration-identifier": opts.IntegrationIdentifier,
		"friendly-name":          opts.FriendlyName,
	}

	relationships := map[string]any{
		"broker": map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.Broker,
			},
		},
		"integration-config": map[string]any{
			"data": map[string]any{
				"type": "integration-configs",
				"id":   opts.IntegrationConfig,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "samsara-integrations",
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

	body, _, err := client.Post(cmd.Context(), "/v1/samsara-integrations", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created Samsara integration %s (%s)\n", details.ID, details.IntegrationID)
	return renderSamsaraIntegrationDetails(cmd, details)
}

func parseDoSamsaraIntegrationsCreateOptions(cmd *cobra.Command) (doSamsaraIntegrationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	integrationIdentifier, _ := cmd.Flags().GetString("integration-identifier")
	friendlyName, _ := cmd.Flags().GetString("friendly-name")
	broker, _ := cmd.Flags().GetString("broker")
	integrationConfig, _ := cmd.Flags().GetString("integration-config")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doSamsaraIntegrationsCreateOptions{
		BaseURL:               baseURL,
		Token:                 token,
		JSON:                  jsonOut,
		IntegrationIdentifier: integrationIdentifier,
		FriendlyName:          friendlyName,
		Broker:                broker,
		IntegrationConfig:     integrationConfig,
	}, nil
}
