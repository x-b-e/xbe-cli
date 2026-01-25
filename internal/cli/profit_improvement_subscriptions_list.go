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

type profitImprovementSubscriptionsListOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	NoAuth            bool
	Limit             int
	Offset            int
	Sort              string
	ContactMethod     string
	User              string
	ProfitImprovement string
}

type profitImprovementSubscriptionRow struct {
	ID                      string `json:"id"`
	ContactMethod           string `json:"contact_method,omitempty"`
	ContactMethodEffective  string `json:"contact_method_effective,omitempty"`
	UserID                  string `json:"user_id,omitempty"`
	User                    string `json:"user,omitempty"`
	UserEmail               string `json:"user_email,omitempty"`
	UserMobile              string `json:"user_mobile,omitempty"`
	ProfitImprovementID     string `json:"profit_improvement_id,omitempty"`
	ProfitImprovementTitle  string `json:"profit_improvement_title,omitempty"`
	ProfitImprovementStatus string `json:"profit_improvement_status,omitempty"`
}

func newProfitImprovementSubscriptionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List profit improvement subscriptions",
		Long: `List profit improvement subscriptions with filtering and pagination.

Output Columns:
  ID                  Subscription identifier
  USER                Subscribed user
  PROFIT IMPROVEMENT  Profit improvement title
  CONTACT             Contact method used for notifications

Filters:
  --contact-method     Filter by contact method (email_address, mobile_number)
  --user               Filter by user ID (comma-separated for multiple)
  --profit-improvement Filter by profit improvement ID (comma-separated for multiple)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List subscriptions
  xbe view profit-improvement-subscriptions list

  # Filter by profit improvement
  xbe view profit-improvement-subscriptions list --profit-improvement 123

  # Filter by user
  xbe view profit-improvement-subscriptions list --user 456

  # Filter by contact method
  xbe view profit-improvement-subscriptions list --contact-method email_address

  # Output as JSON
  xbe view profit-improvement-subscriptions list --json`,
		Args: cobra.NoArgs,
		RunE: runProfitImprovementSubscriptionsList,
	}
	initProfitImprovementSubscriptionsListFlags(cmd)
	return cmd
}

func init() {
	profitImprovementSubscriptionsCmd.AddCommand(newProfitImprovementSubscriptionsListCmd())
}

func initProfitImprovementSubscriptionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("contact-method", "", "Filter by contact method (email_address, mobile_number)")
	cmd.Flags().String("user", "", "Filter by user ID (comma-separated for multiple)")
	cmd.Flags().String("profit-improvement", "", "Filter by profit improvement ID (comma-separated for multiple)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProfitImprovementSubscriptionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProfitImprovementSubscriptionsListOptions(cmd)
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
	query.Set("fields[profit-improvement-subscriptions]", "contact-method,contact-method-effective,user,profit-improvement")
	query.Set("fields[users]", "name,email-address,mobile-number")
	query.Set("fields[profit-improvements]", "title,status")
	query.Set("include", "user,profit-improvement")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[contact_method]", opts.ContactMethod)
	setFilterIfPresent(query, "filter[user]", opts.User)
	setFilterIfPresent(query, "filter[profit_improvement]", opts.ProfitImprovement)

	body, _, err := client.Get(cmd.Context(), "/v1/profit-improvement-subscriptions", query)
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

	rows := buildProfitImprovementSubscriptionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProfitImprovementSubscriptionsTable(cmd, rows)
}

func parseProfitImprovementSubscriptionsListOptions(cmd *cobra.Command) (profitImprovementSubscriptionsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	contactMethod, _ := cmd.Flags().GetString("contact-method")
	user, _ := cmd.Flags().GetString("user")
	profitImprovement, _ := cmd.Flags().GetString("profit-improvement")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return profitImprovementSubscriptionsListOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		NoAuth:            noAuth,
		Limit:             limit,
		Offset:            offset,
		Sort:              sort,
		ContactMethod:     contactMethod,
		User:              user,
		ProfitImprovement: profitImprovement,
	}, nil
}

func buildProfitImprovementSubscriptionRows(resp jsonAPIResponse) []profitImprovementSubscriptionRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]profitImprovementSubscriptionRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildProfitImprovementSubscriptionRowFromResource(resource, included))
	}

	return rows
}

func buildProfitImprovementSubscriptionRowFromResource(resource jsonAPIResource, included map[string]jsonAPIResource) profitImprovementSubscriptionRow {
	attrs := resource.Attributes
	row := profitImprovementSubscriptionRow{
		ID:                     resource.ID,
		ContactMethod:          stringAttr(attrs, "contact-method"),
		ContactMethodEffective: stringAttr(attrs, "contact-method-effective"),
	}

	if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
		row.UserID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.User = stringAttr(user.Attributes, "name")
			row.UserEmail = stringAttr(user.Attributes, "email-address")
			row.UserMobile = stringAttr(user.Attributes, "mobile-number")
		}
	}

	if rel, ok := resource.Relationships["profit-improvement"]; ok && rel.Data != nil {
		row.ProfitImprovementID = rel.Data.ID
		if pi, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.ProfitImprovementTitle = stringAttr(pi.Attributes, "title")
			row.ProfitImprovementStatus = stringAttr(pi.Attributes, "status")
		}
	}

	return row
}

func renderProfitImprovementSubscriptionsTable(cmd *cobra.Command, rows []profitImprovementSubscriptionRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No profit improvement subscriptions found.")
		return nil
	}

	const (
		maxUser        = 26
		maxImprovement = 32
		maxContact     = 16
	)

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tUSER\tPROFIT IMPROVEMENT\tCONTACT")
	for _, row := range rows {
		userLabel := firstNonEmpty(row.User, row.UserEmail, row.UserID)
		improvementLabel := firstNonEmpty(row.ProfitImprovementTitle, row.ProfitImprovementID)
		contactLabel := firstNonEmpty(row.ContactMethodEffective, row.ContactMethod)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(userLabel, maxUser),
			truncateString(improvementLabel, maxImprovement),
			truncateString(contactLabel, maxContact),
		)
	}
	return writer.Flush()
}
