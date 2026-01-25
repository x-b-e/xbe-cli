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

type doApplicationSettingsDeleteOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	Confirm bool
}

func newDoApplicationSettingsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an application setting",
		Long: `Delete an application setting.

Note: The server does not allow application settings to be deleted. This
command will return an error unless the server policy changes.

Arguments:
  <id>    The application setting ID (required)

Flags:
  --confirm    Required flag to confirm deletion`,
		Example: `  # Attempt to delete an application setting
  xbe do application-settings delete 123 --confirm

  # Get JSON output of the record (if deletion succeeds)
  xbe do application-settings delete 123 --confirm --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoApplicationSettingsDelete,
	}
	initDoApplicationSettingsDeleteFlags(cmd)
	return cmd
}

func init() {
	doApplicationSettingsCmd.AddCommand(newDoApplicationSettingsDeleteCmd())
}

func initDoApplicationSettingsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoApplicationSettingsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoApplicationSettingsDeleteOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if !opts.Confirm {
		err := fmt.Errorf("--confirm flag is required for deletion")
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

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[application-settings]", "key,value,description")

	getBody, _, err := client.Get(cmd.Context(), "/v1/application-settings/"+id, query)
	if err != nil {
		if len(getBody) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(getBody))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPISingleResponse
	if err := json.Unmarshal(getBody, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	row := buildApplicationSettingRowFromSingle(resp)

	deleteBody, _, err := client.Delete(cmd.Context(), "/v1/application-settings/"+id)
	if err != nil {
		if len(deleteBody) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(deleteBody))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted application setting %s (%s)\n", row.ID, row.Key)
	return nil
}

func parseDoApplicationSettingsDeleteOptions(cmd *cobra.Command) (doApplicationSettingsDeleteOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return doApplicationSettingsDeleteOptions{}, err
	}
	confirm, err := cmd.Flags().GetBool("confirm")
	if err != nil {
		return doApplicationSettingsDeleteOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return doApplicationSettingsDeleteOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return doApplicationSettingsDeleteOptions{}, err
	}

	return doApplicationSettingsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		Confirm: confirm,
	}, nil
}
