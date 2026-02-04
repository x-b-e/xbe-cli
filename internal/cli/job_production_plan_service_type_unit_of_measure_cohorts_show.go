package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type jobProductionPlanServiceTypeUnitOfMeasureCohortsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type jobProductionPlanServiceTypeUnitOfMeasureCohortDetails struct {
	ID                               string   `json:"id"`
	JobProductionPlanID              string   `json:"job_production_plan_id,omitempty"`
	JobProductionPlan                string   `json:"job_production_plan,omitempty"`
	ServiceTypeUnitOfMeasureCohortID string   `json:"service_type_unit_of_measure_cohort_id,omitempty"`
	ServiceTypeUnitOfMeasureCohort   string   `json:"service_type_unit_of_measure_cohort,omitempty"`
	ServiceTypeUnitOfMeasureActive   bool     `json:"service_type_unit_of_measure_cohort_active,omitempty"`
	ServiceTypeUnitOfMeasureIDs      []string `json:"service_type_unit_of_measure_ids,omitempty"`
}

func newJobProductionPlanServiceTypeUnitOfMeasureCohortsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show job production plan service type unit of measure cohort details",
		Long: `Show the full details of a job production plan service type unit of measure cohort link.

Output Fields:
  ID                         Cohort link identifier
  Job Production Plan        Job production plan
  Service Type Unit Of Measure Cohort  Cohort name
  Cohort Active              Cohort active status
  Service Type Unit Of Measure IDs     Service type unit of measure IDs

Arguments:
  <id>    Job production plan service type unit of measure cohort ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a job production plan service type unit of measure cohort link
  xbe view job-production-plan-service-type-unit-of-measure-cohorts show 123

  # JSON output
  xbe view job-production-plan-service-type-unit-of-measure-cohorts show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runJobProductionPlanServiceTypeUnitOfMeasureCohortsShow,
	}
	initJobProductionPlanServiceTypeUnitOfMeasureCohortsShowFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanServiceTypeUnitOfMeasureCohortsCmd.AddCommand(newJobProductionPlanServiceTypeUnitOfMeasureCohortsShowCmd())
}

func initJobProductionPlanServiceTypeUnitOfMeasureCohortsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanServiceTypeUnitOfMeasureCohortsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseJobProductionPlanServiceTypeUnitOfMeasureCohortsShowOptions(cmd)
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("job production plan service type unit of measure cohort id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[job-production-plan-service-type-unit-of-measure-cohorts]", "job-production-plan,service-type-unit-of-measure-cohort")
	query.Set("fields[job-production-plans]", "job-number,job-name")
	query.Set("fields[service-type-unit-of-measure-cohorts]", "name,is-active,service-type-unit-of-measure-ids")
	query.Set("include", "job-production-plan,service-type-unit-of-measure-cohort")

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-service-type-unit-of-measure-cohorts/"+id, query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPISingleResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildJobProductionPlanServiceTypeUnitOfMeasureCohortDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderJobProductionPlanServiceTypeUnitOfMeasureCohortDetails(cmd, details)
}

func parseJobProductionPlanServiceTypeUnitOfMeasureCohortsShowOptions(cmd *cobra.Command) (jobProductionPlanServiceTypeUnitOfMeasureCohortsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanServiceTypeUnitOfMeasureCohortsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildJobProductionPlanServiceTypeUnitOfMeasureCohortDetails(resp jsonAPISingleResponse) jobProductionPlanServiceTypeUnitOfMeasureCohortDetails {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := jobProductionPlanServiceTypeUnitOfMeasureCohortDetails{
		ID: resp.Data.ID,
	}

	if rel, ok := resp.Data.Relationships["job-production-plan"]; ok && rel.Data != nil {
		details.JobProductionPlanID = rel.Data.ID
		if plan, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			jobNumber := stringAttr(plan.Attributes, "job-number")
			jobName := stringAttr(plan.Attributes, "job-name")
			if jobNumber != "" && jobName != "" {
				details.JobProductionPlan = fmt.Sprintf("%s - %s", jobNumber, jobName)
			} else {
				details.JobProductionPlan = firstNonEmpty(jobNumber, jobName)
			}
		}
	}

	if rel, ok := resp.Data.Relationships["service-type-unit-of-measure-cohort"]; ok && rel.Data != nil {
		details.ServiceTypeUnitOfMeasureCohortID = rel.Data.ID
		if cohort, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.ServiceTypeUnitOfMeasureCohort = stringAttr(cohort.Attributes, "name")
			details.ServiceTypeUnitOfMeasureActive = boolAttr(cohort.Attributes, "is-active")
			details.ServiceTypeUnitOfMeasureIDs = stringSliceAttr(cohort.Attributes, "service-type-unit-of-measure-ids")
		}
	}

	return details
}

func renderJobProductionPlanServiceTypeUnitOfMeasureCohortDetails(cmd *cobra.Command, details jobProductionPlanServiceTypeUnitOfMeasureCohortDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)

	writeLabelWithID(out, "Job Production Plan", details.JobProductionPlan, details.JobProductionPlanID)
	writeLabelWithID(out, "Service Type Unit Of Measure Cohort", details.ServiceTypeUnitOfMeasureCohort, details.ServiceTypeUnitOfMeasureCohortID)
	fmt.Fprintf(out, "Cohort Active: %t\n", details.ServiceTypeUnitOfMeasureActive)

	if len(details.ServiceTypeUnitOfMeasureIDs) > 0 {
		fmt.Fprintf(out, "Service Type Unit Of Measure IDs: %s\n", strings.Join(details.ServiceTypeUnitOfMeasureIDs, ", "))
	}

	return nil
}
