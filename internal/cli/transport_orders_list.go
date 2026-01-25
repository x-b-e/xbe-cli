package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

const transportOrdersDefaultPageLimit = 200

var transportOrderIncludes = []string{
	"customer",
	"project-office",
	"project-category",
	"transport-order-stops",
	"transport-order-stops.location",
	"transport-order-materials",
	"transport-order-materials.material-type",
}

type transportOrdersListOptions struct {
	BaseURL                      string
	Token                        string
	JSON                         bool
	NoAuth                       bool
	Limit                        int
	Offset                       int
	Broker                       string
	StartOn                      string
	EndOn                        string
	OrderNumber                  string
	ProjectOffice                string
	ProjectCategory              string
	Customer                     string
	MaterialType                 string
	IsManaged                    string
	PickupAddressState           string
	DeliveryAddressState         string
	Q                            string
	Status                       string
	DisableDateDefaults          bool
	Project                      string
	Active                       string
	Unplanned                    string
	PickupAtMin                  string
	PickupAtMax                  string
	DeliveryAtMin                string
	DeliveryAtMax                string
	PickupLocation               string
	DeliveryLocation             string
	BillableMiles                string
	BillableMilesMin             string
	BillableMilesMax             string
	OrderedAt                    string
	OrderedAtMin                 string
	OrderedAtMax                 string
	ProjectDivision              string
	ProjectTransportOrganization string
	MaybeActive                  string
	NearPickupLocation           string
	NearDeliveryLocation         string
}

type transportOrderRow struct {
	ID              string  `json:"id"`
	OrderNumber     string  `json:"order_number"`
	Status          string  `json:"status"`
	Customer        string  `json:"customer"`
	ProjectOffice   string  `json:"project_office"`
	ProjectCategory string  `json:"project_category"`
	PickupAt        string  `json:"pickup_at"`
	DeliveryAt      string  `json:"delivery_at"`
	Miles           float64 `json:"miles"`
	Managed         bool    `json:"managed"`
}

func newTransportOrdersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List transport orders",
		Long: `List transport orders with basic filters.

By default, this uses the "Today & Tomorrow" window unless you supply
--start-on/--end-on or an --order-number filter (which disables date filtering).`,
		Example: `  # List transport orders for a broker (defaults to today & tomorrow)
  xbe view transport-orders list --broker 297

  # Filter by order number (disables date filtering)
  xbe view transport-orders list --broker 297 --order-number 4114407

  # Filter by date window
  xbe view transport-orders list --broker 297 --start-on 2026-01-16 --end-on 2026-01-16

  # Filter by office/category/customer
  xbe view transport-orders list --broker 297 --project-office 10 --project-category 5 --customer 22

  # JSON output
  xbe view transport-orders list --broker 297 --json`,
		RunE: runTransportOrdersList,
	}
	initTransportOrdersListFlags(cmd)
	return cmd
}

func init() {
	transportOrdersCmd.AddCommand(newTransportOrdersListCmd())
}

func initTransportOrdersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 0, "Page size (defaults to fetching all pages)")
	cmd.Flags().Int("offset", 0, "Page offset (used with --limit or as start offset for full fetch)")
	cmd.Flags().String("broker", "", "Broker/branch ID (required)")
	cmd.Flags().String("start-on", "", "Start date (YYYY-MM-DD). Defaults to today when not using --order-number")
	cmd.Flags().String("end-on", "", "End date (YYYY-MM-DD). Defaults to tomorrow when not using --order-number")
	cmd.Flags().Bool("no-date-defaults", false, "Disable default today/tomorrow date window")
	cmd.Flags().String("order-number", "", "Filter by transport order number (fuzzy match, disables date filtering)")
	cmd.Flags().String("status", "", "Filter by status (editing/submitted/approved/cancelled/complete/scrapped)")
	cmd.Flags().String("project-office", "", "Filter by project office ID(s), comma-separated")
	cmd.Flags().String("project-category", "", "Filter by project category ID(s), comma-separated")
	cmd.Flags().String("customer", "", "Filter by customer ID(s), comma-separated")
	cmd.Flags().String("material-type", "", "Filter by material type ID(s), comma-separated")
	cmd.Flags().String("is-managed", "", "Filter managed orders (true/false)")
	cmd.Flags().String("pickup-address-state", "", "Filter by pickup City-StateCode (e.g., Dallas-TX)")
	cmd.Flags().String("delivery-address-state", "", "Filter by delivery City-StateCode (e.g., Austin-TX)")
	cmd.Flags().String("q", "", "Server-side search query")
	cmd.Flags().String("project", "", "Filter by project ID (comma-separated for multiple)")
	cmd.Flags().String("active", "", "Filter by active status (true/false)")
	cmd.Flags().String("unplanned", "", "Filter by unplanned status (true/false)")
	cmd.Flags().String("pickup-at-min", "", "Filter by minimum pickup time (RFC3339)")
	cmd.Flags().String("pickup-at-max", "", "Filter by maximum pickup time (RFC3339)")
	cmd.Flags().String("delivery-at-min", "", "Filter by minimum delivery time (RFC3339)")
	cmd.Flags().String("delivery-at-max", "", "Filter by maximum delivery time (RFC3339)")
	cmd.Flags().String("pickup-location", "", "Filter by pickup location ID (comma-separated for multiple)")
	cmd.Flags().String("delivery-location", "", "Filter by delivery location ID (comma-separated for multiple)")
	cmd.Flags().String("billable-miles", "", "Filter by billable miles (exact)")
	cmd.Flags().String("billable-miles-min", "", "Filter by minimum billable miles")
	cmd.Flags().String("billable-miles-max", "", "Filter by maximum billable miles")
	cmd.Flags().String("ordered-at", "", "Filter by ordered at time (RFC3339)")
	cmd.Flags().String("ordered-at-min", "", "Filter by minimum ordered at time (RFC3339)")
	cmd.Flags().String("ordered-at-max", "", "Filter by maximum ordered at time (RFC3339)")
	cmd.Flags().String("project-division", "", "Filter by project division ID (comma-separated for multiple)")
	cmd.Flags().String("project-transport-organization", "", "Filter by project transport organization ID (comma-separated for multiple)")
	cmd.Flags().String("maybe-active", "", "Filter by maybe active status (true/false)")
	cmd.Flags().String("near-pickup-location", "", "Filter by pickup location proximity (lat|lng|miles, e.g. 40.7128|-74.0060|10)")
	cmd.Flags().String("near-delivery-location", "", "Filter by delivery location proximity (lat|lng|miles, e.g. 40.7128|-74.0060|25)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTransportOrdersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTransportOrdersListOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Broker) == "" {
		err := fmt.Errorf("--broker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if err := validateDateInputs(opts.StartOn, opts.EndOn); err != nil {
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
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("include", strings.Join(transportOrderIncludes, ","))
	query.Set("filter[broker]", opts.Broker)
	applyTransportOrderFilters(query, opts)

	data, included, err := fetchTransportOrders(cmd, client, query, opts.Limit, opts.Offset)
	if err != nil {
		return err
	}
	if handled, err := renderSparseListIfRequested(cmd, jsonAPIResponse{Data: data, Included: included}); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	} else if handled {
		return nil
	}

	apiData := newTransportOrdersAPIData(data, included)
	rows := buildTransportOrderRows(apiData, data)

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTransportOrdersTable(cmd, rows)
}

