package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type deviceDiagnosticsListOptions struct {
	BaseURL                                string
	Token                                  string
	JSON                                   bool
	NoAuth                                 bool
	Limit                                  int
	Offset                                 int
	Sort                                   string
	User                                   string
	Device                                 string
	DeviceIdentifier                       string
	PermissionStatus                       string
	NotPermissionStatus                    string
	HasPermissionStatus                    string
	MotionPermissionStatus                 string
	NotMotionPermissionStatus              string
	HasMotionPermissionStatus              string
	LocationAccuracyAuthorizationStatus    string
	HasLocationAccuracyAuthorizationStatus string
	AreLocationServicesEnabled             string
	IsGPSLocationProviderEnabled           string
	IsNetworkLocationProviderEnabled       string
	IsNotTrackingBecauseOfStationaryMode   string
}

type deviceDiagnosticRow struct {
	ID                                  string `json:"id"`
	DeviceID                            string `json:"device_id,omitempty"`
	DeviceIdentifier                    string `json:"device_identifier,omitempty"`
	UserID                              string `json:"user_id,omitempty"`
	IsTracking                          bool   `json:"is_tracking"`
	PermissionStatus                    string `json:"permission_status,omitempty"`
	MotionPermissionStatus              string `json:"motion_permission_status,omitempty"`
	LocationAccuracyAuthorizationStatus string `json:"location_accuracy_authorization_status,omitempty"`
	ChangedAt                           string `json:"changed_at,omitempty"`
	CreatedAt                           string `json:"created_at,omitempty"`
}

func newDeviceDiagnosticsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List device diagnostics",
		Long: `List device diagnostic snapshots with filtering and pagination.

Output Columns:
  ID         Device diagnostic identifier
  DEVICE     Device identifier (or device ID)
  USER       User ID (if present)
  TRACKING   Whether the device is tracking
  PERMISSION Location permission status
  CHANGED    Timestamp for the reported change (if provided)

Filters:
  --user                                      Filter by user ID
  --device                                    Filter by device ID
  --device-identifier                         Filter by device identifier
  --permission-status                         Filter by permission status
  --not-permission-status                     Exclude permission status
  --has-permission-status                     Filter by permission status presence (true/false)
  --motion-permission-status                  Filter by motion permission status
  --not-motion-permission-status              Exclude motion permission status
  --has-motion-permission-status              Filter by motion permission status presence (true/false)
  --location-accuracy-authorization-status    Filter by location accuracy authorization status
  --has-location-accuracy-authorization-status Filter by location accuracy authorization status presence (true/false)
  --are-location-services-enabled             Filter by location services enabled (true/false)
  --is-gps-location-provider-enabled          Filter by GPS location provider enabled (true/false)
  --is-network-location-provider-enabled      Filter by network location provider enabled (true/false)
  --is-not-tracking-because-of-stationary-mode Filter by stationary mode tracking disable (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List recent device diagnostics
  xbe view device-diagnostics list --limit 10

  # Filter by device identifier
  xbe view device-diagnostics list --device-identifier "ABC-123"

  # Filter by permission status
  xbe view device-diagnostics list --permission-status authorized

  # Filter by presence of motion permission status
  xbe view device-diagnostics list --has-motion-permission-status true

  # Output as JSON
  xbe view device-diagnostics list --json`,
		Args: cobra.NoArgs,
		RunE: runDeviceDiagnosticsList,
	}
	initDeviceDiagnosticsListFlags(cmd)
	return cmd
}

func init() {
	deviceDiagnosticsCmd.AddCommand(newDeviceDiagnosticsListCmd())
}

