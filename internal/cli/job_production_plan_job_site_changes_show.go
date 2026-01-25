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

type jobProductionPlanJobSiteChangesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type jobProductionPlanJobSiteChangeDetails struct {
	ID                  string `json:"id"`
	JobProductionPlanID string `json:"job_production_plan_id,omitempty"`
	OldJobSiteID        string `json:"old_job_site_id,omitempty"`
	NewJobSiteID        string `json:"new_job_site_id,omitempty"`
	CreatedByID         string `json:"created_by_id,omitempty"`
	CreatedAt           string `json:"created_at,omitempty"`
	UpdatedAt           string `json:"updated_at,omitempty"`
}

func newJobProductionPlanJobSiteChangesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show job production plan job site change details",
		Long: `Show the full details of a job production plan job site change.

Output Fields:
  ID
  Job Production Plan ID
  Old Job Site ID
  New Job Site ID
  Created By (user ID)
  Created At
  Updated At

Arguments:
  <id>    The job site change ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a job site change
  xbe view job-production-plan-job-site-changes show 123

  # Get JSON output
  xbe view job-production-plan-job-site-changes show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runJobProductionPlanJobSiteChangesShow,
	}
	initJobProductionPlanJobSiteChangesShowFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanJobSiteChangesCmd.AddCommand(newJobProductionPlanJobSiteChangesShowCmd())
}

func initJobProductionPlanJobSiteChangesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanJobSiteChangesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseJobProductionPlanJobSiteChangesShowOptions(cmd)
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
		return fmt.Errorf("job production plan job site change id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-job-site-changes/"+id, nil)
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

	details := buildJobProductionPlanJobSiteChangeDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderJobProductionPlanJobSiteChangeDetails(cmd, details)
}

func parseJobProductionPlanJobSiteChangesShowOptions(cmd *cobra.Command) (jobProductionPlanJobSiteChangesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanJobSiteChangesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildJobProductionPlanJobSiteChangeDetails(resp jsonAPISingleResponse) jobProductionPlanJobSiteChangeDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := jobProductionPlanJobSiteChangeDetails{
		ID:        resource.ID,
		CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt: formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		details.JobProductionPlanID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["old-job-site"]; ok && rel.Data != nil {
		details.OldJobSiteID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["new-job-site"]; ok && rel.Data != nil {
		details.NewJobSiteID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
	}

	return details
}

func renderJobProductionPlanJobSiteChangeDetails(cmd *cobra.Command, details jobProductionPlanJobSiteChangeDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.JobProductionPlanID != "" {
		fmt.Fprintf(out, "Job Production Plan ID: %s\n", details.JobProductionPlanID)
	}
	if details.OldJobSiteID != "" {
		fmt.Fprintf(out, "Old Job Site ID: %s\n", details.OldJobSiteID)
	}
	if details.NewJobSiteID != "" {
		fmt.Fprintf(out, "New Job Site ID: %s\n", details.NewJobSiteID)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By: %s\n", details.CreatedByID)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
