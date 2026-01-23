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

type jobProductionPlanServiceTypeUnitOfMeasureCohortsListOptions struct {
	BaseURL                        string
	Token                          string
	JSON                           bool
	NoAuth                         bool
	Limit                          int
	Offset                         int
	Sort                           string
	JobProductionPlan              string
	ServiceTypeUnitOfMeasureCohort string
}

type jobProductionPlanServiceTypeUnitOfMeasureCohortRow struct {
	ID                               string `json:"id"`
	JobProductionPlanID              string `json:"job_production_plan_id,omitempty"`
	JobProductionPlan                string `json:"job_production_plan,omitempty"`
	ServiceTypeUnitOfMeasureCohortID string `json:"service_type_unit_of_measure_cohort_id,omitempty"`
	ServiceTypeUnitOfMeasureCohort   string `json:"service_type_unit_of_measure_cohort,omitempty"`
	ServiceTypeUnitOfMeasureActive   bool   `json:"service_type_unit_of_measure_cohort_active,omitempty"`
}

func newJobProductionPlanServiceTypeUnitOfMeasureCohortsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List job production plan service type unit of measure cohorts",
		Long: `List job production plan service type unit of measure cohorts with filtering and pagination.

Output Columns:
  ID       Cohort link identifier
  PLAN     Job production plan (job number/name)
  COHORT   Service type unit of measure cohort name
  ACTIVE   Cohort active status

Filters:
  --job-production-plan                  Filter by job production plan ID
  --service-type-unit-of-measure-cohort  Filter by service type unit of measure cohort ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List job production plan service type unit of measure cohorts
  xbe view job-production-plan-service-type-unit-of-measure-cohorts list

  # Filter by job production plan
  xbe view job-production-plan-service-type-unit-of-measure-cohorts list --job-production-plan 123

  # Filter by cohort
  xbe view job-production-plan-service-type-unit-of-measure-cohorts list --service-type-unit-of-measure-cohort 456

  # JSON output
  xbe view job-production-plan-service-type-unit-of-measure-cohorts list --json`,
		Args: cobra.NoArgs,
		RunE: runJobProductionPlanServiceTypeUnitOfMeasureCohortsList,
	}
	initJobProductionPlanServiceTypeUnitOfMeasureCohortsListFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanServiceTypeUnitOfMeasureCohortsCmd.AddCommand(newJobProductionPlanServiceTypeUnitOfMeasureCohortsListCmd())
}

func initJobProductionPlanServiceTypeUnitOfMeasureCohortsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("job-production-plan", "", "Filter by job production plan ID")
	cmd.Flags().String("service-type-unit-of-measure-cohort", "", "Filter by service type unit of measure cohort ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanServiceTypeUnitOfMeasureCohortsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseJobProductionPlanServiceTypeUnitOfMeasureCohortsListOptions(cmd)
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
	query.Set("fields[job-production-plan-service-type-unit-of-measure-cohorts]", "job-production-plan,service-type-unit-of-measure-cohort")
	query.Set("fields[job-production-plans]", "job-number,job-name")
	query.Set("fields[service-type-unit-of-measure-cohorts]", "name,is-active,service-type-unit-of-measure-ids")
	query.Set("include", "job-production-plan,service-type-unit-of-measure-cohort")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[job-production-plan]", opts.JobProductionPlan)
	setFilterIfPresent(query, "filter[service-type-unit-of-measure-cohort]", opts.ServiceTypeUnitOfMeasureCohort)

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-service-type-unit-of-measure-cohorts", query)
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

	rows := buildJobProductionPlanServiceTypeUnitOfMeasureCohortRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderJobProductionPlanServiceTypeUnitOfMeasureCohortsTable(cmd, rows)
}

func parseJobProductionPlanServiceTypeUnitOfMeasureCohortsListOptions(cmd *cobra.Command) (jobProductionPlanServiceTypeUnitOfMeasureCohortsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	serviceTypeUnitOfMeasureCohort, _ := cmd.Flags().GetString("service-type-unit-of-measure-cohort")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanServiceTypeUnitOfMeasureCohortsListOptions{
		BaseURL:                        baseURL,
		Token:                          token,
		JSON:                           jsonOut,
		NoAuth:                         noAuth,
		Limit:                          limit,
		Offset:                         offset,
		Sort:                           sort,
		JobProductionPlan:              jobProductionPlan,
		ServiceTypeUnitOfMeasureCohort: serviceTypeUnitOfMeasureCohort,
	}, nil
}

func buildJobProductionPlanServiceTypeUnitOfMeasureCohortRows(resp jsonAPIResponse) []jobProductionPlanServiceTypeUnitOfMeasureCohortRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]jobProductionPlanServiceTypeUnitOfMeasureCohortRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := jobProductionPlanServiceTypeUnitOfMeasureCohortRow{
			ID: resource.ID,
		}

		if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
			row.JobProductionPlanID = rel.Data.ID
			if plan, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				jobNumber := stringAttr(plan.Attributes, "job-number")
				jobName := stringAttr(plan.Attributes, "job-name")
				if jobNumber != "" && jobName != "" {
					row.JobProductionPlan = fmt.Sprintf("%s - %s", jobNumber, jobName)
				} else {
					row.JobProductionPlan = firstNonEmpty(jobNumber, jobName)
				}
			}
		}

		if rel, ok := resource.Relationships["service-type-unit-of-measure-cohort"]; ok && rel.Data != nil {
			row.ServiceTypeUnitOfMeasureCohortID = rel.Data.ID
			if cohort, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.ServiceTypeUnitOfMeasureCohort = stringAttr(cohort.Attributes, "name")
				row.ServiceTypeUnitOfMeasureActive = boolAttr(cohort.Attributes, "is-active")
			}
		}

		rows = append(rows, row)
	}
	return rows
}

func renderJobProductionPlanServiceTypeUnitOfMeasureCohortsTable(cmd *cobra.Command, rows []jobProductionPlanServiceTypeUnitOfMeasureCohortRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No job production plan service type unit of measure cohorts found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPLAN\tCOHORT\tACTIVE")
	for _, row := range rows {
		plan := row.JobProductionPlan
		if plan == "" {
			plan = row.JobProductionPlanID
		}
		cohort := row.ServiceTypeUnitOfMeasureCohort
		if cohort == "" {
			cohort = row.ServiceTypeUnitOfMeasureCohortID
		}
		active := ""
		if row.ServiceTypeUnitOfMeasureActive {
			active = "Yes"
		}

		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(plan, 28),
			truncateString(cohort, 28),
			active,
		)
	}
	return writer.Flush()
}
