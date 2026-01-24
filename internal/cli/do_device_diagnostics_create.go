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

type doDeviceDiagnosticsCreateOptions struct {
	BaseURL                              string
	Token                                string
	JSON                                 bool
	DeviceIdentifier                     string
	Device                               string
	User                                 string
	IsTrackable                          bool
	IsTracking                           bool
	IsInPowerSaverMode                   bool
	IsIgnoringBatteryOptimizations       bool
	StopOnTerminate                      bool
	PermissionStatus                     string
	LocationAccuracyAuthorizationStatus  string
	MotionPermissionStatus               string
	Changeset                            string
	ChangedAt                            string
	ChangeTriggerSource                  string
	ChangeTriggerContext                 string
	AreLocationServicesEnabled           bool
	IsGPSLocationProviderEnabled         bool
	IsNetworkLocationProviderEnabled     bool
	IsNotTrackingBecauseOfStationaryMode bool
}

func newDoDeviceDiagnosticsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a device diagnostic",
		Long: `Create a device diagnostic snapshot for tracking and permission state.

Required flags:
  --device-identifier   Device identifier (required unless --device is provided)

Optional flags:
  --device                               Device ID (use instead of device identifier)
  --user                                 User ID (defaults to current user)
  --is-trackable                         Device is trackable
  --is-tracking                          Device is currently tracking
  --stop-on-terminate                    Stop tracking when app terminates
  --is-in-power-saver-mode               Device power saver mode
  --is-ignoring-battery-optimizations    Device ignores battery optimizations
  --permission-status                    Location permission status
  --motion-permission-status             Motion permission status
  --location-accuracy-authorization-status Location accuracy authorization status
  --are-location-services-enabled        Location services enabled
  --is-gps-location-provider-enabled     GPS provider enabled
  --is-network-location-provider-enabled Network provider enabled
  --is-not-tracking-because-of-stationary-mode Stationary mode disables tracking
  --changeset                            JSON changeset payload
  --changed-at                           Timestamp of change (ISO8601)
  --change-trigger-source                Change trigger source
  --change-trigger-context               Change trigger context

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a device diagnostic
  xbe do device-diagnostics create \
    --device-identifier "ABC-123" \
    --is-tracking=true \
    --permission-status authorized \
    --are-location-services-enabled=true

  # Create with a JSON changeset
  xbe do device-diagnostics create \
    --device-identifier "ABC-123" \
    --changeset '{"battery_level":85,"network":"wifi"}' \
    --changed-at 2025-01-01T12:00:00Z

  # Output JSON
  xbe do device-diagnostics create --device-identifier "ABC-123" --json`,
		Args: cobra.NoArgs,
		RunE: runDoDeviceDiagnosticsCreate,
	}
	initDoDeviceDiagnosticsCreateFlags(cmd)
	return cmd
}

func init() {
	doDeviceDiagnosticsCmd.AddCommand(newDoDeviceDiagnosticsCreateCmd())
}

func initDoDeviceDiagnosticsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("device-identifier", "", "Device identifier (required unless --device is set)")
	cmd.Flags().String("device", "", "Device ID")
	cmd.Flags().String("user", "", "User ID (defaults to current user)")
	cmd.Flags().Bool("is-trackable", false, "Device is trackable")
	cmd.Flags().Bool("is-tracking", false, "Device is currently tracking")
	cmd.Flags().Bool("stop-on-terminate", false, "Stop tracking on app terminate")
	cmd.Flags().Bool("is-in-power-saver-mode", false, "Device power saver mode")
	cmd.Flags().Bool("is-ignoring-battery-optimizations", false, "Device ignores battery optimizations")
	cmd.Flags().String("permission-status", "", "Location permission status")
	cmd.Flags().String("motion-permission-status", "", "Motion permission status")
	cmd.Flags().String("location-accuracy-authorization-status", "", "Location accuracy authorization status")
	cmd.Flags().Bool("are-location-services-enabled", false, "Location services enabled")
	cmd.Flags().Bool("is-gps-location-provider-enabled", false, "GPS provider enabled")
	cmd.Flags().Bool("is-network-location-provider-enabled", false, "Network provider enabled")
	cmd.Flags().Bool("is-not-tracking-because-of-stationary-mode", false, "Stationary mode disables tracking")
	cmd.Flags().String("changeset", "", "Changeset JSON payload")
	cmd.Flags().String("changed-at", "", "Change timestamp (ISO8601)")
	cmd.Flags().String("change-trigger-source", "", "Change trigger source")
	cmd.Flags().String("change-trigger-context", "", "Change trigger context")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoDeviceDiagnosticsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoDeviceDiagnosticsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.DeviceIdentifier) == "" && strings.TrimSpace(opts.Device) == "" {
		err := fmt.Errorf("--device-identifier or --device is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if strings.TrimSpace(opts.DeviceIdentifier) != "" {
		attributes["device-identifier"] = opts.DeviceIdentifier
	}
	if cmd.Flags().Changed("is-trackable") {
		attributes["is-trackable"] = opts.IsTrackable
	}
	if cmd.Flags().Changed("is-tracking") {
		attributes["is-tracking"] = opts.IsTracking
	}
	if cmd.Flags().Changed("stop-on-terminate") {
		attributes["stop-on-terminate"] = opts.StopOnTerminate
	}
	if cmd.Flags().Changed("is-in-power-saver-mode") {
		attributes["is-in-power-saver-mode"] = opts.IsInPowerSaverMode
	}
	if cmd.Flags().Changed("is-ignoring-battery-optimizations") {
		attributes["is-ignoring-battery-optimizations"] = opts.IsIgnoringBatteryOptimizations
	}
	if strings.TrimSpace(opts.PermissionStatus) != "" {
		attributes["permission-status"] = opts.PermissionStatus
	}
	if strings.TrimSpace(opts.MotionPermissionStatus) != "" {
		attributes["motion-permission-status"] = opts.MotionPermissionStatus
	}
	if strings.TrimSpace(opts.LocationAccuracyAuthorizationStatus) != "" {
		attributes["location-accuracy-authorization-status"] = opts.LocationAccuracyAuthorizationStatus
	}
	if cmd.Flags().Changed("are-location-services-enabled") {
		attributes["are-location-services-enabled"] = opts.AreLocationServicesEnabled
	}
	if cmd.Flags().Changed("is-gps-location-provider-enabled") {
		attributes["is-gps-location-provider-enabled"] = opts.IsGPSLocationProviderEnabled
	}
	if cmd.Flags().Changed("is-network-location-provider-enabled") {
		attributes["is-network-location-provider-enabled"] = opts.IsNetworkLocationProviderEnabled
	}
	if cmd.Flags().Changed("is-not-tracking-because-of-stationary-mode") {
		attributes["is-not-tracking-because-of-stationary-mode"] = opts.IsNotTrackingBecauseOfStationaryMode
	}
	if strings.TrimSpace(opts.ChangedAt) != "" {
		attributes["changed-at"] = opts.ChangedAt
	}
	if strings.TrimSpace(opts.ChangeTriggerSource) != "" {
		attributes["change-trigger-source"] = opts.ChangeTriggerSource
	}
	if strings.TrimSpace(opts.ChangeTriggerContext) != "" {
		attributes["change-trigger-context"] = opts.ChangeTriggerContext
	}
	if cmd.Flags().Changed("changeset") {
		changeset := strings.TrimSpace(opts.Changeset)
		if changeset != "" {
			var parsed any
			if err := json.Unmarshal([]byte(changeset), &parsed); err != nil {
				err := fmt.Errorf("invalid --changeset JSON: %w", err)
				fmt.Fprintln(cmd.ErrOrStderr(), err)
				return err
			}
			attributes["changeset"] = parsed
		} else {
			attributes["changeset"] = map[string]any{}
		}
	}

	relationships := map[string]any{}
	if strings.TrimSpace(opts.Device) != "" {
		relationships["device"] = map[string]any{
			"data": map[string]any{
				"type": "devices",
				"id":   opts.Device,
			},
		}
	}
	if strings.TrimSpace(opts.User) != "" {
		relationships["user"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.User,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "device-diagnostics",
			"attributes": attributes,
		},
	}
	if len(relationships) > 0 {
		requestBody["data"].(map[string]any)["relationships"] = relationships
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/device-diagnostics", jsonBody)
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

	details := buildDeviceDiagnosticDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created device diagnostic %s\n", details.ID)
	return nil
}

func parseDoDeviceDiagnosticsCreateOptions(cmd *cobra.Command) (doDeviceDiagnosticsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	deviceIdentifier, _ := cmd.Flags().GetString("device-identifier")
	device, _ := cmd.Flags().GetString("device")
	user, _ := cmd.Flags().GetString("user")
	isTrackable, _ := cmd.Flags().GetBool("is-trackable")
	isTracking, _ := cmd.Flags().GetBool("is-tracking")
	stopOnTerminate, _ := cmd.Flags().GetBool("stop-on-terminate")
	isInPowerSaverMode, _ := cmd.Flags().GetBool("is-in-power-saver-mode")
	isIgnoringBatteryOptimizations, _ := cmd.Flags().GetBool("is-ignoring-battery-optimizations")
	permissionStatus, _ := cmd.Flags().GetString("permission-status")
	motionPermissionStatus, _ := cmd.Flags().GetString("motion-permission-status")
	locationAccuracyAuthorizationStatus, _ := cmd.Flags().GetString("location-accuracy-authorization-status")
	areLocationServicesEnabled, _ := cmd.Flags().GetBool("are-location-services-enabled")
	isGPSLocationProviderEnabled, _ := cmd.Flags().GetBool("is-gps-location-provider-enabled")
	isNetworkLocationProviderEnabled, _ := cmd.Flags().GetBool("is-network-location-provider-enabled")
	isNotTrackingBecauseOfStationaryMode, _ := cmd.Flags().GetBool("is-not-tracking-because-of-stationary-mode")
	changeset, _ := cmd.Flags().GetString("changeset")
	changedAt, _ := cmd.Flags().GetString("changed-at")
	changeTriggerSource, _ := cmd.Flags().GetString("change-trigger-source")
	changeTriggerContext, _ := cmd.Flags().GetString("change-trigger-context")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doDeviceDiagnosticsCreateOptions{
		BaseURL:                              baseURL,
		Token:                                token,
		JSON:                                 jsonOut,
		DeviceIdentifier:                     deviceIdentifier,
		Device:                               device,
		User:                                 user,
		IsTrackable:                          isTrackable,
		IsTracking:                           isTracking,
		IsInPowerSaverMode:                   isInPowerSaverMode,
		IsIgnoringBatteryOptimizations:       isIgnoringBatteryOptimizations,
		StopOnTerminate:                      stopOnTerminate,
		PermissionStatus:                     permissionStatus,
		MotionPermissionStatus:               motionPermissionStatus,
		LocationAccuracyAuthorizationStatus:  locationAccuracyAuthorizationStatus,
		Changeset:                            changeset,
		ChangedAt:                            changedAt,
		ChangeTriggerSource:                  changeTriggerSource,
		ChangeTriggerContext:                 changeTriggerContext,
		AreLocationServicesEnabled:           areLocationServicesEnabled,
		IsGPSLocationProviderEnabled:         isGPSLocationProviderEnabled,
		IsNetworkLocationProviderEnabled:     isNetworkLocationProviderEnabled,
		IsNotTrackingBecauseOfStationaryMode: isNotTrackingBecauseOfStationaryMode,
	}, nil
}
