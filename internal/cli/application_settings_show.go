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

type applicationSettingsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

func newApplicationSettingsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show application setting details",
		Long: `Show the full details of a specific application setting.

Application settings are global key/value pairs used to configure platform behavior.
Access is restricted to admin users.

Output Fields:
  ID           Setting identifier
  KEY          Setting key
  VALUE        Setting value
  DESCRIPTION  Setting description (if present)

Arguments:
  <id>    The application setting ID (required). You can find IDs using the list command.`,
		Example: `  # View an application setting
  xbe view application-settings show 123

  # Get JSON output
  xbe view application-settings show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runApplicationSettingsShow,
	}
	initApplicationSettingsShowFlags(cmd)
	return cmd
}

func init() {
	applicationSettingsCmd.AddCommand(newApplicationSettingsShowCmd())
}

func initApplicationSettingsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runApplicationSettingsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseApplicationSettingsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
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

	body, _, err := client.Get(cmd.Context(), "/v1/application-settings/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	row := buildApplicationSettingRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	return renderApplicationSettingDetails(cmd, row)
}

func parseApplicationSettingsShowOptions(cmd *cobra.Command) (applicationSettingsShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return applicationSettingsShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return applicationSettingsShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return applicationSettingsShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return applicationSettingsShowOptions{}, err
	}

	return applicationSettingsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func renderApplicationSettingDetails(cmd *cobra.Command, row applicationSettingRow) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", row.ID)
	if row.Key != "" {
		fmt.Fprintf(out, "Key: %s\n", row.Key)
	}
	if row.Value != "" {
		fmt.Fprintf(out, "Value: %s\n", row.Value)
	}
	if row.Description != "" {
		fmt.Fprintf(out, "Description: %s\n", row.Description)
	}

	return nil
}
