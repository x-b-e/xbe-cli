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

type materialTransactionsListOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	NoAuth                 bool
	Limit                  int
	Offset                 int
	Query                  string
	Status                 string
	TicketNumber           string
	Date                   string
	DateMin                string
	DateMax                string
	MaterialType           string
	MaterialSite           string
	MaterialSupplier       string
	JobProductionPlan      string
	Customer               string
	Trucker                string
	Broker                 string
	Project                string
	IsVoided               string
	IncludeAll             bool
	BusinessUnit           string
	JobSite                string
	HasShift               string
	TenderJobScheduleShift string
}

func newMaterialTransactionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List material transactions",
		Long: `List material transactions with filtering and pagination.

Returns a list of material transactions matching the specified criteria,
sorted by transaction date (newest first).

Output Columns (table format):
  ID           Unique transaction identifier
  STATUS       Current workflow state
  DATE         Transaction date
  TIME         Transaction time
  TICKET       Ticket number
  MATERIAL     Material type name
  SUPPLIER     Material supplier name
  TRUCKER      Trucking company name
  DRIVER       Driver name
  TONS         Load quantity in tons
  ORIGIN       Origin site/location
  DESTINATION  Destination site/location

Statuses:
  editing, submitted, accepted, rejected, unmatched, denied, invalidated

  By default, invalidated and denied transactions are excluded.
  Use --include-all to see all statuses.

Pagination:
  Use --limit and --offset to paginate through large result sets.`,
		Example: `  # List recent transactions
  xbe view material-transactions list

  # List transactions for a specific date
  xbe view material-transactions list --date 2025-01-15

  # List transactions in a date range
  xbe view material-transactions list --date-min 2025-01-01 --date-max 2025-01-31

  # Filter by status
  xbe view material-transactions list --status accepted
  xbe view material-transactions list --status editing,submitted

  # Search by ticket number
  xbe view material-transactions list --ticket-number T12345

  # Filter by material type
  xbe view material-transactions list --material-type 456

  # Filter by job production plan
  xbe view material-transactions list --job-production-plan 789

  # Include voided/denied transactions
  xbe view material-transactions list --include-all

  # Full-text search
  xbe view material-transactions list --q "asphalt"

  # Output as JSON
  xbe view material-transactions list --json`,
		RunE: runMaterialTransactionsList,
	}
	initMaterialTransactionsListFlags(cmd)
	return cmd
}

func init() {
	materialTransactionsCmd.AddCommand(newMaterialTransactionsListCmd())
}

func initMaterialTransactionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 0, "Page size (defaults to server default)")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("q", "", "Full-text search")
	cmd.Flags().String("status", "", "Filter by status (editing,submitted,accepted,rejected,unmatched,denied,invalidated)")
	cmd.Flags().String("ticket-number", "", "Filter by ticket number")
	cmd.Flags().String("date", "", "Filter by transaction date (YYYY-MM-DD)")
	cmd.Flags().String("date-min", "", "Filter by minimum transaction date (YYYY-MM-DD)")
	cmd.Flags().String("date-max", "", "Filter by maximum transaction date (YYYY-MM-DD)")
	cmd.Flags().String("material-type", "", "Filter by material type ID (comma-separated for multiple)")
	cmd.Flags().String("material-site", "", "Filter by material site ID (comma-separated for multiple)")
	cmd.Flags().String("material-supplier", "", "Filter by material supplier ID (comma-separated for multiple)")
	cmd.Flags().String("job-production-plan", "", "Filter by job production plan ID (comma-separated for multiple)")
	cmd.Flags().String("customer", "", "Filter by customer ID (comma-separated for multiple)")
	cmd.Flags().String("trucker", "", "Filter by trucker ID (comma-separated for multiple)")
	cmd.Flags().String("broker", "", "Filter by broker ID (comma-separated for multiple)")
	cmd.Flags().String("project", "", "Filter by project ID (comma-separated for multiple)")
	cmd.Flags().String("is-voided", "", "Filter by voided status (true/false)")
	cmd.Flags().Bool("include-all", false, "Include invalidated and denied transactions")
	cmd.Flags().String("business-unit", "", "Filter by business unit ID (comma-separated for multiple)")
	cmd.Flags().String("job-site", "", "Filter by job site ID (comma-separated for multiple)")
	cmd.Flags().String("has-shift", "", "Filter by whether transaction has a shift (true/false)")
	cmd.Flags().String("tender-job-schedule-shift", "", "Filter by tender job schedule shift ID (comma-separated for multiple)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialTransactionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMaterialTransactionsListOptions(cmd)
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
	query.Set("sort", "-transaction-at")
	query.Set("fields[material-transactions]", "status,transaction-at,time-zone-id,ticket-number,net-weight-lbs,is-voided,material-type,material-supplier,trip")
	query.Set("fields[material-types]", "name")
	query.Set("fields[material-suppliers]", "name")
	query.Set("fields[material-sites]", "name")
	query.Set("fields[job-sites]", "name")
	query.Set("fields[trips]", "origin,destination,tender-job-schedule-shift")
	query.Set("fields[tender-job-schedule-shifts]", "accepted-trucker,seller-operations-contact")
	query.Set("fields[truckers]", "company-name")
	query.Set("fields[users]", "name")
	query.Set("include", "material-type,material-supplier,trip.origin,trip.destination,trip.tender-job-schedule-shift.accepted-trucker,trip.tender-job-schedule-shift.seller-operations-contact")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	// Apply filters
	setFilterIfPresent(query, "filter[q]", opts.Query)
	setFilterIfPresent(query, "filter[ticket_number]", opts.TicketNumber)
	setFilterIfPresent(query, "filter[material_type]", opts.MaterialType)
	setFilterIfPresent(query, "filter[material_site]", opts.MaterialSite)
	setFilterIfPresent(query, "filter[material_supplier]", opts.MaterialSupplier)
	setFilterIfPresent(query, "filter[job_production_plan]", opts.JobProductionPlan)
	setFilterIfPresent(query, "filter[customer]", opts.Customer)
	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[project]", opts.Project)
	setFilterIfPresent(query, "filter[is_voided]", opts.IsVoided)
	setFilterIfPresent(query, "filter[business_unit]", opts.BusinessUnit)
	setFilterIfPresent(query, "filter[job_site]", opts.JobSite)
	setFilterIfPresent(query, "filter[has_shift]", opts.HasShift)
	setFilterIfPresent(query, "filter[tender_job_schedule_shift]", opts.TenderJobScheduleShift)

	// Date filters
	if opts.Date != "" {
		query.Set("filter[transaction_date]", opts.Date)
	} else {
		setFilterIfPresent(query, "filter[transaction_at][min]", opts.DateMin)
		setFilterIfPresent(query, "filter[transaction_at][max]", opts.DateMax)
	}

	// Status filter - by default exclude invalidated and denied
	if opts.Status != "" {
		query.Set("filter[status]", opts.Status)
	} else if opts.IncludeAll {
		// Don't set status filter, include all
	} else {
		// Default: exclude invalidated and denied
		query.Set("filter[status]", "editing,submitted,accepted,rejected,unmatched")
	}

	body, _, err := client.Get(cmd.Context(), "/v1/material-transactions", query)
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
		rows := buildMaterialTransactionRows(resp)
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderMaterialTransactionsTable(cmd, resp)
}

