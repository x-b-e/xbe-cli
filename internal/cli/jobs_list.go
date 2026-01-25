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

type jobsListOptions struct {
	BaseURL                     string
	Token                       string
	JSON                        bool
	NoAuth                      bool
	Limit                       int
	Offset                      int
	Sort                        string
	Customer                    string
	JobSite                     string
	StartDate                   string
	StartAtMin                  string
	StartAtMax                  string
	Offered                     string
	Broker                      string
	Trucker                     string
	JobProductionPlan           string
	ExternalJobNumber           string
	ExternalIdentificationValue string
}

type jobRow struct {
	ID                  string `json:"id"`
	ExternalJobNumber   string `json:"external_job_number,omitempty"`
	JobProductionPlan   string `json:"job_production_plan,omitempty"`
	JobProductionPlanID string `json:"job_production_plan_id,omitempty"`
	JobSite             string `json:"job_site,omitempty"`
	JobSiteID           string `json:"job_site_id,omitempty"`
	Customer            string `json:"customer,omitempty"`
	CustomerID          string `json:"customer_id,omitempty"`
}

func newJobsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List jobs",
		Long: `List jobs with filtering and pagination.

Output Columns:
  ID                    Job identifier
  EXTERNAL JOB NUMBER   External job number (if set)
  JOB PRODUCTION PLAN   Job production plan label
  JOB SITE              Job site name
  CUSTOMER              Customer name

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List jobs
  xbe view jobs list

  # Filter by customer and job site
  xbe view jobs list --customer 123 --job-site 456

  # Filter by start date
  xbe view jobs list --start-date 2025-01-01

  # Output as JSON
  xbe view jobs list --json`,
		Args: cobra.NoArgs,
		RunE: runJobsList,
	}
	initJobsListFlags(cmd)
	return cmd
}

func init() {
	jobsCmd.AddCommand(newJobsListCmd())
}

func initJobsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("customer", "", "Filter by customer ID (comma-separated for multiple)")
	cmd.Flags().String("job-site", "", "Filter by job site ID (comma-separated for multiple)")
	cmd.Flags().String("start-date", "", "Filter by start date (YYYY-MM-DD)")
	cmd.Flags().String("start-at-min", "", "Filter by minimum start time (ISO 8601)")
	cmd.Flags().String("start-at-max", "", "Filter by maximum start time (ISO 8601)")
	cmd.Flags().String("offered", "", "Filter by whether job has tenders (true/false)")
	cmd.Flags().String("broker", "", "Filter by broker ID (comma-separated for multiple)")
	cmd.Flags().String("trucker", "", "Filter by trucker ID (comma-separated for multiple)")
	cmd.Flags().String("job-production-plan", "", "Filter by job production plan ID (comma-separated for multiple)")
	cmd.Flags().String("external-job-number", "", "Filter by external job number")
	cmd.Flags().String("external-identification-value", "", "Filter by external identification value")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseJobsListOptions(cmd)
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
	query.Set("fields[jobs]", "external-job-number,job-site,customer,job-production-plan")
	query.Set("include", "job-site,customer,job-production-plan")
	query.Set("fields[job-sites]", "name")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[job-production-plans]", "job-number,job-name")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[customer]", opts.Customer)
	setFilterIfPresent(query, "filter[job-site]", opts.JobSite)
	setFilterIfPresent(query, "filter[start-date]", opts.StartDate)
	setFilterIfPresent(query, "filter[start-at-min]", opts.StartAtMin)
	setFilterIfPresent(query, "filter[start-at-max]", opts.StartAtMax)
	setFilterIfPresent(query, "filter[offered]", opts.Offered)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)
	setFilterIfPresent(query, "filter[job-production-plan]", opts.JobProductionPlan)
	setFilterIfPresent(query, "filter[external-job-number]", opts.ExternalJobNumber)
	setFilterIfPresent(query, "filter[external-identification-value]", opts.ExternalIdentificationValue)

	body, _, err := client.Get(cmd.Context(), "/v1/jobs", query)
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

	rows := buildJobRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderJobsTable(cmd, rows)
}

