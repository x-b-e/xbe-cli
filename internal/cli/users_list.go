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

type usersListOptions struct {
	BaseURL                      string
	Token                        string
	JSON                         bool
	NoAuth                       bool
	Limit                        int
	Offset                       int
	Name                         string
	IsAdmin                      bool
	EmailAddress                 string
	MobileNumber                 string
	SlackID                      string
	IsDriver                     string
	IsSuspendedFromDriving       string
	HavingCustomerMembershipWith string
	HavingTruckerMembershipWith  string
}

func newUsersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List users",
		Long: `List users with filtering and pagination.

Returns a list of users matching the specified criteria.

Pagination:
  Use --limit and --offset to paginate through large result sets.

Filtering:
  Multiple filters can be combined. All filters use AND logic.

Use Case:
  Find user IDs for filtering posts by creator:
    xbe view posts list --creator "User|<id>"`,
		Example: `  # List users
  xbe view users list

  # Search by name
  xbe view users list --name "John"

  # Filter by admin status
  xbe view users list --is-admin

  # Paginate results
  xbe view users list --limit 20 --offset 40

  # Output as JSON
  xbe view users list --json`,
		RunE: runUsersList,
	}
	initUsersListFlags(cmd)
	return cmd
}

func init() {
	usersCmd.AddCommand(newUsersListCmd())
}

func initUsersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 0, "Page size (defaults to server default)")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("name", "", "Filter by name (partial match)")
	cmd.Flags().Bool("is-admin", false, "Filter to only admins")
	cmd.Flags().String("email-address", "", "Filter by email address")
	cmd.Flags().String("mobile-number", "", "Filter by mobile number")
	cmd.Flags().String("slack-id", "", "Filter by Slack ID")
	cmd.Flags().String("is-driver", "", "Filter by driver status (true/false)")
	cmd.Flags().String("is-suspended-from-driving", "", "Filter by driving suspension status (true/false)")
	cmd.Flags().String("having-customer-membership-with", "", "Filter by customer membership (customer ID, comma-separated for multiple)")
	cmd.Flags().String("having-trucker-membership-with", "", "Filter by trucker membership (trucker ID, comma-separated for multiple)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runUsersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseUsersListOptions(cmd)
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
	query.Set("fields[users]", "name,email-address,mobile-number,is-admin")
	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[q]", opts.Name)
	if opts.IsAdmin {
		query.Set("filter[is_admin]", "true")
	}
	setFilterIfPresent(query, "filter[email-address]", opts.EmailAddress)
	setFilterIfPresent(query, "filter[mobile-number]", opts.MobileNumber)
	setFilterIfPresent(query, "filter[slack-id]", opts.SlackID)
	setFilterIfPresent(query, "filter[is-driver]", opts.IsDriver)
	setFilterIfPresent(query, "filter[is-suspended-from-driving]", opts.IsSuspendedFromDriving)
	setFilterIfPresent(query, "filter[having-customer-membership-with]", opts.HavingCustomerMembershipWith)
	setFilterIfPresent(query, "filter[having-trucker-membership-with]", opts.HavingTruckerMembershipWith)

	body, _, err := client.Get(cmd.Context(), "/v1/users", query)
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

	if opts.JSON {
		rows := buildUserRows(resp)
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderUsersList(cmd, resp)
}

func parseUsersListOptions(cmd *cobra.Command) (usersListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return usersListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return usersListOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return usersListOptions{}, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return usersListOptions{}, err
	}
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return usersListOptions{}, err
	}
	isAdmin, err := cmd.Flags().GetBool("is-admin")
	if err != nil {
		return usersListOptions{}, err
	}
	emailAddress, err := cmd.Flags().GetString("email-address")
	if err != nil {
		return usersListOptions{}, err
	}
	mobileNumber, err := cmd.Flags().GetString("mobile-number")
	if err != nil {
		return usersListOptions{}, err
	}
	slackID, err := cmd.Flags().GetString("slack-id")
	if err != nil {
		return usersListOptions{}, err
	}
	isDriver, err := cmd.Flags().GetString("is-driver")
	if err != nil {
		return usersListOptions{}, err
	}
	isSuspendedFromDriving, err := cmd.Flags().GetString("is-suspended-from-driving")
	if err != nil {
		return usersListOptions{}, err
	}
	havingCustomerMembershipWith, err := cmd.Flags().GetString("having-customer-membership-with")
	if err != nil {
		return usersListOptions{}, err
	}
	havingTruckerMembershipWith, err := cmd.Flags().GetString("having-trucker-membership-with")
	if err != nil {
		return usersListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return usersListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return usersListOptions{}, err
	}

	return usersListOptions{
		BaseURL:                      baseURL,
		Token:                        token,
		JSON:                         jsonOut,
		NoAuth:                       noAuth,
		Limit:                        limit,
		Offset:                       offset,
		Name:                         name,
		IsAdmin:                      isAdmin,
		EmailAddress:                 emailAddress,
		MobileNumber:                 mobileNumber,
		SlackID:                      slackID,
		IsDriver:                     isDriver,
		IsSuspendedFromDriving:       isSuspendedFromDriving,
		HavingCustomerMembershipWith: havingCustomerMembershipWith,
		HavingTruckerMembershipWith:  havingTruckerMembershipWith,
	}, nil
}

type userRow struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email,omitempty"`
	Mobile  string `json:"mobile,omitempty"`
	IsAdmin bool   `json:"is_admin"`
}

func buildUserRows(resp jsonAPIResponse) []userRow {
	rows := make([]userRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, userRow{
			ID:      resource.ID,
			Name:    strings.TrimSpace(stringAttr(resource.Attributes, "name")),
			Email:   strings.TrimSpace(stringAttr(resource.Attributes, "email-address")),
			Mobile:  strings.TrimSpace(stringAttr(resource.Attributes, "mobile-number")),
			IsAdmin: boolAttr(resource.Attributes, "is-admin"),
		})
	}
	return rows
}

func renderUsersList(cmd *cobra.Command, resp jsonAPIResponse) error {
	rows := buildUserRows(resp)
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No users found.")
		return nil
	}

	const nameMax = 25
	const emailMax = 35

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tEMAIL\tMOBILE\tADMIN")
	for _, row := range rows {
		admin := ""
		if row.IsAdmin {
			admin = "yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, nameMax),
			truncateString(row.Email, emailMax),
			row.Mobile,
			admin,
		)
	}
	return writer.Flush()
}
