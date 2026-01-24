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

type brokerTenderOfferedSellerNotificationsListOptions struct {
	BaseURL                  string
	Token                    string
	JSON                     bool
	NoAuth                   bool
	Limit                    int
	Offset                   int
	Sort                     string
	User                     string
	Read                     string
	DeliveryDecisionApproach string
	IsReadyForDelivery       string
	DeliverAt                string
	DeliverAtMin             string
	DeliverAtMax             string
	NotificationType         string
	CreatedAtMin             string
	CreatedAtMax             string
	IsCreatedAt              string
	UpdatedAtMin             string
	UpdatedAtMax             string
	IsUpdatedAt              string
	NotID                    string
}

type brokerTenderOfferedSellerNotificationRow struct {
	ID                       string `json:"id"`
	TenderID                 string `json:"tender_id,omitempty"`
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
}

func newBrokerTenderOfferedSellerNotificationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List broker tender offered seller notifications",
		Long: `List broker tender offered seller notifications with filtering and pagination.

Output Columns:
  ID        Notification identifier
  TENDER    Broker tender ID
  USER      Recipient name/email
  READ      Whether the notification has been read
  READY     Ready for delivery status
  DELIVER   Scheduled delivery time
  APPROACH  Delivery decision approach

Filters:
  --user                       Filter by user ID
  --read                       Filter by read status (true/false)
  --delivery-decision-approach Filter by delivery decision approach (static/dynamic/all)
  --is-ready-for-delivery      Filter by ready-for-delivery status (true/false)
  --deliver-at                 Filter by deliver-at time (ISO 8601)
  --deliver-at-min             Filter by deliver-at on/after (ISO 8601)
  --deliver-at-max             Filter by deliver-at on/before (ISO 8601)
  --notification-type          Filter by notification type
  --created-at-min             Filter by created-at on/after (ISO 8601)
  --created-at-max             Filter by created-at on/before (ISO 8601)
  --is-created-at              Filter by has created-at (true/false)
  --updated-at-min             Filter by updated-at on/after (ISO 8601)
  --updated-at-max             Filter by updated-at on/before (ISO 8601)
  --is-updated-at              Filter by has updated-at (true/false)
  --not-id                     Exclude notifications by ID (comma-separated)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List notifications
  xbe view broker-tender-offered-seller-notifications list

  # Filter by user
  xbe view broker-tender-offered-seller-notifications list --user 456

  # Filter by read status
  xbe view broker-tender-offered-seller-notifications list --read false

  # Output as JSON
  xbe view broker-tender-offered-seller-notifications list --json`,
		Args: cobra.NoArgs,
		RunE: runBrokerTenderOfferedSellerNotificationsList,
	}
	initBrokerTenderOfferedSellerNotificationsListFlags(cmd)
	return cmd
}

func init() {
	brokerTenderOfferedSellerNotificationsCmd.AddCommand(newBrokerTenderOfferedSellerNotificationsListCmd())
}

func initBrokerTenderOfferedSellerNotificationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("user", "", "Filter by user ID")
	cmd.Flags().String("read", "", "Filter by read status (true/false)")
	cmd.Flags().String("delivery-decision-approach", "", "Filter by delivery decision approach (static/dynamic/all)")
	cmd.Flags().String("is-ready-for-delivery", "", "Filter by ready-for-delivery status (true/false)")
	cmd.Flags().String("deliver-at", "", "Filter by deliver-at time (ISO 8601)")
	cmd.Flags().String("deliver-at-min", "", "Filter by deliver-at on/after (ISO 8601)")
	cmd.Flags().String("deliver-at-max", "", "Filter by deliver-at on/before (ISO 8601)")
	cmd.Flags().String("notification-type", "", "Filter by notification type")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by has created-at (true/false)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("is-updated-at", "", "Filter by has updated-at (true/false)")
	cmd.Flags().String("not-id", "", "Exclude notifications by ID (comma-separated)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runBrokerTenderOfferedSellerNotificationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseBrokerTenderOfferedSellerNotificationsListOptions(cmd)
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
	query.Set("fields[broker-tender-offered-seller-notifications]", "read,deliver-at,is-ready-for-delivery,delivery-decision-approach,notification-type,created-at,updated-at,user,tender")
	query.Set("include", "user")
	query.Set("fields[users]", "name,email-address")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[user]", opts.User)
	setFilterIfPresent(query, "filter[read]", opts.Read)
	setFilterIfPresent(query, "filter[delivery-decision-approach]", opts.DeliveryDecisionApproach)
	setFilterIfPresent(query, "filter[is-ready-for-delivery]", opts.IsReadyForDelivery)
	setFilterIfPresent(query, "filter[deliver-at]", opts.DeliverAt)
	setFilterIfPresent(query, "filter[deliver-at-min]", opts.DeliverAtMin)
	setFilterIfPresent(query, "filter[deliver-at-max]", opts.DeliverAtMax)
	setFilterIfPresent(query, "filter[notification-type]", opts.NotificationType)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[is-created-at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[is-updated-at]", opts.IsUpdatedAt)
	setFilterIfPresent(query, "filter[not-id]", opts.NotID)

	body, _, err := client.Get(cmd.Context(), "/v1/broker-tender-offered-seller-notifications", query)
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

	rows := buildBrokerTenderOfferedSellerNotificationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderBrokerTenderOfferedSellerNotificationsTable(cmd, rows)
}

func parseBrokerTenderOfferedSellerNotificationsListOptions(cmd *cobra.Command) (brokerTenderOfferedSellerNotificationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	user, _ := cmd.Flags().GetString("user")
	read, _ := cmd.Flags().GetString("read")
	deliveryDecisionApproach, _ := cmd.Flags().GetString("delivery-decision-approach")
	isReadyForDelivery, _ := cmd.Flags().GetString("is-ready-for-delivery")
	deliverAt, _ := cmd.Flags().GetString("deliver-at")
	deliverAtMin, _ := cmd.Flags().GetString("deliver-at-min")
	deliverAtMax, _ := cmd.Flags().GetString("deliver-at-max")
	notificationType, _ := cmd.Flags().GetString("notification-type")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	isCreatedAt, _ := cmd.Flags().GetString("is-created-at")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	isUpdatedAt, _ := cmd.Flags().GetString("is-updated-at")
	notID, _ := cmd.Flags().GetString("not-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return brokerTenderOfferedSellerNotificationsListOptions{
		BaseURL:                  baseURL,
		Token:                    token,
		JSON:                     jsonOut,
		NoAuth:                   noAuth,
		Limit:                    limit,
		Offset:                   offset,
		Sort:                     sort,
		User:                     user,
		Read:                     read,
		DeliveryDecisionApproach: deliveryDecisionApproach,
		IsReadyForDelivery:       isReadyForDelivery,
		DeliverAt:                deliverAt,
		DeliverAtMin:             deliverAtMin,
		DeliverAtMax:             deliverAtMax,
		NotificationType:         notificationType,
		CreatedAtMin:             createdAtMin,
		CreatedAtMax:             createdAtMax,
		IsCreatedAt:              isCreatedAt,
		UpdatedAtMin:             updatedAtMin,
		UpdatedAtMax:             updatedAtMax,
		IsUpdatedAt:              isUpdatedAt,
		NotID:                    notID,
	}, nil
}

func buildBrokerTenderOfferedSellerNotificationRows(resp jsonAPIResponse) []brokerTenderOfferedSellerNotificationRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]brokerTenderOfferedSellerNotificationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildBrokerTenderOfferedSellerNotificationRow(resource, included))
	}
	return rows
}

func brokerTenderOfferedSellerNotificationRowFromSingle(resp jsonAPISingleResponse) brokerTenderOfferedSellerNotificationRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}
	return buildBrokerTenderOfferedSellerNotificationRow(resp.Data, included)
}

func buildBrokerTenderOfferedSellerNotificationRow(resource jsonAPIResource, included map[string]jsonAPIResource) brokerTenderOfferedSellerNotificationRow {
	attrs := resource.Attributes
	row := brokerTenderOfferedSellerNotificationRow{
		ID:                       resource.ID,
		Read:                     boolAttr(attrs, "read"),
		IsReadyForDelivery:       boolAttr(attrs, "is-ready-for-delivery"),
		DeliverAt:                formatDateTime(stringAttr(attrs, "deliver-at")),
		DeliveryDecisionApproach: stringAttr(attrs, "delivery-decision-approach"),
		NotificationType:         stringAttr(attrs, "notification-type"),
		CreatedAt:                formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:                formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resource.Relationships["tender"]; ok && rel.Data != nil {
		row.TenderID = rel.Data.ID
	}

	if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
		row.UserID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.UserName = stringAttr(user.Attributes, "name")
			row.UserEmail = stringAttr(user.Attributes, "email-address")
		}
	}

	return row
}

func renderBrokerTenderOfferedSellerNotificationsTable(cmd *cobra.Command, rows []brokerTenderOfferedSellerNotificationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No broker tender offered seller notifications found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTENDER\tUSER\tREAD\tREADY\tDELIVER\tAPPROACH")
	for _, row := range rows {
		user := firstNonEmpty(row.UserName, row.UserEmail, row.UserID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.TenderID,
			truncateString(user, 24),
			formatBool(row.Read),
			formatBool(row.IsReadyForDelivery),
			truncateString(row.DeliverAt, 19),
			row.DeliveryDecisionApproach,
		)
	}
	return writer.Flush()
}