func parseJobsListOptions(cmd *cobra.Command) (jobsListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return jobsListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return jobsListOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return jobsListOptions{}, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return jobsListOptions{}, err
	}
	sort, err := cmd.Flags().GetString("sort")
	if err != nil {
		return jobsListOptions{}, err
	}
	customer, err := cmd.Flags().GetString("customer")
	if err != nil {
		return jobsListOptions{}, err
	}
	jobSite, err := cmd.Flags().GetString("job-site")
	if err != nil {
		return jobsListOptions{}, err
	}
	startDate, err := cmd.Flags().GetString("start-date")
	if err != nil {
		return jobsListOptions{}, err
	}
	startAtMin, err := cmd.Flags().GetString("start-at-min")
	if err != nil {
		return jobsListOptions{}, err
	}
	startAtMax, err := cmd.Flags().GetString("start-at-max")
	if err != nil {
		return jobsListOptions{}, err
	}
	offered, err := cmd.Flags().GetString("offered")
	if err != nil {
		return jobsListOptions{}, err
	}
	broker, err := cmd.Flags().GetString("broker")
	if err != nil {
		return jobsListOptions{}, err
	}
	trucker, err := cmd.Flags().GetString("trucker")
	if err != nil {
		return jobsListOptions{}, err
	}
	jobProductionPlan, err := cmd.Flags().GetString("job-production-plan")
	if err != nil {
		return jobsListOptions{}, err
	}
	externalJobNumber, err := cmd.Flags().GetString("external-job-number")
	if err != nil {
		return jobsListOptions{}, err
	}
	externalIdentificationValue, err := cmd.Flags().GetString("external-identification-value")
	if err != nil {
		return jobsListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return jobsListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return jobsListOptions{}, err
	}

	return jobsListOptions{
		BaseURL:                     baseURL,
		Token:                       token,
		JSON:                        jsonOut,
		NoAuth:                      noAuth,
		Limit:                       limit,
		Offset:                      offset,
		Sort:                        sort,
		Customer:                    customer,
		JobSite:                     jobSite,
		StartDate:                   startDate,
		StartAtMin:                  startAtMin,
		StartAtMax:                  startAtMax,
		Offered:                     offered,
		Broker:                      broker,
		Trucker:                     trucker,
		JobProductionPlan:           jobProductionPlan,
		ExternalJobNumber:           externalJobNumber,
		ExternalIdentificationValue: externalIdentificationValue,
	}, nil
}

func buildJobRows(resp jsonAPIResponse) []jobRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]jobRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := jobRow{
			ID:                resource.ID,
			ExternalJobNumber: stringAttr(resource.Attributes, "external-job-number"),
		}

		if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
			row.JobProductionPlanID = rel.Data.ID
			if plan, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.JobProductionPlan = jobProductionPlanLabel(plan.Attributes)
			}
		}
		if row.JobProductionPlan == "" {
			row.JobProductionPlan = row.JobProductionPlanID
		}

		if rel, ok := resource.Relationships["job-site"]; ok && rel.Data != nil {
			row.JobSiteID = rel.Data.ID
			if jobSite, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.JobSite = strings.TrimSpace(stringAttr(jobSite.Attributes, "name"))
			}
		}
		if row.JobSite == "" {
			row.JobSite = row.JobSiteID
		}

		if rel, ok := resource.Relationships["customer"]; ok && rel.Data != nil {
			row.CustomerID = rel.Data.ID
			if customer, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.Customer = strings.TrimSpace(stringAttr(customer.Attributes, "company-name"))
			}
		}
		if row.Customer == "" {
			row.Customer = row.CustomerID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderJobsTable(cmd *cobra.Command, rows []jobRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No jobs found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tEXTERNAL JOB NUMBER\tJOB PRODUCTION PLAN\tJOB SITE\tCUSTOMER")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.ExternalJobNumber, 24),
			truncateString(row.JobProductionPlan, 40),
			truncateString(row.JobSite, 28),
			truncateString(row.Customer, 28),
		)
	}
	return writer.Flush()
}

func jobRowFromSingle(resp jsonAPISingleResponse) jobRow {
	row := jobRow{
		ID:                resp.Data.ID,
		ExternalJobNumber: stringAttr(resp.Data.Attributes, "external-job-number"),
	}

	if rel, ok := resp.Data.Relationships["job-production-plan"]; ok && rel.Data != nil {
		row.JobProductionPlanID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["job-site"]; ok && rel.Data != nil {
		row.JobSiteID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["customer"]; ok && rel.Data != nil {
		row.CustomerID = rel.Data.ID
	}

	return row
}
