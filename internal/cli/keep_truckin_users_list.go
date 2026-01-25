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

type keepTruckinUsersListOptions struct {
	BaseURL       string
	Token         string
	JSON          bool
	NoAuth        bool
	Limit         int
	Offset        int
	Sort          string
	UserSetAtMin  string
	UserSetAtMax  string
	IsUserSetAt   string
	AssignedAtMin string
	AssignedAtMax string
	Role          string
	Active        string
	Broker        string
	User          string
	HasUser       string
}

type keepTruckinUserRow struct {
	ID          string `json:"id"`
	DriverID    string `json:"driver_id,omitempty"`
	Name        string `json:"name,omitempty"`
	Role        string `json:"role,omitempty"`
	Active      bool   `json:"active"`
	CarrierName string `json:"carrier_name,omitempty"`
	UserID      string `json:"user_id,omitempty"`
	BrokerID    string `json:"broker_id,omitempty"`
	UserSetAt   string `json:"user_set_at,omitempty"`
}

func newKeepTruckinUsersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List KeepTruckin users",
		Long: `List KeepTruckin users with filtering and pagination.

Output Columns:
  ID          KeepTruckin user identifier
  DRIVER ID   KeepTruckin driver ID
  NAME        Driver name
  ROLE        Driver role
  ACTIVE      Active status
  CARRIER     Carrier name from KeepTruckin
  USER        Linked user ID
  BROKER      Broker ID
  USER SET AT Timestamp when the user was linked

Filters:
  --user-set-at-min  Filter by user-set-at on/after (ISO 8601)
  --user-set-at-max  Filter by user-set-at on/before (ISO 8601)
  --is-user-set-at   Filter by presence of user-set-at (true/false)
  --assigned-at-min  Alias for --user-set-at-min
  --assigned-at-max  Alias for --user-set-at-max
  --role             Filter by role
  --active           Filter by active status (true/false)
  --broker           Filter by broker ID
  --user             Filter by user ID
  --has-user         Filter by presence of user (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List KeepTruckin users
  xbe view keep-truckin-users list

  # Filter by broker
  xbe view keep-truckin-users list --broker 123

  # Filter by role
  xbe view keep-truckin-users list --role driver

  # Filter by assignment date
  xbe view keep-truckin-users list --user-set-at-min 2025-01-01T00:00:00Z

  # Output as JSON
  xbe view keep-truckin-users list --json`,
		Args: cobra.NoArgs,
		RunE: runKeepTruckinUsersList,
	}
	initKeepTruckinUsersListFlags(cmd)
	return cmd
}

func init() {
	keepTruckinUsersCmd.AddCommand(newKeepTruckinUsersListCmd())
}

func initKeepTruckinUsersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("user-set-at-min", "", "Filter by user-set-at on/after (ISO 8601)")
	cmd.Flags().String("user-set-at-max", "", "Filter by user-set-at on/before (ISO 8601)")
	cmd.Flags().String("is-user-set-at", "", "Filter by presence of user-set-at (true/false)")
	cmd.Flags().String("assigned-at-min", "", "Alias for --user-set-at-min (ISO 8601)")
	cmd.Flags().String("assigned-at-max", "", "Alias for --user-set-at-max (ISO 8601)")
	cmd.Flags().String("role", "", "Filter by role")
	cmd.Flags().String("active", "", "Filter by active status (true/false)")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("user", "", "Filter by user ID")
	cmd.Flags().String("has-user", "", "Filter by presence of user (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runKeepTruckinUsersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseKeepTruckinUsersListOptions(cmd)
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
	query.Set("fields[keep-truckin-users]", "driver-id,first-name,last-name,role,active,carrier-name,user-set-at,broker,user")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[user_set_at_min]", opts.UserSetAtMin)
	setFilterIfPresent(query, "filter[user_set_at_max]", opts.UserSetAtMax)
	setFilterIfPresent(query, "filter[is_user_set_at]", opts.IsUserSetAt)
	setFilterIfPresent(query, "filter[assigned_at_min]", opts.AssignedAtMin)
	setFilterIfPresent(query, "filter[assigned_at_max]", opts.AssignedAtMax)
	setFilterIfPresent(query, "filter[role]", opts.Role)
	setFilterIfPresent(query, "filter[active]", opts.Active)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[user]", opts.User)
	setFilterIfPresent(query, "filter[has_user]", opts.HasUser)

	body, _, err := client.Get(cmd.Context(), "/v1/keep-truckin-users", query)
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

	rows := buildKeepTruckinUserRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderKeepTruckinUsersTable(cmd, rows)
}

