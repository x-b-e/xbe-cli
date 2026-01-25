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

type jobProductionPlanServiceTypeUnitOfMeasuresListOptions struct {
	BaseURL                  string
	Token                    string
	JSON                     bool
	NoAuth                   bool
	Limit                    int
	Offset                   int
	Sort                     string
	JobProductionPlan        string
	ServiceTypeUnitOfMeasure string
}

type jobProductionPlanServiceTypeUnitOfMeasureRow struct {
	ID                          string `json:"id"`
	JobProductionPlan           string `json:"job_production_plan_id,omitempty"`
	ServiceTypeUnitOfMeasure    string `json:"service_type_unit_of_measure_id,omitempty"`
	StepSize                    string `json:"step_size,omitempty"`
	ExplicitStepSizeTarget      string `json:"explicit_step_size_target,omitempty"`
	ExcludeFromTimeCardInvoices bool   `json:"exclude_from_time_card_invoices"`
}

func newJobProductionPlanServiceTypeUnitOfMeasuresListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List job production plan service type unit of measures",
		Long: `List job production plan service type unit of measures.

Output Columns:
  ID               Job production plan service type unit of measure ID
  JOB PLAN         Job production plan ID
  SERVICE TYPE UOM Service type unit of measure ID
  STEP SIZE        Step size rule (no_step, ceiling, floor, standard_rounding_up)
  STEP TARGET      Explicit step size target
  EXCLUDE INVOICES Exclude from time card invoices

Filters:
  --job-production-plan          Filter by job production plan ID
  --service-type-unit-of-measure Filter by service type unit of measure ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List job production plan service type unit of measures
  xbe view job-production-plan-service-type-unit-of-measures list

  # Filter by job production plan
  xbe view job-production-plan-service-type-unit-of-measures list --job-production-plan 123

  # Filter by service type unit of measure
  xbe view job-production-plan-service-type-unit-of-measures list --service-type-unit-of-measure 456

  # Output as JSON
  xbe view job-production-plan-service-type-unit-of-measures list --json`,
		Args: cobra.NoArgs,
		RunE: runJobProductionPlanServiceTypeUnitOfMeasuresList,
	}
	initJobProductionPlanServiceTypeUnitOfMeasuresListFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanServiceTypeUnitOfMeasuresCmd.AddCommand(newJobProductionPlanServiceTypeUnitOfMeasuresListCmd())
}

func initJobProductionPlanServiceTypeUnitOfMeasuresListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("job-production-plan", "", "Filter by job production plan ID")
	cmd.Flags().String("service-type-unit-of-measure", "", "Filter by service type unit of measure ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanServiceTypeUnitOfMeasuresList(cmd *cobra.Command, _ []string) error {
	opts, err := parseJobProductionPlanServiceTypeUnitOfMeasuresListOptions(cmd)
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
	query.Set("fields[job-production-plan-service-type-unit-of-measures]", "step-size,explicit-step-size-target,exclude-from-time-card-invoices,job-production-plan,service-type-unit-of-measure")

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
	setFilterIfPresent(query, "filter[service-type-unit-of-measure]", opts.ServiceTypeUnitOfMeasure)

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-service-type-unit-of-measures", query)
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

	rows := buildJobProductionPlanServiceTypeUnitOfMeasureRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderJobProductionPlanServiceTypeUnitOfMeasuresTable(cmd, rows)
}

func parseJobProductionPlanServiceTypeUnitOfMeasuresListOptions(cmd *cobra.Command) (jobProductionPlanServiceTypeUnitOfMeasuresListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	serviceTypeUnitOfMeasure, _ := cmd.Flags().GetString("service-type-unit-of-measure")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanServiceTypeUnitOfMeasuresListOptions{
		BaseURL:                  baseURL,
		Token:                    token,
		JSON:                     jsonOut,
		NoAuth:                   noAuth,
		Limit:                    limit,
		Offset:                   offset,
		Sort:                     sort,
		JobProductionPlan:        jobProductionPlan,
		ServiceTypeUnitOfMeasure: serviceTypeUnitOfMeasure,
	}, nil
}

func buildJobProductionPlanServiceTypeUnitOfMeasureRows(resp jsonAPIResponse) []jobProductionPlanServiceTypeUnitOfMeasureRow {
	rows := make([]jobProductionPlanServiceTypeUnitOfMeasureRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := buildJobProductionPlanServiceTypeUnitOfMeasureRow(resource)
		rows = append(rows, row)
	}
	return rows
}

func buildJobProductionPlanServiceTypeUnitOfMeasureRow(resource jsonAPIResource) jobProductionPlanServiceTypeUnitOfMeasureRow {
	row := jobProductionPlanServiceTypeUnitOfMeasureRow{
		ID:                          resource.ID,
		StepSize:                    stringAttr(resource.Attributes, "step-size"),
		ExplicitStepSizeTarget:      stringAttr(resource.Attributes, "explicit-step-size-target"),
		ExcludeFromTimeCardInvoices: boolAttr(resource.Attributes, "exclude-from-time-card-invoices"),
	}

	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		row.JobProductionPlan = rel.Data.ID
	}
	if rel, ok := resource.Relationships["service-type-unit-of-measure"]; ok && rel.Data != nil {
		row.ServiceTypeUnitOfMeasure = rel.Data.ID
	}

	return row
}

func buildJobProductionPlanServiceTypeUnitOfMeasureRowFromSingle(resp jsonAPISingleResponse) jobProductionPlanServiceTypeUnitOfMeasureRow {
	return buildJobProductionPlanServiceTypeUnitOfMeasureRow(resp.Data)
}

func renderJobProductionPlanServiceTypeUnitOfMeasuresTable(cmd *cobra.Command, rows []jobProductionPlanServiceTypeUnitOfMeasureRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No job production plan service type unit of measures found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tJOB PLAN\tSERVICE TYPE UOM\tSTEP SIZE\tSTEP TARGET\tEXCLUDE INVOICES")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%t\n",
			row.ID,
			row.JobProductionPlan,
			row.ServiceTypeUnitOfMeasure,
			row.StepSize,
			row.ExplicitStepSizeTarget,
			row.ExcludeFromTimeCardInvoices,
		)
	}
	return writer.Flush()
}
