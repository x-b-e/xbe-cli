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

type materialTransactionInspectionsListOptions struct {
	BaseURL                  string
	Token                    string
	JSON                     bool
	NoAuth                   bool
	Limit                    int
	Offset                   int
	MaterialTransaction      string
	ChangedBy                string
	Status                   string
	Strategy                 string
	DeliverySiteType         string
	DeliverySiteID           string
	Trip                     string
	TripID                   string
	TenderJobScheduleShift   string
	TenderJobScheduleShiftID string
	Customer                 string
	CustomerID               string
	Broker                   string
	BrokerID                 string
	MaterialSupplier         string
	MaterialSupplierID       string
	JobProductionPlan        string
	JobProductionPlanID      string
}

type materialTransactionInspectionRow struct {
	ID                    string  `json:"id"`
	Status                string  `json:"status"`
	Strategy              string  `json:"strategy"`
	Note                  string  `json:"note,omitempty"`
	MaterialTransactionID string  `json:"material_transaction_id,omitempty"`
	DeliverySite          string  `json:"delivery_site,omitempty"`
	DeliverySiteID        string  `json:"delivery_site_id,omitempty"`
	DeliverySiteType      string  `json:"delivery_site_type,omitempty"`
	ChangedByName         string  `json:"changed_by_name,omitempty"`
	ChangedByID           string  `json:"changed_by_id,omitempty"`
	TonsAccepted          float64 `json:"tons_accepted,omitempty"`
	TonsRejected          float64 `json:"tons_rejected,omitempty"`
}

func newMaterialTransactionInspectionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List material transaction inspections",
		Long: `List material transaction inspections with filtering and pagination.

Output Columns:
  ID             Inspection identifier
  STATUS         Inspection status (open/closed)
  STRATEGY       Inspection strategy
  TRANSACTION    Material transaction ID
  DELIVERY SITE  Delivery site name (if available)
  CHANGED BY     Last change user
  ACCEPTED       Tons accepted
  REJECTED       Tons rejected

Filters:
  --material-transaction          Filter by material transaction ID
  --changed-by                    Filter by changed-by user ID
  --status                        Filter by status (open,closed)
  --strategy                      Filter by strategy (delivery_site_personnel)
  --delivery-site-type            Filter by delivery site type (JobSite, MaterialSite)
  --delivery-site-id              Filter by delivery site ID
  --trip                          Filter by trip ID
  --trip-id                       Filter by trip ID (join filter)
  --tender-job-schedule-shift     Filter by tender job schedule shift ID
  --tender-job-schedule-shift-id  Filter by tender job schedule shift ID (join filter)
  --customer                      Filter by customer ID
  --customer-id                   Filter by customer ID (join filter)
  --broker                        Filter by broker ID
  --broker-id                     Filter by broker ID (join filter)
  --material-supplier             Filter by material supplier ID
  --material-supplier-id          Filter by material supplier ID (join filter)
  --job-production-plan           Filter by job production plan ID
  --job-production-plan-id        Filter by job production plan ID (join filter)`,
		Example: `  # List inspections
  xbe view material-transaction-inspections list

  # Filter by material transaction
  xbe view material-transaction-inspections list --material-transaction 123

  # Filter by status
  xbe view material-transaction-inspections list --status open

  # Filter by delivery site
  xbe view material-transaction-inspections list --delivery-site-type MaterialSite --delivery-site-id 456

  # Output as JSON
  xbe view material-transaction-inspections list --json`,
		RunE: runMaterialTransactionInspectionsList,
	}
	initMaterialTransactionInspectionsListFlags(cmd)
	return cmd
}

func init() {
	materialTransactionInspectionsCmd.AddCommand(newMaterialTransactionInspectionsListCmd())
}

func initMaterialTransactionInspectionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("material-transaction", "", "Filter by material transaction ID")
	cmd.Flags().String("changed-by", "", "Filter by changed-by user ID")
	cmd.Flags().String("status", "", "Filter by status (open,closed)")
	cmd.Flags().String("strategy", "", "Filter by strategy (delivery_site_personnel)")
	cmd.Flags().String("delivery-site-type", "", "Filter by delivery site type (JobSite, MaterialSite)")
	cmd.Flags().String("delivery-site-id", "", "Filter by delivery site ID")
	cmd.Flags().String("trip", "", "Filter by trip ID")
	cmd.Flags().String("trip-id", "", "Filter by trip ID (join filter)")
	cmd.Flags().String("tender-job-schedule-shift", "", "Filter by tender job schedule shift ID")
	cmd.Flags().String("tender-job-schedule-shift-id", "", "Filter by tender job schedule shift ID (join filter)")
	cmd.Flags().String("customer", "", "Filter by customer ID")
	cmd.Flags().String("customer-id", "", "Filter by customer ID (join filter)")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("broker-id", "", "Filter by broker ID (join filter)")
	cmd.Flags().String("material-supplier", "", "Filter by material supplier ID")
	cmd.Flags().String("material-supplier-id", "", "Filter by material supplier ID (join filter)")
	cmd.Flags().String("job-production-plan", "", "Filter by job production plan ID")
	cmd.Flags().String("job-production-plan-id", "", "Filter by job production plan ID (join filter)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialTransactionInspectionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMaterialTransactionInspectionsListOptions(cmd)
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
	query.Set("fields[material-transaction-inspections]", "note,status,strategy,changed-by-name,tons-accepted,tons-rejected,material-transaction-id,delivery-site,changed-by,material-transaction")
	query.Set("fields[job-sites]", "name")
	query.Set("fields[material-sites]", "name")
	query.Set("include", "delivery-site")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	setFilterIfPresent(query, "filter[material_transaction]", opts.MaterialTransaction)
	setFilterIfPresent(query, "filter[changed_by]", opts.ChangedBy)
	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[strategy]", opts.Strategy)
	if opts.DeliverySiteType != "" && opts.DeliverySiteID != "" {
		query.Set("filter[delivery_site]", opts.DeliverySiteType+"|"+opts.DeliverySiteID)
	}
	setFilterIfPresent(query, "filter[trip]", opts.Trip)
	setFilterIfPresent(query, "filter[trip_id]", opts.TripID)
	setFilterIfPresent(query, "filter[tender_job_schedule_shift]", opts.TenderJobScheduleShift)
	setFilterIfPresent(query, "filter[tender_job_schedule_shift_id]", opts.TenderJobScheduleShiftID)
	setFilterIfPresent(query, "filter[customer]", opts.Customer)
	setFilterIfPresent(query, "filter[customer_id]", opts.CustomerID)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[broker_id]", opts.BrokerID)
	setFilterIfPresent(query, "filter[material_supplier]", opts.MaterialSupplier)
	setFilterIfPresent(query, "filter[material_supplier_id]", opts.MaterialSupplierID)
	setFilterIfPresent(query, "filter[job_production_plan]", opts.JobProductionPlan)
	setFilterIfPresent(query, "filter[job_production_plan_id]", opts.JobProductionPlanID)

	body, _, err := client.Get(cmd.Context(), "/v1/material-transaction-inspections", query)
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

	rows := buildMaterialTransactionInspectionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderMaterialTransactionInspectionsTable(cmd, rows)
}