func initDeviceDiagnosticsListFlags(cmd *cobra.Command) {
	cmd.Flags().String("user", "", "Filter by user ID")
	cmd.Flags().String("device", "", "Filter by device ID")
	cmd.Flags().String("device-identifier", "", "Filter by device identifier")
	cmd.Flags().String("permission-status", "", "Filter by permission status")
	cmd.Flags().String("not-permission-status", "", "Exclude permission status")
	cmd.Flags().String("has-permission-status", "", "Filter by permission status presence (true/false)")
	cmd.Flags().String("motion-permission-status", "", "Filter by motion permission status")
	cmd.Flags().String("not-motion-permission-status", "", "Exclude motion permission status")
	cmd.Flags().String("has-motion-permission-status", "", "Filter by motion permission status presence (true/false)")
	cmd.Flags().String("location-accuracy-authorization-status", "", "Filter by location accuracy authorization status")
	cmd.Flags().String("has-location-accuracy-authorization-status", "", "Filter by location accuracy authorization status presence (true/false)")
	cmd.Flags().String("are-location-services-enabled", "", "Filter by location services enabled (true/false)")
	cmd.Flags().String("is-gps-location-provider-enabled", "", "Filter by GPS location provider enabled (true/false)")
	cmd.Flags().String("is-network-location-provider-enabled", "", "Filter by network location provider enabled (true/false)")
	cmd.Flags().String("is-not-tracking-because-of-stationary-mode", "", "Filter by stationary mode tracking disable (true/false)")

	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDeviceDiagnosticsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseDeviceDiagnosticsListOptions(cmd)
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

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("include", "device,user")
	query.Set("fields[device-diagnostics]", "device-identifier,is-tracking,permission-status,motion-permission-status,location-accuracy-authorization-status,changed-at,created-at")
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	} else {
		query.Set("sort", "-created-at")
	}

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	setFilterIfPresent(query, "filter[user]", opts.User)
	setFilterIfPresent(query, "filter[device]", opts.Device)
	setFilterIfPresent(query, "filter[device-identifier]", opts.DeviceIdentifier)
	setFilterIfPresent(query, "filter[permission-status]", opts.PermissionStatus)
	setFilterIfPresent(query, "filter[not-permission-status]", opts.NotPermissionStatus)
	setFilterIfPresent(query, "filter[has-permission-status]", opts.HasPermissionStatus)
	setFilterIfPresent(query, "filter[motion-permission-status]", opts.MotionPermissionStatus)
	setFilterIfPresent(query, "filter[not-motion-permission-status]", opts.NotMotionPermissionStatus)
	setFilterIfPresent(query, "filter[has-motion-permission-status]", opts.HasMotionPermissionStatus)
	setFilterIfPresent(query, "filter[location-accuracy-authorization-status]", opts.LocationAccuracyAuthorizationStatus)
	setFilterIfPresent(query, "filter[has-location-accuracy-authorization-status]", opts.HasLocationAccuracyAuthorizationStatus)
	setFilterIfPresent(query, "filter[are-location-services-enabled]", opts.AreLocationServicesEnabled)
	setFilterIfPresent(query, "filter[is-gps-location-provider-enabled]", opts.IsGPSLocationProviderEnabled)
	setFilterIfPresent(query, "filter[is-network-location-provider-enabled]", opts.IsNetworkLocationProviderEnabled)
	setFilterIfPresent(query, "filter[is-not-tracking-because-of-stationary-mode]", opts.IsNotTrackingBecauseOfStationaryMode)

	body, _, err := client.Get(cmd.Context(), "/v1/device-diagnostics", query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	handled, err := renderSparseListIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	rows := buildDeviceDiagnosticRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderDeviceDiagnosticsTable(cmd, rows)
}

