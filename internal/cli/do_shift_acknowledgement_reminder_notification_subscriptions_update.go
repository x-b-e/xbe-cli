package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doShiftAcknowledgementReminderNotificationSubscriptionsUpdateOptions struct {
	BaseURL       string
	Token         string
	JSON          bool
	ID            string
	NotifyByTxt   bool
	NotifyByEmail bool
}

func newDoShiftAcknowledgementReminderNotificationSubscriptionsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a shift acknowledgement reminder notification subscription",
		Long: `Update an existing shift acknowledgement reminder notification subscription.

Optional flags:
  --notify-by-txt    Notify by text (true/false)
  --notify-by-email  Notify by email (true/false)`,
		Example: `  # Enable text notifications
  xbe do shift-acknowledgement-reminder-notification-subscriptions update 123 --notify-by-txt

  # Disable email notifications
  xbe do shift-acknowledgement-reminder-notification-subscriptions update 123 --notify-by-email=false`,
		Args: cobra.ExactArgs(1),
		RunE: runDoShiftAcknowledgementReminderNotificationSubscriptionsUpdate,
	}
	initDoShiftAcknowledgementReminderNotificationSubscriptionsUpdateFlags(cmd)
	return cmd
}

func init() {
	doShiftAcknowledgementReminderNotificationSubscriptionsCmd.AddCommand(newDoShiftAcknowledgementReminderNotificationSubscriptionsUpdateCmd())
}

func initDoShiftAcknowledgementReminderNotificationSubscriptionsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("notify-by-txt", false, "Notify by text (true/false)")
	cmd.Flags().Bool("notify-by-email", false, "Notify by email (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoShiftAcknowledgementReminderNotificationSubscriptionsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoShiftAcknowledgementReminderNotificationSubscriptionsUpdateOptions(cmd, args)
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

	attributes := map[string]any{}
	if cmd.Flags().Changed("notify-by-txt") {
		attributes["notify-by-txt"] = opts.NotifyByTxt
	}
	if cmd.Flags().Changed("notify-by-email") {
		attributes["notify-by-email"] = opts.NotifyByEmail
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update; specify at least one field flag")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "shift-acknowledgement-reminder-notification-subscriptions",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/shift-acknowledgement-reminder-notification-subscriptions/"+opts.ID, jsonBody)
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

	row := shiftAcknowledgementReminderNotificationSubscriptionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated shift acknowledgement reminder notification subscription %s\n", row.ID)
	return nil
}

func parseDoShiftAcknowledgementReminderNotificationSubscriptionsUpdateOptions(cmd *cobra.Command, args []string) (doShiftAcknowledgementReminderNotificationSubscriptionsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	notifyByTxt, _ := cmd.Flags().GetBool("notify-by-txt")
	notifyByEmail, _ := cmd.Flags().GetBool("notify-by-email")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doShiftAcknowledgementReminderNotificationSubscriptionsUpdateOptions{
		BaseURL:       baseURL,
		Token:         token,
		JSON:          jsonOut,
		ID:            args[0],
		NotifyByTxt:   notifyByTxt,
		NotifyByEmail: notifyByEmail,
	}, nil
}