func parseTransportOrdersListOptions(cmd *cobra.Command) (transportOrdersListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return transportOrdersListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return transportOrdersListOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return transportOrdersListOptions{}, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return transportOrdersListOptions{}, err
	}
	broker, err := cmd.Flags().GetString("broker")
	if err != nil {
		return transportOrdersListOptions{}, err
	}
	startOn, err := cmd.Flags().GetString("start-on")
	if err != nil {
		return transportOrdersListOptions{}, err
	}
	endOn, err := cmd.Flags().GetString("end-on")
	if err != nil {
		return transportOrdersListOptions{}, err
	}
	noDateDefaults, err := cmd.Flags().GetBool("no-date-defaults")
	if err != nil {
		return transportOrdersListOptions{}, err
	}
	orderNumber, err := cmd.Flags().GetString("order-number")
	if err != nil {
		return transportOrdersListOptions{}, err
	}
	status, err := cmd.Flags().GetString("status")
	if err != nil {
		return transportOrdersListOptions{}, err
	}
	projectOffice, err := cmd.Flags().GetString("project-office")
	if err != nil {
		return transportOrdersListOptions{}, err
	}
	projectCategory, err := cmd.Flags().GetString("project-category")
	if err != nil {
		return transportOrdersListOptions{}, err
	}
	customer, err := cmd.Flags().GetString("customer")
	if err != nil {
		return transportOrdersListOptions{}, err
	}
	materialType, err := cmd.Flags().GetString("material-type")
	if err != nil {
		return transportOrdersListOptions{}, err
	}
	isManaged, err := cmd.Flags().GetString("is-managed")
	if err != nil {
		return transportOrdersListOptions{}, err
	}
	pickupState, err := cmd.Flags().GetString("pickup-address-state")
	if err != nil {
		return transportOrdersListOptions{}, err
	}
	deliveryState, err := cmd.Flags().GetString("delivery-address-state")
	if err != nil {
		return transportOrdersListOptions{}, err
	}
	q, err := cmd.Flags().GetString("q")
	if err != nil {
		return transportOrdersListOptions{}, err
	}
	project, err := cmd.Flags().GetString("project")
	if err != nil {
		return transportOrdersListOptions{}, err
	}
	active, err := cmd.Flags().GetString("active")
	if err != nil {
		return transportOrdersListOptions{}, err
	}
	unplanned, err := cmd.Flags().GetString("unplanned")
	if err != nil {
		return transportOrdersListOptions{}, err
	}
	pickupAtMin, err := cmd.Flags().GetString("pickup-at-min")
	if err != nil {
		return transportOrdersListOptions{}, err
	}
	pickupAtMax, err := cmd.Flags().GetString("pickup-at-max")
	if err != nil {
		return transportOrdersListOptions{}, err
	}
	deliveryAtMin, err := cmd.Flags().GetString("delivery-at-min")
	if err != nil {
		return transportOrdersListOptions{}, err
	}
	deliveryAtMax, err := cmd.Flags().GetString("delivery-at-max")
	if err != nil {
		return transportOrdersListOptions{}, err
	}
	pickupLocation, err := cmd.Flags().GetString("pickup-location")
	if err != nil {
		return transportOrdersListOptions{}, err
	}
	deliveryLocation, err := cmd.Flags().GetString("delivery-location")
	if err != nil {
		return transportOrdersListOptions{}, err
	}
	billableMiles, err := cmd.Flags().GetString("billable-miles")
	if err != nil {
		return transportOrdersListOptions{}, err
	}
	billableMilesMin, err := cmd.Flags().GetString("billable-miles-min")
	if err != nil {
		return transportOrdersListOptions{}, err
	}
	billableMilesMax, err := cmd.Flags().GetString("billable-miles-max")
	if err != nil {
		return transportOrdersListOptions{}, err
	}
	orderedAt, err := cmd.Flags().GetString("ordered-at")
	if err != nil {
		return transportOrdersListOptions{}, err
	}
	orderedAtMin, err := cmd.Flags().GetString("ordered-at-min")
	if err != nil {
		return transportOrdersListOptions{}, err
	}
	orderedAtMax, err := cmd.Flags().GetString("ordered-at-max")
	if err != nil {
		return transportOrdersListOptions{}, err
	}
	projectDivision, err := cmd.Flags().GetString("project-division")
	if err != nil {
		return transportOrdersListOptions{}, err
	}
	projectTransportOrganization, err := cmd.Flags().GetString("project-transport-organization")
	if err != nil {
		return transportOrdersListOptions{}, err
	}
	maybeActive, err := cmd.Flags().GetString("maybe-active")
	if err != nil {
		return transportOrdersListOptions{}, err
	}
	nearPickupLocation, err := cmd.Flags().GetString("near-pickup-location")
	if err != nil {
		return transportOrdersListOptions{}, err
	}
	nearDeliveryLocation, err := cmd.Flags().GetString("near-delivery-location")
	if err != nil {
		return transportOrdersListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return transportOrdersListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return transportOrdersListOptions{}, err
	}

	return transportOrdersListOptions{
		BaseURL:                      baseURL,
		Token:                        token,
		JSON:                         jsonOut,
		NoAuth:                       noAuth,
		Limit:                        limit,
		Offset:                       offset,
		Broker:                       broker,
		StartOn:                      startOn,
		EndOn:                        endOn,
		OrderNumber:                  orderNumber,
		ProjectOffice:                projectOffice,
		ProjectCategory:              projectCategory,
		Customer:                     customer,
		MaterialType:                 materialType,
		IsManaged:                    isManaged,
		PickupAddressState:           pickupState,
		DeliveryAddressState:         deliveryState,
		Q:                            q,
		Status:                       status,
		DisableDateDefaults:          noDateDefaults,
		Project:                      project,
		Active:                       active,
		Unplanned:                    unplanned,
		PickupAtMin:                  pickupAtMin,
		PickupAtMax:                  pickupAtMax,
		DeliveryAtMin:                deliveryAtMin,
		DeliveryAtMax:                deliveryAtMax,
		PickupLocation:               pickupLocation,
		DeliveryLocation:             deliveryLocation,
		BillableMiles:                billableMiles,
		BillableMilesMin:             billableMilesMin,
		BillableMilesMax:             billableMilesMax,
		OrderedAt:                    orderedAt,
		OrderedAtMin:                 orderedAtMin,
		OrderedAtMax:                 orderedAtMax,
		ProjectDivision:              projectDivision,
		ProjectTransportOrganization: projectTransportOrganization,
		MaybeActive:                  maybeActive,
		NearPickupLocation:           nearPickupLocation,
		NearDeliveryLocation:         nearDeliveryLocation,
	}, nil
}

