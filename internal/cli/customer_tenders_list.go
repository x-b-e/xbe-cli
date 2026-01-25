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

type customerTendersListOptions struct {
	BaseURL                                            string
	Token                                              string
	JSON                                               bool
	NoAuth                                             bool
	Limit                                              int
	Offset                                             int
	Sort                                               string
	Buyer                                              string
	Seller                                             string
	Broker                                             string
	Job                                                string
	Status                                             string
	StartAtMin                                         string
	StartAtMax                                         string
	EndAtMax                                           string
	JobSite                                            string
	JobNumber                                          string
	WithAliveShifts                                    string
	HasFlexibleShifts                                  string
	JobProductionPlanNameOrNumberLike                  string
	BusinessUnit                                       string
	JobProductionPlanTrailerClassificationOrEquivalent string
	JobProductionPlanMaterialSites                     string
	JobProductionPlanMaterialTypes                     string
	ExternalIdentificationValue                        string
	CreatedAtMin                                       string
	CreatedAtMax                                       string
	UpdatedAtMin                                       string
	UpdatedAtMax                                       string
}

type customerTenderRow struct {
	ID         string `json:"id"`
	Status     string `json:"status,omitempty"`
	JobID      string `json:"job_id,omitempty"`
	JobNumber  string `json:"job_number,omitempty"`
	CustomerID string `json:"customer_id,omitempty"`
	Customer   string `json:"customer,omitempty"`
	BrokerID   string `json:"broker_id,omitempty"`
	Broker     string `json:"broker,omitempty"`
	ExpiresAt  string `json:"expires_at,omitempty"`
}

func newCustomerTendersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List customer tenders",
		Long: `List customer tenders with filtering and pagination.

Output Columns:
  ID        Customer tender ID
  STATUS    Tender status
  JOB       Job number or ID
  CUSTOMER  Customer company name
  BROKER    Broker company name
  EXPIRES   Expiration time

Filters:
  --buyer                 Filter by buyer ID (polymorphic)
  --seller                Filter by seller ID (polymorphic)
  --broker                Filter by broker ID
  --job                   Filter by job ID
  --status                Filter by status (editing, offered, accepted, rejected, expired, cancelled, returned, sourced)
  --start-at-min          Filter by earliest job shift start time (RFC3339)
  --start-at-max          Filter by latest job shift start time (RFC3339)
  --end-at-max            Filter by latest job shift end time (RFC3339)
  --job-site              Filter by job site ID
  --job-number            Filter by job number
  --with-alive-shifts     Filter by tenders with alive shifts (true/false)
  --has-flexible-shifts   Filter by flexible shifts (true/false)
  --job-production-plan-name-or-number-like  Filter by job production plan name or number (fuzzy)
  --business-unit         Filter by business unit ID(s)
  --job-production-plan-trailer-classification-or-equivalent  Filter by trailer classification ID(s)
  --job-production-plan-material-sites       Filter by job production plan material site ID(s)
  --job-production-plan-material-types       Filter by job production plan material type ID(s)
  --external-identification-value            Filter by external identification value
  --created-at-min        Filter by created-at on/after (RFC3339)
  --created-at-max        Filter by created-at on/before (RFC3339)
  --updated-at-min        Filter by updated-at on/after (RFC3339)
  --updated-at-max        Filter by updated-at on/before (RFC3339)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List customer tenders
  xbe view customer-tenders list

  # Filter by broker and status
  xbe view customer-tenders list --broker 123 --status editing

  # Filter by job number
  xbe view customer-tenders list --job-number "JOB-1001"

  # JSON output
  xbe view customer-tenders list --json`,
		RunE: runCustomerTendersList,
	}
	initCustomerTendersListFlags(cmd)
	return cmd
}

func init() {
	customerTendersCmd.AddCommand(newCustomerTendersListCmd())
}

func initCustomerTendersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort order")
	cmd.Flags().String("buyer", "", "Filter by buyer ID")
	cmd.Flags().String("seller", "", "Filter by seller ID")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("job", "", "Filter by job ID")
	cmd.Flags().String("status", "", "Filter by status (editing, offered, accepted, rejected, expired, cancelled, returned, sourced)")
	cmd.Flags().String("start-at-min", "", "Filter by earliest job shift start time (RFC3339)")
	cmd.Flags().String("start-at-max", "", "Filter by latest job shift start time (RFC3339)")
	cmd.Flags().String("end-at-max", "", "Filter by latest job shift end time (RFC3339)")
	cmd.Flags().String("job-site", "", "Filter by job site ID")
	cmd.Flags().String("job-number", "", "Filter by job number")
	cmd.Flags().String("with-alive-shifts", "", "Filter by tenders with alive shifts (true/false)")
	cmd.Flags().String("has-flexible-shifts", "", "Filter by flexible shifts (true/false)")
	cmd.Flags().String("job-production-plan-name-or-number-like", "", "Filter by job production plan name or number (fuzzy)")
	cmd.Flags().String("business-unit", "", "Filter by business unit ID(s)")
	cmd.Flags().String("job-production-plan-trailer-classification-or-equivalent", "", "Filter by trailer classification ID(s)")
	cmd.Flags().String("job-production-plan-material-sites", "", "Filter by job production plan material site ID(s)")
	cmd.Flags().String("job-production-plan-material-types", "", "Filter by job production plan material type ID(s)")
	cmd.Flags().String("external-identification-value", "", "Filter by external identification value")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (RFC3339)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (RFC3339)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (RFC3339)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (RFC3339)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCustomerTendersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseCustomerTendersListOptions(cmd)
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
	query.Set("fields[customer-tenders]", "status,expires-at,job,buyer,seller")
	query.Set("fields[jobs]", "job-number,job-name")
	query.Set("fields[customers]", "company-name,name")
	query.Set("fields[brokers]", "company-name,name")
	query.Set("include", "job,buyer,seller")

	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}
	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	setFilterIfPresent(query, "filter[buyer]", opts.Buyer)
	setFilterIfPresent(query, "filter[seller]", opts.Seller)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[job]", opts.Job)
	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[start-at-min]", opts.StartAtMin)
	setFilterIfPresent(query, "filter[start-at-max]", opts.StartAtMax)
	setFilterIfPresent(query, "filter[end-at-max]", opts.EndAtMax)
	setFilterIfPresent(query, "filter[job-site]", opts.JobSite)
	setFilterIfPresent(query, "filter[job-number]", opts.JobNumber)
	setFilterIfPresent(query, "filter[with-alive-shifts]", opts.WithAliveShifts)
	setFilterIfPresent(query, "filter[has-flexible-shifts]", opts.HasFlexibleShifts)
	setFilterIfPresent(query, "filter[job-production-plan-name-or-number-like]", opts.JobProductionPlanNameOrNumberLike)
	setFilterIfPresent(query, "filter[business-unit]", opts.BusinessUnit)
	setFilterIfPresent(query, "filter[job-production-plan-trailer-classification-or-equivalent]", opts.JobProductionPlanTrailerClassificationOrEquivalent)
	setFilterIfPresent(query, "filter[job-production-plan-material-sites]", opts.JobProductionPlanMaterialSites)
	setFilterIfPresent(query, "filter[job-production-plan-material-types]", opts.JobProductionPlanMaterialTypes)
	setFilterIfPresent(query, "filter[external-identification-value]", opts.ExternalIdentificationValue)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/customer-tenders", query)
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

	rows := buildCustomerTenderRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderCustomerTendersTable(cmd, rows)
}

