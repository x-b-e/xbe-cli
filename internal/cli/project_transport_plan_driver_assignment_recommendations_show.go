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

type projectTransportPlanDriverAssignmentRecommendationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectTransportPlanDriverAssignmentRecommendationDetails struct {
	ID                           string `json:"id"`
	ProjectTransportPlanDriverID string `json:"project_transport_plan_driver_id,omitempty"`
	Candidates                   any    `json:"candidates,omitempty"`
}

func newProjectTransportPlanDriverAssignmentRecommendationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project transport plan driver assignment recommendation details",
		Long: `Show the full details of a project transport plan driver assignment recommendation.

Output Fields:
  ID        Recommendation identifier
  Driver    Project transport plan driver ID
  Candidates Ranked candidate drivers with scores and probabilities

Arguments:
  <id>    Recommendation ID (required).`,
		Example: `  # Show a recommendation
  xbe view project-transport-plan-driver-assignment-recommendations show 123

  # JSON output
  xbe view project-transport-plan-driver-assignment-recommendations show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectTransportPlanDriverAssignmentRecommendationsShow,
	}
	initProjectTransportPlanDriverAssignmentRecommendationsShowFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanDriverAssignmentRecommendationsCmd.AddCommand(newProjectTransportPlanDriverAssignmentRecommendationsShowCmd())
}

func initProjectTransportPlanDriverAssignmentRecommendationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanDriverAssignmentRecommendationsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseProjectTransportPlanDriverAssignmentRecommendationsShowOptions(cmd)
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
		return fmt.Errorf("project transport plan driver assignment recommendation id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-driver-assignment-recommendations/"+id, nil)
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

	details := buildProjectTransportPlanDriverAssignmentRecommendationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectTransportPlanDriverAssignmentRecommendationDetails(cmd, details)
}

func parseProjectTransportPlanDriverAssignmentRecommendationsShowOptions(cmd *cobra.Command) (projectTransportPlanDriverAssignmentRecommendationsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanDriverAssignmentRecommendationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectTransportPlanDriverAssignmentRecommendationDetails(resp jsonAPISingleResponse) projectTransportPlanDriverAssignmentRecommendationDetails {
	details := projectTransportPlanDriverAssignmentRecommendationDetails{
		ID:         resp.Data.ID,
		Candidates: resp.Data.Attributes["candidates"],
	}

	if rel, ok := resp.Data.Relationships["project-transport-plan-driver"]; ok && rel.Data != nil {
		details.ProjectTransportPlanDriverID = rel.Data.ID
	}

	return details
}

func renderProjectTransportPlanDriverAssignmentRecommendationDetails(cmd *cobra.Command, details projectTransportPlanDriverAssignmentRecommendationDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ProjectTransportPlanDriverID != "" {
		fmt.Fprintf(out, "Project Transport Plan Driver ID: %s\n", details.ProjectTransportPlanDriverID)
	} else {
		fmt.Fprintln(out, "Project Transport Plan Driver ID: (none)")
	}

	fmt.Fprintln(out, "Candidates:")
	formatted := formatRecommendationCandidates(details.Candidates)
	if formatted == "" {
		fmt.Fprintln(out, "  (none)")
	} else {
		fmt.Fprintln(out, indentRecommendationLines(formatted, "  "))
	}

	return nil
}

func formatRecommendationCandidates(value any) string {
	if value == nil {
		return ""
	}
	pretty, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return string(pretty)
}

func indentRecommendationLines(value, prefix string) string {
	if value == "" {
		return ""
	}
	return prefix + strings.ReplaceAll(value, "\n", "\n"+prefix)
}