func validateDateInputs(startOn, endOn string) error {
	if strings.TrimSpace(startOn) != "" {
		if _, err := time.Parse("2006-01-02", startOn); err != nil {
			return fmt.Errorf("invalid --start-on (expected YYYY-MM-DD): %s", startOn)
		}
	}
	if strings.TrimSpace(endOn) != "" {
		if _, err := time.Parse("2006-01-02", endOn); err != nil {
			return fmt.Errorf("invalid --end-on (expected YYYY-MM-DD): %s", endOn)
		}
	}
	return nil
}

func applyTransportOrderFilters(query url.Values, opts transportOrdersListOptions) {
	setFilterIfPresent(query, "filter[project-category]", opts.ProjectCategory)
	setFilterIfPresent(query, "filter[project-office]", opts.ProjectOffice)
	setFilterIfPresent(query, "filter[customer]", opts.Customer)
	setFilterIfPresent(query, "filter[material-type]", opts.MaterialType)
	setFilterIfPresent(query, "filter[status]", opts.Status)

	if opts.OrderNumber != "" {
		query.Set("filter[ext-order-number]", opts.OrderNumber)
	} else {
		applyStopsDateFilters(query, opts)
	}

	if opts.IsManaged != "" {
		query.Set("filter[is-managed]", opts.IsManaged)
	}

	applyCityStateFilter(query, opts.PickupAddressState, "pickup")
	applyCityStateFilter(query, opts.DeliveryAddressState, "delivery")

	if strings.TrimSpace(opts.Q) != "" {
		query.Set("filter[q]", strings.TrimSpace(opts.Q))
	}

	setFilterIfPresent(query, "filter[project]", opts.Project)
	setFilterIfPresent(query, "filter[active]", opts.Active)
	setFilterIfPresent(query, "filter[unplanned]", opts.Unplanned)
	setFilterIfPresent(query, "filter[pickup-at-min]", opts.PickupAtMin)
	setFilterIfPresent(query, "filter[pickup-at-max]", opts.PickupAtMax)
	setFilterIfPresent(query, "filter[delivery-at-min]", opts.DeliveryAtMin)
	setFilterIfPresent(query, "filter[delivery-at-max]", opts.DeliveryAtMax)
	setFilterIfPresent(query, "filter[pickup-location]", opts.PickupLocation)
	setFilterIfPresent(query, "filter[delivery-location]", opts.DeliveryLocation)
	setFilterIfPresent(query, "filter[billable-miles]", opts.BillableMiles)
	setFilterIfPresent(query, "filter[billable-miles-min]", opts.BillableMilesMin)
	setFilterIfPresent(query, "filter[billable-miles-max]", opts.BillableMilesMax)
	setFilterIfPresent(query, "filter[ordered-at]", opts.OrderedAt)
	setFilterIfPresent(query, "filter[ordered-at-min]", opts.OrderedAtMin)
	setFilterIfPresent(query, "filter[ordered-at-max]", opts.OrderedAtMax)
	setFilterIfPresent(query, "filter[project-division]", opts.ProjectDivision)
	setFilterIfPresent(query, "filter[project-transport-organization]", opts.ProjectTransportOrganization)
	setFilterIfPresent(query, "filter[maybe-active]", opts.MaybeActive)
	setFilterIfPresent(query, "filter[near-pickup-location]", opts.NearPickupLocation)
	setFilterIfPresent(query, "filter[near-delivery-location]", opts.NearDeliveryLocation)
}

