package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doProjectsDeleteOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	Confirm bool
}

func newDoProjectsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a project",
		Long: `Delete a project.

The --confirm flag is required to prevent accidental deletion.

Arguments:
  <id>    The project ID (required)

Flags:
  --confirm    Required flag to confirm deletion`,
		Example: `  # Delete a project
  xbe do projects delete 123 --confirm

  # Get JSON output
  xbe do projects delete 123 --confirm --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProjectsDelete,
	}
	initDoProjectsDeleteFlags(cmd)
	return cmd
}

func init() {
	doProjectsCmd.AddCommand(newDoProjectsDeleteCmd())
}

func initDoProjectsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectsDeleteOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Require --confirm flag
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
		return fmt.Errorf("project id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Delete(cmd.Context(), "/v1/projects/"+id)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), map[string]any{
			"id":      id,
			"deleted": true,
		})
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted project %s\n", id)
	return nil
}

func parseDoProjectsDeleteOptions(cmd *cobra.Command) (doProjectsDeleteOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		Confirm: confirm,
	}, nil
}
