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

type doGeofencesUpdateOptions struct {
	BaseURL                                             string
	Token                                               string
	JSON                                                bool
	ID                                                  string
	Name                                                string
	Description                                         string
	PolygonCoordinates                                  string
	Status                                              string
	RestrictionMode                                     string
	ExplicitGeofenceRestrictionViolationNotificationMsg string
}

func newDoGeofencesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a geofence",
		Long: `Update a geofence.

Optional flags:
  --name                      Geofence name
  --description               Geofence description
  --polygon-coordinates       JSON array of coordinate pairs [[lng,lat],[lng,lat],...]
  --status                    Geofence status
  --restriction-mode          Restriction mode
  --notification-message      Explicit notification message for restriction violations`,
		Example: `  # Update geofence name
  xbe do geofences update 123 --name "Updated Office"

  # Update polygon coordinates
  xbe do geofences update 123 \
    --polygon-coordinates '[[-122.4,37.8],[-122.4,37.7],[-122.3,37.7],[-122.3,37.8]]'

  # Update status
  xbe do geofences update 123 --status inactive`,
		Args: cobra.ExactArgs(1),
		RunE: runDoGeofencesUpdate,
	}
	initDoGeofencesUpdateFlags(cmd)
	return cmd
}

func init() {
	doGeofencesCmd.AddCommand(newDoGeofencesUpdateCmd())
}

func initDoGeofencesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Geofence name")
	cmd.Flags().String("description", "", "Geofence description")
	cmd.Flags().String("polygon-coordinates", "", "JSON array of coordinate pairs")
	cmd.Flags().String("status", "", "Geofence status")
	cmd.Flags().String("restriction-mode", "", "Restriction mode")
	cmd.Flags().String("notification-message", "", "Explicit notification message for restriction violations")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoGeofencesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoGeofencesUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("name") {
		attributes["name"] = opts.Name
	}
	if cmd.Flags().Changed("description") {
		attributes["description"] = opts.Description
	}
	if cmd.Flags().Changed("polygon-coordinates") {
		if opts.PolygonCoordinates != "" {
			var coords [][]float64
			if err := json.Unmarshal([]byte(opts.PolygonCoordinates), &coords); err != nil {
				err = fmt.Errorf("invalid polygon-coordinates JSON: %w", err)
				fmt.Fprintln(cmd.ErrOrStderr(), err)
				return err
			}
			attributes["polygon-coordinates"] = coords
		} else {
			attributes["polygon-coordinates"] = nil
		}
	}
	if cmd.Flags().Changed("status") {
		attributes["status"] = opts.Status
	}
	if cmd.Flags().Changed("restriction-mode") {
		attributes["restriction-mode"] = opts.RestrictionMode
	}
	if cmd.Flags().Changed("notification-message") {
		attributes["explicit-geofence-restriction-violation-notification-message"] = opts.ExplicitGeofenceRestrictionViolationNotificationMsg
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "geofences",
			"id":         opts.ID,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/geofences/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated geofence %s\n", row.ID)
	return nil
}

func parseDoGeofencesUpdateOptions(cmd *cobra.Command, args []string) (doGeofencesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	description, _ := cmd.Flags().GetString("description")
	polygonCoordinates, _ := cmd.Flags().GetString("polygon-coordinates")
	status, _ := cmd.Flags().GetString("status")
	restrictionMode, _ := cmd.Flags().GetString("restriction-mode")
	notificationMessage, _ := cmd.Flags().GetString("notification-message")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doGeofencesUpdateOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		ID:                 args[0],
		Name:               name,
		Description:        description,
		PolygonCoordinates: polygonCoordinates,
		Status:             status,
		RestrictionMode:    restrictionMode,
		ExplicitGeofenceRestrictionViolationNotificationMsg: notificationMessage,
	}, nil
}