func applyStopsDateFilters(query url.Values, opts transportOrdersListOptions) {
	startOn := strings.TrimSpace(opts.StartOn)
	endOn := strings.TrimSpace(opts.EndOn)

	if !opts.DisableDateDefaults && startOn == "" && endOn == "" {
		now := time.Now()
		startOn = now.Format("2006-01-02")
		endOn = now.Add(24 * time.Hour).Format("2006-01-02")
	}

	startAt := parseDateWithTime(startOn, false)
	endAt := parseDateWithTime(endOn, true)

	if startAt != nil {
		query.Set("filter[stops-at-max-min]", startAt.Format(time.RFC3339))
	}
	if endAt != nil {
		query.Set("filter[stops-at-min-max]", endAt.Format(time.RFC3339))
	}
}

func applyCityStateFilter(query url.Values, value, prefix string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	city, state := splitCityState(value)
	if city != "" {
		query.Set("filter[q]", city)
	}
	if state != "" {
		key := fmt.Sprintf("filter[%s-address-state-codes]", prefix)
		query.Set(key, state)
	}
}

func splitCityState(value string) (string, string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", ""
	}
	parts := strings.Split(value, "-")
	if len(parts) < 2 {
		return value, ""
	}
	state := strings.TrimSpace(parts[len(parts)-1])
	city := strings.TrimSpace(strings.Join(parts[:len(parts)-1], "-"))
	return city, state
}

func parseDateWithTime(value string, endOfDay bool) *time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	parsed, err := time.ParseInLocation("2006-01-02", value, time.Local)
	if err != nil {
		return nil
	}
	if endOfDay {
		end := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 0, parsed.Location())
		return &end
	}
	start := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, parsed.Location())
	return &start
}

func fetchTransportOrders(cmd *cobra.Command, client *api.Client, query url.Values, limit, offset int) ([]jsonAPIResource, []jsonAPIResource, error) {
	if limit > 0 {
		query.Set("page[limit]", strconv.Itoa(limit))
		if offset > 0 {
			query.Set("page[offset]", strconv.Itoa(offset))
		}
		return fetchTransportOrdersPage(cmd, client, query)
	}

	pageLimit := transportOrdersDefaultPageLimit
	offsetValue := offset

	allData := make([]jsonAPIResource, 0)
	includedMap := make(map[string]jsonAPIResource)

	for {
		pageQuery := cloneValues(query)
		pageQuery.Set("page[limit]", strconv.Itoa(pageLimit))
		if offsetValue > 0 {
			pageQuery.Set("page[offset]", strconv.Itoa(offsetValue))
		}

		data, included, err := fetchTransportOrdersPage(cmd, client, pageQuery)
		if err != nil {
			return nil, nil, err
		}

		allData = append(allData, data...)
		for _, inc := range included {
			includedMap[resourceKey(inc.Type, inc.ID)] = inc
		}

		if len(data) < pageLimit {
			break
		}
		offsetValue += pageLimit
	}

	included := make([]jsonAPIResource, 0, len(includedMap))
	for _, inc := range includedMap {
		included = append(included, inc)
	}

	return allData, included, nil
}

func fetchTransportOrdersPage(cmd *cobra.Command, client *api.Client, query url.Values) ([]jsonAPIResource, []jsonAPIResource, error) {
	body, _, err := client.Get(cmd.Context(), "/v1/transport-orders", query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return nil, nil, err
	}

	var resp jsonAPIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return nil, nil, err
	}

	return resp.Data, resp.Included, nil
}

type transportOrdersAPIData struct {
	resources map[string]jsonAPIResource
}

func newTransportOrdersAPIData(data, included []jsonAPIResource) *transportOrdersAPIData {
	resources := make(map[string]jsonAPIResource)
	for _, inc := range included {
		resources[resourceKey(inc.Type, inc.ID)] = inc
	}
	for _, res := range data {
		resources[resourceKey(res.Type, res.ID)] = res
	}
	return &transportOrdersAPIData{resources: resources}
}

func (d *transportOrdersAPIData) resource(typ, id string) (jsonAPIResource, bool) {
	res, ok := d.resources[resourceKey(typ, id)]
	return res, ok
}

func (d *transportOrdersAPIData) related(res jsonAPIResource, relName string) []jsonAPIResource {
	rel, ok := res.Relationships[relName]
	if !ok || rel.raw == nil {
		return nil
	}
	ids := relationshipIDs(rel)
	if len(ids) == 0 {
		return nil
	}
	items := make([]jsonAPIResource, 0, len(ids))
	for _, id := range ids {
		if item, ok := d.resource(id.Type, id.ID); ok {
			items = append(items, item)
		}
	}
	return items
}

func (d *transportOrdersAPIData) relatedOne(res jsonAPIResource, relName string) *jsonAPIResource {
	rel, ok := res.Relationships[relName]
	if !ok || rel.Data == nil {
		return nil
	}
	item, ok := d.resource(rel.Data.Type, rel.Data.ID)
	if !ok {
		return nil
	}
	return &item
}

