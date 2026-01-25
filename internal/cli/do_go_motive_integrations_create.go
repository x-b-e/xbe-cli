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

type doGoMotiveIntegrationsCreateOptions struct {
	BaseURL               string
	Token                 string
	JSON                  bool
	IntegrationIdentifier string
	FriendlyName          string
	BrokerID              string
	IntegrationConfigID   string
}

func newDoGoMotiveIntegrationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a GoMotive integration",
		Long: `Create a GoMotive integration.

Required flags:
  --integration-identifier  GoMotive integration identifier
  --friendly-name           Friendly name
  --broker                  Broker ID
  --integration-config      Integration config ID`,
		Example: `  # Create a GoMotive integration
  xbe do go-motive-integrations create \
    --integration-identifier "motive-123" \
    --friendly-name "Main GoMotive" \
    --broker 123 \
    --integration-config 456`,
		Args: cobra.NoArgs,
		RunE: runDoGoMotiveIntegrationsCreate,
	}
	initDoGoMotiveIntegrationsCreateFlags(cmd)
	return cmd
}

func init() {
	doGoMotiveIntegrationsCmd.AddCommand(newDoGoMotiveIntegrationsCreateCmd())
}

func initDoGoMotiveIntegrationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("integration-identifier", "", "GoMotive integration identifier")
	cmd.Flags().String("friendly-name", "", "Friendly name")
	cmd.Flags().String("broker", "", "Broker ID")
	cmd.Flags().String("integration-config", "", "Integration config ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoGoMotiveIntegrationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoGoMotiveIntegrationsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.IntegrationIdentifier) == "" {
		err := fmt.Errorf("--integration-identifier is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.FriendlyName) == "" {
		err := fmt.Errorf("--friendly-name is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.BrokerID) == "" {
		err := fmt.Errorf("--broker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.IntegrationConfigID) == "" {
		err := fmt.Errorf("--integration-config is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"broker": map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.BrokerID,
			},
		},
		"integration-config": map[string]any{
			"data": map[string]any{
				"type": "integration-configs",
				"id":   opts.IntegrationConfigID,
			},
		},
	}

	attributes := map[string]any{
		"integration-identifier": opts.IntegrationIdentifier,
		"friendly-name":          opts.FriendlyName,
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "go-motive-integrations",
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

	body, _, err := client.Post(cmd.Context(), "/v1/go-motive-integrations", jsonBody)
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

	row := goMotiveIntegrationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created GoMotive integration %s\n", row.ID)
	return nil
}

func parseDoGoMotiveIntegrationsCreateOptions(cmd *cobra.Command) (doGoMotiveIntegrationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	integrationIdentifier, _ := cmd.Flags().GetString("integration-identifier")
	friendlyName, _ := cmd.Flags().GetString("friendly-name")
	brokerID, _ := cmd.Flags().GetString("broker")
	integrationConfigID, _ := cmd.Flags().GetString("integration-config")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doGoMotiveIntegrationsCreateOptions{
		BaseURL:               baseURL,
		Token:                 token,
		JSON:                  jsonOut,
		IntegrationIdentifier: integrationIdentifier,
		FriendlyName:          friendlyName,
		BrokerID:              brokerID,
		IntegrationConfigID:   integrationConfigID,
	}, nil
}
