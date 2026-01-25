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

type projectTransportPlanSegmentTractorsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectTransportPlanSegmentTractorDetails struct {
	ID                        string `json:"id"`
	ProjectTransportPlanSegID string `json:"project_transport_plan_segment_id,omitempty"`
	TractorID                 string `json:"tractor_id,omitempty"`
	ActualMilesCached         string `json:"actual_miles_cached,omitempty"`
	ActualMilesSource         string `json:"actual_miles_source,omitempty"`
	ActualMilesComputedAt     string `json:"actual_miles_computed_at,omitempty"`
}

func newProjectTransportPlanSegmentTractorsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project transport plan segment tractor details",
		Long: `Show the full details of a project transport plan segment tractor.

Output Fields:
  ID
  Project Transport Plan Segment ID
  Tractor ID
  Actual Miles Cached
  Actual Miles Source
  Actual Miles Computed At

Arguments:
  <id>    The segment tractor ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a project transport plan segment tractor
  xbe view project-transport-plan-segment-tractors show 123

  # JSON output
  xbe view project-transport-plan-segment-tractors show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectTransportPlanSegmentTractorsShow,
	}
	initProjectTransportPlanSegmentTractorsShowFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanSegmentTractorsCmd.AddCommand(newProjectTransportPlanSegmentTractorsShowCmd())
}

func initProjectTransportPlanSegmentTractorsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanSegmentTractorsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseProjectTransportPlanSegmentTractorsShowOptions(cmd)
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
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("project transport plan segment tractor id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-transport-plan-segment-tractors]", "actual-miles-cached,actual-miles-source,actual-miles-computed-at,project-transport-plan-segment,tractor")
	query.Set("include", "project-transport-plan-segment,tractor")

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-segment-tractors/"+id, query)
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

	details := buildProjectTransportPlanSegmentTractorDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectTransportPlanSegmentTractorDetails(cmd, details)
}

func parseProjectTransportPlanSegmentTractorsShowOptions(cmd *cobra.Command) (projectTransportPlanSegmentTractorsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanSegmentTractorsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectTransportPlanSegmentTractorDetails(resp jsonAPISingleResponse) projectTransportPlanSegmentTractorDetails {
	resource := resp.Data
	attrs := resource.Attributes

	return projectTransportPlanSegmentTractorDetails{
		ID:                        resource.ID,
		ProjectTransportPlanSegID: relationshipIDFromMap(resource.Relationships, "project-transport-plan-segment"),
		TractorID:                 relationshipIDFromMap(resource.Relationships, "tractor"),
		ActualMilesCached:         stringAttr(attrs, "actual-miles-cached"),
		ActualMilesSource:         stringAttr(attrs, "actual-miles-source"),
		ActualMilesComputedAt:     formatDateTime(stringAttr(attrs, "actual-miles-computed-at")),
	}
}

func renderProjectTransportPlanSegmentTractorDetails(cmd *cobra.Command, details projectTransportPlanSegmentTractorDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ProjectTransportPlanSegID != "" {
		fmt.Fprintf(out, "Project Transport Plan Segment ID: %s\n", details.ProjectTransportPlanSegID)
	}
	if details.TractorID != "" {
		fmt.Fprintf(out, "Tractor ID: %s\n", details.TractorID)
	}
	if details.ActualMilesCached != "" {
		fmt.Fprintf(out, "Actual Miles Cached: %s\n", details.ActualMilesCached)
	}
	if details.ActualMilesSource != "" {
		fmt.Fprintf(out, "Actual Miles Source: %s\n", details.ActualMilesSource)
	}
	if details.ActualMilesComputedAt != "" {
		fmt.Fprintf(out, "Actual Miles Computed At: %s\n", details.ActualMilesComputedAt)
	}

	return nil
}