func relationshipIDs(rel jsonAPIRelationship) []jsonAPIResourceIdentifier {
	if rel.raw == nil {
		return nil
	}
	var ids []jsonAPIResourceIdentifier
	if err := json.Unmarshal(rel.raw, &ids); err != nil {
		return nil
	}
	return ids
}

func buildTransportOrderRows(apiData *transportOrdersAPIData, orders []jsonAPIResource) []transportOrderRow {
	rows := make([]transportOrderRow, 0, len(orders))
	for _, order := range orders {
		rows = append(rows, buildTransportOrderRow(apiData, order))
	}
	return rows
}

func buildTransportOrderRow(apiData *transportOrdersAPIData, order jsonAPIResource) transportOrderRow {
	orderNumber := stringAttr(order.Attributes, "external-order-number")
	if orderNumber == "" {
		orderNumber = order.ID
	}

	customerName := ""
	if customer := apiData.relatedOne(order, "customer"); customer != nil {
		customerName = firstNonEmpty(
			stringAttr(customer.Attributes, "company-name"),
			stringAttr(customer.Attributes, "name"),
		)
	}

	projectOffice := ""
	if office := apiData.relatedOne(order, "project-office"); office != nil {
		projectOffice = stringAttr(office.Attributes, "name")
	}

	projectCategory := ""
	if category := apiData.relatedOne(order, "project-category"); category != nil {
		projectCategory = stringAttr(category.Attributes, "name")
	}

	pickupAt, deliveryAt := pickupDeliveryTimes(apiData, order)

	return transportOrderRow{
		ID:              order.ID,
		OrderNumber:     orderNumber,
		Status:          stringAttr(order.Attributes, "status"),
		Customer:        customerName,
		ProjectOffice:   projectOffice,
		ProjectCategory: projectCategory,
		PickupAt:        formatTimeValue(pickupAt),
		DeliveryAt:      formatTimeValue(deliveryAt),
		Miles:           floatAttr(order.Attributes, "billable-miles"),
		Managed:         boolAttr(order.Attributes, "is-managed"),
	}
}

func pickupDeliveryTimes(apiData *transportOrdersAPIData, order jsonAPIResource) (*time.Time, *time.Time) {
	var pickup *time.Time
	var delivery *time.Time
	stops := apiData.related(order, "transport-order-stops")
	for _, stop := range stops {
		role := stringAttr(stop.Attributes, "role")
		atMin := parseTimeAttr(stop.Attributes, "at-min")
		if atMin == nil {
			continue
		}
		if role == "pickup" {
			if pickup == nil || atMin.Before(*pickup) {
				pickup = atMin
			}
		}
		if role == "delivery" {
			if delivery == nil || atMin.Before(*delivery) {
				delivery = atMin
			}
		}
	}
	return pickup, delivery
}

func formatTimeValue(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format("2006-01-02 15:04")
}

func parseTimeAttr(attrs map[string]any, key string) *time.Time {
	value := stringAttr(attrs, key)
	if value == "" {
		return nil
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return &parsed
	}
	if parsed, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return &parsed
	}
	if parsed, err := time.ParseInLocation("2006-01-02", value, time.Local); err == nil {
		return &parsed
	}
	return nil
}

func renderTransportOrdersTable(cmd *cobra.Command, rows []transportOrderRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No transport orders found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, 32, 0)
	fmt.Fprintln(writer, "ID\tORDER\tSTATUS\tCUSTOMER\tOFFICE\tCATEGORY\tPICKUP\tDELIVERY\tMILES\tMANAGED")

	for _, row := range rows {
		managed := ""
		if row.Managed {
			managed = "Y"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.OrderNumber, 18),
			row.Status,
			truncateString(row.Customer, 24),
			truncateString(row.ProjectOffice, 18),
			truncateString(row.ProjectCategory, 18),
			row.PickupAt,
			row.DeliveryAt,
			formatMiles(row.Miles),
			managed,
		)
	}

	return writer.Flush()
}

func formatMiles(miles float64) string {
	if miles == 0 {
		return ""
	}
	return fmt.Sprintf("%.1f", miles)
}

func cloneValues(values url.Values) url.Values {
	clone := url.Values{}
	for key, vals := range values {
		out := make([]string, len(vals))
		copy(out, vals)
		clone[key] = out
	}
	return clone
}

func parseCommaList(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
