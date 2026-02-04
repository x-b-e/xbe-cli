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

type notificationDeliveryDecisionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type notificationDeliveryDecisionDetails struct {
	ID                 string `json:"id"`
	NotificationID     string `json:"notification_id,omitempty"`
	NotifyByValueMin   string `json:"notify_by_value_min,omitempty"`
	NotifyByEmailValue string `json:"notify_by_email_value,omitempty"`
	NotifyByTxtValue   string `json:"notify_by_txt_value,omitempty"`
	DeliverAt          string `json:"deliver_at,omitempty"`
	NotifyByEmail      bool   `json:"notify_by_email"`
	NotifyByTxt        bool   `json:"notify_by_txt"`
}

func newNotificationDeliveryDecisionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show notification delivery decision details",
		Long: `Show the full details of a notification delivery decision.

Output Fields:
  ID                 Decision identifier
  Notification ID    Related notification
  Notify By Min      Minimum value required for delivery
  Notify By Email    Value if delivered by email
  Notify By Txt      Value if delivered by txt
  Deliver At         Scheduled delivery time
  Email Enabled      Whether email delivery is enabled
  Txt Enabled        Whether txt delivery is enabled

Arguments:
  <id>               The notification delivery decision ID (required).`,
		Example: `  # View a decision by ID
  xbe view notification-delivery-decisions show 123

  # Get decision details as JSON
  xbe view notification-delivery-decisions show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runNotificationDeliveryDecisionsShow,
	}
	initNotificationDeliveryDecisionsShowFlags(cmd)
	return cmd
}

func init() {
	notificationDeliveryDecisionsCmd.AddCommand(newNotificationDeliveryDecisionsShowCmd())
}

func initNotificationDeliveryDecisionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runNotificationDeliveryDecisionsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseNotificationDeliveryDecisionsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
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
		return fmt.Errorf("notification delivery decision id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[notification-delivery-decisions]", strings.Join([]string{
		"notify-by-value-min",
		"notify-by-email-value",
		"notify-by-txt-value",
		"deliver-at",
		"notify-by-email",
		"notify-by-txt",
		"notification",
	}, ","))

	body, _, err := client.Get(cmd.Context(), "/v1/notification-delivery-decisions/"+id, query)
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

	details := buildNotificationDeliveryDecisionDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderNotificationDeliveryDecisionDetails(cmd, details)
}

func parseNotificationDeliveryDecisionsShowOptions(cmd *cobra.Command) (notificationDeliveryDecisionsShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return notificationDeliveryDecisionsShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return notificationDeliveryDecisionsShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return notificationDeliveryDecisionsShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return notificationDeliveryDecisionsShowOptions{}, err
	}

	return notificationDeliveryDecisionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildNotificationDeliveryDecisionDetails(resp jsonAPISingleResponse) notificationDeliveryDecisionDetails {
	attrs := resp.Data.Attributes
	details := notificationDeliveryDecisionDetails{
		ID:                 resp.Data.ID,
		NotifyByValueMin:   stringAttr(attrs, "notify-by-value-min"),
		NotifyByEmailValue: stringAttr(attrs, "notify-by-email-value"),
		NotifyByTxtValue:   stringAttr(attrs, "notify-by-txt-value"),
		DeliverAt:          formatDateTime(stringAttr(attrs, "deliver-at")),
		NotifyByEmail:      boolAttr(attrs, "notify-by-email"),
		NotifyByTxt:        boolAttr(attrs, "notify-by-txt"),
	}

	if rel, ok := resp.Data.Relationships["notification"]; ok && rel.Data != nil {
		details.NotificationID = rel.Data.ID
	}

	return details
}

func renderNotificationDeliveryDecisionDetails(cmd *cobra.Command, details notificationDeliveryDecisionDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.NotificationID != "" {
		fmt.Fprintf(out, "Notification ID: %s\n", details.NotificationID)
	}
	if details.NotifyByValueMin != "" {
		fmt.Fprintf(out, "Notify By Min: %s\n", details.NotifyByValueMin)
	}
	if details.NotifyByEmailValue != "" {
		fmt.Fprintf(out, "Notify By Email: %s\n", details.NotifyByEmailValue)
	}
	if details.NotifyByTxtValue != "" {
		fmt.Fprintf(out, "Notify By Txt: %s\n", details.NotifyByTxtValue)
	}
	if details.DeliverAt != "" {
		fmt.Fprintf(out, "Deliver At: %s\n", details.DeliverAt)
	}
	fmt.Fprintf(out, "Email Enabled: %s\n", formatBool(details.NotifyByEmail))
	fmt.Fprintf(out, "Txt Enabled: %s\n", formatBool(details.NotifyByTxt))

	return nil
}
