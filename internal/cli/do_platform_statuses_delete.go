package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doPlatformStatusesDeleteOptions struct {
	BaseURL string
	Token   string
	ID      string
	Confirm bool
}

func newDoPlatformStatusesDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a platform status",
		Long: `Delete a platform status.

Provide the status ID as an argument. The --confirm flag is required
to prevent accidental deletions.`,
		Example: `  # Delete a platform status
  xbe do platform-statuses delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoPlatformStatusesDelete,
	}
	initDoPlatformStatusesDeleteFlags(cmd)
	return cmd
}

func init() {
	doPlatformStatusesCmd.AddCommand(newDoPlatformStatusesDeleteCmd())
}

func initDoPlatformStatusesDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoPlatformStatusesDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoPlatformStatusesDeleteOptions(cmd, args)
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
		err := fmt.Errorf("--confirm flag is required to delete a platform status")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Delete(cmd.Context(), "/v1/platform-statuses/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted platform status %s\n", opts.ID)
	return nil
}

func parseDoPlatformStatusesDeleteOptions(cmd *cobra.Command, args []string) (doPlatformStatusesDeleteOptions, error) {
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doPlatformStatusesDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
