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

type projectPhaseDatesEstimatesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectPhaseDatesEstimateDetails struct {
	ID                   string `json:"id"`
	ProjectPhaseID       string `json:"project_phase_id,omitempty"`
	ProjectEstimateSetID string `json:"project_estimate_set_id,omitempty"`
	CreatedByID          string `json:"created_by_id,omitempty"`
	StartDate            string `json:"start_date,omitempty"`
	EndDate              string `json:"end_date,omitempty"`
}

func newProjectPhaseDatesEstimatesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project phase dates estimate details",
		Long: `Show the full details of a project phase dates estimate.

Output Fields:
  ID
  Project Phase
  Project Estimate Set
  Created By
  Start Date
  End Date

Arguments:
  <id>  Project phase dates estimate ID (required). Use the list command to find IDs.`,
		Example: `  # Show a date estimate
  xbe view project-phase-dates-estimates show 123

  # Output as JSON
  xbe view project-phase-dates-estimates show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectPhaseDatesEstimatesShow,
	}
	initProjectPhaseDatesEstimatesShowFlags(cmd)
	return cmd
}

func init() {
	projectPhaseDatesEstimatesCmd.AddCommand(newProjectPhaseDatesEstimatesShowCmd())
}

func initProjectPhaseDatesEstimatesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectPhaseDatesEstimatesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseProjectPhaseDatesEstimatesShowOptions(cmd)
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
		return fmt.Errorf("project phase dates estimate id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-phase-dates-estimates]", "start-date,end-date,project-phase,project-estimate-set,created-by")

	body, _, err := client.Get(cmd.Context(), "/v1/project-phase-dates-estimates/"+id, query)
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

	details := buildProjectPhaseDatesEstimateDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectPhaseDatesEstimateDetails(cmd, details)
}

func parseProjectPhaseDatesEstimatesShowOptions(cmd *cobra.Command) (projectPhaseDatesEstimatesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectPhaseDatesEstimatesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectPhaseDatesEstimateDetails(resp jsonAPISingleResponse) projectPhaseDatesEstimateDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := projectPhaseDatesEstimateDetails{
		ID:        resource.ID,
		StartDate: stringAttr(attrs, "start-date"),
		EndDate:   stringAttr(attrs, "end-date"),
	}

	if rel, ok := resource.Relationships["project-phase"]; ok && rel.Data != nil {
		details.ProjectPhaseID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["project-estimate-set"]; ok && rel.Data != nil {
		details.ProjectEstimateSetID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
	}

	return details
}

func renderProjectPhaseDatesEstimateDetails(cmd *cobra.Command, details projectPhaseDatesEstimateDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ProjectPhaseID != "" {
		fmt.Fprintf(out, "Project Phase: %s\n", details.ProjectPhaseID)
	}
	if details.ProjectEstimateSetID != "" {
		fmt.Fprintf(out, "Project Estimate Set: %s\n", details.ProjectEstimateSetID)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By: %s\n", details.CreatedByID)
	}
	if details.StartDate != "" {
		fmt.Fprintf(out, "Start Date: %s\n", details.StartDate)
	}
	if details.EndDate != "" {
		fmt.Fprintf(out, "End Date: %s\n", details.EndDate)
	}

	return nil
}
