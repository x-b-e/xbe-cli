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

type devicesListOptions struct {
	BaseURL                       string
	Token                         string
	JSON                          bool
	Limit                         int
	Offset                        int
	Sort                          string
	User                          string
	Identifier                    string
	IsPushable                    string
	HasPushToken                  string
	PushableStatus                string
	FirstDeviceLocationEventAtMin string
	FirstDeviceLocationEventAtMax string
	LastDeviceLocationEventAtMin  string
	LastDeviceLocationEventAtMax  string
	NoAuth                        bool
}

type deviceRow struct {
	ID                         string `json:"id"`
	Identifier                 string `json:"identifier,omitempty"`
	Nickname                   string `json:"nickname,omitempty"`
	IsPushable                 bool   `json:"is_pushable"`
	IsPreferred                bool   `json:"is_preferred"`
	HasPushToken               bool   `json:"has_push_token"`
	PushableStatus             string `json:"pushable_status,omitempty"`
	LastNativeAppVersion       string `json:"last_native_app_version,omitempty"`
	FirstDeviceLocationEventAt string `json:"first_device_location_event_at,omitempty"`
	LastDeviceLocationEventAt  string `json:"last_device_location_event_at,omitempty"`
	UserID                     string `json:"user_id,omitempty"`
}

func newDevicesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List devices",
		Long: `List devices (mobile app instances).

Filter flags:
  --user                                 Filter by user ID
  --identifier                           Filter by device identifier
  --is-pushable                          Filter by pushable status (true/false)
  --has-push-token                       Filter by presence of push token (true/false)
  --pushable-status                      Filter by pushable status enum
  --first-device-location-event-at-min   Filter by minimum first location event date (ISO8601)
  --first-device-location-event-at-max   Filter by maximum first location event date (ISO8601)
  --last-device-location-event-at-min    Filter by minimum last location event date (ISO8601)
  --last-device-location-event-at-max    Filter by maximum last location event date (ISO8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List all devices
  xbe view devices list

  # List devices for a specific user
  xbe view devices list --user 123

  # List only pushable devices
  xbe view devices list --is-pushable true

  # List devices with push tokens
  xbe view devices list --has-push-token true`,
		Args: cobra.NoArgs,
		RunE: runDevicesList,
	}
	initDevicesListFlags(cmd)
	return cmd
}

func init() {
	devicesCmd.AddCommand(newDevicesListCmd())
}

func initDevicesListFlags(cmd *cobra.Command) {
	// Filter flags
	cmd.Flags().String("user", "", "Filter by user ID")
	cmd.Flags().String("identifier", "", "Filter by device identifier")
	cmd.Flags().String("is-pushable", "", "Filter by pushable status (true/false)")
	cmd.Flags().String("has-push-token", "", "Filter by presence of push token (true/false)")
	cmd.Flags().String("pushable-status", "", "Filter by pushable status enum")
	cmd.Flags().String("first-device-location-event-at-min", "", "Filter by minimum first location event date (ISO8601)")
	cmd.Flags().String("first-device-location-event-at-max", "", "Filter by maximum first location event date (ISO8601)")
	cmd.Flags().String("last-device-location-event-at-min", "", "Filter by minimum last location event date (ISO8601)")
	cmd.Flags().String("last-device-location-event-at-max", "", "Filter by maximum last location event date (ISO8601)")

	// Global flags
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Int("limit", 0, "Limit results")
	cmd.Flags().Int("offset", 0, "Offset results")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
	cmd.Flags().Bool("no-auth", false, "Skip authentication")
}

