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

type projectTransportPlanSegmentSetsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectTransportPlanSegmentSetDetails struct {
	ID                             string   `json:"id"`
	Position                       string   `json:"position,omitempty"`
	ExternalTmsLegNumber           string   `json:"external_tms_leg_number,omitempty"`
	SegmentMilesSum                any      `json:"segment_miles_sum,omitempty"`
	ProjectTransportPlanID         string   `json:"project_transport_plan_id,omitempty"`
	TruckerID                      string   `json:"trucker_id,omitempty"`
	ProjectTransportPlanSegmentIDs []string `json:"project_transport_plan_segment_ids,omitempty"`
}

func newProjectTransportPlanSegmentSetsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project transport plan segment set details",
		Long: `Show the full details of a project transport plan segment set.

Output Fields:
  ID         Segment set identifier
  POSITION   Sequence position within the plan
  PLAN       Project transport plan ID
  TRUCKER    Trucker ID (if assigned)
  EXT LEG    External TMS leg number
  SEG MI     Cached total segment miles
  SEGMENTS   Segment IDs associated with the set

Arguments:
  <id>  Project transport plan segment set ID (required).`,
		Example: `  # Show a project transport plan segment set
  xbe view project-transport-plan-segment-sets show 123

  # Output as JSON
  xbe view project-transport-plan-segment-sets show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectTransportPlanSegmentSetsShow,
	}
	initProjectTransportPlanSegmentSetsShowFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanSegmentSetsCmd.AddCommand(newProjectTransportPlanSegmentSetsShowCmd())
}

func initProjectTransportPlanSegmentSetsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanSegmentSetsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseProjectTransportPlanSegmentSetsShowOptions(cmd)
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
		return fmt.Errorf("project transport plan segment set id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-transport-plan-segment-sets]", "position,external-tms-leg-number,segment-miles-sum,project-transport-plan,trucker,project-transport-plan-segments")

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-segment-sets/"+id, query)
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

	details := buildProjectTransportPlanSegmentSetDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectTransportPlanSegmentSetDetails(cmd, details)
}

func parseProjectTransportPlanSegmentSetsShowOptions(cmd *cobra.Command) (projectTransportPlanSegmentSetsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanSegmentSetsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectTransportPlanSegmentSetDetails(resp jsonAPISingleResponse) projectTransportPlanSegmentSetDetails {
	attrs := resp.Data.Attributes

	details := projectTransportPlanSegmentSetDetails{
		ID:                   resp.Data.ID,
		Position:             stringAttr(attrs, "position"),
		ExternalTmsLegNumber: stringAttr(attrs, "external-tms-leg-number"),
		SegmentMilesSum:      anyAttr(attrs, "segment-miles-sum"),
	}

	if rel, ok := resp.Data.Relationships["project-transport-plan"]; ok && rel.Data != nil {
		details.ProjectTransportPlanID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["trucker"]; ok && rel.Data != nil {
		details.TruckerID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["project-transport-plan-segments"]; ok {
		details.ProjectTransportPlanSegmentIDs = relationshipIDList(rel)
	}

	return details
}

func renderProjectTransportPlanSegmentSetDetails(cmd *cobra.Command, details projectTransportPlanSegmentSetDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Position != "" {
		fmt.Fprintf(out, "Position: %s\n", details.Position)
	}
	if details.ExternalTmsLegNumber != "" {
		fmt.Fprintf(out, "External TMS Leg Number: %s\n", details.ExternalTmsLegNumber)
	}
	if details.SegmentMilesSum != nil {
		fmt.Fprintf(out, "Segment Miles Sum: %s\n", formatDistanceMiles(details.SegmentMilesSum))
	}
	if details.ProjectTransportPlanID != "" {
		fmt.Fprintf(out, "Project Transport Plan: %s\n", details.ProjectTransportPlanID)
	}
	if details.TruckerID != "" {
		fmt.Fprintf(out, "Trucker: %s\n", details.TruckerID)
	}
	if len(details.ProjectTransportPlanSegmentIDs) > 0 {
		fmt.Fprintf(out, "Segments: %s\n", strings.Join(details.ProjectTransportPlanSegmentIDs, ", "))
	}

	return nil
}
