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

type doDeviceLocationEventsCreateOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	DeviceIdentifier string
	Payload          string
	EventID          string
	EventAt          string
	EventDescription string
	EventLatitude    string
	EventLongitude   string
}

type deviceLocationEventDetails struct {
	ID               string `json:"id"`
	DeviceIdentifier string `json:"device_identifier,omitempty"`
	EventID          string `json:"event_id,omitempty"`
	EventAt          string `json:"event_at,omitempty"`
	EventDescription string `json:"event_description,omitempty"`
	EventLatitude    string `json:"event_latitude,omitempty"`
	EventLongitude   string `json:"event_longitude,omitempty"`
	Payload          any    `json:"payload,omitempty"`
}

func newDoDeviceLocationEventsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a device location event",
		Long: `Create a device location event.

Required flags:
  --device-identifier   Device identifier (hardware/app identifier)

Optional flags:
  --payload             Event payload (JSON string)
  --event-id            Event UUID
  --event-at            Event timestamp (ISO 8601)
  --event-description   Event description or activity type
  --event-latitude      Event latitude
  --event-longitude     Event longitude

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create an event from payload
  xbe do device-location-events create --device-identifier "ios:ABC123" \
    --payload '{"uuid":"evt-1","timestamp":"2025-01-01T00:00:00Z","activity":{"type":"walking"},"coords":{"latitude":40.0,"longitude":-74.0}}'

  # Create an event with explicit fields
  xbe do device-location-events create --device-identifier "ios:ABC123" \
    --event-id "evt-2" --event-at 2025-01-01T00:05:00Z --event-description "moving" \
    --event-latitude 40.1 --event-longitude -74.1`,
		Args: cobra.NoArgs,
		RunE: runDoDeviceLocationEventsCreate,
	}
	initDoDeviceLocationEventsCreateFlags(cmd)
	return cmd
}

func init() {
	doDeviceLocationEventsCmd.AddCommand(newDoDeviceLocationEventsCreateCmd())
}

func initDoDeviceLocationEventsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("device-identifier", "", "Device identifier (required)")
	cmd.Flags().String("payload", "", "Event payload (JSON string)")
	cmd.Flags().String("event-id", "", "Event UUID")
	cmd.Flags().String("event-at", "", "Event timestamp (ISO 8601)")
	cmd.Flags().String("event-description", "", "Event description or activity type")
	cmd.Flags().String("event-latitude", "", "Event latitude")
	cmd.Flags().String("event-longitude", "", "Event longitude")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("device-identifier")
}

func runDoDeviceLocationEventsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoDeviceLocationEventsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.DeviceIdentifier) == "" {
		err := fmt.Errorf("--device-identifier is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"device-identifier": opts.DeviceIdentifier,
	}

	if strings.TrimSpace(opts.Payload) != "" {
		var payload any
		if err := json.Unmarshal([]byte(opts.Payload), &payload); err != nil {
			err := fmt.Errorf("invalid payload JSON: %w", err)
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["payload"] = payload
	}
	if strings.TrimSpace(opts.EventID) != "" {
		attributes["event-id"] = opts.EventID
	}
	if strings.TrimSpace(opts.EventAt) != "" {
		attributes["event-at"] = opts.EventAt
	}
	if strings.TrimSpace(opts.EventDescription) != "" {
		attributes["event-description"] = opts.EventDescription
	}
	if strings.TrimSpace(opts.EventLatitude) != "" {
		attributes["event-latitude"] = opts.EventLatitude
	}
	if strings.TrimSpace(opts.EventLongitude) != "" {
		attributes["event-longitude"] = opts.EventLongitude
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "device-location-events",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/device-location-events", jsonBody)
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

	details := buildDeviceLocationEventDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created device location event %s\n", details.ID)
	return nil
}

func parseDoDeviceLocationEventsCreateOptions(cmd *cobra.Command) (doDeviceLocationEventsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	deviceIdentifier, _ := cmd.Flags().GetString("device-identifier")
	payload, _ := cmd.Flags().GetString("payload")
	eventID, _ := cmd.Flags().GetString("event-id")
	eventAt, _ := cmd.Flags().GetString("event-at")
	eventDescription, _ := cmd.Flags().GetString("event-description")
	eventLatitude, _ := cmd.Flags().GetString("event-latitude")
	eventLongitude, _ := cmd.Flags().GetString("event-longitude")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doDeviceLocationEventsCreateOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		DeviceIdentifier: deviceIdentifier,
		Payload:          payload,
		EventID:          eventID,
		EventAt:          eventAt,
		EventDescription: eventDescription,
		EventLatitude:    eventLatitude,
		EventLongitude:   eventLongitude,
	}, nil
}

func buildDeviceLocationEventDetails(resp jsonAPISingleResponse) deviceLocationEventDetails {
	attrs := resp.Data.Attributes

	return deviceLocationEventDetails{
		ID:               resp.Data.ID,
		DeviceIdentifier: stringAttr(attrs, "device-identifier"),
		EventID:          stringAttr(attrs, "event-id"),
		EventAt:          formatDateTime(stringAttr(attrs, "event-at")),
		EventDescription: stringAttr(attrs, "event-description"),
		EventLatitude:    stringAttr(attrs, "event-latitude"),
		EventLongitude:   stringAttr(attrs, "event-longitude"),
		Payload:          attrs["payload"],
	}
}
