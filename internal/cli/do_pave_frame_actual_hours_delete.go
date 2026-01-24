package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doPaveFrameActualHoursDeleteOptions struct {
	BaseURL string
	Token   string
	ID      string
	Confirm bool
}

func newDoPaveFrameActualHoursDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a pave frame actual hour",
		Long: `Delete a pave frame actual hour.

Provide the record ID as an argument. The --confirm flag is required
to prevent accidental deletions.

Note: Only admin users can delete pave frame actual hours.`,
		Example: `  # Delete a pave frame actual hour
  xbe do pave-frame-actual-hours delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoPaveFrameActualHoursDelete,
	}
	initDoPaveFrameActualHoursDeleteFlags(cmd)
	return cmd
}

func init() {
	doPaveFrameActualHoursCmd.AddCommand(newDoPaveFrameActualHoursDeleteCmd())
}

func initDoPaveFrameActualHoursDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoPaveFrameActualHoursDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoPaveFrameActualHoursDeleteOptions(cmd, args)
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
		err := fmt.Errorf("--confirm flag is required to delete a pave frame actual hour")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Delete(cmd.Context(), "/v1/pave-frame-actual-hours/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted pave frame actual hour %s\n", opts.ID)
	return nil
}

func parseDoPaveFrameActualHoursDeleteOptions(cmd *cobra.Command, args []string) (doPaveFrameActualHoursDeleteOptions, error) {
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doPaveFrameActualHoursDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
