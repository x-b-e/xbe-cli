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

type jobProductionPlansShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type jobProductionPlanDetails struct {
	ID                   string   `json:"id"`
	Status               string   `json:"status"`
	JobNumber            string   `json:"job_number"`
	JobName              string   `json:"job_name"`
	StartOn              string   `json:"start_on"`
	StartTime            string   `json:"start_time"`
	EndTime              string   `json:"end_time,omitempty"`
	Customer             string   `json:"customer,omitempty"`
	Planner              string   `json:"planner,omitempty"`
	ProjectManager       string   `json:"project_manager,omitempty"`
	JobSite              string   `json:"job_site,omitempty"`
	MaterialSites        []string `json:"material_sites,omitempty"`
	GoalTons             float64  `json:"goal_tons"`
	GoalHours            float64  `json:"goal_hours,omitempty"`
	RemainingQuantity    float64  `json:"remaining_quantity,omitempty"`
	DispatchInstructions string   `json:"dispatch_instructions,omitempty"`
	Notes                string   `json:"notes,omitempty"`
	IsOnHold             bool     `json:"is_on_hold"`
	OnHoldComment        string   `json:"on_hold_comment,omitempty"`
	IsScheduleLocked     bool     `json:"is_schedule_locked"`
}

func newJobProductionPlansShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show job production plan details",
		Long: `Show the full details of a specific job production plan.

Retrieves and displays comprehensive information about a plan including
scheduling, assignments, and production targets.

Output Fields:
  ID                    Plan identifier
  Status                Current status
  Job Number            Job number
  Job Name              Job name
  Start Date/Time       When the plan starts
  Customer              Customer name
  Planner               Assigned planner
  Project Manager       Assigned project manager
  Job Site              Job site location
  Material Sites        Material source sites
  Goal (Tons/Hours)     Production targets
  Dispatch Instructions Instructions for drivers
  Notes                 Plan notes

Arguments:
  <id>    The job production plan ID (required)`,
		Example: `  # View a plan by ID
  xbe view job-production-plans show 12345

  # Get plan as JSON
  xbe view job-production-plans show 12345 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runJobProductionPlansShow,
	}
	initJobProductionPlansShowFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlansCmd.AddCommand(newJobProductionPlansShowCmd())
}

func initJobProductionPlansShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlansShow(cmd *cobra.Command, args []string) error {
	opts, err := parseJobProductionPlansShowOptions(cmd)
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("job production plan id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[job-production-plans]", "job-number,job-name,status,start-on,start-time,end-time,goal-quantity,goal-hours,remaining-quantity,dispatch-instructions,notes,is-on-hold,on-hold-comment,is-schedule-locked,customer,planner,project-manager,job-site")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[users]", "name")
	query.Set("fields[job-sites]", "name")
	query.Set("fields[material-sites]", "name")
	query.Set("include", "customer,planner,project-manager,job-site,job-production-plan-material-sites.material-site")

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plans/"+id, query)
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

	details := buildJobProductionPlanDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderJobProductionPlanDetails(cmd, details)
}

func parseJobProductionPlansShowOptions(cmd *cobra.Command) (jobProductionPlansShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return jobProductionPlansShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return jobProductionPlansShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return jobProductionPlansShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return jobProductionPlansShowOptions{}, err
	}

	return jobProductionPlansShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildJobProductionPlanDetails(resp jsonAPISingleResponse) jobProductionPlanDetails {
	attrs := resp.Data.Attributes

	// Build included lookup
	included := make(map[string]map[string]any)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc.Attributes
	}

	details := jobProductionPlanDetails{
		ID:                   resp.Data.ID,
		Status:               stringAttr(attrs, "status"),
		JobNumber:            stringAttr(attrs, "job-number"),
		JobName:              stringAttr(attrs, "job-name"),
		StartOn:              formatDate(stringAttr(attrs, "start-on")),
		StartTime:            formatTime(stringAttr(attrs, "start-time")),
		EndTime:              formatTime(stringAttr(attrs, "end-time")),
		GoalTons:             floatAttr(attrs, "goal-quantity"),
		GoalHours:            floatAttr(attrs, "goal-hours"),
		RemainingQuantity:    floatAttr(attrs, "remaining-quantity"),
		DispatchInstructions: stringAttr(attrs, "dispatch-instructions"),
		Notes:                stringAttr(attrs, "notes"),
		IsOnHold:             boolAttr(attrs, "is-on-hold"),
		OnHoldComment:        stringAttr(attrs, "on-hold-comment"),
		IsScheduleLocked:     boolAttr(attrs, "is-schedule-locked"),
	}

	// Resolve customer
	if rel, ok := resp.Data.Relationships["customer"]; ok && rel.Data != nil {
		key := resourceKey(rel.Data.Type, rel.Data.ID)
		if attrs, ok := included[key]; ok {
			details.Customer = stringAttr(attrs, "company-name")
		}
	}

	// Resolve planner
	if rel, ok := resp.Data.Relationships["planner"]; ok && rel.Data != nil {
		key := resourceKey(rel.Data.Type, rel.Data.ID)
		if attrs, ok := included[key]; ok {
			details.Planner = stringAttr(attrs, "name")
		}
	}

	// Resolve project manager
	if rel, ok := resp.Data.Relationships["project-manager"]; ok && rel.Data != nil {
		key := resourceKey(rel.Data.Type, rel.Data.ID)
		if attrs, ok := included[key]; ok {
			details.ProjectManager = stringAttr(attrs, "name")
		}
	}

	// Resolve job site
	if rel, ok := resp.Data.Relationships["job-site"]; ok && rel.Data != nil {
		key := resourceKey(rel.Data.Type, rel.Data.ID)
		if attrs, ok := included[key]; ok {
			details.JobSite = stringAttr(attrs, "name")
		}
	}

	// Resolve material sites
	details.MaterialSites = resolveShowMaterialSites(resp.Data, included)

	return details
}

func resolveShowMaterialSites(resource jsonAPIResource, included map[string]map[string]any) []string {
	var sites []string
	seen := make(map[string]bool)

	// Look for material-sites in included
	for key, attrs := range included {
		if strings.HasPrefix(key, "material-sites|") {
			siteName := stringAttr(attrs, "name")
			if siteName != "" && !seen[siteName] {
				seen[siteName] = true
				sites = append(sites, siteName)
			}
		}
	}

	return sites
}

func renderJobProductionPlanDetails(cmd *cobra.Command, details jobProductionPlanDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	fmt.Fprintf(out, "Status: %s\n", details.Status)

	if details.JobNumber != "" {
		fmt.Fprintf(out, "Job Number: %s\n", details.JobNumber)
	}
	if details.JobName != "" {
		fmt.Fprintf(out, "Job Name: %s\n", details.JobName)
	}

	fmt.Fprintf(out, "Start: %s %s\n", details.StartOn, details.StartTime)
	if details.EndTime != "" {
		fmt.Fprintf(out, "End Time: %s\n", details.EndTime)
	}

	if details.Customer != "" {
		fmt.Fprintf(out, "Customer: %s\n", details.Customer)
	}
	if details.Planner != "" {
		fmt.Fprintf(out, "Planner: %s\n", details.Planner)
	}
	if details.ProjectManager != "" {
		fmt.Fprintf(out, "Project Manager: %s\n", details.ProjectManager)
	}
	if details.JobSite != "" {
		fmt.Fprintf(out, "Job Site: %s\n", details.JobSite)
	}
	if len(details.MaterialSites) > 0 {
		fmt.Fprintf(out, "Material Sites: %s\n", strings.Join(details.MaterialSites, ", "))
	}

	if details.GoalTons > 0 {
		fmt.Fprintf(out, "Goal (Tons): %.0f\n", details.GoalTons)
	}
	if details.GoalHours > 0 {
		fmt.Fprintf(out, "Goal (Hours): %.1f\n", details.GoalHours)
	}
	if details.RemainingQuantity > 0 {
		fmt.Fprintf(out, "Remaining: %.0f\n", details.RemainingQuantity)
	}

	if details.IsOnHold {
		fmt.Fprintln(out, "On Hold: Yes")
		if details.OnHoldComment != "" {
			fmt.Fprintf(out, "Hold Comment: %s\n", details.OnHoldComment)
		}
	}
	if details.IsScheduleLocked {
		fmt.Fprintln(out, "Schedule Locked: Yes")
	}

	if details.DispatchInstructions != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Dispatch Instructions:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, details.DispatchInstructions)
	}

	if details.Notes != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Notes:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, details.Notes)
	}

	return nil
}
