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

type doUserDeviceLocationTrackingRequestsCreateOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	User                   string
	LocationTrackingKind   string
	LocationTrackingAction string
}

type userDeviceLocationTrackingRequestRow struct {
	ID                     string   `json:"id"`
	UserID                 string   `json:"user_id,omitempty"`
	LocationTrackingKind   string   `json:"location_tracking_kind,omitempty"`
	LocationTrackingAction string   `json:"location_tracking_action,omitempty"`
	DevicesToNotifyIDs     []string `json:"devices_to_notify_ids,omitempty"`
}

func newDoUserDeviceLocationTrackingRequestsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Send a location tracking request",
		Long: `Send a location tracking request to a user's devices.

Required flags:
  --user   User ID to notify

Optional flags:
  --location-tracking-kind    Tracking kind (normal or continuous, default: normal)
  --location-tracking-action  Tracking action (start or stop, default: start)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Request normal tracking start
  xbe do user-device-location-tracking-requests create --user 123

  # Request continuous tracking start
  xbe do user-device-location-tracking-requests create \
    --user 123 \
    --location-tracking-kind continuous \
    --location-tracking-action start

  # Request tracking stop
  xbe do user-device-location-tracking-requests create \
    --user 123 \
    --location-tracking-action stop

  # JSON output
  xbe do user-device-location-tracking-requests create --user 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoUserDeviceLocationTrackingRequestsCreate,
	}
	initDoUserDeviceLocationTrackingRequestsCreateFlags(cmd)
	return cmd
}

func init() {
	doUserDeviceLocationTrackingRequestsCmd.AddCommand(newDoUserDeviceLocationTrackingRequestsCreateCmd())
}

func initDoUserDeviceLocationTrackingRequestsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("user", "", "User ID to notify (required)")
	cmd.Flags().String("location-tracking-kind", "", "Tracking kind (normal or continuous)")
	cmd.Flags().String("location-tracking-action", "", "Tracking action (start or stop)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("user")
}

func runDoUserDeviceLocationTrackingRequestsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoUserDeviceLocationTrackingRequestsCreateOptions(cmd)
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

	userID := strings.TrimSpace(opts.User)
	if userID == "" {
		err := fmt.Errorf("--user is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if strings.TrimSpace(opts.LocationTrackingKind) != "" {
		attributes["location-tracking-kind"] = strings.TrimSpace(opts.LocationTrackingKind)
	}
	if strings.TrimSpace(opts.LocationTrackingAction) != "" {
		attributes["location-tracking-action"] = strings.TrimSpace(opts.LocationTrackingAction)
	}

	relationships := map[string]any{
		"user": map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   userID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "user-device-location-tracking-requests",
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

	body, _, err := client.Post(cmd.Context(), "/v1/user-device-location-tracking-requests", jsonBody)
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

	row := buildUserDeviceLocationTrackingRequestRowFromSingle(resp)
	if row.UserID == "" {
		row.UserID = userID
	}
	if row.LocationTrackingKind == "" {
		row.LocationTrackingKind = strings.TrimSpace(opts.LocationTrackingKind)
	}
	if row.LocationTrackingAction == "" {
		row.LocationTrackingAction = strings.TrimSpace(opts.LocationTrackingAction)
	}

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	return renderUserDeviceLocationTrackingRequest(cmd, row)
}

func parseDoUserDeviceLocationTrackingRequestsCreateOptions(cmd *cobra.Command) (doUserDeviceLocationTrackingRequestsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	user, _ := cmd.Flags().GetString("user")
	locationTrackingKind, _ := cmd.Flags().GetString("location-tracking-kind")
	locationTrackingAction, _ := cmd.Flags().GetString("location-tracking-action")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doUserDeviceLocationTrackingRequestsCreateOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		User:                   user,
		LocationTrackingKind:   locationTrackingKind,
		LocationTrackingAction: locationTrackingAction,
	}, nil
}

func buildUserDeviceLocationTrackingRequestRowFromSingle(resp jsonAPISingleResponse) userDeviceLocationTrackingRequestRow {
	resource := resp.Data
	return userDeviceLocationTrackingRequestRow{
		ID:                     resource.ID,
		UserID:                 relationshipIDFromMap(resource.Relationships, "user"),
		LocationTrackingKind:   stringAttr(resource.Attributes, "location-tracking-kind"),
		LocationTrackingAction: stringAttr(resource.Attributes, "location-tracking-action"),
		DevicesToNotifyIDs:     relationshipIDsFromMap(resource.Relationships, "devices-to-notify"),
	}
}

func renderUserDeviceLocationTrackingRequest(cmd *cobra.Command, row userDeviceLocationTrackingRequestRow) error {
	out := cmd.OutOrStdout()

	if row.ID != "" {
		fmt.Fprintf(out, "Created user device location tracking request %s\n", row.ID)
	} else {
		fmt.Fprintln(out, "Created user device location tracking request")
	}

	if row.UserID != "" {
		fmt.Fprintf(out, "User: %s\n", row.UserID)
	}
	if row.LocationTrackingKind != "" {
		fmt.Fprintf(out, "Tracking kind: %s\n", row.LocationTrackingKind)
	}
	if row.LocationTrackingAction != "" {
		fmt.Fprintf(out, "Tracking action: %s\n", row.LocationTrackingAction)
	}

	if len(row.DevicesToNotifyIDs) == 0 {
		fmt.Fprintln(out, "Devices to notify: none")
		return nil
	}

	fmt.Fprintf(out, "Devices to notify (%d):\n", len(row.DevicesToNotifyIDs))
	for _, deviceID := range row.DevicesToNotifyIDs {
		fmt.Fprintf(out, "  %s\n", deviceID)
	}

	return nil
}
