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

type projectTransportPlanSegmentsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectTransportPlanSegmentDetails struct {
	ID                                    string   `json:"id"`
	Position                              *int     `json:"position,omitempty"`
	Miles                                 *float64 `json:"miles,omitempty"`
	MilesSource                           string   `json:"miles_source,omitempty"`
	ActualMinutesCached                   *int     `json:"actual_minutes_cached,omitempty"`
	ExternalTmsOrderNumber                string   `json:"external_tms_order_number,omitempty"`
	ExternalTmsMovementNumber             string   `json:"external_tms_movement_number,omitempty"`
	ProjectTransportPlanID                string   `json:"project_transport_plan_id,omitempty"`
	OriginID                              string   `json:"origin_id,omitempty"`
	DestinationID                         string   `json:"destination_id,omitempty"`
	ProjectTransportPlanSegmentSetID      string   `json:"project_transport_plan_segment_set_id,omitempty"`
	TruckerID                             string   `json:"trucker_id,omitempty"`
	TransportRouteID                      string   `json:"transport_route_id,omitempty"`
	ProjectTransportPlanSegmentDriverIDs  []string `json:"project_transport_plan_segment_driver_ids,omitempty"`
	ProjectTransportPlanSegmentTrailerIDs []string `json:"project_transport_plan_segment_trailer_ids,omitempty"`
	ProjectTransportPlanSegmentTractorIDs []string `json:"project_transport_plan_segment_tractor_ids,omitempty"`
	ExternalIdentificationIDs             []string `json:"external_identification_ids,omitempty"`
}

func newProjectTransportPlanSegmentsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project transport plan segment details",
		Long: `Show the full details of a project transport plan segment.

Output Fields:
  ID
  Position
  Miles
  Miles Source
  Actual Minutes Cached
  External TMS Order Number
  External TMS Movement Number
  Project Transport Plan ID
  Origin Stop ID
  Destination Stop ID
  Project Transport Plan Segment Set ID
  Trucker ID
  Transport Route ID
  Segment Driver IDs
  Segment Trailer IDs
  Segment Tractor IDs
  External Identification IDs

Arguments:
  <id>    Segment ID (required)

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show segment details
  xbe view project-transport-plan-segments show 123

  # JSON output
  xbe view project-transport-plan-segments show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectTransportPlanSegmentsShow,
	}
	initProjectTransportPlanSegmentsShowFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanSegmentsCmd.AddCommand(newProjectTransportPlanSegmentsShowCmd())
}

func initProjectTransportPlanSegmentsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanSegmentsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseProjectTransportPlanSegmentsShowOptions(cmd)
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
		return fmt.Errorf("project transport plan segment id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-transport-plan-segments]", strings.Join([]string{
		"position",
		"miles",
		"miles-source",
		"actual-minutes-cached",
		"external-tms-order-number",
		"external-tms-movement-number",
		"project-transport-plan",
		"origin",
		"destination",
		"project-transport-plan-segment-set",
		"trucker",
		"project-transport-plan-segment-drivers",
		"project-transport-plan-segment-trailers",
		"project-transport-plan-segment-tractors",
		"transport-route",
		"external-identifications",
	}, ","))

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-segments/"+id, query)
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

	details := buildProjectTransportPlanSegmentDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectTransportPlanSegmentDetails(cmd, details)
}

func parseProjectTransportPlanSegmentsShowOptions(cmd *cobra.Command) (projectTransportPlanSegmentsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanSegmentsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectTransportPlanSegmentDetails(resp jsonAPISingleResponse) projectTransportPlanSegmentDetails {
	attrs := resp.Data.Attributes
	details := projectTransportPlanSegmentDetails{
		ID:                        resp.Data.ID,
		MilesSource:               strings.TrimSpace(stringAttr(attrs, "miles-source")),
		ExternalTmsOrderNumber:    strings.TrimSpace(stringAttr(attrs, "external-tms-order-number")),
		ExternalTmsMovementNumber: strings.TrimSpace(stringAttr(attrs, "external-tms-movement-number")),
	}

	if position, ok := intAttrValue(attrs, "position"); ok {
		details.Position = &position
	}
	if miles, ok := floatAttrValue(attrs, "miles"); ok {
		details.Miles = &miles
	}
	if actualMinutes, ok := intAttrValue(attrs, "actual-minutes-cached"); ok {
		details.ActualMinutesCached = &actualMinutes
	}

	if rel, ok := resp.Data.Relationships["project-transport-plan"]; ok && rel.Data != nil {
		details.ProjectTransportPlanID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["origin"]; ok && rel.Data != nil {
		details.OriginID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["destination"]; ok && rel.Data != nil {
		details.DestinationID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["project-transport-plan-segment-set"]; ok && rel.Data != nil {
		details.ProjectTransportPlanSegmentSetID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["trucker"]; ok && rel.Data != nil {
		details.TruckerID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["transport-route"]; ok && rel.Data != nil {
		details.TransportRouteID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["project-transport-plan-segment-drivers"]; ok {
		details.ProjectTransportPlanSegmentDriverIDs = relationshipIDStrings(rel)
	}
	if rel, ok := resp.Data.Relationships["project-transport-plan-segment-trailers"]; ok {
		details.ProjectTransportPlanSegmentTrailerIDs = relationshipIDStrings(rel)
	}
	if rel, ok := resp.Data.Relationships["project-transport-plan-segment-tractors"]; ok {
		details.ProjectTransportPlanSegmentTractorIDs = relationshipIDStrings(rel)
	}
	if rel, ok := resp.Data.Relationships["external-identifications"]; ok {
		details.ExternalIdentificationIDs = relationshipIDStrings(rel)
	}

	return details
}

func renderProjectTransportPlanSegmentDetails(cmd *cobra.Command, details projectTransportPlanSegmentDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Position != nil {
		fmt.Fprintf(out, "Position: %d\n", *details.Position)
	}
	if details.Miles != nil {
		fmt.Fprintf(out, "Miles: %s\n", formatMiles(*details.Miles))
	}
	if details.MilesSource != "" {
		fmt.Fprintf(out, "Miles Source: %s\n", details.MilesSource)
	}
	if details.ActualMinutesCached != nil {
		fmt.Fprintf(out, "Actual Minutes Cached: %d\n", *details.ActualMinutesCached)
	}
	if details.ExternalTmsOrderNumber != "" {
		fmt.Fprintf(out, "External TMS Order Number: %s\n", details.ExternalTmsOrderNumber)
	}
	if details.ExternalTmsMovementNumber != "" {
		fmt.Fprintf(out, "External TMS Movement Number: %s\n", details.ExternalTmsMovementNumber)
	}
	if details.ProjectTransportPlanID != "" {
		fmt.Fprintf(out, "Project Transport Plan ID: %s\n", details.ProjectTransportPlanID)
	}
	if details.OriginID != "" {
		fmt.Fprintf(out, "Origin Stop ID: %s\n", details.OriginID)
	}
	if details.DestinationID != "" {
		fmt.Fprintf(out, "Destination Stop ID: %s\n", details.DestinationID)
	}
	if details.ProjectTransportPlanSegmentSetID != "" {
		fmt.Fprintf(out, "Project Transport Plan Segment Set ID: %s\n", details.ProjectTransportPlanSegmentSetID)
	}
	if details.TruckerID != "" {
		fmt.Fprintf(out, "Trucker ID: %s\n", details.TruckerID)
	}
	if details.TransportRouteID != "" {
		fmt.Fprintf(out, "Transport Route ID: %s\n", details.TransportRouteID)
	}
	if len(details.ProjectTransportPlanSegmentDriverIDs) > 0 {
		fmt.Fprintf(out, "Segment Driver IDs: %s\n", strings.Join(details.ProjectTransportPlanSegmentDriverIDs, ", "))
	}
	if len(details.ProjectTransportPlanSegmentTrailerIDs) > 0 {
		fmt.Fprintf(out, "Segment Trailer IDs: %s\n", strings.Join(details.ProjectTransportPlanSegmentTrailerIDs, ", "))
	}
	if len(details.ProjectTransportPlanSegmentTractorIDs) > 0 {
		fmt.Fprintf(out, "Segment Tractor IDs: %s\n", strings.Join(details.ProjectTransportPlanSegmentTractorIDs, ", "))
	}
	if len(details.ExternalIdentificationIDs) > 0 {
		fmt.Fprintf(out, "External Identification IDs: %s\n", strings.Join(details.ExternalIdentificationIDs, ", "))
	}

	return nil
}
