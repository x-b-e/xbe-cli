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

type doProjectTransportPlanSegmentsUpdateOptions struct {
	BaseURL                   string
	Token                     string
	JSON                      bool
	ID                        string
	Position                  int
	Miles                     float64
	MilesSource               string
	ExternalTmsOrderNumber    string
	ExternalTmsMovementNumber string
	ProjectTransportPlanSet   string
	Trucker                   string
}

func newDoProjectTransportPlanSegmentsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a project transport plan segment",
		Long: `Update a project transport plan segment.

Optional:
  --position                        Segment position within the plan
  --miles                           Segment miles
  --miles-source                    Miles source (unknown, transport_route)
  --project-transport-plan-segment-set  Segment set ID
  --trucker                         Trucker ID
  --external-tms-order-number       External TMS order number (transport-only projects)
  --external-tms-movement-number    External TMS movement number (transport-only projects)`,
		Example: `  # Update segment miles
  xbe do project-transport-plan-segments update 123 --miles 15.2

  # Update segment set
  xbe do project-transport-plan-segments update 123 --project-transport-plan-segment-set 456`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProjectTransportPlanSegmentsUpdate,
	}
	initDoProjectTransportPlanSegmentsUpdateFlags(cmd)
	return cmd
}

func init() {
	doProjectTransportPlanSegmentsCmd.AddCommand(newDoProjectTransportPlanSegmentsUpdateCmd())
}

func initDoProjectTransportPlanSegmentsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Int("position", 0, "Segment position within the plan")
	cmd.Flags().Float64("miles", 0, "Segment miles")
	cmd.Flags().String("miles-source", "", "Miles source (unknown, transport_route)")
	cmd.Flags().String("project-transport-plan-segment-set", "", "Segment set ID")
	cmd.Flags().String("trucker", "", "Trucker ID")
	cmd.Flags().String("external-tms-order-number", "", "External TMS order number (transport-only projects)")
	cmd.Flags().String("external-tms-movement-number", "", "External TMS movement number (transport-only projects)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectTransportPlanSegmentsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectTransportPlanSegmentsUpdateOptions(cmd, args)
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

	attributes := map[string]any{}
	if cmd.Flags().Changed("position") {
		attributes["position"] = opts.Position
	}
	if cmd.Flags().Changed("miles") {
		attributes["miles"] = opts.Miles
	}
	if cmd.Flags().Changed("miles-source") {
		attributes["miles-source"] = opts.MilesSource
	}
	if cmd.Flags().Changed("external-tms-order-number") {
		attributes["external-tms-order-number"] = opts.ExternalTmsOrderNumber
	}
	if cmd.Flags().Changed("external-tms-movement-number") {
		attributes["external-tms-movement-number"] = opts.ExternalTmsMovementNumber
	}

	relationships := map[string]any{}
	if opts.ProjectTransportPlanSet != "" {
		relationships["project-transport-plan-segment-set"] = map[string]any{
			"data": map[string]any{
				"type": "project-transport-plan-segment-sets",
				"id":   opts.ProjectTransportPlanSet,
			},
		}
	}
	if opts.Trucker != "" {
		relationships["trucker"] = map[string]any{
			"data": map[string]any{
				"type": "truckers",
				"id":   opts.Trucker,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes or relationships to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "project-transport-plan-segments",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
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

	body, _, err := client.Patch(cmd.Context(), "/v1/project-transport-plan-segments/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated project transport plan segment %s\n", details.ID)
	return nil
}

func parseDoProjectTransportPlanSegmentsUpdateOptions(cmd *cobra.Command, args []string) (doProjectTransportPlanSegmentsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	position, _ := cmd.Flags().GetInt("position")
	miles, _ := cmd.Flags().GetFloat64("miles")
	milesSource, _ := cmd.Flags().GetString("miles-source")
	externalTmsOrderNumber, _ := cmd.Flags().GetString("external-tms-order-number")
	externalTmsMovementNumber, _ := cmd.Flags().GetString("external-tms-movement-number")
	segmentSet, _ := cmd.Flags().GetString("project-transport-plan-segment-set")
	trucker, _ := cmd.Flags().GetString("trucker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectTransportPlanSegmentsUpdateOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		ID:                        args[0],
		Position:                  position,
		Miles:                     miles,
		MilesSource:               milesSource,
		ExternalTmsOrderNumber:    externalTmsOrderNumber,
		ExternalTmsMovementNumber: externalTmsMovementNumber,
		ProjectTransportPlanSet:   segmentSet,
		Trucker:                   trucker,
	}, nil
}
