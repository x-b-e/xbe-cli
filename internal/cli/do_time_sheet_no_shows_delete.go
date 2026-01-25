package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doTimeSheetNoShowsDeleteOptions struct {
	BaseURL string
	Token   string
	ID      string
	Confirm bool
}

func newDoTimeSheetNoShowsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a time sheet no-show",
		Long: `Delete a time sheet no-show.

Provide the no-show ID as an argument. The --confirm flag is required
to prevent accidental deletions.`,
		Example: `  # Delete a time sheet no-show
  xbe do time-sheet-no-shows delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTimeSheetNoShowsDelete,
	}
	initDoTimeSheetNoShowsDeleteFlags(cmd)
	return cmd
}

func init() {
	doTimeSheetNoShowsCmd.AddCommand(newDoTimeSheetNoShowsDeleteCmd())
}

func initDoTimeSheetNoShowsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTimeSheetNoShowsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTimeSheetNoShowsDeleteOptions(cmd, args)
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
		err := fmt.Errorf("--confirm flag is required to delete a time sheet no-show")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Delete(cmd.Context(), "/v1/time-sheet-no-shows/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted time sheet no-show %s\n", opts.ID)
	return nil
}

func parseDoTimeSheetNoShowsDeleteOptions(cmd *cobra.Command, args []string) (doTimeSheetNoShowsDeleteOptions, error) {
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTimeSheetNoShowsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
