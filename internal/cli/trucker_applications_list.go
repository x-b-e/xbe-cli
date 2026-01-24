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

type truckerApplicationsListOptions struct {
	BaseURL              string
	Token                string
	JSON                 bool
	NoAuth               bool
	Limit                int
	Offset               int
	Sort                 string
	Broker               string
	Status               string
	Q                    string
	PhoneNumber          string
	CompanyAddressWithin string
}

type truckerApplicationRow struct {
	ID                      string `json:"id"`
	CompanyName             string `json:"company_name,omitempty"`
	Status                  string `json:"status,omitempty"`
	BrokerID                string `json:"broker_id,omitempty"`
	BrokerName              string `json:"broker_name,omitempty"`
	UserID                  string `json:"user_id,omitempty"`
	UserName                string `json:"user_name,omitempty"`
	UserEmail               string `json:"user_email,omitempty"`
	TruckerID               string `json:"trucker_id,omitempty"`
	TruckerName             string `json:"trucker_name,omitempty"`
	DistanceFromSearchMiles any    `json:"distance_from_search_miles,omitempty"`
}

func newTruckerApplicationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List trucker applications",
		Long: `List trucker applications with filtering and pagination.

Output Columns:
  ID       Application identifier
  COMPANY  Company name
  STATUS   Application status
  BROKER   Broker name or ID
  USER     User name or email
  TRUCKER  Trucker name or ID (if approved)
  DIST MI  Distance in miles (when using --company-address-within)

Filters:
  --broker                 Filter by broker ID
  --status                 Filter by status (pending, reviewing, denied, approved)
  --q                      Search applications (company name)
  --phone-number           Filter by phone number
  --company-address-within Filter by company address proximity (lat,lng:miles)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List trucker applications
  xbe view trucker-applications list

  # Filter by broker
  xbe view trucker-applications list --broker 123

  # Filter by status
  xbe view trucker-applications list --status pending

  # Filter by proximity (lat,lng:miles)
  xbe view trucker-applications list --company-address-within "41.8781,-87.6298:25"

  # Output as JSON
  xbe view trucker-applications list --json`,
		Args: cobra.NoArgs,
		RunE: runTruckerApplicationsList,
	}
	initTruckerApplicationsListFlags(cmd)
	return cmd
}

func init() {
	truckerApplicationsCmd.AddCommand(newTruckerApplicationsListCmd())
}

func initTruckerApplicationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("status", "", "Filter by status (pending, reviewing, denied, approved)")
	cmd.Flags().String("q", "", "Search applications")
	cmd.Flags().String("phone-number", "", "Filter by phone number")
	cmd.Flags().String("company-address-within", "", "Filter by company address proximity (lat,lng:miles)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTruckerApplicationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTruckerApplicationsListOptions(cmd)
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
	query.Set("fields[trucker-applications]", "company-name,status,broker,user,trucker,distance-from-search-miles")
	query.Set("include", "broker,user,trucker")
	query.Set("fields[brokers]", "company-name")
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

	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[q]", opts.Q)
	setFilterIfPresent(query, "filter[phone-number]", opts.PhoneNumber)
	setFilterIfPresent(query, "filter[company-address-within]", opts.CompanyAddressWithin)

	body, _, err := client.Get(cmd.Context(), "/v1/trucker-applications", query)
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

	rows := buildTruckerApplicationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTruckerApplicationsTable(cmd, rows)
}

func parseTruckerApplicationsListOptions(cmd *cobra.Command) (truckerApplicationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	broker, _ := cmd.Flags().GetString("broker")
	status, _ := cmd.Flags().GetString("status")
	q, _ := cmd.Flags().GetString("q")
	phoneNumber, _ := cmd.Flags().GetString("phone-number")
	companyAddressWithin, _ := cmd.Flags().GetString("company-address-within")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return truckerApplicationsListOptions{
		BaseURL:              baseURL,
		Token:                token,
		JSON:                 jsonOut,
		NoAuth:               noAuth,
		Limit:                limit,
		Offset:               offset,
		Sort:                 sort,
		Broker:               broker,
		Status:               status,
		Q:                    q,
		PhoneNumber:          phoneNumber,
		CompanyAddressWithin: companyAddressWithin,
	}, nil
}

func buildTruckerApplicationRows(resp jsonAPIResponse) []truckerApplicationRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]truckerApplicationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildTruckerApplicationRow(resource, included))
	}
	return rows
}

func truckerApplicationRowFromSingle(resp jsonAPISingleResponse) truckerApplicationRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}
	return buildTruckerApplicationRow(resp.Data, included)
}

func buildTruckerApplicationRow(resource jsonAPIResource, included map[string]jsonAPIResource) truckerApplicationRow {
	row := truckerApplicationRow{
		ID:                      resource.ID,
		CompanyName:             stringAttr(resource.Attributes, "company-name"),
		Status:                  stringAttr(resource.Attributes, "status"),
		DistanceFromSearchMiles: resource.Attributes["distance-from-search-miles"],
	}

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

	if rel, ok := resource.Relationships["trucker"]; ok && rel.Data != nil {
		row.TruckerID = rel.Data.ID
		if trucker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.TruckerName = stringAttr(trucker.Attributes, "company-name")
		}
	}

	return row
}

func renderTruckerApplicationsTable(cmd *cobra.Command, rows []truckerApplicationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No trucker applications found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tCOMPANY\tSTATUS\tBROKER\tUSER\tTRUCKER\tDIST MI")
	for _, row := range rows {
		brokerLabel := row.BrokerName
		if brokerLabel == "" {
			brokerLabel = row.BrokerID
		}

		userLabel := row.UserName
		if userLabel == "" {
			userLabel = row.UserEmail
		}
		if userLabel == "" {
			userLabel = row.UserID
		}

		truckerLabel := row.TruckerName
		if truckerLabel == "" {
			truckerLabel = row.TruckerID
		}

		distance := formatDistanceMiles(row.DistanceFromSearchMiles)

		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.CompanyName, 28),
			truncateString(row.Status, 12),
			truncateString(brokerLabel, 20),
			truncateString(userLabel, 20),
			truncateString(truckerLabel, 20),
			distance,
		)
	}
	return writer.Flush()
}
