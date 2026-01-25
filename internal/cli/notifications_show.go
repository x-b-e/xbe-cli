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

type notificationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type notificationDetails struct {
	ID                       string `json:"id"`
	UserID                   string `json:"user_id,omitempty"`
	UserName                 string `json:"user_name,omitempty"`
	UserEmail                string `json:"user_email,omitempty"`
	Read                     bool   `json:"read"`
	IsReadyForDelivery       bool   `json:"is_ready_for_delivery"`
	DeliverAt                string `json:"deliver_at,omitempty"`
	DeliveryDecisionApproach string `json:"delivery_decision_approach,omitempty"`
	NotificationType         string `json:"notification_type,omitempty"`
	CreatedAt                string `json:"created_at,omitempty"`
	UpdatedAt                string `json:"updated_at,omitempty"`
	Details                  any    `json:"details,omitempty"`
}

func newNotificationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show notification details",
		Long: `Show the full details of a notification.

Output Fields:
  ID        Notification identifier
  User      Recipient information
  Read      Whether the notification has been read
  Ready     Ready for delivery status
  Deliver   Scheduled delivery time
  Approach  Delivery decision approach
  Type      Notification type
  Created   Created at time
  Updated   Updated at time
  Details   Notification payload details

Arguments:
  <id>  The notification ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a notification
  xbe view notifications show 123

  # Output as JSON
  xbe view notifications show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runNotificationsShow,
	}
	initNotificationsShowFlags(cmd)
	return cmd
}

func init() {
	notificationsCmd.AddCommand(newNotificationsShowCmd())
}

func initNotificationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runNotificationsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseNotificationsShowOptions(cmd)
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
		return fmt.Errorf("notification id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[notifications]", "read,details,deliver-at,is-ready-for-delivery,delivery-decision-approach,notification-type,created-at,updated-at,user")
	query.Set("include", "user")
	query.Set("fields[users]", "name,email-address")

	body, _, err := client.Get(cmd.Context(), "/v1/notifications/"+id, query)
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

	details := buildNotificationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderNotificationDetails(cmd, details)
}

func parseNotificationsShowOptions(cmd *cobra.Command) (notificationsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return notificationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildNotificationDetails(resp jsonAPISingleResponse) notificationDetails {
	row := notificationRowFromSingle(resp)

	details := notificationDetails{
		ID:                       row.ID,
		UserID:                   row.UserID,
		UserName:                 row.UserName,
		UserEmail:                row.UserEmail,
		Read:                     row.Read,
		IsReadyForDelivery:       row.IsReadyForDelivery,
		DeliverAt:                row.DeliverAt,
		DeliveryDecisionApproach: row.DeliveryDecisionApproach,
		NotificationType:         row.NotificationType,
		CreatedAt:                row.CreatedAt,
		UpdatedAt:                row.UpdatedAt,
		Details:                  anyAttr(resp.Data.Attributes, "details"),
	}

	return details
}

func renderNotificationDetails(cmd *cobra.Command, details notificationDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.UserID != "" {
		fmt.Fprintf(out, "User ID: %s\n", details.UserID)
	}
	if details.UserName != "" {
		fmt.Fprintf(out, "User Name: %s\n", details.UserName)
	}
	if details.UserEmail != "" {
		fmt.Fprintf(out, "User Email: %s\n", details.UserEmail)
	}
	fmt.Fprintf(out, "Read: %s\n", formatBool(details.Read))
	fmt.Fprintf(out, "Ready For Delivery: %s\n", formatBool(details.IsReadyForDelivery))
	if details.DeliverAt != "" {
		fmt.Fprintf(out, "Deliver At: %s\n", details.DeliverAt)
	}
	if details.DeliveryDecisionApproach != "" {
		fmt.Fprintf(out, "Delivery Decision Approach: %s\n", details.DeliveryDecisionApproach)
	}
	if details.NotificationType != "" {
		fmt.Fprintf(out, "Notification Type: %s\n", details.NotificationType)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}
	if details.Details != nil {
		if formatted := formatAnyJSON(details.Details); formatted != "" {
			fmt.Fprintln(out, "")
			fmt.Fprintln(out, "Details:")
			fmt.Fprintln(out, strings.Repeat("-", 40))
			fmt.Fprintln(out, formatted)
		}
	}

	return nil
}
