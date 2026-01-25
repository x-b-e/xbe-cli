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

type notificationSubscriptionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type notificationSubscriptionDetails struct {
	ID            string `json:"id"`
	Type          string `json:"type,omitempty"`
	UserID        string `json:"user_id,omitempty"`
	NotifyByEmail bool   `json:"notify_by_email"`
	NotifyByTxt   bool   `json:"notify_by_txt"`
	CreatedAt     string `json:"created_at,omitempty"`
	UpdatedAt     string `json:"updated_at,omitempty"`
}

func newNotificationSubscriptionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show notification subscription details",
		Long: `Show the full details of a notification subscription.

Output Fields:
  ID
  Type
  User ID
  Notify By Email
  Notify By Txt
  Created At
  Updated At

Global flags (see xbe --help): --json, --base-url, --token, --no-auth

Arguments:
  <id>    The notification subscription ID (required). You can find IDs using the list command.`,
		Example: `  # Show a notification subscription
  xbe view notification-subscriptions show 123

  # Get JSON output
  xbe view notification-subscriptions show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runNotificationSubscriptionsShow,
	}
	initNotificationSubscriptionsShowFlags(cmd)
	return cmd
}

func init() {
	notificationSubscriptionsCmd.AddCommand(newNotificationSubscriptionsShowCmd())
}

func initNotificationSubscriptionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runNotificationSubscriptionsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseNotificationSubscriptionsShowOptions(cmd)
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
		return fmt.Errorf("notification subscription id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[notification-subscriptions]", "polymorphic-type,notify-by-email,notify-by-txt,created-at,updated-at,user")

	body, _, err := client.Get(cmd.Context(), "/v1/notification-subscriptions/"+id, query)
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

	details := buildNotificationSubscriptionDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderNotificationSubscriptionDetails(cmd, details)
}

func parseNotificationSubscriptionsShowOptions(cmd *cobra.Command) (notificationSubscriptionsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return notificationSubscriptionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildNotificationSubscriptionDetails(resp jsonAPISingleResponse) notificationSubscriptionDetails {
	resource := resp.Data
	attrs := resource.Attributes
	return notificationSubscriptionDetails{
		ID:            resource.ID,
		Type:          stringAttr(attrs, "polymorphic-type"),
		UserID:        relationshipIDFromMap(resource.Relationships, "user"),
		NotifyByEmail: boolAttr(attrs, "notify-by-email"),
		NotifyByTxt:   boolAttr(attrs, "notify-by-txt"),
		CreatedAt:     formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:     formatDateTime(stringAttr(attrs, "updated-at")),
	}
}

func renderNotificationSubscriptionDetails(cmd *cobra.Command, details notificationSubscriptionDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Type != "" {
		fmt.Fprintf(out, "Type: %s\n", details.Type)
	}
	if details.UserID != "" {
		fmt.Fprintf(out, "User ID: %s\n", details.UserID)
	}
	fmt.Fprintf(out, "Notify By Email: %t\n", details.NotifyByEmail)
	fmt.Fprintf(out, "Notify By Txt: %t\n", details.NotifyByTxt)
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