func parseMaterialTransactionInspectionsListOptions(cmd *cobra.Command) (materialTransactionInspectionsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	materialTransaction, _ := cmd.Flags().GetString("material-transaction")
	changedBy, _ := cmd.Flags().GetString("changed-by")
	status, _ := cmd.Flags().GetString("status")
	strategy, _ := cmd.Flags().GetString("strategy")
	deliverySiteType, _ := cmd.Flags().GetString("delivery-site-type")
	deliverySiteID, _ := cmd.Flags().GetString("delivery-site-id")
	trip, _ := cmd.Flags().GetString("trip")
	tripID, _ := cmd.Flags().GetString("trip-id")
	tenderJobScheduleShift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	tenderJobScheduleShiftID, _ := cmd.Flags().GetString("tender-job-schedule-shift-id")
	customer, _ := cmd.Flags().GetString("customer")
	customerID, _ := cmd.Flags().GetString("customer-id")
	broker, _ := cmd.Flags().GetString("broker")
	brokerID, _ := cmd.Flags().GetString("broker-id")
	materialSupplier, _ := cmd.Flags().GetString("material-supplier")
	materialSupplierID, _ := cmd.Flags().GetString("material-supplier-id")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	jobProductionPlanID, _ := cmd.Flags().GetString("job-production-plan-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialTransactionInspectionsListOptions{
		BaseURL:                  baseURL,
		Token:                    token,
		JSON:                     jsonOut,
		NoAuth:                   noAuth,
		Limit:                    limit,
		Offset:                   offset,
		MaterialTransaction:      materialTransaction,
		ChangedBy:                changedBy,
		Status:                   status,
		Strategy:                 strategy,
		DeliverySiteType:         deliverySiteType,
		DeliverySiteID:           deliverySiteID,
		Trip:                     trip,
		TripID:                   tripID,
		TenderJobScheduleShift:   tenderJobScheduleShift,
		TenderJobScheduleShiftID: tenderJobScheduleShiftID,
		Customer:                 customer,
		CustomerID:               customerID,
		Broker:                   broker,
		BrokerID:                 brokerID,
		MaterialSupplier:         materialSupplier,
		MaterialSupplierID:       materialSupplierID,
		JobProductionPlan:        jobProductionPlan,
		JobProductionPlanID:      jobProductionPlanID,
	}, nil
}

func buildMaterialTransactionInspectionRows(resp jsonAPIResponse) []materialTransactionInspectionRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]materialTransactionInspectionRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := materialTransactionInspectionRow{
			ID:                    resource.ID,
			Status:                stringAttr(resource.Attributes, "status"),
			Strategy:              stringAttr(resource.Attributes, "strategy"),
			Note:                  stringAttr(resource.Attributes, "note"),
			ChangedByName:         stringAttr(resource.Attributes, "changed-by-name"),
			TonsAccepted:          floatAttr(resource.Attributes, "tons-accepted"),
			TonsRejected:          floatAttr(resource.Attributes, "tons-rejected"),
			MaterialTransactionID: stringAttr(resource.Attributes, "material-transaction-id"),
		}

		if rel, ok := resource.Relationships["material-transaction"]; ok && rel.Data != nil {
			row.MaterialTransactionID = rel.Data.ID
		}

		if rel, ok := resource.Relationships["changed-by"]; ok && rel.Data != nil {
			row.ChangedByID = rel.Data.ID
			if row.ChangedByName == "" {
				if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
					row.ChangedByName = stringAttr(user.Attributes, "name")
				}
			}
		}

		if rel, ok := resource.Relationships["delivery-site"]; ok && rel.Data != nil {
			row.DeliverySiteType = rel.Data.Type
			row.DeliverySiteID = rel.Data.ID
			if site, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.DeliverySite = stringAttr(site.Attributes, "name")
			}
		}

		rows = append(rows, row)
	}

	return rows
}

func renderMaterialTransactionInspectionsTable(cmd *cobra.Command, rows []materialTransactionInspectionRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No material transaction inspections found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, 32, 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tSTRATEGY\tTRANSACTION\tDELIVERY SITE\tCHANGED BY\tACCEPTED\tREJECTED")

	for _, row := range rows {
		deliverySite := formatInspectionDeliverySite(row)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Status,
			row.Strategy,
			row.MaterialTransactionID,
			truncateString(deliverySite, 30),
			truncateString(row.ChangedByName, 20),
			formatInspectionTons(row.TonsAccepted),
			formatInspectionTons(row.TonsRejected),
		)
	}

	return writer.Flush()
}

func formatInspectionDeliverySite(row materialTransactionInspectionRow) string {
	if row.DeliverySite != "" && row.DeliverySiteType != "" {
		return fmt.Sprintf("%s (%s)", row.DeliverySite, row.DeliverySiteType)
	}
	if row.DeliverySite != "" {
		return row.DeliverySite
	}
	if row.DeliverySiteType != "" && row.DeliverySiteID != "" {
		return fmt.Sprintf("%s:%s", row.DeliverySiteType, row.DeliverySiteID)
	}
	return row.DeliverySiteID
}

func formatInspectionTons(value float64) string {
	return fmt.Sprintf("%.2f", value)
}
