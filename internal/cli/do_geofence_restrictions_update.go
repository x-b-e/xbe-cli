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

type doGeofenceRestrictionsUpdateOptions struct {
	BaseURL                       string
	Token                         string
	JSON                          bool
	ID                            string
	Geofence                      string
	Status                        string
	MaxSecondsBetweenNotification int
}

func newDoGeofenceRestrictionsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a geofence restriction",
		Long: `Update an existing geofence restriction.

Provide the restriction ID as an argument, then use flags to specify which
fields to update. Only specified fields will be modified.

Updatable fields:
  --geofence                        Geofence ID
  --status                          Restriction status
  --max-seconds-between-notification  Max seconds between notifications`,
		Example: `  # Update status
  xbe do geofence-restrictions update 123 --status inactive

  # Update notification pacing
  xbe do geofence-restrictions update 123 --max-seconds-between-notification 600

  # Update geofence
  xbe do geofence-restrictions update 123 --geofence 456`,
		Args: cobra.ExactArgs(1),
		RunE: runDoGeofenceRestrictionsUpdate,
	}
	initDoGeofenceRestrictionsUpdateFlags(cmd)
	return cmd
}

func init() {
	doGeofenceRestrictionsCmd.AddCommand(newDoGeofenceRestrictionsUpdateCmd())
}

func initDoGeofenceRestrictionsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("geofence", "", "Geofence ID")
	cmd.Flags().String("status", "", "Restriction status")
	cmd.Flags().Int("max-seconds-between-notification", 0, "Max seconds between notifications")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoGeofenceRestrictionsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoGeofenceRestrictionsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("status") {
		attributes["status"] = opts.Status
	}
	if cmd.Flags().Changed("max-seconds-between-notification") {
		attributes["max-seconds-between-notification"] = opts.MaxSecondsBetweenNotification
	}

	relationships := map[string]any{}
	if cmd.Flags().Changed("geofence") {
		relationships["geofence"] = map[string]any{
			"data": map[string]any{
				"type": "geofences",
				"id":   opts.Geofence,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no fields to update; specify at least one of --geofence, --status, --max-seconds-between-notification")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "geofence-restrictions",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/geofence-restrictions/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated geofence restriction %s\n", row.ID)
	return nil
}

func parseDoGeofenceRestrictionsUpdateOptions(cmd *cobra.Command, args []string) (doGeofenceRestrictionsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	geofence, _ := cmd.Flags().GetString("geofence")
	status, _ := cmd.Flags().GetString("status")
	maxSeconds, _ := cmd.Flags().GetInt("max-seconds-between-notification")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doGeofenceRestrictionsUpdateOptions{
		BaseURL:                       baseURL,
		Token:                         token,
		JSON:                          jsonOut,
		ID:                            args[0],
		Geofence:                      geofence,
		Status:                        status,
		MaxSecondsBetweenNotification: maxSeconds,
	}, nil
}
