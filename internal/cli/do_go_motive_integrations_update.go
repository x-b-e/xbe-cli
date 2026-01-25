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

type doGoMotiveIntegrationsUpdateOptions struct {
	BaseURL               string
	Token                 string
	JSON                  bool
	ID                    string
	IntegrationIdentifier string
	FriendlyName          string
}

func newDoGoMotiveIntegrationsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a GoMotive integration",
		Long: `Update GoMotive integration attributes.

Optional flags:
  --integration-identifier  GoMotive integration identifier
  --friendly-name           Friendly name`,
		Example: `  # Update the friendly name
  xbe do go-motive-integrations update 123 --friendly-name "New Name"

  # Update identifier and name
  xbe do go-motive-integrations update 123 \
    --integration-identifier "motive-456" \
    --friendly-name "Updated GoMotive"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoGoMotiveIntegrationsUpdate,
	}
	initDoGoMotiveIntegrationsUpdateFlags(cmd)
	return cmd
}

func init() {
	doGoMotiveIntegrationsCmd.AddCommand(newDoGoMotiveIntegrationsUpdateCmd())
}

func initDoGoMotiveIntegrationsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("integration-identifier", "", "GoMotive integration identifier")
	cmd.Flags().String("friendly-name", "", "Friendly name")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoGoMotiveIntegrationsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoGoMotiveIntegrationsUpdateOptions(cmd, args)
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
			"type":       "go-motive-integrations",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/go-motive-integrations/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated GoMotive integration %s\n", row.ID)
	return nil
}

func parseDoGoMotiveIntegrationsUpdateOptions(cmd *cobra.Command, args []string) (doGoMotiveIntegrationsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	integrationIdentifier, _ := cmd.Flags().GetString("integration-identifier")
	friendlyName, _ := cmd.Flags().GetString("friendly-name")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doGoMotiveIntegrationsUpdateOptions{
		BaseURL:               baseURL,
		Token:                 token,
		JSON:                  jsonOut,
		ID:                    args[0],
		IntegrationIdentifier: integrationIdentifier,
		FriendlyName:          friendlyName,
	}, nil
}
