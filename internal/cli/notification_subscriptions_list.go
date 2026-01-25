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

type notificationSubscriptionsListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	User         string
	CreatedAtMin string
	CreatedAtMax string
	UpdatedAtMin string
	UpdatedAtMax string
	IsCreatedAt  string
	IsUpdatedAt  string
}

type notificationSubscriptionRow struct {
	ID            string `json:"id"`
	Type          string `json:"type,omitempty"`
	UserID        string `json:"user_id,omitempty"`
	NotifyByEmail bool   `json:"notify_by_email"`
	NotifyByTxt   bool   `json:"notify_by_txt"`
}

func newNotificationSubscriptionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List notification subscriptions",
		Long: `List notification subscriptions with filtering and pagination.

Notification subscriptions define which users receive specific notification types
and delivery channels.

Output Columns:
  ID     Subscription identifier
  TYPE   Subscription type
  USER   User ID
  EMAIL  Email notifications enabled
  TXT    Text notifications enabled

Filters:
  --user           Filter by user ID
  --created-at-min Filter by created-at on/after (ISO 8601)
  --created-at-max Filter by created-at on/before (ISO 8601)
  --updated-at-min Filter by updated-at on/after (ISO 8601)
  --updated-at-max Filter by updated-at on/before (ISO 8601)
  --is-created-at  Filter by presence of created-at (true/false)
  --is-updated-at  Filter by presence of updated-at (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List notification subscriptions
  xbe view notification-subscriptions list

  # Filter by user
  xbe view notification-subscriptions list --user 123

  # Filter by created-at range
  xbe view notification-subscriptions list --created-at-min 2024-01-01T00:00:00Z --created-at-max 2024-12-31T23:59:59Z

  # Output as JSON
  xbe view notification-subscriptions list --json`,
		Args: cobra.NoArgs,
		RunE: runNotificationSubscriptionsList,
	}
	initNotificationSubscriptionsListFlags(cmd)
	return cmd
}

func init() {
	notificationSubscriptionsCmd.AddCommand(newNotificationSubscriptionsListCmd())
}

func initNotificationSubscriptionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("user", "", "Filter by user ID")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by presence of created-at (true/false)")
	cmd.Flags().String("is-updated-at", "", "Filter by presence of updated-at (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runNotificationSubscriptionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseNotificationSubscriptionsListOptions(cmd)
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
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[notification-subscriptions]", "polymorphic-type,notify-by-email,notify-by-txt,user")

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
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[is-created-at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[is-updated-at]", opts.IsUpdatedAt)

	body, _, err := client.Get(cmd.Context(), "/v1/notification-subscriptions", query)
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

	rows := buildNotificationSubscriptionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderNotificationSubscriptionsTable(cmd, rows)
}

func parseNotificationSubscriptionsListOptions(cmd *cobra.Command) (notificationSubscriptionsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	user, _ := cmd.Flags().GetString("user")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	isCreatedAt, _ := cmd.Flags().GetString("is-created-at")
	isUpdatedAt, _ := cmd.Flags().GetString("is-updated-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return notificationSubscriptionsListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		User:         user,
		CreatedAtMin: createdAtMin,
		CreatedAtMax: createdAtMax,
		UpdatedAtMin: updatedAtMin,
		UpdatedAtMax: updatedAtMax,
		IsCreatedAt:  isCreatedAt,
		IsUpdatedAt:  isUpdatedAt,
	}, nil
}

func buildNotificationSubscriptionRows(resp jsonAPIResponse) []notificationSubscriptionRow {
	rows := make([]notificationSubscriptionRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := notificationSubscriptionRow{
			ID:            resource.ID,
			Type:          stringAttr(attrs, "polymorphic-type"),
			UserID:        relationshipIDFromMap(resource.Relationships, "user"),
			NotifyByEmail: boolAttr(attrs, "notify-by-email"),
			NotifyByTxt:   boolAttr(attrs, "notify-by-txt"),
		}
		rows = append(rows, row)
	}
	return rows
}

func renderNotificationSubscriptionsTable(cmd *cobra.Command, rows []notificationSubscriptionRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No notification subscriptions found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTYPE\tUSER\tEMAIL\tTXT")
	for _, row := range rows {
		notifyByEmail := "no"
		if row.NotifyByEmail {
			notifyByEmail = "yes"
		}
		notifyByTxt := "no"
		if row.NotifyByTxt {
			notifyByTxt = "yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Type, 32),
			truncateString(row.UserID, 20),
			notifyByEmail,
			notifyByTxt,
		)
	}
	return writer.Flush()
}
