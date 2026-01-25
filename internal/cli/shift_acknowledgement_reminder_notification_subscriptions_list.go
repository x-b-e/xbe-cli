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

type shiftAcknowledgementReminderNotificationSubscriptionsListOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	NoAuth    bool
	Limit     int
	Offset    int
	Sort      string
	User      string
	TruckerID string
	Trucker   string
}

type shiftAcknowledgementReminderNotificationSubscriptionRow struct {
	ID            string `json:"id"`
	TruckerID     string `json:"trucker_id,omitempty"`
	TruckerName   string `json:"trucker_name,omitempty"`
	UserID        string `json:"user_id,omitempty"`
	UserName      string `json:"user_name,omitempty"`
	UserEmail     string `json:"user_email,omitempty"`
	NotifyByTxt   bool   `json:"notify_by_txt"`
	NotifyByEmail bool   `json:"notify_by_email"`
}

func newShiftAcknowledgementReminderNotificationSubscriptionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List shift acknowledgement reminder notification subscriptions",
		Long: `List shift acknowledgement reminder notification subscriptions with filtering and pagination.

Output Columns:
  ID       Subscription identifier
  TRUCKER  Trucker name
  USER     Subscriber name/email
  TXT      Notify by text (true/false)
  EMAIL    Notify by email (true/false)

Filters:
  --user        Filter by user ID
  --trucker-id  Filter by trucker ID
  --trucker     Filter by trucker ID

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List subscriptions
  xbe view shift-acknowledgement-reminder-notification-subscriptions list

  # Filter by user
  xbe view shift-acknowledgement-reminder-notification-subscriptions list --user 456

  # Filter by trucker
  xbe view shift-acknowledgement-reminder-notification-subscriptions list --trucker 123

  # Output as JSON
  xbe view shift-acknowledgement-reminder-notification-subscriptions list --json`,
		Args: cobra.NoArgs,
		RunE: runShiftAcknowledgementReminderNotificationSubscriptionsList,
	}
	initShiftAcknowledgementReminderNotificationSubscriptionsListFlags(cmd)
	return cmd
}

func init() {
	shiftAcknowledgementReminderNotificationSubscriptionsCmd.AddCommand(newShiftAcknowledgementReminderNotificationSubscriptionsListCmd())
}

func initShiftAcknowledgementReminderNotificationSubscriptionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("user", "", "Filter by user ID")
	cmd.Flags().String("trucker-id", "", "Filter by trucker ID")
	cmd.Flags().String("trucker", "", "Filter by trucker ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runShiftAcknowledgementReminderNotificationSubscriptionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseShiftAcknowledgementReminderNotificationSubscriptionsListOptions(cmd)
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
	query.Set("fields[shift-acknowledgement-reminder-notification-subscriptions]", "trucker,user,notify-by-txt,notify-by-email")
	query.Set("include", "trucker,user")
	query.Set("fields[truckers]", "company-name")
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
	setFilterIfPresent(query, "filter[trucker_id]", opts.TruckerID)
	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)

	body, _, err := client.Get(cmd.Context(), "/v1/shift-acknowledgement-reminder-notification-subscriptions", query)
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

	rows := buildShiftAcknowledgementReminderNotificationSubscriptionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderShiftAcknowledgementReminderNotificationSubscriptionsTable(cmd, rows)
}

func parseShiftAcknowledgementReminderNotificationSubscriptionsListOptions(cmd *cobra.Command) (shiftAcknowledgementReminderNotificationSubscriptionsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	user, _ := cmd.Flags().GetString("user")
	truckerID, _ := cmd.Flags().GetString("trucker-id")
	trucker, _ := cmd.Flags().GetString("trucker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return shiftAcknowledgementReminderNotificationSubscriptionsListOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		NoAuth:    noAuth,
		Limit:     limit,
		Offset:    offset,
		Sort:      sort,
		User:      user,
		TruckerID: truckerID,
		Trucker:   trucker,
	}, nil
}

func buildShiftAcknowledgementReminderNotificationSubscriptionRows(resp jsonAPIResponse) []shiftAcknowledgementReminderNotificationSubscriptionRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]shiftAcknowledgementReminderNotificationSubscriptionRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildShiftAcknowledgementReminderNotificationSubscriptionRow(resource, included))
	}
	return rows
}

func shiftAcknowledgementReminderNotificationSubscriptionRowFromSingle(resp jsonAPISingleResponse) shiftAcknowledgementReminderNotificationSubscriptionRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}
	return buildShiftAcknowledgementReminderNotificationSubscriptionRow(resp.Data, included)
}

func buildShiftAcknowledgementReminderNotificationSubscriptionRow(resource jsonAPIResource, included map[string]jsonAPIResource) shiftAcknowledgementReminderNotificationSubscriptionRow {
	row := shiftAcknowledgementReminderNotificationSubscriptionRow{
		ID: resource.ID,
	}

	row.NotifyByTxt = boolAttr(resource.Attributes, "notify-by-txt")
	row.NotifyByEmail = boolAttr(resource.Attributes, "notify-by-email")

	if rel, ok := resource.Relationships["trucker"]; ok && rel.Data != nil {
		row.TruckerID = rel.Data.ID
		if trucker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.TruckerName = stringAttr(trucker.Attributes, "company-name")
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

func renderShiftAcknowledgementReminderNotificationSubscriptionsTable(cmd *cobra.Command, rows []shiftAcknowledgementReminderNotificationSubscriptionRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No shift acknowledgement reminder notification subscriptions found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTRUCKER\tUSER\tTXT\tEMAIL")
	for _, row := range rows {
		trucker := formatShiftAcknowledgementReminderNotificationSubscriptionTruckerLabel(row)
		user := firstNonEmpty(row.UserName, row.UserEmail, row.UserID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(trucker, 28),
			truncateString(user, 24),
			formatBool(row.NotifyByTxt),
			formatBool(row.NotifyByEmail),
		)
	}
	return writer.Flush()
}

func formatShiftAcknowledgementReminderNotificationSubscriptionTruckerLabel(row shiftAcknowledgementReminderNotificationSubscriptionRow) string {
	if row.TruckerName != "" {
		return row.TruckerName
	}
	return row.TruckerID
}
