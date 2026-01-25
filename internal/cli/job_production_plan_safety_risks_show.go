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

type jobProductionPlanSafetyRisksShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type jobProductionPlanSafetyRiskDetails struct {
	ID                  string `json:"id"`
	JobProductionPlanID string `json:"job_production_plan_id,omitempty"`
	JobProductionPlan   string `json:"job_production_plan,omitempty"`
	Description         string `json:"description,omitempty"`
}

func newJobProductionPlanSafetyRisksShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show job production plan safety risk details",
		Long: `Show the full details of a job production plan safety risk.

Output Fields:
  ID                   Resource identifier
  Job Production Plan  Job production plan (job number or name)
  Description          Safety risk description

Arguments:
  <id>          The job production plan safety risk ID (required).`,
		Example: `  # Show details
  xbe view job-production-plan-safety-risks show 123

  # Output as JSON
  xbe view job-production-plan-safety-risks show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runJobProductionPlanSafetyRisksShow,
	}
	initJobProductionPlanSafetyRisksShowFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanSafetyRisksCmd.AddCommand(newJobProductionPlanSafetyRisksShowCmd())
}

func initJobProductionPlanSafetyRisksShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanSafetyRisksShow(cmd *cobra.Command, args []string) error {
	opts, err := parseJobProductionPlanSafetyRisksShowOptions(cmd)
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
		return fmt.Errorf("job production plan safety risk id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[job-production-plan-safety-risks]", "description,job-production-plan")
	query.Set("include", "job-production-plan")
	query.Set("fields[job-production-plans]", "job-number,job-name")

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-safety-risks/"+id, query)
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

	details := buildJobProductionPlanSafetyRiskDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderJobProductionPlanSafetyRiskDetails(cmd, details)
}

func parseJobProductionPlanSafetyRisksShowOptions(cmd *cobra.Command) (jobProductionPlanSafetyRisksShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return jobProductionPlanSafetyRisksShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return jobProductionPlanSafetyRisksShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return jobProductionPlanSafetyRisksShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return jobProductionPlanSafetyRisksShowOptions{}, err
	}

	return jobProductionPlanSafetyRisksShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildJobProductionPlanSafetyRiskDetails(resp jsonAPISingleResponse) jobProductionPlanSafetyRiskDetails {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	attrs := resp.Data.Attributes
	details := jobProductionPlanSafetyRiskDetails{
		ID:          resp.Data.ID,
		Description: stringAttr(attrs, "description"),
	}

	if rel, ok := resp.Data.Relationships["job-production-plan"]; ok && rel.Data != nil {
		details.JobProductionPlanID = rel.Data.ID
		if plan, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.JobProductionPlan = firstNonEmpty(
				stringAttr(plan.Attributes, "job-number"),
				stringAttr(plan.Attributes, "job-name"),
			)
		}
	}

	return details
}

func renderJobProductionPlanSafetyRiskDetails(cmd *cobra.Command, details jobProductionPlanSafetyRiskDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.JobProductionPlanID != "" || details.JobProductionPlan != "" {
		fmt.Fprintf(out, "Job Production Plan: %s\n", formatRelated(details.JobProductionPlan, details.JobProductionPlanID))
	}
	if details.Description != "" {
		fmt.Fprintf(out, "Description: %s\n", details.Description)
	}

	return nil
}
