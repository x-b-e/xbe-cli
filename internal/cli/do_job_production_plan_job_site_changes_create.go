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

type doJobProductionPlanJobSiteChangesCreateOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	JobProductionPlan string
	OldJobSite        string
	NewJobSite        string
}

type jobProductionPlanJobSiteChangeRow struct {
	ID                  string `json:"id"`
	JobProductionPlanID string `json:"job_production_plan_id,omitempty"`
	OldJobSiteID        string `json:"old_job_site_id,omitempty"`
	NewJobSiteID        string `json:"new_job_site_id,omitempty"`
	CreatedByID         string `json:"created_by_id,omitempty"`
	CreatedAt           string `json:"created_at,omitempty"`
}

func newDoJobProductionPlanJobSiteChangesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a job production plan job site change",
		Long: `Create a job production plan job site change.

Job site changes update the job production plan's job site and associated jobs.

Required flags:
  --job-production-plan   Job production plan ID (required)
  --old-job-site          Current job site ID on the plan (required)
  --new-job-site          New job site ID (required)

Notes:
  - Old and new job sites must be different.
  - Job sites must belong to the plan's customer.

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Change a job production plan job site
  xbe do job-production-plan-job-site-changes create \\
    --job-production-plan 123 \\
    --old-job-site 456 \\
    --new-job-site 789`,
		Args: cobra.NoArgs,
		RunE: runDoJobProductionPlanJobSiteChangesCreate,
	}
	initDoJobProductionPlanJobSiteChangesCreateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanJobSiteChangesCmd.AddCommand(newDoJobProductionPlanJobSiteChangesCreateCmd())
}

func initDoJobProductionPlanJobSiteChangesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID (required)")
	cmd.Flags().String("old-job-site", "", "Current job site ID on the plan (required)")
	cmd.Flags().String("new-job-site", "", "New job site ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlanJobSiteChangesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoJobProductionPlanJobSiteChangesCreateOptions(cmd)
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

	jobProductionPlanID := strings.TrimSpace(opts.JobProductionPlan)
	oldJobSiteID := strings.TrimSpace(opts.OldJobSite)
	newJobSiteID := strings.TrimSpace(opts.NewJobSite)

	if jobProductionPlanID == "" {
		err := fmt.Errorf("--job-production-plan is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if oldJobSiteID == "" {
		err := fmt.Errorf("--old-job-site is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if newJobSiteID == "" {
		err := fmt.Errorf("--new-job-site is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"job-production-plan": map[string]any{
			"data": map[string]any{
				"type": "job-production-plans",
				"id":   jobProductionPlanID,
			},
		},
		"old-job-site": map[string]any{
			"data": map[string]any{
				"type": "job-sites",
				"id":   oldJobSiteID,
			},
		},
		"new-job-site": map[string]any{
			"data": map[string]any{
				"type": "job-sites",
				"id":   newJobSiteID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "job-production-plan-job-site-changes",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/job-production-plan-job-site-changes", jsonBody)
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

	row := buildJobProductionPlanJobSiteChangeRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created job production plan job site change %s\n", row.ID)
	return nil
}

func parseDoJobProductionPlanJobSiteChangesCreateOptions(cmd *cobra.Command) (doJobProductionPlanJobSiteChangesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	oldJobSite, _ := cmd.Flags().GetString("old-job-site")
	newJobSite, _ := cmd.Flags().GetString("new-job-site")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanJobSiteChangesCreateOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		JobProductionPlan: jobProductionPlan,
		OldJobSite:        oldJobSite,
		NewJobSite:        newJobSite,
	}, nil
}

func buildJobProductionPlanJobSiteChangeRowFromSingle(resp jsonAPISingleResponse) jobProductionPlanJobSiteChangeRow {
	resource := resp.Data
	attrs := resource.Attributes
	row := jobProductionPlanJobSiteChangeRow{
		ID:        resource.ID,
		CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
	}

	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		row.JobProductionPlanID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["old-job-site"]; ok && rel.Data != nil {
		row.OldJobSiteID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["new-job-site"]; ok && rel.Data != nil {
		row.NewJobSiteID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		row.CreatedByID = rel.Data.ID
	}

	return row
}
