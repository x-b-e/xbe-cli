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

type doApplicationSettingsCreateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	Key         string
	Value       string
	Description string
}

func newDoApplicationSettingsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an application setting",
		Long: `Create a new application setting.

Required flags:
  --key     Setting key (required)
  --value   Setting value (required)

Optional flags:
  --description   Setting description`,
		Example: `  # Create a new application setting
  xbe do application-settings create --key "FEATURE_FLAG" --value "true"

  # Create with description
  xbe do application-settings create --key "SYNC_TIMEOUT" --value "30" --description "Timeout in seconds"

  # Get JSON output
  xbe do application-settings create --key "EXAMPLE" --value "1" --json`,
		Args: cobra.NoArgs,
		RunE: runDoApplicationSettingsCreate,
	}
	initDoApplicationSettingsCreateFlags(cmd)
	return cmd
}

func init() {
	doApplicationSettingsCmd.AddCommand(newDoApplicationSettingsCreateCmd())
}

func initDoApplicationSettingsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("key", "", "Setting key (required)")
	cmd.Flags().String("value", "", "Setting value (required)")
	cmd.Flags().String("description", "", "Setting description")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoApplicationSettingsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoApplicationSettingsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.Key) == "" {
		err := fmt.Errorf("--key is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.Value) == "" {
		err := fmt.Errorf("--value is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"key":   opts.Key,
		"value": opts.Value,
	}
	if opts.Description != "" {
		attributes["description"] = opts.Description
	}

	requestBody := map[string]any{
		"data": map[string]any{
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

	body, _, err := client.Post(cmd.Context(), "/v1/application-settings", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created application setting %s (%s)\n", row.ID, row.Key)
	return nil
}

func parseDoApplicationSettingsCreateOptions(cmd *cobra.Command) (doApplicationSettingsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	key, _ := cmd.Flags().GetString("key")
	value, _ := cmd.Flags().GetString("value")
	description, _ := cmd.Flags().GetString("description")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doApplicationSettingsCreateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		Key:         key,
		Value:       value,
		Description: description,
	}, nil
}
