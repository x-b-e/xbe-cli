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

type doProjectTransportPlanSegmentsCreateOptions struct {
	BaseURL                   string
	Token                     string
	JSON                      bool
	ProjectTransportPlan      string
	Origin                    string
	Destination               string
	ProjectTransportPlanSet   string
	Trucker                   string
	Position                  int
	Miles                     float64
	MilesSource               string
	ExternalTmsOrderNumber    string
	ExternalTmsMovementNumber string
}

func newDoProjectTransportPlanSegmentsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project transport plan segment",
		Long: `Create a project transport plan segment.

Required:
  --project-transport-plan  Project transport plan ID
  --origin                  Origin stop ID
  --destination             Destination stop ID

Optional:
  --position                        Segment position within the plan
  --miles                           Segment miles
  --miles-source                    Miles source (unknown, transport_route)
  --project-transport-plan-segment-set  Segment set ID
  --trucker                         Trucker ID
  --external-tms-order-number       External TMS order number (transport-only projects)
  --external-tms-movement-number    External TMS movement number (transport-only projects)`,
		Example: `  # Create a segment with required relationships
  xbe do project-transport-plan-segments create \
    --project-transport-plan 123 \
    --origin 456 \
    --destination 789

  # Create a segment with position and miles
  xbe do project-transport-plan-segments create \
    --project-transport-plan 123 \
    --origin 456 \
    --destination 789 \
    --position 2 \
    --miles 12.5 \
    --miles-source transport_route`,
		Args: cobra.NoArgs,
		RunE: runDoProjectTransportPlanSegmentsCreate,
	}
	initDoProjectTransportPlanSegmentsCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectTransportPlanSegmentsCmd.AddCommand(newDoProjectTransportPlanSegmentsCreateCmd())
}

func initDoProjectTransportPlanSegmentsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project-transport-plan", "", "Project transport plan ID (required)")
	cmd.Flags().String("origin", "", "Origin stop ID (required)")
	cmd.Flags().String("destination", "", "Destination stop ID (required)")
	cmd.Flags().String("project-transport-plan-segment-set", "", "Segment set ID")
	cmd.Flags().String("trucker", "", "Trucker ID")
	cmd.Flags().Int("position", 0, "Segment position within the plan")
	cmd.Flags().Float64("miles", 0, "Segment miles")
	cmd.Flags().String("miles-source", "", "Miles source (unknown, transport_route)")
	cmd.Flags().String("external-tms-order-number", "", "External TMS order number (transport-only projects)")
	cmd.Flags().String("external-tms-movement-number", "", "External TMS movement number (transport-only projects)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectTransportPlanSegmentsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectTransportPlanSegmentsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.ProjectTransportPlan) == "" {
		err := fmt.Errorf("--project-transport-plan is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.Origin) == "" {
		err := fmt.Errorf("--origin is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.Destination) == "" {
		err := fmt.Errorf("--destination is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if cmd.Flags().Changed("position") {
		attributes["position"] = opts.Position
	}
	if cmd.Flags().Changed("miles") {
		attributes["miles"] = opts.Miles
	}
	if opts.MilesSource != "" {
		attributes["miles-source"] = opts.MilesSource
	}
	if opts.ExternalTmsOrderNumber != "" {
		attributes["external-tms-order-number"] = opts.ExternalTmsOrderNumber
	}
	if opts.ExternalTmsMovementNumber != "" {
		attributes["external-tms-movement-number"] = opts.ExternalTmsMovementNumber
	}

	relationships := map[string]any{
		"project-transport-plan": map[string]any{
			"data": map[string]string{
				"type": "project-transport-plans",
				"id":   opts.ProjectTransportPlan,
			},
		},
		"origin": map[string]any{
			"data": map[string]string{
				"type": "project-transport-plan-stops",
				"id":   opts.Origin,
			},
		},
		"destination": map[string]any{
			"data": map[string]string{
				"type": "project-transport-plan-stops",
				"id":   opts.Destination,
			},
		},
	}
	if opts.ProjectTransportPlanSet != "" {
		relationships["project-transport-plan-segment-set"] = map[string]any{
			"data": map[string]string{
				"type": "project-transport-plan-segment-sets",
				"id":   opts.ProjectTransportPlanSet,
			},
		}
	}
	if opts.Trucker != "" {
		relationships["trucker"] = map[string]any{
			"data": map[string]string{
				"type": "truckers",
				"id":   opts.Trucker,
			},
		}
	}

	data := map[string]any{
		"type":          "project-transport-plan-segments",
		"relationships": relationships,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/project-transport-plan-segments", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created project transport plan segment %s\n", details.ID)
	return nil
}

func parseDoProjectTransportPlanSegmentsCreateOptions(cmd *cobra.Command) (doProjectTransportPlanSegmentsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	projectTransportPlan, _ := cmd.Flags().GetString("project-transport-plan")
	origin, _ := cmd.Flags().GetString("origin")
	destination, _ := cmd.Flags().GetString("destination")
	segmentSet, _ := cmd.Flags().GetString("project-transport-plan-segment-set")
	trucker, _ := cmd.Flags().GetString("trucker")
	position, _ := cmd.Flags().GetInt("position")
	miles, _ := cmd.Flags().GetFloat64("miles")
	milesSource, _ := cmd.Flags().GetString("miles-source")
	externalTmsOrderNumber, _ := cmd.Flags().GetString("external-tms-order-number")
	externalTmsMovementNumber, _ := cmd.Flags().GetString("external-tms-movement-number")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectTransportPlanSegmentsCreateOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		ProjectTransportPlan:      projectTransportPlan,
		Origin:                    origin,
		Destination:               destination,
		ProjectTransportPlanSet:   segmentSet,
		Trucker:                   trucker,
		Position:                  position,
		Miles:                     miles,
		MilesSource:               milesSource,
		ExternalTmsOrderNumber:    externalTmsOrderNumber,
		ExternalTmsMovementNumber: externalTmsMovementNumber,
	}, nil
}
