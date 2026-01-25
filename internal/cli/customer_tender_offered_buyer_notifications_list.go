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

type customerTenderOfferedBuyerNotificationsListOptions struct {
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
	NotificationType         string
}

type customerTenderOfferedBuyerNotificationRow struct {
	ID        string `json:"id"`
	TenderID  string `json:"tender_id,omitempty"`
	UserID    string `json:"user_id,omitempty"`
	Read      bool   `json:"read"`
	DeliverAt string `json:"deliver_at,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

func newCustomerTenderOfferedBuyerNotificationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List customer tender offered buyer notifications",
		Long: `List customer tender offered buyer notifications with filtering and pagination.

Output Columns:
  ID         Notification identifier
  TENDER     Customer tender ID
  READ       Read status
  DELIVER AT Scheduled delivery time
  CREATED    Creation time

Filters:
  --user                       Filter by user ID
  --read                       Filter by read status (true/false)
  --delivery-decision-approach Filter by delivery decision approach (static/dynamic)
  --is-ready-for-delivery      Filter by ready for delivery (true/false)
  --deliver-at                 Filter by deliver-at timestamp (ISO8601)
  --notification-type          Filter by notification type

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List notifications
  xbe view customer-tender-offered-buyer-notifications list

  # Filter by read status
  xbe view customer-tender-offered-buyer-notifications list --read false

  # Filter by user
  xbe view customer-tender-offered-buyer-notifications list --user 123

  # Output as JSON
  xbe view customer-tender-offered-buyer-notifications list --json`,
		Args: cobra.NoArgs,
		RunE: runCustomerTenderOfferedBuyerNotificationsList,
	}
	initCustomerTenderOfferedBuyerNotificationsListFlags(cmd)
	return cmd
}

func init() {
	customerTenderOfferedBuyerNotificationsCmd.AddCommand(newCustomerTenderOfferedBuyerNotificationsListCmd())
}

func initCustomerTenderOfferedBuyerNotificationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 0, "Page size (defaults to server default)")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("user", "", "Filter by user ID")
	cmd.Flags().String("read", "", "Filter by read status (true/false)")
	cmd.Flags().String("delivery-decision-approach", "", "Filter by delivery decision approach (static/dynamic)")
	cmd.Flags().String("is-ready-for-delivery", "", "Filter by ready for delivery (true/false)")
	cmd.Flags().String("deliver-at", "", "Filter by deliver-at timestamp (ISO8601)")
	cmd.Flags().String("notification-type", "", "Filter by notification type")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCustomerTenderOfferedBuyerNotificationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseCustomerTenderOfferedBuyerNotificationsListOptions(cmd)
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

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[user]", opts.User)
	setFilterIfPresent(query, "filter[read]", opts.Read)
	setFilterIfPresent(query, "filter[delivery-decision-approach]", opts.DeliveryDecisionApproach)
	setFilterIfPresent(query, "filter[is-ready-for-delivery]", opts.IsReadyForDelivery)
	setFilterIfPresent(query, "filter[deliver-at]", opts.DeliverAt)
	setFilterIfPresent(query, "filter[notification-type]", opts.NotificationType)

	body, _, err := client.Get(cmd.Context(), "/v1/customer-tender-offered-buyer-notifications", query)
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

	handled, err := renderSparseListIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	rows := buildCustomerTenderOfferedBuyerNotificationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderCustomerTenderOfferedBuyerNotificationsTable(cmd, rows)
}

func parseCustomerTenderOfferedBuyerNotificationsListOptions(cmd *cobra.Command) (customerTenderOfferedBuyerNotificationsListOptions, error) {
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
	notificationType, _ := cmd.Flags().GetString("notification-type")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return customerTenderOfferedBuyerNotificationsListOptions{
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
		NotificationType:         notificationType,
	}, nil
}

func buildCustomerTenderOfferedBuyerNotificationRows(resp jsonAPIResponse) []customerTenderOfferedBuyerNotificationRow {
	rows := make([]customerTenderOfferedBuyerNotificationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := customerTenderOfferedBuyerNotificationRow{
			ID:        resource.ID,
			TenderID:  relationshipIDFromMap(resource.Relationships, "tender"),
			UserID:    relationshipIDFromMap(resource.Relationships, "user"),
			Read:      boolAttr(attrs, "read"),
			DeliverAt: formatDateTime(stringAttr(attrs, "deliver-at")),
			CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
		}
		rows = append(rows, row)
	}
	return rows
}

func renderCustomerTenderOfferedBuyerNotificationsTable(cmd *cobra.Command, rows []customerTenderOfferedBuyerNotificationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No customer tender offered buyer notifications found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTENDER\tREAD\tDELIVER AT\tCREATED")
	for _, row := range rows {
		read := "no"
		if row.Read {
			read = "yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.TenderID, 20),
			read,
			row.DeliverAt,
			row.CreatedAt,
		)
	}
	return writer.Flush()
}
