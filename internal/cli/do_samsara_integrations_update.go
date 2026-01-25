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

type doSamsaraIntegrationsUpdateOptions struct {
	BaseURL               string
	Token                 string
	JSON                  bool
	ID                    string
	IntegrationIdentifier string
	FriendlyName          string
}

func newDoSamsaraIntegrationsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a Samsara integration",
		Long: `Update an existing Samsara integration.

Provide the integration ID as an argument, then use flags to specify
which fields to update. Only specified fields will be modified.

Updatable fields:
  --integration-identifier   Integration identifier from Samsara
  --friendly-name            Friendly name for the integration

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update integration identifier
  xbe do samsara-integrations update 123 --integration-identifier "samsara-456"

  # Update friendly name
  xbe do samsara-integrations update 123 --friendly-name "Updated Samsara Account"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoSamsaraIntegrationsUpdate,
	}
	initDoSamsaraIntegrationsUpdateFlags(cmd)
	return cmd
}

func init() {
	doSamsaraIntegrationsCmd.AddCommand(newDoSamsaraIntegrationsUpdateCmd())
}

func initDoSamsaraIntegrationsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("integration-identifier", "", "Integration identifier from Samsara")
	cmd.Flags().String("friendly-name", "", "Friendly name for the integration")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoSamsaraIntegrationsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoSamsaraIntegrationsUpdateOptions(cmd, args)
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

	attributes := map[string]any{}
	if cmd.Flags().Changed("integration-identifier") {
		attributes["integration-identifier"] = opts.IntegrationIdentifier
	}
	if cmd.Flags().Changed("friendly-name") {
		attributes["friendly-name"] = opts.FriendlyName
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update; specify --integration-identifier or --friendly-name")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "samsara-integrations",
			"id":         opts.ID,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/samsara-integrations/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated Samsara integration %s (%s)\n", details.ID, details.IntegrationID)
	return nil
}

func parseDoSamsaraIntegrationsUpdateOptions(cmd *cobra.Command, args []string) (doSamsaraIntegrationsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	integrationIdentifier, _ := cmd.Flags().GetString("integration-identifier")
	friendlyName, _ := cmd.Flags().GetString("friendly-name")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doSamsaraIntegrationsUpdateOptions{
		BaseURL:               baseURL,
		Token:                 token,
		JSON:                  jsonOut,
		ID:                    args[0],
		IntegrationIdentifier: integrationIdentifier,
		FriendlyName:          friendlyName,
	}, nil
}