func parseMaterialTransactionsListOptions(cmd *cobra.Command) (materialTransactionsListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	query, err := cmd.Flags().GetString("q")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	status, err := cmd.Flags().GetString("status")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	ticketNumber, err := cmd.Flags().GetString("ticket-number")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	date, err := cmd.Flags().GetString("date")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	dateMin, err := cmd.Flags().GetString("date-min")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	dateMax, err := cmd.Flags().GetString("date-max")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	materialType, err := cmd.Flags().GetString("material-type")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	materialSite, err := cmd.Flags().GetString("material-site")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	materialSupplier, err := cmd.Flags().GetString("material-supplier")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	jobProductionPlan, err := cmd.Flags().GetString("job-production-plan")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	customer, err := cmd.Flags().GetString("customer")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	trucker, err := cmd.Flags().GetString("trucker")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	broker, err := cmd.Flags().GetString("broker")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	project, err := cmd.Flags().GetString("project")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	isVoided, err := cmd.Flags().GetString("is-voided")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	includeAll, err := cmd.Flags().GetBool("include-all")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	businessUnit, err := cmd.Flags().GetString("business-unit")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	jobSite, err := cmd.Flags().GetString("job-site")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	hasShift, err := cmd.Flags().GetString("has-shift")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	tenderJobScheduleShift, err := cmd.Flags().GetString("tender-job-schedule-shift")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}

	return materialTransactionsListOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		NoAuth:                 noAuth,
		Limit:                  limit,
		Offset:                 offset,
		Query:                  query,
		Status:                 status,
		TicketNumber:           ticketNumber,
		Date:                   date,
		DateMin:                dateMin,
		DateMax:                dateMax,
		MaterialType:           materialType,
		MaterialSite:           materialSite,
		MaterialSupplier:       materialSupplier,
		JobProductionPlan:      jobProductionPlan,
		Customer:               customer,
		Trucker:                trucker,
		Broker:                 broker,
		Project:                project,
		IsVoided:               isVoided,
		IncludeAll:             includeAll,
		BusinessUnit:           businessUnit,
		JobSite:                jobSite,
		HasShift:               hasShift,
		TenderJobScheduleShift: tenderJobScheduleShift,
	}, nil
}

func renderMaterialTransactionsTable(cmd *cobra.Command, resp jsonAPIResponse) error {
	rows := buildMaterialTransactionRows(resp)
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No material transactions found.")
		return nil
	}

	// Build included lookup
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, 32, 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tDATE\tTIME\tTICKET\tMATERIAL\tSUPPLIER\tTRUCKER\tDRIVER\tTONS\tORIGIN\tDESTINATION")

	for _, resource := range resp.Data {
		row := buildMaterialTransactionRowFromResource(resource, included)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Status,
			row.TransactionDate,
			row.TransactionTime,
			row.TicketNumber,
			truncateString(row.MaterialType, 25),
			truncateString(row.Supplier, 25),
			truncateString(row.Trucker, 20),
			truncateString(row.Driver, 20),
			row.Quantity,
			truncateString(row.Origin, 35),
			truncateString(row.Destination, 35),
		)
	}

	return writer.Flush()
}

type materialTransactionRow struct {
	ID                      string  `json:"id"`
	Status                  string  `json:"status"`
	TransactionAt           string  `json:"transaction_at"`
	TransactionDate         string  `json:"transaction_date"`
	TransactionTime         string  `json:"transaction_time"`
	TimeZoneID              string  `json:"time_zone_id,omitempty"`
	LocationTransactionAt   string  `json:"location_transaction_at"`
	LocationTransactionDate string  `json:"location_transaction_date"`
	TicketNumber            string  `json:"ticket_number"`
	MaterialType            string  `json:"material_type"`
	MaterialTypeID          string  `json:"material_type_id,omitempty"`
	Supplier                string  `json:"supplier"`
	SupplierID              string  `json:"supplier_id,omitempty"`
	Trucker                 string  `json:"trucker"`
	TruckerID               string  `json:"trucker_id,omitempty"`
	Driver                  string  `json:"driver"`
	DriverID                string  `json:"driver_id,omitempty"`
	NetWeightLbs            float64 `json:"net_weight_lbs"`
	Quantity                string  `json:"quantity"`
	Origin                  string  `json:"origin"`
	OriginID                string  `json:"origin_id,omitempty"`
	OriginType              string  `json:"origin_type,omitempty"`
	Destination             string  `json:"destination"`
	DestinationID           string  `json:"destination_id,omitempty"`
	DestinationType         string  `json:"destination_type,omitempty"`
	IsVoided                bool    `json:"is_voided"`
}

func buildMaterialTransactionRows(resp jsonAPIResponse) []materialTransactionRow {
	// Build included lookup
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	rows := make([]materialTransactionRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildMaterialTransactionRowFromResource(resource, included))
	}

	return rows
}

