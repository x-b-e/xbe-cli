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

type doProjectTransportPlanSegmentSetsCreateOptions struct {
	BaseURL              string
	Token                string
	JSON                 bool
	ProjectTransportPlan string
	Trucker              string
	ExternalTmsLegNumber string
	Position             int
}

func newDoProjectTransportPlanSegmentSetsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project transport plan segment set",
		Long: `Create a project transport plan segment set.

Required:
  --project-transport-plan  Project transport plan ID

Optional:
  --trucker                 Trucker ID
  --external-tms-leg-number External TMS leg number
  --position                Sequence position within the plan`,
		Example: `  # Create a segment set
  xbe do project-transport-plan-segment-sets create --project-transport-plan 123

  # Create with a trucker and external leg number
  xbe do project-transport-plan-segment-sets create \\
    --project-transport-plan 123 \\
    --trucker 456 \\
    --external-tms-leg-number LEG-01 \\
    --position 1`,
		RunE: runDoProjectTransportPlanSegmentSetsCreate,
	}
	initDoProjectTransportPlanSegmentSetsCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectTransportPlanSegmentSetsCmd.AddCommand(newDoProjectTransportPlanSegmentSetsCreateCmd())
}

func initDoProjectTransportPlanSegmentSetsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project-transport-plan", "", "Project transport plan ID")
	cmd.Flags().String("trucker", "", "Trucker ID")
	cmd.Flags().String("external-tms-leg-number", "", "External TMS leg number")
	cmd.Flags().Int("position", 0, "Sequence position within the plan")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("project-transport-plan")
}

func runDoProjectTransportPlanSegmentSetsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectTransportPlanSegmentSetsCreateOptions(cmd)
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

	attributes := map[string]any{}
	if cmd.Flags().Changed("external-tms-leg-number") {
		attributes["external-tms-leg-number"] = opts.ExternalTmsLegNumber
	}
	if cmd.Flags().Changed("position") {
		attributes["position"] = opts.Position
	}

	relationships := map[string]any{
		"project-transport-plan": map[string]any{
			"data": map[string]any{
				"type": "project-transport-plans",
				"id":   opts.ProjectTransportPlan,
			},
		},
	}
	if opts.Trucker != "" {
		relationships["trucker"] = map[string]any{
			"data": map[string]any{
				"type": "truckers",
				"id":   opts.Trucker,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "project-transport-plan-segment-sets",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/project-transport-plan-segment-sets", jsonBody)
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

	if opts.JSON {
		row := buildProjectTransportPlanSegmentSetRow(resp.Data)
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project transport plan segment set %s\n", resp.Data.ID)
	return nil
}

func parseDoProjectTransportPlanSegmentSetsCreateOptions(cmd *cobra.Command) (doProjectTransportPlanSegmentSetsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	projectTransportPlan, _ := cmd.Flags().GetString("project-transport-plan")
	trucker, _ := cmd.Flags().GetString("trucker")
	externalTmsLegNumber, _ := cmd.Flags().GetString("external-tms-leg-number")
	position, _ := cmd.Flags().GetInt("position")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectTransportPlanSegmentSetsCreateOptions{
		BaseURL:              baseURL,
		Token:                token,
		JSON:                 jsonOut,
		ProjectTransportPlan: projectTransportPlan,
		Trucker:              trucker,
		ExternalTmsLegNumber: externalTmsLegNumber,
		Position:             position,
	}, nil
}
