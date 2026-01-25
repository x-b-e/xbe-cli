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

type doDeereIntegrationsUpdateOptions struct {
	BaseURL               string
	Token                 string
	JSON                  bool
	ID                    string
	IntegrationIdentifier string
	FriendlyName          string
}

func newDoDeereIntegrationsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a Deere integration",
		Long: `Update an existing Deere integration.

Provide the integration ID as an argument, then use flags to specify
which fields to update. Only specified fields will be modified.

Updatable fields:
  --integration-identifier   Integration identifier from Deere
  --friendly-name            Friendly name for the integration`,
		Example: `  # Update integration identifier
  xbe do deere-integrations update 123 --integration-identifier "deere-456"

  # Update friendly name
  xbe do deere-integrations update 123 --friendly-name "Updated Deere Account"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoDeereIntegrationsUpdate,
	}
	initDoDeereIntegrationsUpdateFlags(cmd)
	return cmd
}

func init() {
	doDeereIntegrationsCmd.AddCommand(newDoDeereIntegrationsUpdateCmd())
}

func initDoDeereIntegrationsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("integration-identifier", "", "Integration identifier from Deere")
	cmd.Flags().String("friendly-name", "", "Friendly name for the integration")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoDeereIntegrationsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoDeereIntegrationsUpdateOptions(cmd, args)
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
			"type":       "deere-integrations",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/deere-integrations/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated Deere integration %s (%s)\n", details.ID, details.IntegrationID)
	return nil
}

func parseDoDeereIntegrationsUpdateOptions(cmd *cobra.Command, args []string) (doDeereIntegrationsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	integrationIdentifier, _ := cmd.Flags().GetString("integration-identifier")
	friendlyName, _ := cmd.Flags().GetString("friendly-name")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doDeereIntegrationsUpdateOptions{
		BaseURL:               baseURL,
		Token:                 token,
		JSON:                  jsonOut,
		ID:                    args[0],
		IntegrationIdentifier: integrationIdentifier,
		FriendlyName:          friendlyName,
	}, nil
}
