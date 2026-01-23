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

type jobScheduleShiftsListOptions struct {
	BaseURL                               string
	Token                                 string
	JSON                                  bool
	NoAuth                                bool
	Limit                                 int
	Offset                                int
	Sort                                  string
	Job                                   string
	BusinessUnit                          string
	MatchesMaterialPurchaseOrderRelease   string
	OnAcceptedBrokerTender                string
	OnAcceptedCustomerTender              string
	ActiveOnTender                        string
	IsCancelled                           string
	IsManaged                             string
	IsManagedOrAlive                      string
	IsSubsequentShiftInDriverDay          string
	Unsourced                             string
	Customer                              string
	CustomerID                            string
	Broker                                string
	BrokerID                              string
	Ordered                               string
	JobProductionPlanStatus               string
	StartDate                             string
	StartDateMin                          string
	StartDateMax                          string
	HasStartDate                          string
	StartAtMin                            string
	StartAtMax                            string
	IsStartAt                             string
	EndAtMin                              string
	EndAtMax                              string
	IsEndAt                               string
	RelatedToTruckerThroughAcceptedTender string
}

type jobScheduleShiftRow struct {
	ID         string `json:"id"`
	StartAt    string `json:"start_at,omitempty"`
	EndAt      string `json:"end_at,omitempty"`
	StartDate  string `json:"start_date,omitempty"`
	IsManaged  bool   `json:"is_managed"`
	Cancelled  bool   `json:"is_cancelled"`
	JobID      string `json:"job_id,omitempty"`
	JobSite    string `json:"job_site,omitempty"`
	JobSiteID  string `json:"job_site_id,omitempty"`
	Customer   string `json:"customer,omitempty"`
	CustomerID string `json:"customer_id,omitempty"`
	Broker     string `json:"broker,omitempty"`
	BrokerID   string `json:"broker_id,omitempty"`
}

func newJobScheduleShiftsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List job schedule shifts",
		Long: `List job schedule shifts with filtering.

Output Columns:
  ID        Job schedule shift identifier
  START AT  Shift start time
  END AT    Shift end time
  MANAGED   Whether the shift is managed
  CANCELLED Whether the shift is cancelled
  JOB       Job identifier
  JOB SITE  Job site name
  CUSTOMER  Customer name
  BROKER    Broker name

Filters:
  --job                                     Filter by job ID
  --business-unit                           Filter by business unit ID
  --matches-material-purchase-order-release Filter by material purchase order release ID
  --on-accepted-broker-tender               Filter by accepted broker tender (true/false)
  --on-accepted-customer-tender             Filter by accepted customer tender (true/false)
  --active-on-tender                        Filter by tender ID with active shifts
  --is-cancelled                            Filter by cancelled status (true/false)
  --is-managed                              Filter by managed status (true/false)
  --is-managed-or-alive                     Filter by managed-or-alive status (true/false)
  --is-subsequent-shift-in-driver-day       Filter by subsequent shift in driver day (true/false)
  --unsourced                               Filter by unsourced status (true/false)
  --customer                                Filter by customer ID
  --customer-id                             Filter by customer ID (job-based)
  --broker                                  Filter by broker ID
  --broker-id                               Filter by broker ID (job-based)
  --ordered                                 Filter by ordered status (true/false)
  --job-production-plan-status              Filter by job production plan status
  --start-date                              Filter by start date (YYYY-MM-DD)
  --start-date-min                          Filter by minimum start date (YYYY-MM-DD)
  --start-date-max                          Filter by maximum start date (YYYY-MM-DD)
  --has-start-date                          Filter by presence of start date (true/false)
  --start-at-min                            Filter by minimum start time (ISO 8601)
  --start-at-max                            Filter by maximum start time (ISO 8601)
  --is-start-at                             Filter by presence of start time (true/false)
  --end-at-min                              Filter by minimum end time (ISO 8601)
  --end-at-max                              Filter by maximum end time (ISO 8601)
  --is-end-at                               Filter by presence of end time (true/false)
  --related-to-trucker-through-accepted-tender Filter by trucker ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List job schedule shifts
  xbe view job-schedule-shifts list

  # Filter by job
  xbe view job-schedule-shifts list --job 123

  # Filter by start date
  xbe view job-schedule-shifts list --start-date 2025-01-01

  # Output as JSON
  xbe view job-schedule-shifts list --json`,
		Args: cobra.NoArgs,
		RunE: runJobScheduleShiftsList,
	}
	initJobScheduleShiftsListFlags(cmd)
	return cmd
}

func init() {
	jobScheduleShiftsCmd.AddCommand(newJobScheduleShiftsListCmd())
}

func initJobScheduleShiftsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("job", "", "Filter by job ID")
	cmd.Flags().String("business-unit", "", "Filter by business unit ID")
	cmd.Flags().String("matches-material-purchase-order-release", "", "Filter by material purchase order release ID")
	cmd.Flags().String("on-accepted-broker-tender", "", "Filter by accepted broker tender (true/false)")
	cmd.Flags().String("on-accepted-customer-tender", "", "Filter by accepted customer tender (true/false)")
	cmd.Flags().String("active-on-tender", "", "Filter by tender ID with active shifts")
	cmd.Flags().String("is-cancelled", "", "Filter by cancelled status (true/false)")
	cmd.Flags().String("is-managed", "", "Filter by managed status (true/false)")
	cmd.Flags().String("is-managed-or-alive", "", "Filter by managed-or-alive status (true/false)")
	cmd.Flags().String("is-subsequent-shift-in-driver-day", "", "Filter by subsequent shift in driver day (true/false)")
	cmd.Flags().String("unsourced", "", "Filter by unsourced status (true/false)")
	cmd.Flags().String("customer", "", "Filter by customer ID")
	cmd.Flags().String("customer-id", "", "Filter by customer ID (job-based)")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("broker-id", "", "Filter by broker ID (job-based)")
	cmd.Flags().String("ordered", "", "Filter by ordered status (true/false)")
	cmd.Flags().String("job-production-plan-status", "", "Filter by job production plan status")
	cmd.Flags().String("start-date", "", "Filter by start date (YYYY-MM-DD)")
	cmd.Flags().String("start-date-min", "", "Filter by minimum start date (YYYY-MM-DD)")
	cmd.Flags().String("start-date-max", "", "Filter by maximum start date (YYYY-MM-DD)")
	cmd.Flags().String("has-start-date", "", "Filter by presence of start date (true/false)")
	cmd.Flags().String("start-at-min", "", "Filter by minimum start time (ISO 8601)")
	cmd.Flags().String("start-at-max", "", "Filter by maximum start time (ISO 8601)")
	cmd.Flags().String("is-start-at", "", "Filter by presence of start time (true/false)")
	cmd.Flags().String("end-at-min", "", "Filter by minimum end time (ISO 8601)")
	cmd.Flags().String("end-at-max", "", "Filter by maximum end time (ISO 8601)")
	cmd.Flags().String("is-end-at", "", "Filter by presence of end time (true/false)")
	cmd.Flags().String("related-to-trucker-through-accepted-tender", "", "Filter by trucker ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobScheduleShiftsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseJobScheduleShiftsListOptions(cmd)
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
	query.Set("fields[job-schedule-shifts]", "start-at,end-at,start-date,is-managed,cancelled-at,job,job-site,customer,broker")
	query.Set("fields[job-sites]", "name")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[brokers]", "company-name")
	query.Set("include", "job-site,customer,broker")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[job]", opts.Job)
	setFilterIfPresent(query, "filter[business-unit]", opts.BusinessUnit)
	setFilterIfPresent(query, "filter[matches-material-purchase-order-release]", opts.MatchesMaterialPurchaseOrderRelease)
	setFilterIfPresent(query, "filter[on-accepted-broker-tender]", opts.OnAcceptedBrokerTender)
	setFilterIfPresent(query, "filter[on-accepted-customer-tender]", opts.OnAcceptedCustomerTender)
	setFilterIfPresent(query, "filter[active-on-tender]", opts.ActiveOnTender)
	setFilterIfPresent(query, "filter[is-cancelled]", opts.IsCancelled)
	setFilterIfPresent(query, "filter[is-managed]", opts.IsManaged)
	setFilterIfPresent(query, "filter[is-managed-or-alive]", opts.IsManagedOrAlive)
	setFilterIfPresent(query, "filter[is-subsequent-shift-in-driver-day]", opts.IsSubsequentShiftInDriverDay)
	setFilterIfPresent(query, "filter[unsourced]", opts.Unsourced)
	setFilterIfPresent(query, "filter[customer]", opts.Customer)
	setFilterIfPresent(query, "filter[customer-id]", opts.CustomerID)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[broker-id]", opts.BrokerID)
	setFilterIfPresent(query, "filter[ordered]", opts.Ordered)
	setFilterIfPresent(query, "filter[job-production-plan-status]", opts.JobProductionPlanStatus)
	setFilterIfPresent(query, "filter[start-date]", opts.StartDate)
	setFilterIfPresent(query, "filter[start-date-min]", opts.StartDateMin)
	setFilterIfPresent(query, "filter[start-date-max]", opts.StartDateMax)
	setFilterIfPresent(query, "filter[has-start-date]", opts.HasStartDate)
	setFilterIfPresent(query, "filter[start-at-min]", opts.StartAtMin)
	setFilterIfPresent(query, "filter[start-at-max]", opts.StartAtMax)
	setFilterIfPresent(query, "filter[is-start-at]", opts.IsStartAt)
	setFilterIfPresent(query, "filter[end-at-min]", opts.EndAtMin)
	setFilterIfPresent(query, "filter[end-at-max]", opts.EndAtMax)
	setFilterIfPresent(query, "filter[is-end-at]", opts.IsEndAt)
	setFilterIfPresent(query, "filter[related-to-trucker-through-accepted-tender]", opts.RelatedToTruckerThroughAcceptedTender)

	body, _, err := client.Get(cmd.Context(), "/v1/job-schedule-shifts", query)
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

	rows := buildJobScheduleShiftRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderJobScheduleShiftsTable(cmd, rows)
}

