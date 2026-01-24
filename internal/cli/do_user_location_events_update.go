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

type doUserLocationEventsUpdateOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	ID             string
	User           string
	EventAt        string
	EventLatitude  string
	EventLongitude string
	Provenance     string
}

func newDoUserLocationEventsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a user location event",
		Long: `Update a user location event.

Optional flags:
  --user              User ID for the event
  --event-at          Event timestamp (ISO 8601)
  --event-latitude    Event latitude
  --event-longitude   Event longitude
  --provenance        Event provenance (gps,map)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update coordinates
  xbe do user-location-events update 123 --event-latitude 40.0 --event-longitude -74.0

  # Update provenance
  xbe do user-location-events update 123 --provenance map`,
		Args: cobra.ExactArgs(1),
		RunE: runDoUserLocationEventsUpdate,
	}
	initDoUserLocationEventsUpdateFlags(cmd)
	return cmd
}

func init() {
	doUserLocationEventsCmd.AddCommand(newDoUserLocationEventsUpdateCmd())
}

func initDoUserLocationEventsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("user", "", "User ID")
	cmd.Flags().String("event-at", "", "Event timestamp (ISO 8601)")
	cmd.Flags().String("event-latitude", "", "Event latitude")
	cmd.Flags().String("event-longitude", "", "Event longitude")
	cmd.Flags().String("provenance", "", "Event provenance (gps,map)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoUserLocationEventsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoUserLocationEventsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("event-at") {
		attributes["event-at"] = opts.EventAt
	}
	if cmd.Flags().Changed("event-latitude") {
		attributes["event-latitude"] = opts.EventLatitude
	}
	if cmd.Flags().Changed("event-longitude") {
		attributes["event-longitude"] = opts.EventLongitude
	}
	if cmd.Flags().Changed("provenance") {
		attributes["provenance"] = opts.Provenance
	}

	relationships := map[string]any{}
	if cmd.Flags().Changed("user") {
		relationships["user"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.User,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes or relationships to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "user-location-events",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/user-location-events/"+opts.ID, jsonBody)
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

	row := buildUserLocationEventRow(resp.Data)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated user location event %s\n", row.ID)
	return nil
}

func parseDoUserLocationEventsUpdateOptions(cmd *cobra.Command, args []string) (doUserLocationEventsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	user, _ := cmd.Flags().GetString("user")
	eventAt, _ := cmd.Flags().GetString("event-at")
	eventLatitude, _ := cmd.Flags().GetString("event-latitude")
	eventLongitude, _ := cmd.Flags().GetString("event-longitude")
	provenance, _ := cmd.Flags().GetString("provenance")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doUserLocationEventsUpdateOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		ID:             args[0],
		User:           user,
		EventAt:        eventAt,
		EventLatitude:  eventLatitude,
		EventLongitude: eventLongitude,
		Provenance:     provenance,
	}, nil
}
