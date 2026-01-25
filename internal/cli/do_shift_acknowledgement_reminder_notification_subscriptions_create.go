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

type doShiftAcknowledgementReminderNotificationSubscriptionsCreateOptions struct {
	BaseURL       string
	Token         string
	JSON          bool
	Trucker       string
	User          string
	NotifyByTxt   bool
	NotifyByEmail bool
}

func newDoShiftAcknowledgementReminderNotificationSubscriptionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a shift acknowledgement reminder notification subscription",
		Long: `Create a shift acknowledgement reminder notification subscription.

Required flags:
  --trucker  Trucker ID (required)
  --user     User ID (required)

Optional flags:
  --notify-by-txt    Notify by text (true/false)
  --notify-by-email  Notify by email (true/false)`,
		Example: `  # Subscribe a user to shift acknowledgement reminders
  xbe do shift-acknowledgement-reminder-notification-subscriptions create --trucker 123 --user 456

  # Enable text notifications
  xbe do shift-acknowledgement-reminder-notification-subscriptions create --trucker 123 --user 456 --notify-by-txt`,
		Args: cobra.NoArgs,
		RunE: runDoShiftAcknowledgementReminderNotificationSubscriptionsCreate,
	}
	initDoShiftAcknowledgementReminderNotificationSubscriptionsCreateFlags(cmd)
	return cmd
}

func init() {
	doShiftAcknowledgementReminderNotificationSubscriptionsCmd.AddCommand(newDoShiftAcknowledgementReminderNotificationSubscriptionsCreateCmd())
}

func initDoShiftAcknowledgementReminderNotificationSubscriptionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("trucker", "", "Trucker ID (required)")
	cmd.Flags().String("user", "", "User ID (required)")
	cmd.Flags().Bool("notify-by-txt", false, "Notify by text (true/false)")
	cmd.Flags().Bool("notify-by-email", false, "Notify by email (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoShiftAcknowledgementReminderNotificationSubscriptionsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoShiftAcknowledgementReminderNotificationSubscriptionsCreateOptions(cmd)
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

	if opts.Trucker == "" {
		err := fmt.Errorf("--trucker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.User == "" {
		err := fmt.Errorf("--user is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"trucker": map[string]any{
			"data": map[string]any{
				"type": "truckers",
				"id":   opts.Trucker,
			},
		},
		"user": map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.User,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "shift-acknowledgement-reminder-notification-subscriptions",
			"relationships": relationships,
		},
	}

	attributes := map[string]any{}
	if cmd.Flags().Changed("notify-by-txt") {
		attributes["notify-by-txt"] = opts.NotifyByTxt
	}
	if cmd.Flags().Changed("notify-by-email") {
		attributes["notify-by-email"] = opts.NotifyByEmail
	}
	if len(attributes) > 0 {
		requestBody["data"].(map[string]any)["attributes"] = attributes
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/shift-acknowledgement-reminder-notification-subscriptions", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created shift acknowledgement reminder notification subscription %s\n", row.ID)
	return nil
}

func parseDoShiftAcknowledgementReminderNotificationSubscriptionsCreateOptions(cmd *cobra.Command) (doShiftAcknowledgementReminderNotificationSubscriptionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	trucker, _ := cmd.Flags().GetString("trucker")
	user, _ := cmd.Flags().GetString("user")
	notifyByTxt, _ := cmd.Flags().GetBool("notify-by-txt")
	notifyByEmail, _ := cmd.Flags().GetBool("notify-by-email")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doShiftAcknowledgementReminderNotificationSubscriptionsCreateOptions{
		BaseURL:       baseURL,
		Token:         token,
		JSON:          jsonOut,
		Trucker:       trucker,
		User:          user,
		NotifyByTxt:   notifyByTxt,
		NotifyByEmail: notifyByEmail,
	}, nil
}
