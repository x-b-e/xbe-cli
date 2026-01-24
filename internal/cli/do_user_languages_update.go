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

type doUserLanguagesUpdateOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	ID        string
	IsDefault string
}

func newDoUserLanguagesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a user language",
		Long: `Update a user language preference.

Optional flags:
  --is-default  Default language flag (true/false)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Mark as default
  xbe do user-languages update 123 --is-default true

  # Unset default
  xbe do user-languages update 123 --is-default false`,
		Args: cobra.ExactArgs(1),
		RunE: runDoUserLanguagesUpdate,
	}
	initDoUserLanguagesUpdateFlags(cmd)
	return cmd
}

func init() {
	doUserLanguagesCmd.AddCommand(newDoUserLanguagesUpdateCmd())
}

func initDoUserLanguagesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("is-default", "", "Default language flag (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoUserLanguagesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoUserLanguagesUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("is-default") {
		attributes["is-default"] = opts.IsDefault == "true"
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "user-languages",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/user-languages/"+opts.ID, jsonBody)
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

	row := buildUserLanguageRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated user language %s\n", row.ID)
	return nil
}

func parseDoUserLanguagesUpdateOptions(cmd *cobra.Command, args []string) (doUserLanguagesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	isDefault, _ := cmd.Flags().GetString("is-default")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doUserLanguagesUpdateOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		ID:        args[0],
		IsDefault: isDefault,
	}, nil
}
