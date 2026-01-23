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

type doEquipmentLocationEventsUpdateOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	ID             string
	EquipmentID    string
	EventLatitude  string
	EventLongitude string
	EventAt        string
	Provenance     string
}

func newDoEquipmentLocationEventsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an equipment location event",
		Long: `Update an equipment location event.

Optional flags:
  --equipment        Equipment ID
  --event-at         Event timestamp (ISO 8601)
  --event-latitude   Event latitude
  --event-longitude  Event longitude
  --provenance       Event provenance (gps, map)`,
		Example: `  # Update event location
  xbe do equipment-location-events update 123 \
    --event-at 2025-01-16T12:00:00Z \
    --event-latitude 41.0000 \
    --event-longitude -73.9000

  # Update equipment and provenance
  xbe do equipment-location-events update 123 --equipment 456 --provenance map`,
		Args: cobra.ExactArgs(1),
		RunE: runDoEquipmentLocationEventsUpdate,
	}
	initDoEquipmentLocationEventsUpdateFlags(cmd)
	return cmd
}

func init() {
	doEquipmentLocationEventsCmd.AddCommand(newDoEquipmentLocationEventsUpdateCmd())
}

func initDoEquipmentLocationEventsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("equipment", "", "Equipment ID")
	cmd.Flags().String("event-at", "", "Event timestamp (ISO 8601)")
	cmd.Flags().String("event-latitude", "", "Event latitude")
	cmd.Flags().String("event-longitude", "", "Event longitude")
	cmd.Flags().String("provenance", "", "Event provenance (gps, map)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoEquipmentLocationEventsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoEquipmentLocationEventsUpdateOptions(cmd, args)
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
	var hasChanges bool

	if cmd.Flags().Changed("event-at") {
		attributes["event-at"] = opts.EventAt
		hasChanges = true
	}
	if cmd.Flags().Changed("event-latitude") {
		if opts.EventLatitude == "" {
			attributes["event-latitude"] = nil
		} else {
			attributes["event-latitude"] = opts.EventLatitude
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("event-longitude") {
		if opts.EventLongitude == "" {
			attributes["event-longitude"] = nil
		} else {
			attributes["event-longitude"] = opts.EventLongitude
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("provenance") {
		attributes["provenance"] = opts.Provenance
		hasChanges = true
	}
	if cmd.Flags().Changed("equipment") {
		if opts.EquipmentID == "" {
			relationships["equipment"] = map[string]any{"data": nil}
		} else {
			relationships["equipment"] = map[string]any{
				"data": map[string]any{
					"type": "equipment",
					"id":   opts.EquipmentID,
				},
			}
		}
		hasChanges = true
	}

	if !hasChanges {
		err := fmt.Errorf("no fields to update; specify at least one of --equipment, --event-at, --event-latitude, --event-longitude, --provenance")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "equipment-location-events",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/equipment-location-events/"+opts.ID, jsonBody)
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

	row := buildEquipmentLocationEventRow(resp.Data)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated equipment location event %s\n", row.ID)
	return nil
}

func parseDoEquipmentLocationEventsUpdateOptions(cmd *cobra.Command, args []string) (doEquipmentLocationEventsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	equipmentID, _ := cmd.Flags().GetString("equipment")
	eventAt, _ := cmd.Flags().GetString("event-at")
	eventLatitude, _ := cmd.Flags().GetString("event-latitude")
	eventLongitude, _ := cmd.Flags().GetString("event-longitude")
	provenance, _ := cmd.Flags().GetString("provenance")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doEquipmentLocationEventsUpdateOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		ID:             args[0],
		EquipmentID:    equipmentID,
		EventAt:        eventAt,
		EventLatitude:  eventLatitude,
		EventLongitude: eventLongitude,
		Provenance:     provenance,
	}, nil
}
