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

type doJobProductionPlanServiceTypeUnitOfMeasuresCreateOptions struct {
	BaseURL                  string
	Token                    string
	JSON                     bool
	JobProductionPlan        string
	ServiceTypeUnitOfMeasure string
	StepSize                 string
	ExplicitStepSizeTarget   string
	ExcludeFromInvoices      bool
}

func newDoJobProductionPlanServiceTypeUnitOfMeasuresCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Add a service type unit of measure to a job production plan",
		Long: `Add a service type unit of measure to a job production plan.

Required flags:
  --job-production-plan          Job production plan ID
  --service-type-unit-of-measure Service type unit of measure ID

Optional flags:
  --step-size                    Step size rule (no_step, ceiling, floor, standard_rounding_up)
  --explicit-step-size-target    Explicit step size target label
  --exclude-from-time-card-invoices Exclude from time card invoices

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Add a service type unit of measure to a job production plan
  xbe do job-production-plan-service-type-unit-of-measures create \
    --job-production-plan 123 \
    --service-type-unit-of-measure 456

  # Add with step size and explicit target
  xbe do job-production-plan-service-type-unit-of-measures create \
    --job-production-plan 123 \
    --service-type-unit-of-measure 456 \
    --step-size ceiling \
    --explicit-step-size-target "Tons"

  # Output as JSON
  xbe do job-production-plan-service-type-unit-of-measures create \
    --job-production-plan 123 \
    --service-type-unit-of-measure 456 \
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoJobProductionPlanServiceTypeUnitOfMeasuresCreate,
	}
	initDoJobProductionPlanServiceTypeUnitOfMeasuresCreateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanServiceTypeUnitOfMeasuresCmd.AddCommand(newDoJobProductionPlanServiceTypeUnitOfMeasuresCreateCmd())
}

func initDoJobProductionPlanServiceTypeUnitOfMeasuresCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID")
	cmd.Flags().String("service-type-unit-of-measure", "", "Service type unit of measure ID")
	cmd.Flags().String("step-size", "", "Step size rule (no_step, ceiling, floor, standard_rounding_up)")
	cmd.Flags().String("explicit-step-size-target", "", "Explicit step size target label")
	cmd.Flags().Bool("exclude-from-time-card-invoices", false, "Exclude from time card invoices")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("job-production-plan")
	_ = cmd.MarkFlagRequired("service-type-unit-of-measure")
}

func runDoJobProductionPlanServiceTypeUnitOfMeasuresCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoJobProductionPlanServiceTypeUnitOfMeasuresCreateOptions(cmd)
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

	if strings.TrimSpace(opts.JobProductionPlan) == "" {
		err := fmt.Errorf("--job-production-plan is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.ServiceTypeUnitOfMeasure) == "" {
		err := fmt.Errorf("--service-type-unit-of-measure is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if cmd.Flags().Changed("step-size") {
		attributes["step-size"] = opts.StepSize
	}
	if cmd.Flags().Changed("explicit-step-size-target") {
		attributes["explicit-step-size-target"] = opts.ExplicitStepSizeTarget
	}
	if cmd.Flags().Changed("exclude-from-time-card-invoices") {
		attributes["exclude-from-time-card-invoices"] = opts.ExcludeFromInvoices
	}

	relationships := map[string]any{
		"job-production-plan": map[string]any{
			"data": map[string]any{
				"type": "job-production-plans",
				"id":   opts.JobProductionPlan,
			},
		},
		"service-type-unit-of-measure": map[string]any{
			"data": map[string]any{
				"type": "service-type-unit-of-measures",
				"id":   opts.ServiceTypeUnitOfMeasure,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "job-production-plan-service-type-unit-of-measures",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/job-production-plan-service-type-unit-of-measures", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created job production plan service type unit of measure %s\n", row.ID)
	return nil
}

func parseDoJobProductionPlanServiceTypeUnitOfMeasuresCreateOptions(cmd *cobra.Command) (doJobProductionPlanServiceTypeUnitOfMeasuresCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	serviceTypeUnitOfMeasure, _ := cmd.Flags().GetString("service-type-unit-of-measure")
	stepSize, _ := cmd.Flags().GetString("step-size")
	explicitStepSizeTarget, _ := cmd.Flags().GetString("explicit-step-size-target")
	excludeFromInvoices, _ := cmd.Flags().GetBool("exclude-from-time-card-invoices")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanServiceTypeUnitOfMeasuresCreateOptions{
		BaseURL:                  baseURL,
		Token:                    token,
		JSON:                     jsonOut,
		JobProductionPlan:        jobProductionPlan,
		ServiceTypeUnitOfMeasure: serviceTypeUnitOfMeasure,
		StepSize:                 stepSize,
		ExplicitStepSizeTarget:   explicitStepSizeTarget,
		ExcludeFromInvoices:      excludeFromInvoices,
	}, nil
}