func parseCustomerTendersListOptions(cmd *cobra.Command) (customerTendersListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	buyer, _ := cmd.Flags().GetString("buyer")
	seller, _ := cmd.Flags().GetString("seller")
	broker, _ := cmd.Flags().GetString("broker")
	job, _ := cmd.Flags().GetString("job")
	status, _ := cmd.Flags().GetString("status")
	startAtMin, _ := cmd.Flags().GetString("start-at-min")
	startAtMax, _ := cmd.Flags().GetString("start-at-max")
	endAtMax, _ := cmd.Flags().GetString("end-at-max")
	jobSite, _ := cmd.Flags().GetString("job-site")
	jobNumber, _ := cmd.Flags().GetString("job-number")
	withAliveShifts, _ := cmd.Flags().GetString("with-alive-shifts")
	hasFlexibleShifts, _ := cmd.Flags().GetString("has-flexible-shifts")
	jppNameOrNumberLike, _ := cmd.Flags().GetString("job-production-plan-name-or-number-like")
	businessUnit, _ := cmd.Flags().GetString("business-unit")
	jppTrailerClassification, _ := cmd.Flags().GetString("job-production-plan-trailer-classification-or-equivalent")
	jppMaterialSites, _ := cmd.Flags().GetString("job-production-plan-material-sites")
	jppMaterialTypes, _ := cmd.Flags().GetString("job-production-plan-material-types")
	externalIdentificationValue, _ := cmd.Flags().GetString("external-identification-value")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return customerTendersListOptions{
		BaseURL:                           baseURL,
		Token:                             token,
		JSON:                              jsonOut,
		NoAuth:                            noAuth,
		Limit:                             limit,
		Offset:                            offset,
		Sort:                              sort,
		Buyer:                             buyer,
		Seller:                            seller,
		Broker:                            broker,
		Job:                               job,
		Status:                            status,
		StartAtMin:                        startAtMin,
		StartAtMax:                        startAtMax,
		EndAtMax:                          endAtMax,
		JobSite:                           jobSite,
		JobNumber:                         jobNumber,
		WithAliveShifts:                   withAliveShifts,
		HasFlexibleShifts:                 hasFlexibleShifts,
		JobProductionPlanNameOrNumberLike: jppNameOrNumberLike,
		BusinessUnit:                      businessUnit,
		JobProductionPlanTrailerClassificationOrEquivalent: jppTrailerClassification,
		JobProductionPlanMaterialSites:                     jppMaterialSites,
		JobProductionPlanMaterialTypes:                     jppMaterialTypes,
		ExternalIdentificationValue:                        externalIdentificationValue,
		CreatedAtMin:                                       createdAtMin,
		CreatedAtMax:                                       createdAtMax,
		UpdatedAtMin:                                       updatedAtMin,
		UpdatedAtMax:                                       updatedAtMax,
	}, nil
}

func buildCustomerTenderRows(resp jsonAPIResponse) []customerTenderRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]customerTenderRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := customerTenderRow{
			ID:        resource.ID,
			Status:    stringAttr(attrs, "status"),
			ExpiresAt: formatDateTime(stringAttr(attrs, "expires-at")),
			JobID:     relationshipIDFromMap(resource.Relationships, "job"),
		}

		if rel, ok := resource.Relationships["buyer"]; ok && rel.Data != nil {
			if rel.Data.Type == "customers" {
				row.CustomerID = rel.Data.ID
			}
		}
		if rel, ok := resource.Relationships["seller"]; ok && rel.Data != nil {
			if rel.Data.Type == "brokers" {
				row.BrokerID = rel.Data.ID
			}
		}

		if row.JobID != "" {
			if job, ok := included[resourceKey("jobs", row.JobID)]; ok {
				row.JobNumber = firstNonEmpty(
					stringAttr(job.Attributes, "job-number"),
					stringAttr(job.Attributes, "job-name"),
				)
			}
		}

		if row.CustomerID != "" {
			if customer, ok := included[resourceKey("customers", row.CustomerID)]; ok {
				row.Customer = firstNonEmpty(
					stringAttr(customer.Attributes, "company-name"),
					stringAttr(customer.Attributes, "name"),
				)
			}
		}

		if row.BrokerID != "" {
			if broker, ok := included[resourceKey("brokers", row.BrokerID)]; ok {
				row.Broker = firstNonEmpty(
					stringAttr(broker.Attributes, "company-name"),
					stringAttr(broker.Attributes, "name"),
				)
			}
		}

		rows = append(rows, row)
	}

	return rows
}

func buildCustomerTenderRowFromSingle(resp jsonAPISingleResponse) customerTenderRow {
	rows := buildCustomerTenderRows(jsonAPIResponse{Data: []jsonAPIResource{resp.Data}, Included: resp.Included})
	if len(rows) == 0 {
		return customerTenderRow{}
	}
	return rows[0]
}

func renderCustomerTendersTable(cmd *cobra.Command, rows []customerTenderRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No customer tenders found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tJOB\tCUSTOMER\tBROKER\tEXPIRES")
	for _, row := range rows {
		jobLabel := firstNonEmpty(row.JobNumber, row.JobID)
		customerLabel := firstNonEmpty(row.Customer, row.CustomerID)
		brokerLabel := firstNonEmpty(row.Broker, row.BrokerID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Status,
			truncateString(jobLabel, 24),
			truncateString(customerLabel, 20),
			truncateString(brokerLabel, 20),
			row.ExpiresAt,
		)
	}
	return writer.Flush()
}
