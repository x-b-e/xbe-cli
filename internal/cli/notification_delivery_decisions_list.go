package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type notificationDeliveryDecisionsListOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	NoAuth           bool
	Limit            int
	Offset           int
	Sort             string
	Notification     string
	NotificationUser string
}

type notificationDeliveryDecisionRow struct {
	ID                 string `json:"id"`
	NotificationID     string `json:"notification_id,omitempty"`
	NotificationUserID string `json:"notification_user_id,omitempty"`
	NotifyByValueMin   string `json:"notify_by_value_min,omitempty"`
	NotifyByEmailValue string `json:"notify_by_email_value,omitempty"`
	NotifyByTxtValue   string `json:"notify_by_txt_value,omitempty"`
	DeliverAt          string `json:"deliver_at,omitempty"`
	NotifyByEmail      bool   `json:"notify_by_email"`
	NotifyByTxt        bool   `json:"notify_by_txt"`
}

func newNotificationDeliveryDecisionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List notification delivery decisions",
		Long: `List notification delivery decisions with filtering and pagination.

Output Columns:
  ID           Decision identifier
  NOTIFICATION Notification ID
  MIN          Minimum value required for delivery
  EMAIL        Value if delivered by email
  TXT          Value if delivered by txt
  DELIVER      Scheduled delivery time
  CHANNELS     Delivery channels enabled

Filters:
  --notification      Filter by notification ID
  --notification-user Filter by notification user ID

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List notification delivery decisions
  xbe view notification-delivery-decisions list

  # Filter by notification
  xbe view notification-delivery-decisions list --notification 123

  # Filter by notification user
  xbe view notification-delivery-decisions list --notification-user 456

  # Output as JSON
  xbe view notification-delivery-decisions list --json`,
		Args: cobra.NoArgs,
		RunE: runNotificationDeliveryDecisionsList,
	}
	initNotificationDeliveryDecisionsListFlags(cmd)
	return cmd
}

func init() {
	notificationDeliveryDecisionsCmd.AddCommand(newNotificationDeliveryDecisionsListCmd())
}

func initNotificationDeliveryDecisionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("notification", "", "Filter by notification ID")
	cmd.Flags().String("notification-user", "", "Filter by notification user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runNotificationDeliveryDecisionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseNotificationDeliveryDecisionsListOptions(cmd)
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
	query.Set("include", "notification")
	query.Set("fields[notifications]", "user")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[notification]", opts.Notification)
	setFilterIfPresent(query, "filter[notification-user]", opts.NotificationUser)

	body, _, err := client.Get(cmd.Context(), "/v1/notification-delivery-decisions", query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	rows := buildNotificationDeliveryDecisionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderNotificationDeliveryDecisionsTable(cmd, rows)
}

func parseNotificationDeliveryDecisionsListOptions(cmd *cobra.Command) (notificationDeliveryDecisionsListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return notificationDeliveryDecisionsListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return notificationDeliveryDecisionsListOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return notificationDeliveryDecisionsListOptions{}, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return notificationDeliveryDecisionsListOptions{}, err
	}
	sort, err := cmd.Flags().GetString("sort")
	if err != nil {
		return notificationDeliveryDecisionsListOptions{}, err
	}
	notification, err := cmd.Flags().GetString("notification")
	if err != nil {
		return notificationDeliveryDecisionsListOptions{}, err
	}
	notificationUser, err := cmd.Flags().GetString("notification-user")
	if err != nil {
		return notificationDeliveryDecisionsListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return notificationDeliveryDecisionsListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return notificationDeliveryDecisionsListOptions{}, err
	}

	return notificationDeliveryDecisionsListOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		NoAuth:           noAuth,
		Limit:            limit,
		Offset:           offset,
		Sort:             sort,
		Notification:     notification,
		NotificationUser: notificationUser,
	}, nil
}

func buildNotificationDeliveryDecisionRows(resp jsonAPIResponse) []notificationDeliveryDecisionRow {
	rows := make([]notificationDeliveryDecisionRow, 0, len(resp.Data))
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := notificationDeliveryDecisionRow{
			ID:                 resource.ID,
			NotifyByValueMin:   stringAttr(attrs, "notify-by-value-min"),
			NotifyByEmailValue: stringAttr(attrs, "notify-by-email-value"),
			NotifyByTxtValue:   stringAttr(attrs, "notify-by-txt-value"),
			DeliverAt:          formatDateTime(stringAttr(attrs, "deliver-at")),
			NotifyByEmail:      boolAttr(attrs, "notify-by-email"),
			NotifyByTxt:        boolAttr(attrs, "notify-by-txt"),
		}

		if rel, ok := resource.Relationships["notification"]; ok && rel.Data != nil {
			row.NotificationID = rel.Data.ID
			if notification, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				if userRel, ok := notification.Relationships["user"]; ok && userRel.Data != nil {
					row.NotificationUserID = userRel.Data.ID
				}
			}
		}

		rows = append(rows, row)
	}
	return rows
}

func renderNotificationDeliveryDecisionsTable(cmd *cobra.Command, rows []notificationDeliveryDecisionRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No notification delivery decisions found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNOTIFICATION\tMIN\tEMAIL\tTXT\tDELIVER\tCHANNELS")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.NotificationID,
			truncateString(row.NotifyByValueMin, 12),
			truncateString(row.NotifyByEmailValue, 12),
			truncateString(row.NotifyByTxtValue, 12),
			truncateString(row.DeliverAt, 19),
			truncateString(formatNotificationChannels(row.NotifyByEmail, row.NotifyByTxt), 14),
		)
	}
	return writer.Flush()
}

func formatNotificationChannels(notifyByEmail, notifyByTxt bool) string {
	channels := make([]string, 0, 2)
	if notifyByEmail {
		channels = append(channels, "email")
	}
	if notifyByTxt {
		channels = append(channels, "txt")
	}
	if len(channels) == 0 {
		return "none"
	}
	return strings.Join(channels, ",")
}