func buildMaterialTransactionRowFromResource(resource jsonAPIResource, included map[string]jsonAPIResource) materialTransactionRow {
	attrs := resource.Attributes
	tzID := stringAttr(attrs, "time-zone-id")

	row := materialTransactionRow{
		ID:                      resource.ID,
		Status:                  stringAttr(attrs, "status"),
		TransactionAt:           stringAttr(attrs, "transaction-at"),
		TransactionDate:         formatDate(stringAttr(attrs, "transaction-at")),
		TransactionTime:         formatTransactionTime(stringAttr(attrs, "transaction-at"), tzID),
		TimeZoneID:              tzID,
		LocationTransactionAt:   stringAttr(attrs, "location-transaction-at"),
		LocationTransactionDate: formatDate(stringAttr(attrs, "location-transaction-at")),
		TicketNumber:            stringAttr(attrs, "ticket-number"),
		IsVoided:                boolAttr(attrs, "is-voided"),
	}

	// Calculate tons from net weight (lbs / 2000)
	netWeightLbs := floatAttr(attrs, "net-weight-lbs")
	row.NetWeightLbs = netWeightLbs
	if netWeightLbs > 0 {
		tons := netWeightLbs / 2000.0
		row.Quantity = fmt.Sprintf("%.2f", tons)
	}

	// Resolve material type
	if rel, ok := resource.Relationships["material-type"]; ok && rel.Data != nil {
		row.MaterialTypeID = rel.Data.ID
		if mt, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.MaterialType = stringAttr(mt.Attributes, "name")
		}
	}

	// Resolve material supplier
	if rel, ok := resource.Relationships["material-supplier"]; ok && rel.Data != nil {
		row.SupplierID = rel.Data.ID
		if ms, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.Supplier = stringAttr(ms.Attributes, "name")
		}
	}

	// Resolve trip and nested relationships (origin, destination, trucker, driver)
	if rel, ok := resource.Relationships["trip"]; ok && rel.Data != nil {
		if trip, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			// Origin
			if originRel, ok := trip.Relationships["origin"]; ok && originRel.Data != nil {
				row.OriginID = originRel.Data.ID
				row.OriginType = originRel.Data.Type
				if origin, ok := included[resourceKey(originRel.Data.Type, originRel.Data.ID)]; ok {
					row.Origin = stringAttr(origin.Attributes, "name")
				}
			}
			// Destination
			if destRel, ok := trip.Relationships["destination"]; ok && destRel.Data != nil {
				row.DestinationID = destRel.Data.ID
				row.DestinationType = destRel.Data.Type
				if dest, ok := included[resourceKey(destRel.Data.Type, destRel.Data.ID)]; ok {
					row.Destination = stringAttr(dest.Attributes, "name")
				}
			}
			// Trucker and Driver via tender-job-schedule-shift
			if tjssRel, ok := trip.Relationships["tender-job-schedule-shift"]; ok && tjssRel.Data != nil {
				if tjss, ok := included[resourceKey(tjssRel.Data.Type, tjssRel.Data.ID)]; ok {
					// Trucker (accepted-trucker)
					if truckerRel, ok := tjss.Relationships["accepted-trucker"]; ok && truckerRel.Data != nil {
						row.TruckerID = truckerRel.Data.ID
						if trucker, ok := included[resourceKey(truckerRel.Data.Type, truckerRel.Data.ID)]; ok {
							row.Trucker = stringAttr(trucker.Attributes, "company-name")
						}
					}
					// Driver (seller-operations-contact)
					if driverRel, ok := tjss.Relationships["seller-operations-contact"]; ok && driverRel.Data != nil {
						row.DriverID = driverRel.Data.ID
						if driver, ok := included[resourceKey(driverRel.Data.Type, driverRel.Data.ID)]; ok {
							row.Driver = stringAttr(driver.Attributes, "name")
						}
					}
				}
			}
		}
	}

	return row
}

func formatTransactionTime(value, tzID string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return ""
	}
	// Convert to local timezone if provided
	if tzID != "" {
		if loc, err := time.LoadLocation(tzID); err == nil {
			parsed = parsed.In(loc)
		}
	}
	return parsed.Format("15:04")
}