func parseKeepTruckinUsersListOptions(cmd *cobra.Command) (keepTruckinUsersListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	userSetAtMin, _ := cmd.Flags().GetString("user-set-at-min")
	userSetAtMax, _ := cmd.Flags().GetString("user-set-at-max")
	isUserSetAt, _ := cmd.Flags().GetString("is-user-set-at")
	assignedAtMin, _ := cmd.Flags().GetString("assigned-at-min")
	assignedAtMax, _ := cmd.Flags().GetString("assigned-at-max")
	role, _ := cmd.Flags().GetString("role")
	active, _ := cmd.Flags().GetString("active")
	broker, _ := cmd.Flags().GetString("broker")
	user, _ := cmd.Flags().GetString("user")
	hasUser, _ := cmd.Flags().GetString("has-user")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return keepTruckinUsersListOptions{
		BaseURL:       baseURL,
		Token:         token,
		JSON:          jsonOut,
		NoAuth:        noAuth,
		Limit:         limit,
		Offset:        offset,
		Sort:          sort,
		UserSetAtMin:  userSetAtMin,
		UserSetAtMax:  userSetAtMax,
		IsUserSetAt:   isUserSetAt,
		AssignedAtMin: assignedAtMin,
		AssignedAtMax: assignedAtMax,
		Role:          role,
		Active:        active,
		Broker:        broker,
		User:          user,
		HasUser:       hasUser,
	}, nil
}

func buildKeepTruckinUserRows(resp jsonAPIResponse) []keepTruckinUserRow {
	rows := make([]keepTruckinUserRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildKeepTruckinUserRow(resource))
	}
	return rows
}

func buildKeepTruckinUserRow(resource jsonAPIResource) keepTruckinUserRow {
	attrs := resource.Attributes
	firstName := stringAttr(attrs, "first-name")
	lastName := stringAttr(attrs, "last-name")

	return keepTruckinUserRow{
		ID:          resource.ID,
		DriverID:    stringAttr(attrs, "driver-id"),
		Name:        formatKeepTruckinUserName(firstName, lastName),
		Role:        stringAttr(attrs, "role"),
		Active:      boolAttr(attrs, "active"),
		CarrierName: stringAttr(attrs, "carrier-name"),
		UserID:      relationshipIDFromMap(resource.Relationships, "user"),
		BrokerID:    relationshipIDFromMap(resource.Relationships, "broker"),
		UserSetAt:   formatDateTime(stringAttr(attrs, "user-set-at")),
	}
}

func formatKeepTruckinUserName(firstName, lastName string) string {
	firstName = strings.TrimSpace(firstName)
	lastName = strings.TrimSpace(lastName)
	if firstName == "" {
		return lastName
	}
	if lastName == "" {
		return firstName
	}
	return firstName + " " + lastName
}

func renderKeepTruckinUsersTable(cmd *cobra.Command, rows []keepTruckinUserRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No keep-truckin users found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tDRIVER ID\tNAME\tROLE\tACTIVE\tCARRIER\tUSER\tBROKER\tUSER SET AT")
	for _, row := range rows {
		active := "no"
		if row.Active {
			active = "yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.DriverID,
			truncateString(row.Name, 28),
			row.Role,
			active,
			truncateString(row.CarrierName, 24),
			row.UserID,
			row.BrokerID,
			row.UserSetAt,
		)
	}
	return writer.Flush()
}
