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

type doEquipmentLocationEventsCreateOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	EquipmentID    string
	EventLatitude  string
	EventLongitude string
	EventAt        string
	Provenance     string
}

func newDoEquipmentLocationEventsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an equipment location event",
		Long: `Create an equipment location event.

Required flags:
  --equipment    Equipment ID (required)
  --provenance   Event provenance (gps, map) (required)

Optional flags:
  --event-at         Event timestamp (ISO 8601)
  --event-latitude   Event latitude
  --event-longitude  Event longitude`,
		Example: `  # Create a location event for equipment
  xbe do equipment-location-events create \
    --equipment 123 \
    --event-at 2025-01-15T12:00:00Z \
    --event-latitude 40.7128 \
    --event-longitude -74.0060 \
    --provenance gps`,
		Args: cobra.NoArgs,
		RunE: runDoEquipmentLocationEventsCreate,
	}
	initDoEquipmentLocationEventsCreateFlags(cmd)
	return cmd
}

func init() {
	doEquipmentLocationEventsCmd.AddCommand(newDoEquipmentLocationEventsCreateCmd())
}

func initDoEquipmentLocationEventsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("equipment", "", "Equipment ID (required)")
	cmd.Flags().String("event-at", "", "Event timestamp (ISO 8601)")
	cmd.Flags().String("event-latitude", "", "Event latitude")
	cmd.Flags().String("event-longitude", "", "Event longitude")
	cmd.Flags().String("provenance", "", "Event provenance (gps, map) (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoEquipmentLocationEventsCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoEquipmentLocationEventsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.EquipmentID) == "" {
		err := fmt.Errorf("--equipment is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.Provenance) == "" {
		err := fmt.Errorf("--provenance is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"provenance": opts.Provenance,
	}

	if strings.TrimSpace(opts.EventAt) != "" {
		attributes["event-at"] = opts.EventAt
	}
	if strings.TrimSpace(opts.EventLatitude) != "" {
		attributes["event-latitude"] = opts.EventLatitude
	}
	if strings.TrimSpace(opts.EventLongitude) != "" {
		attributes["event-longitude"] = opts.EventLongitude
	}

	relationships := map[string]any{
		"equipment": map[string]any{
			"data": map[string]any{
				"type": "equipment",
				"id":   opts.EquipmentID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "equipment-location-events",
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

	body, _, err := client.Post(cmd.Context(), "/v1/equipment-location-events", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created equipment location event %s\n", row.ID)
	return nil
}

func parseDoEquipmentLocationEventsCreateOptions(cmd *cobra.Command) (doEquipmentLocationEventsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	equipmentID, _ := cmd.Flags().GetString("equipment")
	eventAt, _ := cmd.Flags().GetString("event-at")
	eventLatitude, _ := cmd.Flags().GetString("event-latitude")
	eventLongitude, _ := cmd.Flags().GetString("event-longitude")
	provenance, _ := cmd.Flags().GetString("provenance")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doEquipmentLocationEventsCreateOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		EquipmentID:    equipmentID,
		EventAt:        eventAt,
		EventLatitude:  eventLatitude,
		EventLongitude: eventLongitude,
		Provenance:     provenance,
	}, nil
}