func parseJobScheduleShiftsListOptions(cmd *cobra.Command) (jobScheduleShiftsListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return jobScheduleShiftsListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return jobScheduleShiftsListOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return jobScheduleShiftsListOptions{}, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return jobScheduleShiftsListOptions{}, err
	}
	sort, err := cmd.Flags().GetString("sort")
	if err != nil {
		return jobScheduleShiftsListOptions{}, err
	}
	job, err := cmd.Flags().GetString("job")
	if err != nil {
		return jobScheduleShiftsListOptions{}, err
	}
	businessUnit, err := cmd.Flags().GetString("business-unit")
	if err != nil {
		return jobScheduleShiftsListOptions{}, err
	}
	matchesRelease, err := cmd.Flags().GetString("matches-material-purchase-order-release")
	if err != nil {
		return jobScheduleShiftsListOptions{}, err
	}
	onAcceptedBrokerTender, err := cmd.Flags().GetString("on-accepted-broker-tender")
	if err != nil {
		return jobScheduleShiftsListOptions{}, err
	}
	onAcceptedCustomerTender, err := cmd.Flags().GetString("on-accepted-customer-tender")
	if err != nil {
		return jobScheduleShiftsListOptions{}, err
	}
	activeOnTender, err := cmd.Flags().GetString("active-on-tender")
	if err != nil {
		return jobScheduleShiftsListOptions{}, err
	}
	isCancelled, err := cmd.Flags().GetString("is-cancelled")
	if err != nil {
		return jobScheduleShiftsListOptions{}, err
	}
	isManaged, err := cmd.Flags().GetString("is-managed")
	if err != nil {
		return jobScheduleShiftsListOptions{}, err
	}
	isManagedOrAlive, err := cmd.Flags().GetString("is-managed-or-alive")
	if err != nil {
		return jobScheduleShiftsListOptions{}, err
	}
	isSubsequentShiftInDriverDay, err := cmd.Flags().GetString("is-subsequent-shift-in-driver-day")
	if err != nil {
		return jobScheduleShiftsListOptions{}, err
	}
	unsourced, err := cmd.Flags().GetString("unsourced")
	if err != nil {
		return jobScheduleShiftsListOptions{}, err
	}
	customer, err := cmd.Flags().GetString("customer")
	if err != nil {
		return jobScheduleShiftsListOptions{}, err
	}
	customerID, err := cmd.Flags().GetString("customer-id")
	if err != nil {
		return jobScheduleShiftsListOptions{}, err
	}
	broker, err := cmd.Flags().GetString("broker")
	if err != nil {
		return jobScheduleShiftsListOptions{}, err
	}
	brokerID, err := cmd.Flags().GetString("broker-id")
	if err != nil {
		return jobScheduleShiftsListOptions{}, err
	}
	ordered, err := cmd.Flags().GetString("ordered")
	if err != nil {
		return jobScheduleShiftsListOptions{}, err
	}
	jobProductionPlanStatus, err := cmd.Flags().GetString("job-production-plan-status")
	if err != nil {
		return jobScheduleShiftsListOptions{}, err
	}
	startDate, err := cmd.Flags().GetString("start-date")
	if err != nil {
		return jobScheduleShiftsListOptions{}, err
	}
	startDateMin, err := cmd.Flags().GetString("start-date-min")
	if err != nil {
		return jobScheduleShiftsListOptions{}, err
	}
	startDateMax, err := cmd.Flags().GetString("start-date-max")
	if err != nil {
		return jobScheduleShiftsListOptions{}, err
	}
	hasStartDate, err := cmd.Flags().GetString("has-start-date")
	if err != nil {
		return jobScheduleShiftsListOptions{}, err
	}
	startAtMin, err := cmd.Flags().GetString("start-at-min")
	if err != nil {
		return jobScheduleShiftsListOptions{}, err
	}
	startAtMax, err := cmd.Flags().GetString("start-at-max")
	if err != nil {
		return jobScheduleShiftsListOptions{}, err
	}
	isStartAt, err := cmd.Flags().GetString("is-start-at")
	if err != nil {
		return jobScheduleShiftsListOptions{}, err
	}
	endAtMin, err := cmd.Flags().GetString("end-at-min")
	if err != nil {
		return jobScheduleShiftsListOptions{}, err
	}
	endAtMax, err := cmd.Flags().GetString("end-at-max")
	if err != nil {
		return jobScheduleShiftsListOptions{}, err
	}
	isEndAt, err := cmd.Flags().GetString("is-end-at")
	if err != nil {
		return jobScheduleShiftsListOptions{}, err
	}
	relatedToTruckerThroughAcceptedTender, err := cmd.Flags().GetString("related-to-trucker-through-accepted-tender")
	if err != nil {
		return jobScheduleShiftsListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return jobScheduleShiftsListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return jobScheduleShiftsListOptions{}, err
	}

	return jobScheduleShiftsListOptions{
		BaseURL:                               baseURL,
		Token:                                 token,
		JSON:                                  jsonOut,
		NoAuth:                                noAuth,
		Limit:                                 limit,
		Offset:                                offset,
		Sort:                                  sort,
		Job:                                   job,
		BusinessUnit:                          businessUnit,
		MatchesMaterialPurchaseOrderRelease:   matchesRelease,
		OnAcceptedBrokerTender:                onAcceptedBrokerTender,
		OnAcceptedCustomerTender:              onAcceptedCustomerTender,
		ActiveOnTender:                        activeOnTender,
		IsCancelled:                           isCancelled,
		IsManaged:                             isManaged,
		IsManagedOrAlive:                      isManagedOrAlive,
		IsSubsequentShiftInDriverDay:          isSubsequentShiftInDriverDay,
		Unsourced:                             unsourced,
		Customer:                              customer,
		CustomerID:                            customerID,
		Broker:                                broker,
		BrokerID:                              brokerID,
		Ordered:                               ordered,
		JobProductionPlanStatus:               jobProductionPlanStatus,
		StartDate:                             startDate,
		StartDateMin:                          startDateMin,
		StartDateMax:                          startDateMax,
		HasStartDate:                          hasStartDate,
		StartAtMin:                            startAtMin,
		StartAtMax:                            startAtMax,
		IsStartAt:                             isStartAt,
		EndAtMin:                              endAtMin,
		EndAtMax:                              endAtMax,
		IsEndAt:                               isEndAt,
		RelatedToTruckerThroughAcceptedTender: relatedToTruckerThroughAcceptedTender,
	}, nil
}