func runDevicesList(cmd *cobra.Command, args []string) error {
	opts, err := parseDevicesListOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if !opts.NoAuth && strings.TrimSpace(opts.Token) == "" {
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
	query.Set("include", "user")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	// Apply filters
	if opts.User != "" {
		query.Set("filter[user]", opts.User)
	}
	if opts.Identifier != "" {
		query.Set("filter[identifier]", opts.Identifier)
	}
	if opts.IsPushable != "" {
		query.Set("filter[is-pushable]", opts.IsPushable)
	}
	if opts.HasPushToken != "" {
		query.Set("filter[has-push-token]", opts.HasPushToken)
	}
	if opts.PushableStatus != "" {
		query.Set("filter[pushable-status]", opts.PushableStatus)
	}
	if opts.FirstDeviceLocationEventAtMin != "" {
		query.Set("filter[first-device-location-event-at-min]", opts.FirstDeviceLocationEventAtMin)
	}
	if opts.FirstDeviceLocationEventAtMax != "" {
		query.Set("filter[first-device-location-event-at-max]", opts.FirstDeviceLocationEventAtMax)
	}
	if opts.LastDeviceLocationEventAtMin != "" {
		query.Set("filter[last-device-location-event-at-min]", opts.LastDeviceLocationEventAtMin)
	}
	if opts.LastDeviceLocationEventAtMax != "" {
		query.Set("filter[last-device-location-event-at-max]", opts.LastDeviceLocationEventAtMax)
	}

	body, _, err := client.Get(cmd.Context(), "/v1/devices", query)
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

	rows := buildDeviceRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderDevicesTable(cmd, rows)
}

func renderDevicesTable(cmd *cobra.Command, rows []deviceRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No devices found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tIDENTIFIER\tNICKNAME\tPUSHABLE\tPREFERRED\tAPP VERSION\tUSER ID")
	for _, row := range rows {
		pushable := "no"
		if row.IsPushable {
			pushable = "yes"
		}
		preferred := "no"
		if row.IsPreferred {
			preferred = "yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Identifier, 20),
			truncateString(row.Nickname, 20),
			pushable,
			preferred,
			truncateString(row.LastNativeAppVersion, 15),
			row.UserID,
		)
	}
	return writer.Flush()
}

func parseDevicesListOptions(cmd *cobra.Command) (devicesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	user, _ := cmd.Flags().GetString("user")
	identifier, _ := cmd.Flags().GetString("identifier")
	isPushable, _ := cmd.Flags().GetString("is-pushable")
	hasPushToken, _ := cmd.Flags().GetString("has-push-token")
	pushableStatus, _ := cmd.Flags().GetString("pushable-status")
	firstDeviceLocationEventAtMin, _ := cmd.Flags().GetString("first-device-location-event-at-min")
	firstDeviceLocationEventAtMax, _ := cmd.Flags().GetString("first-device-location-event-at-max")
	lastDeviceLocationEventAtMin, _ := cmd.Flags().GetString("last-device-location-event-at-min")
	lastDeviceLocationEventAtMax, _ := cmd.Flags().GetString("last-device-location-event-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")
	noAuth, _ := cmd.Flags().GetBool("no-auth")

	return devicesListOptions{
		BaseURL:                       baseURL,
		Token:                         token,
		JSON:                          jsonOut,
		Limit:                         limit,
		Offset:                        offset,
		Sort:                          sort,
		User:                          user,
		Identifier:                    identifier,
		IsPushable:                    isPushable,
		HasPushToken:                  hasPushToken,
		PushableStatus:                pushableStatus,
		FirstDeviceLocationEventAtMin: firstDeviceLocationEventAtMin,
		FirstDeviceLocationEventAtMax: firstDeviceLocationEventAtMax,
		LastDeviceLocationEventAtMin:  lastDeviceLocationEventAtMin,
		LastDeviceLocationEventAtMax:  lastDeviceLocationEventAtMax,
		NoAuth:                        noAuth,
	}, nil
}

func buildDeviceRows(resp jsonAPIResponse) []deviceRow {
	rows := make([]deviceRow, 0, len(resp.Data))
	for _, item := range resp.Data {
		attrs := item.Attributes
		row := deviceRow{
			ID:                         item.ID,
			Identifier:                 stringAttr(attrs, "identifier"),
			Nickname:                   stringAttr(attrs, "nickname"),
			IsPushable:                 boolAttr(attrs, "is-pushable"),
			IsPreferred:                boolAttr(attrs, "is-preferred"),
			HasPushToken:               boolAttr(attrs, "has-push-token"),
			PushableStatus:             stringAttr(attrs, "pushable-status"),
			LastNativeAppVersion:       stringAttr(attrs, "last-native-app-version"),
			FirstDeviceLocationEventAt: stringAttr(attrs, "first-device-location-event-at"),
			LastDeviceLocationEventAt:  stringAttr(attrs, "last-device-location-event-at"),
		}

		if rel, ok := item.Relationships["user"]; ok && rel.Data != nil {
			row.UserID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}
