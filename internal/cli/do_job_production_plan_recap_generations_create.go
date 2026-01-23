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

type doJobProductionPlanRecapGenerationsCreateOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	JobProductionPlan string
}

type jobProductionPlanRecapGenerationRow struct {
	ID                  string `json:"id"`
	JobProductionPlanID string `json:"job_production_plan_id"`
}

func newDoJobProductionPlanRecapGenerationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Generate a job production plan recap",
		Long: `Generate a recap for a job production plan.

Recaps can only be generated when the job production plan is approved or complete.

Required flags:
  --job-production-plan   Job production plan ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Generate a recap for a job production plan
  xbe do job-production-plan-recap-generations create --job-production-plan 123

  # JSON output
  xbe do job-production-plan-recap-generations create --job-production-plan 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoJobProductionPlanRecapGenerationsCreate,
	}
	initDoJobProductionPlanRecapGenerationsCreateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanRecapGenerationsCmd.AddCommand(newDoJobProductionPlanRecapGenerationsCreateCmd())
}

func initDoJobProductionPlanRecapGenerationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlanRecapGenerationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoJobProductionPlanRecapGenerationsCreateOptions(cmd)
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

	relationships := map[string]any{
		"job-production-plan": map[string]any{
			"data": map[string]any{
				"type": "job-production-plans",
				"id":   opts.JobProductionPlan,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "job-production-plan-recap-generations",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/job-production-plan-recap-generations", jsonBody)
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

	row := buildJobProductionPlanRecapGenerationRow(resp.Data)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Queued job production plan recap generation %s\n", row.ID)
	return nil
}

func parseDoJobProductionPlanRecapGenerationsCreateOptions(cmd *cobra.Command) (doJobProductionPlanRecapGenerationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanRecapGenerationsCreateOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		JobProductionPlan: jobProductionPlan,
	}, nil
}

func buildJobProductionPlanRecapGenerationRow(resource jsonAPIResource) jobProductionPlanRecapGenerationRow {
	row := jobProductionPlanRecapGenerationRow{ID: resource.ID}
	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		row.JobProductionPlanID = rel.Data.ID
	}
	return row
}
