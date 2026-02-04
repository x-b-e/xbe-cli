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

type jobProductionPlanServiceTypeUnitOfMeasuresShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

func newJobProductionPlanServiceTypeUnitOfMeasuresShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show job production plan service type unit of measure details",
		Long: `Show the full details of a job production plan service type unit of measure.

Output Fields:
  ID
  Job Production Plan ID
  Service Type Unit of Measure ID
  Step Size
  Explicit Step Size Target
  Exclude From Time Card Invoices

Arguments:
  <id>    The job production plan service type unit of measure ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a job production plan service type unit of measure
  xbe view job-production-plan-service-type-unit-of-measures show 123

  # Output as JSON
  xbe view job-production-plan-service-type-unit-of-measures show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runJobProductionPlanServiceTypeUnitOfMeasuresShow,
	}
	initJobProductionPlanServiceTypeUnitOfMeasuresShowFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanServiceTypeUnitOfMeasuresCmd.AddCommand(newJobProductionPlanServiceTypeUnitOfMeasuresShowCmd())
}

func initJobProductionPlanServiceTypeUnitOfMeasuresShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanServiceTypeUnitOfMeasuresShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseJobProductionPlanServiceTypeUnitOfMeasuresShowOptions(cmd)
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
		return fmt.Errorf("job production plan service type unit of measure id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[job-production-plan-service-type-unit-of-measures]", "step-size,explicit-step-size-target,exclude-from-time-card-invoices,job-production-plan,service-type-unit-of-measure")

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-service-type-unit-of-measures/"+id, query)
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

	details := buildJobProductionPlanServiceTypeUnitOfMeasureRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderJobProductionPlanServiceTypeUnitOfMeasureDetails(cmd, details)
}

func parseJobProductionPlanServiceTypeUnitOfMeasuresShowOptions(cmd *cobra.Command) (jobProductionPlanServiceTypeUnitOfMeasuresShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanServiceTypeUnitOfMeasuresShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func renderJobProductionPlanServiceTypeUnitOfMeasureDetails(cmd *cobra.Command, details jobProductionPlanServiceTypeUnitOfMeasureRow) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.JobProductionPlan != "" {
		fmt.Fprintf(out, "Job Production Plan ID: %s\n", details.JobProductionPlan)
	}
	if details.ServiceTypeUnitOfMeasure != "" {
		fmt.Fprintf(out, "Service Type Unit of Measure ID: %s\n", details.ServiceTypeUnitOfMeasure)
	}
	if details.StepSize != "" {
		fmt.Fprintf(out, "Step Size: %s\n", details.StepSize)
	}
	if details.ExplicitStepSizeTarget != "" {
		fmt.Fprintf(out, "Explicit Step Size Target: %s\n", details.ExplicitStepSizeTarget)
	}
	fmt.Fprintf(out, "Exclude From Time Card Invoices: %t\n", details.ExcludeFromTimeCardInvoices)

	return nil
}
