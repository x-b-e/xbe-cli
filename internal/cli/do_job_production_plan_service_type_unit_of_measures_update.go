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

type doJobProductionPlanServiceTypeUnitOfMeasuresUpdateOptions struct {
	BaseURL                  string
	Token                    string
	JSON                     bool
	ID                       string
	JobProductionPlan        string
	ServiceTypeUnitOfMeasure string
	StepSize                 string
	ExplicitStepSizeTarget   string
	ExcludeFromInvoices      bool
}

func newDoJobProductionPlanServiceTypeUnitOfMeasuresUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a job production plan service type unit of measure",
		Long: `Update a job production plan service type unit of measure.

Optional flags:
  --job-production-plan          Job production plan ID
  --service-type-unit-of-measure Service type unit of measure ID
  --step-size                    Step size rule (no_step, ceiling, floor, standard_rounding_up)
  --explicit-step-size-target    Explicit step size target label
  --exclude-from-time-card-invoices Exclude from time card invoices (true/false)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update step size
  xbe do job-production-plan-service-type-unit-of-measures update 123 --step-size floor

  # Update explicit step size target
  xbe do job-production-plan-service-type-unit-of-measures update 123 --explicit-step-size-target "Loads"

  # Update relationships
  xbe do job-production-plan-service-type-unit-of-measures update 123 \
    --job-production-plan 456 \
    --service-type-unit-of-measure 789`,
		Args: cobra.ExactArgs(1),
		RunE: runDoJobProductionPlanServiceTypeUnitOfMeasuresUpdate,
	}
	initDoJobProductionPlanServiceTypeUnitOfMeasuresUpdateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanServiceTypeUnitOfMeasuresCmd.AddCommand(newDoJobProductionPlanServiceTypeUnitOfMeasuresUpdateCmd())
}

func initDoJobProductionPlanServiceTypeUnitOfMeasuresUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID")
	cmd.Flags().String("service-type-unit-of-measure", "", "Service type unit of measure ID")
	cmd.Flags().String("step-size", "", "Step size rule (no_step, ceiling, floor, standard_rounding_up)")
	cmd.Flags().String("explicit-step-size-target", "", "Explicit step size target label")
	cmd.Flags().Bool("exclude-from-time-card-invoices", false, "Exclude from time card invoices")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlanServiceTypeUnitOfMeasuresUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoJobProductionPlanServiceTypeUnitOfMeasuresUpdateOptions(cmd, args)
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

	attributes := map[string]any{}
	relationships := map[string]any{}

	if cmd.Flags().Changed("step-size") {
		attributes["step-size"] = opts.StepSize
	}
	if cmd.Flags().Changed("explicit-step-size-target") {
		attributes["explicit-step-size-target"] = opts.ExplicitStepSizeTarget
	}
	if cmd.Flags().Changed("exclude-from-time-card-invoices") {
		attributes["exclude-from-time-card-invoices"] = opts.ExcludeFromInvoices
	}
	if cmd.Flags().Changed("job-production-plan") {
		relationships["job-production-plan"] = map[string]any{
			"data": map[string]any{
				"type": "job-production-plans",
				"id":   opts.JobProductionPlan,
			},
		}
	}
	if cmd.Flags().Changed("service-type-unit-of-measure") {
		relationships["service-type-unit-of-measure"] = map[string]any{
			"data": map[string]any{
				"type": "service-type-unit-of-measures",
				"id":   opts.ServiceTypeUnitOfMeasure,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes or relationships to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "job-production-plan-service-type-unit-of-measures",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/job-production-plan-service-type-unit-of-measures/"+opts.ID, jsonBody)
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

	row := buildJobProductionPlanServiceTypeUnitOfMeasureRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated job production plan service type unit of measure %s\n", row.ID)
	return nil
}

func parseDoJobProductionPlanServiceTypeUnitOfMeasuresUpdateOptions(cmd *cobra.Command, args []string) (doJobProductionPlanServiceTypeUnitOfMeasuresUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	serviceTypeUnitOfMeasure, _ := cmd.Flags().GetString("service-type-unit-of-measure")
	stepSize, _ := cmd.Flags().GetString("step-size")
	explicitStepSizeTarget, _ := cmd.Flags().GetString("explicit-step-size-target")
	excludeFromInvoices, _ := cmd.Flags().GetBool("exclude-from-time-card-invoices")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanServiceTypeUnitOfMeasuresUpdateOptions{
		BaseURL:                  baseURL,
		Token:                    token,
		JSON:                     jsonOut,
		ID:                       args[0],
		JobProductionPlan:        jobProductionPlan,
		ServiceTypeUnitOfMeasure: serviceTypeUnitOfMeasure,
		StepSize:                 stepSize,
		ExplicitStepSizeTarget:   explicitStepSizeTarget,
		ExcludeFromInvoices:      excludeFromInvoices,
	}, nil
}
