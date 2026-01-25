package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type deviceDiagnosticsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type deviceDiagnosticDetails struct {
	ID                                   string `json:"id"`
	DeviceID                             string `json:"device_id,omitempty"`
	DeviceIdentifier                     string `json:"device_identifier,omitempty"`
	UserID                               string `json:"user_id,omitempty"`
	IsTrackable                          bool   `json:"is_trackable"`
	IsTracking                           bool   `json:"is_tracking"`
	StopOnTerminate                      bool   `json:"stop_on_terminate"`
	IsInPowerSaverMode                   bool   `json:"is_in_power_saver_mode"`
	IsIgnoringBatteryOptimizations       bool   `json:"is_ignoring_battery_optimizations"`
	AreLocationServicesEnabled           bool   `json:"are_location_services_enabled"`
	IsGPSLocationProviderEnabled         bool   `json:"is_gps_location_provider_enabled"`
	IsNetworkLocationProviderEnabled     bool   `json:"is_network_location_provider_enabled"`
	IsNotTrackingBecauseOfStationaryMode bool   `json:"is_not_tracking_because_of_stationary_mode"`
	PermissionStatus                     string `json:"permission_status,omitempty"`
	MotionPermissionStatus               string `json:"motion_permission_status,omitempty"`
	LocationAccuracyAuthorizationStatus  string `json:"location_accuracy_authorization_status,omitempty"`
	NativeAppVersion                     string `json:"native_app_version,omitempty"`
	NativeOTAVersion                     string `json:"native_ota_version,omitempty"`
	OTADeviceIdentifier                  string `json:"ota_device_identifier,omitempty"`
	DeviceInfo                           any    `json:"device_info,omitempty"`
	Changeset                            any    `json:"changeset,omitempty"`
	ChangeTriggerSource                  string `json:"change_trigger_source,omitempty"`
	ChangeTriggerContext                 string `json:"change_trigger_context,omitempty"`
	ChangedAt                            string `json:"changed_at,omitempty"`
	CreatedAt                            string `json:"created_at,omitempty"`
	UpdatedAt                            string `json:"updated_at,omitempty"`
}

func newDeviceDiagnosticsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show device diagnostic details",
		Long: `Show the full details of a device diagnostic snapshot.

Includes tracking state, permissions, power settings, and device info
captured at the time of the diagnostic report.

Arguments:
  <id>    Device diagnostic ID (required). Use the list command to find IDs.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a device diagnostic
  xbe view device-diagnostics show 123

  # Output JSON for scripting
  xbe view device-diagnostics show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDeviceDiagnosticsShow,
	}
	initDeviceDiagnosticsShowFlags(cmd)
	return cmd
}

func init() {
	deviceDiagnosticsCmd.AddCommand(newDeviceDiagnosticsShowCmd())
}

func initDeviceDiagnosticsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDeviceDiagnosticsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseDeviceDiagnosticsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("device diagnostic id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("include", "device,user")
	query.Set("fields[devices]", "identifier")

	body, _, err := client.Get(cmd.Context(), "/v1/device-diagnostics/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildDeviceDiagnosticDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderDeviceDiagnosticDetails(cmd, details)
}

func parseDeviceDiagnosticsShowOptions(cmd *cobra.Command) (deviceDiagnosticsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return deviceDiagnosticsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildDeviceDiagnosticDetails(resp jsonAPISingleResponse) deviceDiagnosticDetails {
	resource := resp.Data
	attrs := resource.Attributes

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := deviceDiagnosticDetails{
		ID:                                   resource.ID,
		DeviceIdentifier:                     stringAttr(attrs, "device-identifier"),
		IsTrackable:                          boolAttr(attrs, "is-trackable"),
		IsTracking:                           boolAttr(attrs, "is-tracking"),
		StopOnTerminate:                      boolAttr(attrs, "stop-on-terminate"),
		IsInPowerSaverMode:                   boolAttr(attrs, "is-in-power-saver-mode"),
		IsIgnoringBatteryOptimizations:       boolAttr(attrs, "is-ignoring-battery-optimizations"),
		AreLocationServicesEnabled:           boolAttr(attrs, "are-location-services-enabled"),
		IsGPSLocationProviderEnabled:         boolAttr(attrs, "is-gps-location-provider-enabled"),
		IsNetworkLocationProviderEnabled:     boolAttr(attrs, "is-network-location-provider-enabled"),
		IsNotTrackingBecauseOfStationaryMode: boolAttr(attrs, "is-not-tracking-because-of-stationary-mode"),
		PermissionStatus:                     stringAttr(attrs, "permission-status"),
		MotionPermissionStatus:               stringAttr(attrs, "motion-permission-status"),
		LocationAccuracyAuthorizationStatus:  stringAttr(attrs, "location-accuracy-authorization-status"),
		NativeAppVersion:                     stringAttr(attrs, "native-app-version"),
		NativeOTAVersion:                     stringAttr(attrs, "native-ota-version"),
		OTADeviceIdentifier:                  stringAttr(attrs, "ota-device-identifier"),
		ChangeTriggerSource:                  stringAttr(attrs, "change-trigger-source"),
		ChangeTriggerContext:                 stringAttr(attrs, "change-trigger-context"),
		ChangedAt:                            stringAttr(attrs, "changed-at"),
		CreatedAt:                            stringAttr(attrs, "created-at"),
		UpdatedAt:                            stringAttr(attrs, "updated-at"),
	}

	if value, ok := attrs["device-info"]; ok {
		details.DeviceInfo = value
	}
	if value, ok := attrs["changeset"]; ok {
		details.Changeset = value
	}

	if rel, ok := resource.Relationships["device"]; ok && rel.Data != nil {
		details.DeviceID = rel.Data.ID
		if details.DeviceIdentifier == "" {
			if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				details.DeviceIdentifier = stringAttr(inc.Attributes, "identifier")
			}
		}
	}
	if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
		details.UserID = rel.Data.ID
	}

	return details
}

func renderDeviceDiagnosticDetails(cmd *cobra.Command, details deviceDiagnosticDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	fmt.Fprintln(out, "")

	fmt.Fprintln(out, "Device:")
	if details.DeviceID != "" {
		fmt.Fprintf(out, "  ID: %s\n", details.DeviceID)
	}
	if details.DeviceIdentifier != "" {
		fmt.Fprintf(out, "  Identifier: %s\n", details.DeviceIdentifier)
	}
	fmt.Fprintln(out, "")

	if details.UserID != "" {
		fmt.Fprintln(out, "User:")
		fmt.Fprintf(out, "  ID: %s\n", details.UserID)
		fmt.Fprintln(out, "")
	}

	fmt.Fprintln(out, "Tracking:")
	fmt.Fprintf(out, "  Trackable: %s\n", formatBool(details.IsTrackable))
	fmt.Fprintf(out, "  Tracking: %s\n", formatBool(details.IsTracking))
	fmt.Fprintf(out, "  Stop On Terminate: %s\n", formatBool(details.StopOnTerminate))
	fmt.Fprintf(out, "  Stationary Mode Disabled: %s\n", formatBool(details.IsNotTrackingBecauseOfStationaryMode))
	fmt.Fprintln(out, "")

	fmt.Fprintln(out, "Permissions:")
	if details.PermissionStatus != "" {
		fmt.Fprintf(out, "  Location Permission: %s\n", details.PermissionStatus)
	}
	if details.MotionPermissionStatus != "" {
		fmt.Fprintf(out, "  Motion Permission: %s\n", details.MotionPermissionStatus)
	}
	if details.LocationAccuracyAuthorizationStatus != "" {
		fmt.Fprintf(out, "  Location Accuracy Authorization: %s\n", details.LocationAccuracyAuthorizationStatus)
	}
	fmt.Fprintln(out, "")

	fmt.Fprintln(out, "Location Services:")
	fmt.Fprintf(out, "  Location Services Enabled: %s\n", formatBool(details.AreLocationServicesEnabled))
	fmt.Fprintf(out, "  GPS Provider Enabled: %s\n", formatBool(details.IsGPSLocationProviderEnabled))
	fmt.Fprintf(out, "  Network Provider Enabled: %s\n", formatBool(details.IsNetworkLocationProviderEnabled))
	fmt.Fprintln(out, "")

	fmt.Fprintln(out, "Power:")
	fmt.Fprintf(out, "  Power Saver Mode: %s\n", formatBool(details.IsInPowerSaverMode))
	fmt.Fprintf(out, "  Ignoring Battery Optimizations: %s\n", formatBool(details.IsIgnoringBatteryOptimizations))
	fmt.Fprintln(out, "")

	if details.ChangeTriggerSource != "" || details.ChangeTriggerContext != "" || details.ChangedAt != "" {
		fmt.Fprintln(out, "Change Trigger:")
		if details.ChangeTriggerSource != "" {
			fmt.Fprintf(out, "  Source: %s\n", details.ChangeTriggerSource)
		}
		if details.ChangeTriggerContext != "" {
			fmt.Fprintf(out, "  Context: %s\n", details.ChangeTriggerContext)
		}
		if details.ChangedAt != "" {
			fmt.Fprintf(out, "  Changed At: %s\n", details.ChangedAt)
		}
		fmt.Fprintln(out, "")
	}

	if details.NativeAppVersion != "" || details.NativeOTAVersion != "" || details.OTADeviceIdentifier != "" {
		fmt.Fprintln(out, "App/OTA:")
		if details.NativeAppVersion != "" {
			fmt.Fprintf(out, "  Native App Version: %s\n", details.NativeAppVersion)
		}
		if details.NativeOTAVersion != "" {
			fmt.Fprintf(out, "  Native OTA Version: %s\n", details.NativeOTAVersion)
		}
		if details.OTADeviceIdentifier != "" {
			fmt.Fprintf(out, "  OTA Device Identifier: %s\n", details.OTADeviceIdentifier)
		}
		fmt.Fprintln(out, "")
	}

	if details.DeviceInfo != nil {
		fmt.Fprintln(out, "Device Info:")
		fmt.Fprintln(out, formatDeviceDiagnosticJSONBlock(details.DeviceInfo, "  "))
		fmt.Fprintln(out, "")
	}

	if details.Changeset != nil {
		fmt.Fprintln(out, "Changeset:")
		fmt.Fprintln(out, formatDeviceDiagnosticJSONBlock(details.Changeset, "  "))
		fmt.Fprintln(out, "")
	}

	fmt.Fprintln(out, "Timestamps:")
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "  Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "  Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}

func formatDeviceDiagnosticJSONBlock(value any, indent string) string {
	if value == nil {
		return ""
	}
	pretty, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	if indent == "" {
		return string(pretty)
	}
	lines := strings.Split(string(pretty), "\n")
	for i, line := range lines {
		lines[i] = indent + line
	}
	return strings.Join(lines, "\n")
}
