package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doShiftAcknowledgementReminderNotificationSubscriptionsDeleteOptions struct {
	BaseURL string
	Token   string
	ID      string
	Confirm bool
}

func newDoShiftAcknowledgementReminderNotificationSubscriptionsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a shift acknowledgement reminder notification subscription",
		Long: `Delete a shift acknowledgement reminder notification subscription.

Provide the subscription ID as an argument. The --confirm flag is required
for destructive actions.`,
		Example: `  # Delete a subscription
  xbe do shift-acknowledgement-reminder-notification-subscriptions delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoShiftAcknowledgementReminderNotificationSubscriptionsDelete,
	}
	initDoShiftAcknowledgementReminderNotificationSubscriptionsDeleteFlags(cmd)
	return cmd
}

func init() {
	doShiftAcknowledgementReminderNotificationSubscriptionsCmd.AddCommand(newDoShiftAcknowledgementReminderNotificationSubscriptionsDeleteCmd())
}

func initDoShiftAcknowledgementReminderNotificationSubscriptionsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoShiftAcknowledgementReminderNotificationSubscriptionsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoShiftAcknowledgementReminderNotificationSubscriptionsDeleteOptions(cmd, args)
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
		err := fmt.Errorf("--confirm flag is required to delete a shift acknowledgement reminder notification subscription")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Delete(cmd.Context(), "/v1/shift-acknowledgement-reminder-notification-subscriptions/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted shift acknowledgement reminder notification subscription %s\n", opts.ID)
	return nil
}

func parseDoShiftAcknowledgementReminderNotificationSubscriptionsDeleteOptions(cmd *cobra.Command, args []string) (doShiftAcknowledgementReminderNotificationSubscriptionsDeleteOptions, error) {
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doShiftAcknowledgementReminderNotificationSubscriptionsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
