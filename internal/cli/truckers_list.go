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

type truckersListOptions struct {
	BaseURL                        string
	Token                          string
	JSON                           bool
	NoAuth                         bool
	Limit                          int
	Offset                         int
	Name                           string
	IsActive                       bool
	Broker                         string
	Q                              string
	PhoneNumber                    string
	Favorite                       string
	TrailerClassifications         string
	TaxIdentifier                  string
	ManagingCustomer               string
	CompanyAddressWithin           string
	WithinCustomerTruckersOf       string
	WithUninvoicedApprovedTimeCard string
	BrokerVendorID                 string
	BrokerRating                   string
	// NOTE: earliest-tender-accepted-at-within-previous-year-min removed due to server-side SQL bug
	LastShiftStartAtMin string
	LastShiftStartAtMax string
}

func newTruckersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List truckers",
		Long: `List truckers with filtering and pagination.

Returns a list of truckers matching the specified criteria.

Pagination:
  Use --limit and --offset to paginate through large result sets.

Filtering:
  Multiple filters can be combined. All filters use AND logic.

Use Case:
  Find trucker IDs for filtering posts by creator:
    xbe view posts list --creator "Trucker|<id>"`,
		Example: `  # List truckers
  xbe view truckers list

  # Search by company name
  xbe view truckers list --name "Acme"

  # Filter by active status
  xbe view truckers list --active

  # Paginate results
  xbe view truckers list --limit 20 --offset 40

  # Output as JSON
  xbe view truckers list --json`,
		RunE: runTruckersList,
	}
	initTruckersListFlags(cmd)
	return cmd
}

func init() {
	truckersCmd.AddCommand(newTruckersListCmd())
}

func initTruckersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 0, "Page size (defaults to server default)")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("name", "", "Filter by company name (partial match)")
	cmd.Flags().Bool("active", false, "Filter to only active truckers")
	cmd.Flags().String("broker", "", "Filter by broker ID (comma-separated for multiple)")
	cmd.Flags().String("q", "", "Full-text search")
	cmd.Flags().String("phone-number", "", "Filter by phone number")
	cmd.Flags().String("favorite", "", "Filter by favorite status (true/false)")
	cmd.Flags().String("trailer-classifications", "", "Filter by trailer classifications (comma-separated)")
	cmd.Flags().String("tax-identifier", "", "Filter by tax identifier")
	cmd.Flags().String("managing-customer", "", "Filter by managing customer ID (comma-separated for multiple)")
	cmd.Flags().String("company-address-within", "", "Filter by company address proximity (lat,lng:miles)")
	cmd.Flags().String("within-customer-truckers-of", "", "Filter by customer truckers (customer ID, comma-separated for multiple)")
	cmd.Flags().String("with-uninvoiced-approved-time-card", "", "Filter by having uninvoiced approved time card (true/false)")
	cmd.Flags().String("broker-vendor-id", "", "Filter by broker vendor ID")
	cmd.Flags().String("broker-rating", "", "Filter by broker rating (format: broker_id;rating1|rating2, e.g., 123;1|2|3)")
	cmd.Flags().String("last-shift-start-at-min", "", "Filter by minimum last shift start datetime (ISO 8601)")
	cmd.Flags().String("last-shift-start-at-max", "", "Filter by maximum last shift start datetime (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTruckersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTruckersListOptions(cmd)
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
	query.Set("fields[truckers]", "company-name,is-active,broker")
	query.Set("fields[brokers]", "company-name")
	query.Set("include", "broker")
	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[company_name]", opts.Name)
	if opts.IsActive {
		query.Set("filter[is_active]", "true")
	}
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[q]", opts.Q)
	setFilterIfPresent(query, "filter[phone-number]", opts.PhoneNumber)
	setFilterIfPresent(query, "filter[favorite]", opts.Favorite)
	setFilterIfPresent(query, "filter[trailer-classifications]", opts.TrailerClassifications)
	setFilterIfPresent(query, "filter[tax-identifier]", opts.TaxIdentifier)
	setFilterIfPresent(query, "filter[managing-customer]", opts.ManagingCustomer)
	setFilterIfPresent(query, "filter[company-address-within]", opts.CompanyAddressWithin)
	setFilterIfPresent(query, "filter[within-customer-truckers-of]", opts.WithinCustomerTruckersOf)
	setFilterIfPresent(query, "filter[with-uninvoiced-approved-time-card]", opts.WithUninvoicedApprovedTimeCard)
	setFilterIfPresent(query, "filter[broker-vendor-id]", opts.BrokerVendorID)
	setFilterIfPresent(query, "filter[broker-rating]", opts.BrokerRating)
	setFilterIfPresent(query, "filter[last-shift-start-at-min]", opts.LastShiftStartAtMin)
	setFilterIfPresent(query, "filter[last-shift-start-at-max]", opts.LastShiftStartAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/truckers", query)
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
		rows := buildTruckerRows(resp)
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTruckersList(cmd, resp)
}

func parseTruckersListOptions(cmd *cobra.Command) (truckersListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return truckersListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return truckersListOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return truckersListOptions{}, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return truckersListOptions{}, err
	}
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return truckersListOptions{}, err
	}
	isActive, err := cmd.Flags().GetBool("active")
	if err != nil {
		return truckersListOptions{}, err
	}
	broker, err := cmd.Flags().GetString("broker")
	if err != nil {
		return truckersListOptions{}, err
	}
	q, err := cmd.Flags().GetString("q")
	if err != nil {
		return truckersListOptions{}, err
	}
	phoneNumber, err := cmd.Flags().GetString("phone-number")
	if err != nil {
		return truckersListOptions{}, err
	}
	favorite, err := cmd.Flags().GetString("favorite")
	if err != nil {
		return truckersListOptions{}, err
	}
	trailerClassifications, err := cmd.Flags().GetString("trailer-classifications")
	if err != nil {
		return truckersListOptions{}, err
	}
	taxIdentifier, err := cmd.Flags().GetString("tax-identifier")
	if err != nil {
		return truckersListOptions{}, err
	}
	managingCustomer, err := cmd.Flags().GetString("managing-customer")
	if err != nil {
		return truckersListOptions{}, err
	}
	companyAddressWithin, err := cmd.Flags().GetString("company-address-within")
	if err != nil {
		return truckersListOptions{}, err
	}
	withinCustomerTruckersOf, err := cmd.Flags().GetString("within-customer-truckers-of")
	if err != nil {
		return truckersListOptions{}, err
	}
	withUninvoicedApprovedTimeCard, err := cmd.Flags().GetString("with-uninvoiced-approved-time-card")
	if err != nil {
		return truckersListOptions{}, err
	}
	brokerVendorID, err := cmd.Flags().GetString("broker-vendor-id")
	if err != nil {
		return truckersListOptions{}, err
	}
	brokerRating, err := cmd.Flags().GetString("broker-rating")
	if err != nil {
		return truckersListOptions{}, err
	}
	lastShiftStartAtMin, err := cmd.Flags().GetString("last-shift-start-at-min")
	if err != nil {
		return truckersListOptions{}, err
	}
	lastShiftStartAtMax, err := cmd.Flags().GetString("last-shift-start-at-max")
	if err != nil {
		return truckersListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return truckersListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return truckersListOptions{}, err
	}

	return truckersListOptions{
		BaseURL:                        baseURL,
		Token:                          token,
		JSON:                           jsonOut,
		NoAuth:                         noAuth,
		Limit:                          limit,
		Offset:                         offset,
		Name:                           name,
		IsActive:                       isActive,
		Broker:                         broker,
		Q:                              q,
		PhoneNumber:                    phoneNumber,
		Favorite:                       favorite,
		TrailerClassifications:         trailerClassifications,
		TaxIdentifier:                  taxIdentifier,
		ManagingCustomer:               managingCustomer,
		CompanyAddressWithin:           companyAddressWithin,
		WithinCustomerTruckersOf:       withinCustomerTruckersOf,
		WithUninvoicedApprovedTimeCard: withUninvoicedApprovedTimeCard,
		BrokerVendorID:                 brokerVendorID,
		BrokerRating:                   brokerRating,
		LastShiftStartAtMin:            lastShiftStartAtMin,
		LastShiftStartAtMax:            lastShiftStartAtMax,
	}, nil
}

type truckerRow struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Broker   string `json:"broker"`
	IsActive bool   `json:"is_active"`
}

func buildTruckerRows(resp jsonAPIResponse) []truckerRow {
	// Build included map for broker lookup
	included := make(map[string]map[string]any)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc.Attributes
	}

	rows := make([]truckerRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		brokerName := ""
		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			key := resourceKey(rel.Data.Type, rel.Data.ID)
			if attrs, ok := included[key]; ok {
				brokerName = strings.TrimSpace(stringAttr(attrs, "company-name"))
			}
		}

		rows = append(rows, truckerRow{
			ID:       resource.ID,
			Name:     strings.TrimSpace(stringAttr(resource.Attributes, "company-name")),
			Broker:   brokerName,
			IsActive: boolAttr(resource.Attributes, "is-active"),
		})
	}
	return rows
}

func renderTruckersList(cmd *cobra.Command, resp jsonAPIResponse) error {
	rows := buildTruckerRows(resp)
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No truckers found.")
		return nil
	}

	const nameMax = 50
	const brokerMax = 30

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tBROKER")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, nameMax),
			truncateString(row.Broker, brokerMax),
		)
	}
	return writer.Flush()
}
