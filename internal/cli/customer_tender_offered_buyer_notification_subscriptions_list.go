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

type customerTenderOfferedBuyerNotificationSubscriptionsListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Sort    string
	User    string
}

type customerTenderOfferedBuyerNotificationSubscriptionRow struct {
	ID            string `json:"id"`
	BrokerID      string `json:"broker_id,omitempty"`
	BrokerName    string `json:"broker_name,omitempty"`
	UserID        string `json:"user_id,omitempty"`
	UserName      string `json:"user_name,omitempty"`
	UserEmail     string `json:"user_email,omitempty"`
	NotifyByTxt   bool   `json:"notify_by_txt"`
	NotifyByEmail bool   `json:"notify_by_email"`
}

func newCustomerTenderOfferedBuyerNotificationSubscriptionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List customer tender offered buyer notification subscriptions",
		Long: `List customer tender offered buyer notification subscriptions with filtering and pagination.

Output Columns:
  ID       Subscription identifier
  BROKER   Broker name
  USER     Subscriber name/email
  TXT      Notify by text (true/false)
  EMAIL    Notify by email (true/false)

Filters:
  --user  Filter by user ID

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List subscriptions
  xbe view customer-tender-offered-buyer-notification-subscriptions list

  # Filter by user
  xbe view customer-tender-offered-buyer-notification-subscriptions list --user 456

  # Output as JSON
  xbe view customer-tender-offered-buyer-notification-subscriptions list --json`,
		Args: cobra.NoArgs,
		RunE: runCustomerTenderOfferedBuyerNotificationSubscriptionsList,
	}
	initCustomerTenderOfferedBuyerNotificationSubscriptionsListFlags(cmd)
	return cmd
}

func init() {
	customerTenderOfferedBuyerNotificationSubscriptionsCmd.AddCommand(newCustomerTenderOfferedBuyerNotificationSubscriptionsListCmd())
}

func initCustomerTenderOfferedBuyerNotificationSubscriptionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("user", "", "Filter by user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCustomerTenderOfferedBuyerNotificationSubscriptionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseCustomerTenderOfferedBuyerNotificationSubscriptionsListOptions(cmd)
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
	query.Set("fields[customer-tender-offered-buyer-notification-subscriptions]", "broker,user,notify-by-txt,notify-by-email")
	query.Set("include", "broker,user")
	query.Set("fields[brokers]", "company-name")
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

	body, _, err := client.Get(cmd.Context(), "/v1/customer-tender-offered-buyer-notification-subscriptions", query)
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

	rows := buildCustomerTenderOfferedBuyerNotificationSubscriptionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderCustomerTenderOfferedBuyerNotificationSubscriptionsTable(cmd, rows)
}

func parseCustomerTenderOfferedBuyerNotificationSubscriptionsListOptions(cmd *cobra.Command) (customerTenderOfferedBuyerNotificationSubscriptionsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	user, _ := cmd.Flags().GetString("user")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return customerTenderOfferedBuyerNotificationSubscriptionsListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
		User:    user,
	}, nil
}

func buildCustomerTenderOfferedBuyerNotificationSubscriptionRows(resp jsonAPIResponse) []customerTenderOfferedBuyerNotificationSubscriptionRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]customerTenderOfferedBuyerNotificationSubscriptionRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildCustomerTenderOfferedBuyerNotificationSubscriptionRow(resource, included))
	}
	return rows
}

func customerTenderOfferedBuyerNotificationSubscriptionRowFromSingle(resp jsonAPISingleResponse) customerTenderOfferedBuyerNotificationSubscriptionRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}
	return buildCustomerTenderOfferedBuyerNotificationSubscriptionRow(resp.Data, included)
}

func buildCustomerTenderOfferedBuyerNotificationSubscriptionRow(resource jsonAPIResource, included map[string]jsonAPIResource) customerTenderOfferedBuyerNotificationSubscriptionRow {
	row := customerTenderOfferedBuyerNotificationSubscriptionRow{
		ID: resource.ID,
	}

	row.NotifyByTxt = boolAttr(resource.Attributes, "notify-by-txt")
	row.NotifyByEmail = boolAttr(resource.Attributes, "notify-by-email")

	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
		if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.BrokerName = stringAttr(broker.Attributes, "company-name")
		}
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

func renderCustomerTenderOfferedBuyerNotificationSubscriptionsTable(cmd *cobra.Command, rows []customerTenderOfferedBuyerNotificationSubscriptionRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No customer tender offered buyer notification subscriptions found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tBROKER\tUSER\tTXT\tEMAIL")
	for _, row := range rows {
		broker := formatCustomerTenderOfferedBuyerNotificationSubscriptionBrokerLabel(row)
		user := firstNonEmpty(row.UserName, row.UserEmail, row.UserID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(broker, 28),
			truncateString(user, 24),
			formatBool(row.NotifyByTxt),
			formatBool(row.NotifyByEmail),
		)
	}
	return writer.Flush()
}

func formatCustomerTenderOfferedBuyerNotificationSubscriptionBrokerLabel(row customerTenderOfferedBuyerNotificationSubscriptionRow) string {
	if row.BrokerName != "" {
		return row.BrokerName
	}
	return row.BrokerID
}
