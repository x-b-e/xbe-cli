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

type doEquipmentMovementStopsCreateOptions struct {
	BaseURL            string
	Token              string
	JSON               bool
	TripID             string
	LocationID         string
	SequencePosition   int
	ScheduledArrivalAt string
}

func newDoEquipmentMovementStopsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an equipment movement stop",
		Long: `Create an equipment movement stop.

Required flags:
  --trip      Equipment movement trip ID (required)
  --location  Requirement location ID (required)

Optional flags:
  --sequence-position     Position in trip sequence
  --scheduled-arrival-at  Scheduled arrival time (ISO 8601)`,
		Example: `  # Create a stop
  xbe do equipment-movement-stops create \
    --trip 123 \
    --location 456 \
    --sequence-position 1 \
    --scheduled-arrival-at 2025-01-01T08:00:00Z`,
		Args: cobra.NoArgs,
		RunE: runDoEquipmentMovementStopsCreate,
	}
	initDoEquipmentMovementStopsCreateFlags(cmd)
	return cmd
}

func init() {
	doEquipmentMovementStopsCmd.AddCommand(newDoEquipmentMovementStopsCreateCmd())
}

func initDoEquipmentMovementStopsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("trip", "", "Equipment movement trip ID (required)")
	cmd.Flags().String("location", "", "Requirement location ID (required)")
	cmd.Flags().Int("sequence-position", 0, "Position in trip sequence")
	cmd.Flags().String("scheduled-arrival-at", "", "Scheduled arrival time (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("trip")
	_ = cmd.MarkFlagRequired("location")
}

func runDoEquipmentMovementStopsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoEquipmentMovementStopsCreateOptions(cmd)
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

	if opts.TripID == "" {
		err := fmt.Errorf("--trip is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.LocationID == "" {
		err := fmt.Errorf("--location is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if cmd.Flags().Changed("sequence-position") {
		attributes["sequence-position"] = opts.SequencePosition
	}
	if opts.ScheduledArrivalAt != "" {
		attributes["scheduled-arrival-at"] = opts.ScheduledArrivalAt
	}

	relationships := map[string]any{
		"trip": map[string]any{
			"data": map[string]any{
				"type": "equipment-movement-trips",
				"id":   opts.TripID,
			},
		},
		"location": map[string]any{
			"data": map[string]any{
				"type": "equipment-movement-requirement-locations",
				"id":   opts.LocationID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "equipment-movement-stops",
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

	body, _, err := client.Post(cmd.Context(), "/v1/equipment-movement-stops", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created equipment movement stop %s\n", row.ID)
	return nil
}

func parseDoEquipmentMovementStopsCreateOptions(cmd *cobra.Command) (doEquipmentMovementStopsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	tripID, _ := cmd.Flags().GetString("trip")
	locationID, _ := cmd.Flags().GetString("location")
	sequencePosition, _ := cmd.Flags().GetInt("sequence-position")
	scheduledArrivalAt, _ := cmd.Flags().GetString("scheduled-arrival-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doEquipmentMovementStopsCreateOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		TripID:             tripID,
		LocationID:         locationID,
		SequencePosition:   sequencePosition,
		ScheduledArrivalAt: scheduledArrivalAt,
	}, nil
}
