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

type doApplicationSettingsUpdateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	Value       string
	Description string
}

func newDoApplicationSettingsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an application setting",
		Long: `Update an existing application setting.

Only the fields you specify will be updated. Fields not provided will remain unchanged.

Arguments:
  <id>    The application setting ID (required)

Flags:
  --value         Update the setting value
  --description   Update the setting description`,
		Example: `  # Update the value
  xbe do application-settings update 123 --value "false"

  # Update the description
  xbe do application-settings update 123 --description "Updated description"

  # Get JSON output
  xbe do application-settings update 123 --value "false" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoApplicationSettingsUpdate,
	}
	initDoApplicationSettingsUpdateFlags(cmd)
	return cmd
}

func init() {
	doApplicationSettingsCmd.AddCommand(newDoApplicationSettingsUpdateCmd())
}

func initDoApplicationSettingsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("value", "", "New setting value")
	cmd.Flags().String("description", "", "New setting description")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoApplicationSettingsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoApplicationSettingsUpdateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Require authentication for write operations
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("application setting id is required")
	}

	if opts.Value == "" && opts.Description == "" {
		err := fmt.Errorf("at least one of --value or --description is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.Value != "" {
		attributes["value"] = opts.Value
	}
	if opts.Description != "" {
		attributes["description"] = opts.Description
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"id":         id,
			"type":       "application-settings",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/application-settings/"+id, jsonBody)
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

	row := buildApplicationSettingRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated application setting %s (%s)\n", row.ID, row.Key)
	return nil
}

func parseDoApplicationSettingsUpdateOptions(cmd *cobra.Command) (doApplicationSettingsUpdateOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return doApplicationSettingsUpdateOptions{}, err
	}
	value, err := cmd.Flags().GetString("value")
	if err != nil {
		return doApplicationSettingsUpdateOptions{}, err
	}
	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return doApplicationSettingsUpdateOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return doApplicationSettingsUpdateOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return doApplicationSettingsUpdateOptions{}, err
	}

	return doApplicationSettingsUpdateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		Value:       value,
		Description: description,
	}, nil
}
