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

type doGeofencesCreateOptions struct {
	BaseURL                                             string
	Token                                               string
	JSON                                                bool
	Name                                                string
	Description                                         string
	PolygonCoordinates                                  string
	Status                                              string
	RestrictionMode                                     string
	ExplicitGeofenceRestrictionViolationNotificationMsg string
	BrokerID                                            string
}

func newDoGeofencesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new geofence",
		Long: `Create a new geofence.

Required flags:
  --name                      Geofence name (required)
  --broker                    Broker ID (required)

Optional flags:
  --description               Geofence description
  --polygon-coordinates       JSON array of coordinate pairs [[lng,lat],[lng,lat],...]
  --status                    Geofence status
  --restriction-mode          Restriction mode
  --notification-message      Explicit notification message for restriction violations`,
		Example: `  # Create a basic geofence
  xbe do geofences create --name "Main Office" --broker 123

  # Create with polygon coordinates
  xbe do geofences create --name "Job Site A" --broker 123 \
    --polygon-coordinates '[[-122.4,37.8],[-122.4,37.7],[-122.3,37.7],[-122.3,37.8]]'

  # Create with description and status
  xbe do geofences create --name "Warehouse" --broker 123 \
    --description "Main warehouse area" --status active`,
		Args: cobra.NoArgs,
		RunE: runDoGeofencesCreate,
	}
	initDoGeofencesCreateFlags(cmd)
	return cmd
}

func init() {
	doGeofencesCmd.AddCommand(newDoGeofencesCreateCmd())
}

func initDoGeofencesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Geofence name (required)")
	cmd.Flags().String("description", "", "Geofence description")
	cmd.Flags().String("polygon-coordinates", "", "JSON array of coordinate pairs")
	cmd.Flags().String("status", "", "Geofence status")
	cmd.Flags().String("restriction-mode", "", "Restriction mode")
	cmd.Flags().String("notification-message", "", "Explicit notification message for restriction violations")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoGeofencesCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoGeofencesCreateOptions(cmd)
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

	if opts.Name == "" {
		err := fmt.Errorf("--name is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.BrokerID == "" {
		err := fmt.Errorf("--broker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"name": opts.Name,
	}

	if opts.Description != "" {
		attributes["description"] = opts.Description
	}
	if opts.PolygonCoordinates != "" {
		// Parse the JSON array
		var coords [][]float64
		if err := json.Unmarshal([]byte(opts.PolygonCoordinates), &coords); err != nil {
			err = fmt.Errorf("invalid polygon-coordinates JSON: %w", err)
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["polygon-coordinates"] = coords
	}
	if opts.Status != "" {
		attributes["status"] = opts.Status
	}
	if opts.RestrictionMode != "" {
		attributes["restriction-mode"] = opts.RestrictionMode
	}
	if opts.ExplicitGeofenceRestrictionViolationNotificationMsg != "" {
		attributes["explicit-geofence-restriction-violation-notification-message"] = opts.ExplicitGeofenceRestrictionViolationNotificationMsg
	}

	relationships := map[string]any{
		"broker": map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.BrokerID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "geofences",
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

	body, _, err := client.Post(cmd.Context(), "/v1/geofences", jsonBody)
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

	row := buildGeofenceRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created geofence %s\n", row.ID)
	return nil
}

func parseDoGeofencesCreateOptions(cmd *cobra.Command) (doGeofencesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	description, _ := cmd.Flags().GetString("description")
	polygonCoordinates, _ := cmd.Flags().GetString("polygon-coordinates")
	status, _ := cmd.Flags().GetString("status")
	restrictionMode, _ := cmd.Flags().GetString("restriction-mode")
	notificationMessage, _ := cmd.Flags().GetString("notification-message")
	brokerID, _ := cmd.Flags().GetString("broker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doGeofencesCreateOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		Name:               name,
		Description:        description,
		PolygonCoordinates: polygonCoordinates,
		Status:             status,
		RestrictionMode:    restrictionMode,
		ExplicitGeofenceRestrictionViolationNotificationMsg: notificationMessage,
		BrokerID: brokerID,
	}, nil
}

func buildGeofenceRowFromSingle(resp jsonAPISingleResponse) geofenceRow {
	attrs := resp.Data.Attributes

	row := geofenceRow{
		ID:              resp.Data.ID,
		Name:            stringAttr(attrs, "name"),
		Description:     stringAttr(attrs, "description"),
		Status:          stringAttr(attrs, "status"),
		RestrictionMode: stringAttr(attrs, "restriction-mode"),
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}

	return row
}
