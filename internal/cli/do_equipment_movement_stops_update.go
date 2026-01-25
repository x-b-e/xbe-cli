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

type doEquipmentMovementStopsUpdateOptions struct {
	BaseURL            string
	Token              string
	JSON               bool
	ID                 string
	LocationID         string
	SequencePosition   int
	ScheduledArrivalAt string
}

func newDoEquipmentMovementStopsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an equipment movement stop",
		Long: `Update an equipment movement stop.

Optional flags:
  --location              Requirement location ID
  --sequence-position     Position in trip sequence
  --scheduled-arrival-at  Scheduled arrival time (ISO 8601)`,
		Example: `  # Update sequence position
  xbe do equipment-movement-stops update 123 --sequence-position 2

  # Update scheduled arrival
  xbe do equipment-movement-stops update 123 --scheduled-arrival-at 2025-01-01T09:00:00Z

  # Update location
  xbe do equipment-movement-stops update 123 --location 456`,
		Args: cobra.ExactArgs(1),
		RunE: runDoEquipmentMovementStopsUpdate,
	}
	initDoEquipmentMovementStopsUpdateFlags(cmd)
	return cmd
}

func init() {
	doEquipmentMovementStopsCmd.AddCommand(newDoEquipmentMovementStopsUpdateCmd())
}

func initDoEquipmentMovementStopsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("location", "", "Requirement location ID")
	cmd.Flags().Int("sequence-position", 0, "Position in trip sequence")
	cmd.Flags().String("scheduled-arrival-at", "", "Scheduled arrival time (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoEquipmentMovementStopsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoEquipmentMovementStopsUpdateOptions(cmd, args)
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
	relationships := map[string]any{}

	if cmd.Flags().Changed("sequence-position") {
		attributes["sequence-position"] = opts.SequencePosition
	}
	if cmd.Flags().Changed("scheduled-arrival-at") {
		attributes["scheduled-arrival-at"] = opts.ScheduledArrivalAt
	}
	if cmd.Flags().Changed("location") {
		if opts.LocationID == "" {
			relationships["location"] = map[string]any{"data": nil}
		} else {
			relationships["location"] = map[string]any{
				"data": map[string]any{
					"type": "equipment-movement-requirement-locations",
					"id":   opts.LocationID,
				},
			}
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "equipment-movement-stops",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/equipment-movement-stops/"+opts.ID, jsonBody)
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

	row := buildEquipmentMovementStopRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated equipment movement stop %s\n", row.ID)
	return nil
}

func parseDoEquipmentMovementStopsUpdateOptions(cmd *cobra.Command, args []string) (doEquipmentMovementStopsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	locationID, _ := cmd.Flags().GetString("location")
	sequencePosition, _ := cmd.Flags().GetInt("sequence-position")
	scheduledArrivalAt, _ := cmd.Flags().GetString("scheduled-arrival-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doEquipmentMovementStopsUpdateOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		ID:                 args[0],
		LocationID:         locationID,
		SequencePosition:   sequencePosition,
		ScheduledArrivalAt: scheduledArrivalAt,
	}, nil
}