func buildJobScheduleShiftRows(resp jsonAPIResponse) []jobScheduleShiftRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]jobScheduleShiftRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := jobScheduleShiftRow{
			ID:        resource.ID,
			StartAt:   formatDateTime(stringAttr(attrs, "start-at")),
			EndAt:     formatDateTime(stringAttr(attrs, "end-at")),
			StartDate: stringAttr(attrs, "start-date"),
			IsManaged: boolAttr(attrs, "is-managed"),
			Cancelled: strings.TrimSpace(stringAttr(attrs, "cancelled-at")) != "",
		}

		if rel, ok := resource.Relationships["job"]; ok && rel.Data != nil {
			row.JobID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["job-site"]; ok && rel.Data != nil {
			row.JobSiteID = rel.Data.ID
			if site, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.JobSite = stringAttr(site.Attributes, "name")
			}
		}
		if rel, ok := resource.Relationships["customer"]; ok && rel.Data != nil {
			row.CustomerID = rel.Data.ID
			if customer, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.Customer = stringAttr(customer.Attributes, "company-name")
			}
		}
		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			row.BrokerID = rel.Data.ID
			if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.Broker = stringAttr(broker.Attributes, "company-name")
			}
		}

		rows = append(rows, row)
	}

	return rows
}

func renderJobScheduleShiftsTable(cmd *cobra.Command, rows []jobScheduleShiftRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No job schedule shifts found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTART AT\tEND AT\tMANAGED\tCANCELLED\tJOB\tJOB SITE\tCUSTOMER\tBROKER")
	for _, row := range rows {
		managedStr := ""
		if row.IsManaged {
			managedStr = "Yes"
		}
		cancelledStr := ""
		if row.Cancelled {
			cancelledStr = "Yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.StartAt,
			row.EndAt,
			managedStr,
			cancelledStr,
			row.JobID,
			truncateString(row.JobSite, 25),
			truncateString(row.Customer, 25),
			truncateString(row.Broker, 25),
		)
	}
	return writer.Flush()
}
