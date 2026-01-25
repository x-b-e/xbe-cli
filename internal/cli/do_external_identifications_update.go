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

type doExternalIdentificationsUpdateOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	ID                  string
	Value               string
	SkipValueValidation bool
}

func newDoExternalIdentificationsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an external identification",
		Long: `Update an external identification.

Optional flags:
  --value                    The external identification value
  --skip-value-validation    Skip format and uniqueness validation`,
		Example: `  # Update value
  xbe do external-identifications update 123 --value "DL-99999"

  # Update with validation skipped
  xbe do external-identifications update 123 --value "TEMP-ID" --skip-value-validation`,
		Args: cobra.ExactArgs(1),
		RunE: runDoExternalIdentificationsUpdate,
	}
	initDoExternalIdentificationsUpdateFlags(cmd)
	return cmd
}

func init() {
	doExternalIdentificationsCmd.AddCommand(newDoExternalIdentificationsUpdateCmd())
}

func initDoExternalIdentificationsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("value", "", "The external identification value")
	cmd.Flags().Bool("skip-value-validation", false, "Skip format and uniqueness validation")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoExternalIdentificationsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoExternalIdentificationsUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("value") {
		attributes["value"] = opts.Value
	}
	if cmd.Flags().Changed("skip-value-validation") {
		attributes["skip-value-validation"] = opts.SkipValueValidation
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "external-identifications",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/external-identifications/"+opts.ID, jsonBody)
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

	row := buildExternalIdentificationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated external identification %s\n", row.ID)
	return nil
}

func parseDoExternalIdentificationsUpdateOptions(cmd *cobra.Command, args []string) (doExternalIdentificationsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	value, _ := cmd.Flags().GetString("value")
	skipValueValidation, _ := cmd.Flags().GetBool("skip-value-validation")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doExternalIdentificationsUpdateOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		ID:                  args[0],
		Value:               value,
		SkipValueValidation: skipValueValidation,
	}, nil
}
