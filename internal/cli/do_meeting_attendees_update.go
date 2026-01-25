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

type doMeetingAttendeesUpdateOptions struct {
	BaseURL            string
	Token              string
	JSON               bool
	ID                 string
	LocationKind       string
	IsPresenceRequired bool
	IsPresent          bool
	LocationLatitude   string
	LocationLongitude  string
}

func newDoMeetingAttendeesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a meeting attendee",
		Long: `Update a meeting attendee.

Optional flags:
  --location-kind         Update location kind (on_site, remote)
  --is-presence-required  Update presence requirement (true/false)
  --is-present            Update present status (true/false)
  --location-latitude     Update location latitude
  --location-longitude    Update location longitude`,
		Example: `  # Update presence
  xbe do meeting-attendees update 123 --is-present true

  # Update location
  xbe do meeting-attendees update 123 --location-kind remote --location-latitude "41.0" --location-longitude "-87.0"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoMeetingAttendeesUpdate,
	}
	initDoMeetingAttendeesUpdateFlags(cmd)
	return cmd
}

func init() {
	doMeetingAttendeesCmd.AddCommand(newDoMeetingAttendeesUpdateCmd())
}

func initDoMeetingAttendeesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("location-kind", "", "Location kind (on_site, remote)")
	cmd.Flags().Bool("is-presence-required", false, "Presence required")
	cmd.Flags().Bool("is-present", false, "Present status")
	cmd.Flags().String("location-latitude", "", "Location latitude")
	cmd.Flags().String("location-longitude", "", "Location longitude")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMeetingAttendeesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoMeetingAttendeesUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("location-kind") {
		attributes["location-kind"] = opts.LocationKind
	}
	if cmd.Flags().Changed("is-presence-required") {
		attributes["is-presence-required"] = opts.IsPresenceRequired
	}
	if cmd.Flags().Changed("is-present") {
		attributes["is-present"] = opts.IsPresent
	}
	if cmd.Flags().Changed("location-latitude") {
		attributes["location-latitude"] = opts.LocationLatitude
	}
	if cmd.Flags().Changed("location-longitude") {
		attributes["location-longitude"] = opts.LocationLongitude
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "meeting-attendees",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/meeting-attendees/"+opts.ID, jsonBody)
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

	row := buildMeetingAttendeeRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated meeting attendee %s\n", row.ID)
	return nil
}

func parseDoMeetingAttendeesUpdateOptions(cmd *cobra.Command, args []string) (doMeetingAttendeesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	locationKind, _ := cmd.Flags().GetString("location-kind")
	isPresenceRequired, _ := cmd.Flags().GetBool("is-presence-required")
	isPresent, _ := cmd.Flags().GetBool("is-present")
	locationLatitude, _ := cmd.Flags().GetString("location-latitude")
	locationLongitude, _ := cmd.Flags().GetString("location-longitude")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMeetingAttendeesUpdateOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		ID:                 args[0],
		LocationKind:       locationKind,
		IsPresenceRequired: isPresenceRequired,
		IsPresent:          isPresent,
		LocationLatitude:   locationLatitude,
		LocationLongitude:  locationLongitude,
	}, nil
}
