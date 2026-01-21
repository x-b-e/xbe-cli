package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doCustomWorkOrderStatusesDeleteOptions struct {
	BaseURL string
	Token   string
	ID      string
	Confirm bool
}

func newDoCustomWorkOrderStatusesDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a custom work order status",
		Long: `Delete a custom work order status.

Provide the status ID as an argument. The --confirm flag is required
to prevent accidental deletions.`,
		Example: `  # Delete a custom work order status
  xbe do custom-work-order-statuses delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoCustomWorkOrderStatusesDelete,
	}
	initDoCustomWorkOrderStatusesDeleteFlags(cmd)
	return cmd
}

func init() {
	doCustomWorkOrderStatusesCmd.AddCommand(newDoCustomWorkOrderStatusesDeleteCmd())
}

func initDoCustomWorkOrderStatusesDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCustomWorkOrderStatusesDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoCustomWorkOrderStatusesDeleteOptions(cmd, args)
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
		err := fmt.Errorf("--confirm flag is required to delete a custom work order status")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Delete(cmd.Context(), "/v1/custom-work-order-statuses/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted custom work order status %s\n", opts.ID)
	return nil
}

func parseDoCustomWorkOrderStatusesDeleteOptions(cmd *cobra.Command, args []string) (doCustomWorkOrderStatusesDeleteOptions, error) {
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCustomWorkOrderStatusesDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
