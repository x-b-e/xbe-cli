package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type shiftAcknowledgementReminderNotificationSubscriptionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

func newShiftAcknowledgementReminderNotificationSubscriptionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show shift acknowledgement reminder notification subscription details",
		Long: `Show the full details of a shift acknowledgement reminder notification subscription.

Includes the associated trucker and user information.

Arguments:
  <id>  The subscription ID (required).`,
		Example: `  # Show a subscription
  xbe view shift-acknowledgement-reminder-notification-subscriptions show 123

  # Output as JSON
  xbe view shift-acknowledgement-reminder-notification-subscriptions show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runShiftAcknowledgementReminderNotificationSubscriptionsShow,
	}
	initShiftAcknowledgementReminderNotificationSubscriptionsShowFlags(cmd)
	return cmd
}

func init() {
	shiftAcknowledgementReminderNotificationSubscriptionsCmd.AddCommand(newShiftAcknowledgementReminderNotificationSubscriptionsShowCmd())
}

func initShiftAcknowledgementReminderNotificationSubscriptionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runShiftAcknowledgementReminderNotificationSubscriptionsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseShiftAcknowledgementReminderNotificationSubscriptionsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("subscription id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[shift-acknowledgement-reminder-notification-subscriptions]", "trucker,user,notify-by-txt,notify-by-email")
	query.Set("include", "trucker,user")
	query.Set("fields[truckers]", "company-name")
	query.Set("fields[users]", "name,email-address")

	body, _, err := client.Get(cmd.Context(), "/v1/shift-acknowledgement-reminder-notification-subscriptions/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := shiftAcknowledgementReminderNotificationSubscriptionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderShiftAcknowledgementReminderNotificationSubscriptionDetails(cmd, details)
}

func parseShiftAcknowledgementReminderNotificationSubscriptionsShowOptions(cmd *cobra.Command) (shiftAcknowledgementReminderNotificationSubscriptionsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return shiftAcknowledgementReminderNotificationSubscriptionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func renderShiftAcknowledgementReminderNotificationSubscriptionDetails(cmd *cobra.Command, details shiftAcknowledgementReminderNotificationSubscriptionRow) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TruckerID != "" {
		fmt.Fprintf(out, "Trucker ID: %s\n", details.TruckerID)
	}
	if details.TruckerName != "" {
		fmt.Fprintf(out, "Trucker Name: %s\n", details.TruckerName)
	}
	if details.UserID != "" {
		fmt.Fprintf(out, "User ID: %s\n", details.UserID)
	}
	if details.UserName != "" {
		fmt.Fprintf(out, "User Name: %s\n", details.UserName)
	}
	if details.UserEmail != "" {
		fmt.Fprintf(out, "User Email: %s\n", details.UserEmail)
	}
	fmt.Fprintf(out, "Notify By Txt: %s\n", formatBool(details.NotifyByTxt))
	fmt.Fprintf(out, "Notify By Email: %s\n", formatBool(details.NotifyByEmail))

	return nil
}
