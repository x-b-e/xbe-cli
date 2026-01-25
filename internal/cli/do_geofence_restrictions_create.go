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

type doGeofenceRestrictionsCreateOptions struct {
	BaseURL                       string
	Token                         string
	JSON                          bool
	Geofence                      string
	Trucker                       string
	Status                        string
	MaxSecondsBetweenNotification int
}

func newDoGeofenceRestrictionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a geofence restriction",
		Long: `Create a geofence restriction.

Geofence restrictions require a geofence in custom_truckers mode and a trucker
from the same broker.

Required flags:
  --geofence                        Geofence ID (required)
  --trucker                         Trucker ID (required)

Optional flags:
  --status                          Restriction status (active/inactive)
  --max-seconds-between-notification  Max seconds between notifications`,
		Example: `  # Create a geofence restriction
  xbe do geofence-restrictions create --geofence 123 --trucker 456

  # Create with custom notification pacing
  xbe do geofence-restrictions create --geofence 123 --trucker 456 --max-seconds-between-notification 600

  # Create inactive restriction
  xbe do geofence-restrictions create --geofence 123 --trucker 456 --status inactive`,
		Args: cobra.NoArgs,
		RunE: runDoGeofenceRestrictionsCreate,
	}
	initDoGeofenceRestrictionsCreateFlags(cmd)
	return cmd
}

func init() {
	doGeofenceRestrictionsCmd.AddCommand(newDoGeofenceRestrictionsCreateCmd())
}

func initDoGeofenceRestrictionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("geofence", "", "Geofence ID (required)")
	cmd.Flags().String("trucker", "", "Trucker ID (required)")
	cmd.Flags().String("status", "", "Restriction status (active/inactive)")
	cmd.Flags().Int("max-seconds-between-notification", 0, "Max seconds between notifications")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoGeofenceRestrictionsCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoGeofenceRestrictionsCreateOptions(cmd)
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

	if opts.Geofence == "" {
		err := fmt.Errorf("--geofence is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.Trucker == "" {
		err := fmt.Errorf("--trucker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.Status != "" {
		attributes["status"] = opts.Status
	}
	if cmd.Flags().Changed("max-seconds-between-notification") {
		attributes["max-seconds-between-notification"] = opts.MaxSecondsBetweenNotification
	}

	relationships := map[string]any{
		"geofence": map[string]any{
			"data": map[string]any{
				"type": "geofences",
				"id":   opts.Geofence,
			},
		},
		"trucker": map[string]any{
			"data": map[string]any{
				"type": "truckers",
				"id":   opts.Trucker,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "geofence-restrictions",
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

	body, _, err := client.Post(cmd.Context(), "/v1/geofence-restrictions", jsonBody)
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

	row := buildGeofenceRestrictionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created geofence restriction %s\n", row.ID)
	return nil
}

func parseDoGeofenceRestrictionsCreateOptions(cmd *cobra.Command) (doGeofenceRestrictionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	geofence, _ := cmd.Flags().GetString("geofence")
	trucker, _ := cmd.Flags().GetString("trucker")
	status, _ := cmd.Flags().GetString("status")
	maxSeconds, _ := cmd.Flags().GetInt("max-seconds-between-notification")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doGeofenceRestrictionsCreateOptions{
		BaseURL:                       baseURL,
		Token:                         token,
		JSON:                          jsonOut,
		Geofence:                      geofence,
		Trucker:                       trucker,
		Status:                        status,
		MaxSecondsBetweenNotification: maxSeconds,
	}, nil
}

func buildGeofenceRestrictionRowFromSingle(resp jsonAPISingleResponse) geofenceRestrictionRow {
	attrs := resp.Data.Attributes

	row := geofenceRestrictionRow{
		ID:                            resp.Data.ID,
		Status:                        stringAttr(attrs, "status"),
		MaxSecondsBetweenNotification: stringAttr(attrs, "max-seconds-between-notification"),
	}

	if rel, ok := resp.Data.Relationships["geofence"]; ok && rel.Data != nil {
		row.GeofenceID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["trucker"]; ok && rel.Data != nil {
		row.TruckerID = rel.Data.ID
	}

	return row
}
