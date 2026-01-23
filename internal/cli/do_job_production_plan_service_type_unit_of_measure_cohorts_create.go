package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doJobProductionPlanServiceTypeUnitOfMeasureCohortsCreateOptions struct {
	BaseURL                        string
	Token                          string
	JSON                           bool
	JobProductionPlanID            string
	ServiceTypeUnitOfMeasureCohort string
}

func newDoJobProductionPlanServiceTypeUnitOfMeasureCohortsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a job production plan service type unit of measure cohort link",
		Long: `Create a job production plan service type unit of measure cohort link.

Required flags:
  --job-production-plan                  Job production plan ID (required)
  --service-type-unit-of-measure-cohort  Service type unit of measure cohort ID (required)`,
		Example: `  # Create a cohort link for a job production plan
  xbe do job-production-plan-service-type-unit-of-measure-cohorts create \
    --job-production-plan 123 \
    --service-type-unit-of-measure-cohort 456

  # JSON output
  xbe do job-production-plan-service-type-unit-of-measure-cohorts create \
    --job-production-plan 123 \
    --service-type-unit-of-measure-cohort 456 \
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoJobProductionPlanServiceTypeUnitOfMeasureCohortsCreate,
	}
	initDoJobProductionPlanServiceTypeUnitOfMeasureCohortsCreateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanServiceTypeUnitOfMeasureCohortsCmd.AddCommand(newDoJobProductionPlanServiceTypeUnitOfMeasureCohortsCreateCmd())
}

func initDoJobProductionPlanServiceTypeUnitOfMeasureCohortsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID (required)")
	cmd.Flags().String("service-type-unit-of-measure-cohort", "", "Service type unit of measure cohort ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlanServiceTypeUnitOfMeasureCohortsCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoJobProductionPlanServiceTypeUnitOfMeasureCohortsCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
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

	if strings.TrimSpace(opts.JobProductionPlanID) == "" {
		err := fmt.Errorf("--job-production-plan is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.ServiceTypeUnitOfMeasureCohort) == "" {
		err := fmt.Errorf("--service-type-unit-of-measure-cohort is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"job-production-plan": map[string]any{
			"data": map[string]any{
				"type": "job-production-plans",
				"id":   opts.JobProductionPlanID,
			},
		},
		"service-type-unit-of-measure-cohort": map[string]any{
			"data": map[string]any{
				"type": "service-type-unit-of-measure-cohorts",
				"id":   opts.ServiceTypeUnitOfMeasureCohort,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "job-production-plan-service-type-unit-of-measure-cohorts",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/job-production-plan-service-type-unit-of-measure-cohorts", jsonBody)
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

	row := buildJobProductionPlanServiceTypeUnitOfMeasureCohortRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created job production plan service type unit of measure cohort %s\n", row.ID)
	return nil
}

func parseDoJobProductionPlanServiceTypeUnitOfMeasureCohortsCreateOptions(cmd *cobra.Command) (doJobProductionPlanServiceTypeUnitOfMeasureCohortsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobProductionPlanID, _ := cmd.Flags().GetString("job-production-plan")
	serviceTypeUnitOfMeasureCohort, _ := cmd.Flags().GetString("service-type-unit-of-measure-cohort")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanServiceTypeUnitOfMeasureCohortsCreateOptions{
		BaseURL:                        baseURL,
		Token:                          token,
		JSON:                           jsonOut,
		JobProductionPlanID:            jobProductionPlanID,
		ServiceTypeUnitOfMeasureCohort: serviceTypeUnitOfMeasureCohort,
	}, nil
}

func buildJobProductionPlanServiceTypeUnitOfMeasureCohortRowFromSingle(resp jsonAPISingleResponse) jobProductionPlanServiceTypeUnitOfMeasureCohortRow {
	row := jobProductionPlanServiceTypeUnitOfMeasureCohortRow{
		ID: resp.Data.ID,
	}

	if rel, ok := resp.Data.Relationships["job-production-plan"]; ok && rel.Data != nil {
		row.JobProductionPlanID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["service-type-unit-of-measure-cohort"]; ok && rel.Data != nil {
		row.ServiceTypeUnitOfMeasureCohortID = rel.Data.ID
	}

	return row
}
