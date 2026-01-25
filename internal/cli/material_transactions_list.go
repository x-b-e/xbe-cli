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
	BaseURL                                         string
	Token                                           string
	JSON                                            bool
	NoAuth                                          bool
	Limit                                           int
	Offset                                          int
	Query                                           string
	Status                                          string
	TicketNumber                                    string
	Date                                            string
	DateMin                                         string
	DateMax                                         string
	MaterialType                                    string
	MaterialSite                                    string
	MaterialSupplier                                string
	JobProductionPlan                               string
	Customer                                        string
	Trucker                                         string
	Broker                                          string
	Project                                         string
	IsVoided                                        string
	IncludeAll                                      bool
	BusinessUnit                                    string
	JobSite                                         string
	HasShift                                        string
	TenderJobScheduleShift                          string
	Source                                          string
	SourceType                                      string
	FromImportSource                                string
	NotFromImportSource                             string
	MaterialMixDesign                               string
	MaterialTypeHierarchyLike                       string
	MaterialTypeUltimateParentMaterialType          string
	MaterialTypeNotUltimateParentMaterialType       string
	LikelyJobProductionPlan                         string
	ShiftTruckNumber                                string
	ShiftJobNumber                                  string
	RawTruckName                                    string
	RawTruckerName                                  string
	RawMaterialID                                   string
	RawJobNumber                                    string
	RawHaulerType                                   string
	RawIsVoided                                     string
	RawIsMillings                                   string
	SalesCustomer                                   string
	JobOrSalesCustomer                              string
	TruckerShiftSet                                 string
	Invoice                                         string
	IsPlanDriverExpectingMaterialTransactionInspect string
	ExplicitConfirmationOfMatchAccuracy             string
	HasMaterialMixDesign                            string
	MissingRequiredMixDesign                        string
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
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
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
	cmd.Flags().String("source", "", "Filter by source (Type|ID, comma-separated for multiple)")
	cmd.Flags().String("source-type", "", "Filter by source type")
	cmd.Flags().String("from-import-source", "", "Filter by import source (comma-separated for multiple)")
	cmd.Flags().String("not-from-import-source", "", "Exclude import source (comma-separated for multiple)")
	cmd.Flags().String("material-mix-design", "", "Filter by material mix design ID (comma-separated for multiple)")
	cmd.Flags().String("material-type-hierarchy-like", "", "Filter by material type hierarchy (partial match)")
	cmd.Flags().String("material-type-ultimate-parent", "", "Filter by ultimate parent material type ID (comma-separated for multiple)")
	cmd.Flags().String("material-type-not-ultimate-parent", "", "Exclude ultimate parent material type ID (comma-separated for multiple)")
	cmd.Flags().String("likely-job-production-plan", "", "Filter by likely job production plan ID (comma-separated for multiple)")
	cmd.Flags().String("shift-truck-number", "", "Filter by shift truck number")
	cmd.Flags().String("shift-job-number", "", "Filter by shift job number")
	cmd.Flags().String("raw-truck-name", "", "Filter by raw truck name")
	cmd.Flags().String("raw-trucker-name", "", "Filter by raw trucker name")
	cmd.Flags().String("raw-material-id", "", "Filter by raw material ID")
	cmd.Flags().String("raw-job-number", "", "Filter by raw job number")
	cmd.Flags().String("raw-hauler-type", "", "Filter by raw hauler type")
	cmd.Flags().String("raw-is-voided", "", "Filter by raw voided status (true/false)")
	cmd.Flags().String("raw-is-millings", "", "Filter by raw millings status (true/false)")
	cmd.Flags().String("sales-customer", "", "Filter by sales customer ID (comma-separated for multiple)")
	cmd.Flags().String("job-or-sales-customer", "", "Filter by job or sales customer ID (comma-separated for multiple)")
	cmd.Flags().String("trucker-shift-set", "", "Filter by trucker shift set (driver day)")
	cmd.Flags().String("invoice", "", "Filter by invoice ID (comma-separated for multiple)")
	cmd.Flags().String("is-plan-driver-expecting-inspection", "", "Filter by driver expecting inspection (true/false)")
	cmd.Flags().String("explicit-confirmation-of-match-accuracy", "", "Filter by explicit confirmation (true/false)")
	cmd.Flags().String("has-material-mix-design", "", "Filter by having material mix design (true/false)")
	cmd.Flags().String("missing-required-mix-design", "", "Filter by missing required mix design (true/false)")
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
	setFilterIfPresent(query, "filter[source]", opts.Source)
	setFilterIfPresent(query, "filter[source_type]", opts.SourceType)
	setFilterIfPresent(query, "filter[from_import_source]", opts.FromImportSource)
	setFilterIfPresent(query, "filter[not_from_import_source]", opts.NotFromImportSource)
	setFilterIfPresent(query, "filter[material_mix_design]", opts.MaterialMixDesign)
	setFilterIfPresent(query, "filter[material_type_hierarchy_like]", opts.MaterialTypeHierarchyLike)
	setFilterIfPresent(query, "filter[material_type_ultimate_parent_material_type]", opts.MaterialTypeUltimateParentMaterialType)
	setFilterIfPresent(query, "filter[material_type_not_ultimate_parent_material_type]", opts.MaterialTypeNotUltimateParentMaterialType)
	setFilterIfPresent(query, "filter[likely_job_production_plan]", opts.LikelyJobProductionPlan)
	setFilterIfPresent(query, "filter[shift_truck_number]", opts.ShiftTruckNumber)
	setFilterIfPresent(query, "filter[shift_job_number]", opts.ShiftJobNumber)
	setFilterIfPresent(query, "filter[raw_truck_name]", opts.RawTruckName)
	setFilterIfPresent(query, "filter[raw_trucker_name]", opts.RawTruckerName)
	setFilterIfPresent(query, "filter[raw_material_id]", opts.RawMaterialID)
	setFilterIfPresent(query, "filter[raw_job_number]", opts.RawJobNumber)
	setFilterIfPresent(query, "filter[raw_hauler_type]", opts.RawHaulerType)
	setFilterIfPresent(query, "filter[raw_is_voided]", opts.RawIsVoided)
	setFilterIfPresent(query, "filter[raw_is_millings]", opts.RawIsMillings)
	setFilterIfPresent(query, "filter[sales_customer]", opts.SalesCustomer)
	setFilterIfPresent(query, "filter[job_or_sales_customer]", opts.JobOrSalesCustomer)
	setFilterIfPresent(query, "filter[trucker_shift_set]", opts.TruckerShiftSet)
	setFilterIfPresent(query, "filter[invoice]", opts.Invoice)
	setFilterIfPresent(query, "filter[is_plan_driver_expecting_material_transaction_inspection]", opts.IsPlanDriverExpectingMaterialTransactionInspect)
	setFilterIfPresent(query, "filter[explicit_confirmation_of_match_accuracy]", opts.ExplicitConfirmationOfMatchAccuracy)
	setFilterIfPresent(query, "filter[has_material_mix_design]", opts.HasMaterialMixDesign)
	setFilterIfPresent(query, "filter[missing_required_mix_design]", opts.MissingRequiredMixDesign)

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

	handled, err := renderSparseListIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
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
	source, err := cmd.Flags().GetString("source")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	sourceType, err := cmd.Flags().GetString("source-type")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	fromImportSource, err := cmd.Flags().GetString("from-import-source")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	notFromImportSource, err := cmd.Flags().GetString("not-from-import-source")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	materialMixDesign, err := cmd.Flags().GetString("material-mix-design")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	materialTypeHierarchyLike, err := cmd.Flags().GetString("material-type-hierarchy-like")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	materialTypeUltimateParent, err := cmd.Flags().GetString("material-type-ultimate-parent")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	materialTypeNotUltimateParent, err := cmd.Flags().GetString("material-type-not-ultimate-parent")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	likelyJobProductionPlan, err := cmd.Flags().GetString("likely-job-production-plan")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	shiftTruckNumber, err := cmd.Flags().GetString("shift-truck-number")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	shiftJobNumber, err := cmd.Flags().GetString("shift-job-number")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	rawTruckName, err := cmd.Flags().GetString("raw-truck-name")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	rawTruckerName, err := cmd.Flags().GetString("raw-trucker-name")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	rawMaterialID, err := cmd.Flags().GetString("raw-material-id")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	rawJobNumber, err := cmd.Flags().GetString("raw-job-number")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	rawHaulerType, err := cmd.Flags().GetString("raw-hauler-type")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	rawIsVoided, err := cmd.Flags().GetString("raw-is-voided")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	rawIsMillings, err := cmd.Flags().GetString("raw-is-millings")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	salesCustomer, err := cmd.Flags().GetString("sales-customer")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	jobOrSalesCustomer, err := cmd.Flags().GetString("job-or-sales-customer")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	truckerShiftSet, err := cmd.Flags().GetString("trucker-shift-set")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	invoice, err := cmd.Flags().GetString("invoice")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	isPlanDriverExpectingInspection, err := cmd.Flags().GetString("is-plan-driver-expecting-inspection")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	explicitConfirmationOfMatchAccuracy, err := cmd.Flags().GetString("explicit-confirmation-of-match-accuracy")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	hasMaterialMixDesign, err := cmd.Flags().GetString("has-material-mix-design")
	if err != nil {
		return materialTransactionsListOptions{}, err
	}
	missingRequiredMixDesign, err := cmd.Flags().GetString("missing-required-mix-design")
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
		BaseURL:                                baseURL,
		Token:                                  token,
		JSON:                                   jsonOut,
		NoAuth:                                 noAuth,
		Limit:                                  limit,
		Offset:                                 offset,
		Query:                                  query,
		Status:                                 status,
		TicketNumber:                           ticketNumber,
		Date:                                   date,
		DateMin:                                dateMin,
		DateMax:                                dateMax,
		MaterialType:                           materialType,
		MaterialSite:                           materialSite,
		MaterialSupplier:                       materialSupplier,
		JobProductionPlan:                      jobProductionPlan,
		Customer:                               customer,
		Trucker:                                trucker,
		Broker:                                 broker,
		Project:                                project,
		IsVoided:                               isVoided,
		IncludeAll:                             includeAll,
		BusinessUnit:                           businessUnit,
		JobSite:                                jobSite,
		HasShift:                               hasShift,
		TenderJobScheduleShift:                 tenderJobScheduleShift,
		Source:                                 source,
		SourceType:                             sourceType,
		FromImportSource:                       fromImportSource,
		NotFromImportSource:                    notFromImportSource,
		MaterialMixDesign:                      materialMixDesign,
		MaterialTypeHierarchyLike:              materialTypeHierarchyLike,
		MaterialTypeUltimateParentMaterialType: materialTypeUltimateParent,
		MaterialTypeNotUltimateParentMaterialType: materialTypeNotUltimateParent,
		LikelyJobProductionPlan:                   likelyJobProductionPlan,
		ShiftTruckNumber:                          shiftTruckNumber,
		ShiftJobNumber:                            shiftJobNumber,
		RawTruckName:                              rawTruckName,
		RawTruckerName:                            rawTruckerName,
		RawMaterialID:                             rawMaterialID,
		RawJobNumber:                              rawJobNumber,
		RawHaulerType:                             rawHaulerType,
		RawIsVoided:                               rawIsVoided,
		RawIsMillings:                             rawIsMillings,
		SalesCustomer:                             salesCustomer,
		JobOrSalesCustomer:                        jobOrSalesCustomer,
		TruckerShiftSet:                           truckerShiftSet,
		Invoice:                                   invoice,
		IsPlanDriverExpectingMaterialTransactionInspect: isPlanDriverExpectingInspection,
		ExplicitConfirmationOfMatchAccuracy:             explicitConfirmationOfMatchAccuracy,
		HasMaterialMixDesign:                            hasMaterialMixDesign,
		MissingRequiredMixDesign:                        missingRequiredMixDesign,
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
