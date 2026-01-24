package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doActionItemTrackerUpdateRequestsDeleteOptions struct {
	BaseURL string
	Token   string
	ID      string
	Confirm bool
}

func newDoActionItemTrackerUpdateRequestsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an action item tracker update request",
		Long: `Delete an action item tracker update request.

Provide the update request ID as an argument. The --confirm flag is required
to prevent accidental deletions.`,
		Example: `  # Delete an update request
  xbe do action-item-tracker-update-requests delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoActionItemTrackerUpdateRequestsDelete,
	}
	initDoActionItemTrackerUpdateRequestsDeleteFlags(cmd)
	return cmd
}

func init() {
	doActionItemTrackerUpdateRequestsCmd.AddCommand(newDoActionItemTrackerUpdateRequestsDeleteCmd())
}

func initDoActionItemTrackerUpdateRequestsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoActionItemTrackerUpdateRequestsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoActionItemTrackerUpdateRequestsDeleteOptions(cmd, args)
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

	if !opts.Confirm {
		err := fmt.Errorf("--confirm flag is required to delete an action item tracker update request")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Delete(cmd.Context(), "/v1/action-item-tracker-update-requests/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted action item tracker update request %s\n", opts.ID)
	return nil
}

func parseDoActionItemTrackerUpdateRequestsDeleteOptions(cmd *cobra.Command, args []string) (doActionItemTrackerUpdateRequestsDeleteOptions, error) {
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doActionItemTrackerUpdateRequestsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
