package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doAdministrativeIncidentsDeleteOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	Confirm bool
}

func newDoAdministrativeIncidentsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an administrative incident",
		Long: `Delete an administrative incident.

Requires the --confirm flag to prevent accidental deletion.

Arguments:
  <id>    The administrative incident ID (required)

Flags:
  --confirm    Required flag to confirm deletion`,
		Example: `  # Delete an administrative incident
  xbe do administrative-incidents delete 123 --confirm

  # Delete and return JSON
  xbe do administrative-incidents delete 123 --confirm --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoAdministrativeIncidentsDelete,
	}
	initDoAdministrativeIncidentsDeleteFlags(cmd)
	return cmd
}

func init() {
	doAdministrativeIncidentsCmd.AddCommand(newDoAdministrativeIncidentsDeleteCmd())
}

func initDoAdministrativeIncidentsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoAdministrativeIncidentsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoAdministrativeIncidentsDeleteOptions(cmd, args)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if !opts.Confirm {
		err := fmt.Errorf("--confirm flag is required for deletion")
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

	client := api.NewClient(opts.BaseURL, opts.Token)

	path := fmt.Sprintf("/v1/administrative-incidents/%s", opts.ID)
	body, _, err := client.Delete(cmd.Context(), path)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), map[string]any{
			"id":      opts.ID,
			"deleted": true,
		})
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted administrative incident %s\n", opts.ID)
	return nil
}

func parseDoAdministrativeIncidentsDeleteOptions(cmd *cobra.Command, args []string) (doAdministrativeIncidentsDeleteOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doAdministrativeIncidentsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
