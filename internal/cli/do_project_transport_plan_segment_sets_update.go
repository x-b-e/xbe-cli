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

type doProjectTransportPlanSegmentSetsUpdateOptions struct {
	BaseURL              string
	Token                string
	JSON                 bool
	ID                   string
	ExternalTmsLegNumber string
	Position             int
	Trucker              string
}

func newDoProjectTransportPlanSegmentSetsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a project transport plan segment set",
		Long: `Update a project transport plan segment set.

Optional:
  --external-tms-leg-number External TMS leg number (empty to clear)
  --position                Sequence position within the plan
  --trucker                 Trucker ID (empty to clear)`,
		Example: `  # Update the external leg number
  xbe do project-transport-plan-segment-sets update 123 --external-tms-leg-number LEG-02

  # Update position
  xbe do project-transport-plan-segment-sets update 123 --position 2

  # Clear trucker assignment
  xbe do project-transport-plan-segment-sets update 123 --trucker ""`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProjectTransportPlanSegmentSetsUpdate,
	}
	initDoProjectTransportPlanSegmentSetsUpdateFlags(cmd)
	return cmd
}

func init() {
	doProjectTransportPlanSegmentSetsCmd.AddCommand(newDoProjectTransportPlanSegmentSetsUpdateCmd())
}

func initDoProjectTransportPlanSegmentSetsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("external-tms-leg-number", "", "External TMS leg number")
	cmd.Flags().Int("position", 0, "Sequence position within the plan")
	cmd.Flags().String("trucker", "", "Trucker ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectTransportPlanSegmentSetsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectTransportPlanSegmentSetsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("external-tms-leg-number") {
		attributes["external-tms-leg-number"] = opts.ExternalTmsLegNumber
	}
	if cmd.Flags().Changed("position") {
		attributes["position"] = opts.Position
	}

	relationships := map[string]any{}
	if cmd.Flags().Changed("trucker") {
		if strings.TrimSpace(opts.Trucker) == "" {
			relationships["trucker"] = map[string]any{
				"data": nil,
			}
		} else {
			relationships["trucker"] = map[string]any{
				"data": map[string]any{
					"type": "truckers",
					"id":   opts.Trucker,
				},
			}
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no fields to update; specify at least one field")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "project-transport-plan-segment-sets",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/project-transport-plan-segment-sets/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated project transport plan segment set %s\n", resp.Data.ID)
	return nil
}

func parseDoProjectTransportPlanSegmentSetsUpdateOptions(cmd *cobra.Command, args []string) (doProjectTransportPlanSegmentSetsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	externalTmsLegNumber, _ := cmd.Flags().GetString("external-tms-leg-number")
	position, _ := cmd.Flags().GetInt("position")
	trucker, _ := cmd.Flags().GetString("trucker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectTransportPlanSegmentSetsUpdateOptions{
		BaseURL:              baseURL,
		Token:                token,
		JSON:                 jsonOut,
		ID:                   args[0],
		ExternalTmsLegNumber: externalTmsLegNumber,
		Position:             position,
		Trucker:              trucker,
	}, nil
}