func parseDeviceDiagnosticsListOptions(cmd *cobra.Command) (deviceDiagnosticsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	user, _ := cmd.Flags().GetString("user")
	device, _ := cmd.Flags().GetString("device")
	deviceIdentifier, _ := cmd.Flags().GetString("device-identifier")
	permissionStatus, _ := cmd.Flags().GetString("permission-status")
	notPermissionStatus, _ := cmd.Flags().GetString("not-permission-status")
	hasPermissionStatus, _ := cmd.Flags().GetString("has-permission-status")
	motionPermissionStatus, _ := cmd.Flags().GetString("motion-permission-status")
	notMotionPermissionStatus, _ := cmd.Flags().GetString("not-motion-permission-status")
	hasMotionPermissionStatus, _ := cmd.Flags().GetString("has-motion-permission-status")
	locationAccuracyAuthorizationStatus, _ := cmd.Flags().GetString("location-accuracy-authorization-status")
	hasLocationAccuracyAuthorizationStatus, _ := cmd.Flags().GetString("has-location-accuracy-authorization-status")
	areLocationServicesEnabled, _ := cmd.Flags().GetString("are-location-services-enabled")
	isGPSLocationProviderEnabled, _ := cmd.Flags().GetString("is-gps-location-provider-enabled")
	isNetworkLocationProviderEnabled, _ := cmd.Flags().GetString("is-network-location-provider-enabled")
	isNotTrackingBecauseOfStationaryMode, _ := cmd.Flags().GetString("is-not-tracking-because-of-stationary-mode")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return deviceDiagnosticsListOptions{
		BaseURL:                                baseURL,
		Token:                                  token,
		JSON:                                   jsonOut,
		NoAuth:                                 noAuth,
		Limit:                                  limit,
		Offset:                                 offset,
		Sort:                                   sort,
		User:                                   user,
		Device:                                 device,
		DeviceIdentifier:                       deviceIdentifier,
		PermissionStatus:                       permissionStatus,
		NotPermissionStatus:                    notPermissionStatus,
		HasPermissionStatus:                    hasPermissionStatus,
		MotionPermissionStatus:                 motionPermissionStatus,
		NotMotionPermissionStatus:              notMotionPermissionStatus,
		HasMotionPermissionStatus:              hasMotionPermissionStatus,
		LocationAccuracyAuthorizationStatus:    locationAccuracyAuthorizationStatus,
		HasLocationAccuracyAuthorizationStatus: hasLocationAccuracyAuthorizationStatus,
		AreLocationServicesEnabled:             areLocationServicesEnabled,
		IsGPSLocationProviderEnabled:           isGPSLocationProviderEnabled,
		IsNetworkLocationProviderEnabled:       isNetworkLocationProviderEnabled,
		IsNotTrackingBecauseOfStationaryMode:   isNotTrackingBecauseOfStationaryMode,
	}, nil
}

func buildDeviceDiagnosticRows(resp jsonAPIResponse) []deviceDiagnosticRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]deviceDiagnosticRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := deviceDiagnosticRow{
			ID:                                  resource.ID,
			DeviceIdentifier:                    stringAttr(attrs, "device-identifier"),
			IsTracking:                          boolAttr(attrs, "is-tracking"),
			PermissionStatus:                    stringAttr(attrs, "permission-status"),
			MotionPermissionStatus:              stringAttr(attrs, "motion-permission-status"),
			LocationAccuracyAuthorizationStatus: stringAttr(attrs, "location-accuracy-authorization-status"),
			ChangedAt:                           stringAttr(attrs, "changed-at"),
			CreatedAt:                           stringAttr(attrs, "created-at"),
		}

		if rel, ok := resource.Relationships["device"]; ok && rel.Data != nil {
			row.DeviceID = rel.Data.ID
			if row.DeviceIdentifier == "" {
				if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
					row.DeviceIdentifier = stringAttr(inc.Attributes, "identifier")
				}
			}
		}
		if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
			row.UserID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderDeviceDiagnosticsTable(cmd *cobra.Command, rows []deviceDiagnosticRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No device diagnostics found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tDEVICE\tUSER\tTRACKING\tPERMISSION\tCHANGED")
	for _, row := range rows {
		tracking := "no"
		if row.IsTracking {
			tracking = "yes"
		}
		deviceLabel := firstNonEmpty(row.DeviceIdentifier, row.DeviceID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(deviceLabel, 24),
			row.UserID,
			tracking,
			truncateString(row.PermissionStatus, 16),
			formatDate(row.ChangedAt),
		)
	}
	return writer.Flush()
}
