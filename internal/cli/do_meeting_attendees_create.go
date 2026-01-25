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

type doMeetingAttendeesCreateOptions struct {
	BaseURL            string
	Token              string
	JSON               bool
	Meeting            string
	User               string
	LocationKind       string
	IsPresenceRequired bool
	IsPresent          bool
	LocationLatitude   string
	LocationLongitude  string
}

func newDoMeetingAttendeesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a meeting attendee",
		Long: `Create a meeting attendee.

Required flags:
  --meeting        Meeting ID (required)
  --user           User ID (required)
  --location-kind  Location kind (on_site, remote) (required)

Optional flags:
  --is-presence-required  Presence required (true/false)
  --is-present            Present status (true/false)
  --location-latitude     Location latitude
  --location-longitude    Location longitude`,
		Example: `  # Create a meeting attendee
  xbe do meeting-attendees create \
    --meeting 123 \
    --user 456 \
    --location-kind on_site

  # Create and mark present
  xbe do meeting-attendees create \
    --meeting 123 \
    --user 456 \
    --location-kind remote \
    --is-present true \
    --is-presence-required false

  # JSON output
  xbe do meeting-attendees create \
    --meeting 123 \
    --user 456 \
    --location-kind on_site \
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoMeetingAttendeesCreate,
	}
	initDoMeetingAttendeesCreateFlags(cmd)
	return cmd
}

func init() {
	doMeetingAttendeesCmd.AddCommand(newDoMeetingAttendeesCreateCmd())
}

func initDoMeetingAttendeesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("meeting", "", "Meeting ID (required)")
	cmd.Flags().String("user", "", "User ID (required)")
	cmd.Flags().String("location-kind", "", "Location kind (on_site, remote) (required)")
	cmd.Flags().Bool("is-presence-required", false, "Presence required")
	cmd.Flags().Bool("is-present", false, "Present status")
	cmd.Flags().String("location-latitude", "", "Location latitude")
	cmd.Flags().String("location-longitude", "", "Location longitude")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMeetingAttendeesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMeetingAttendeesCreateOptions(cmd)
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

	if strings.TrimSpace(opts.Meeting) == "" {
		err := fmt.Errorf("--meeting is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.User) == "" {
		err := fmt.Errorf("--user is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.LocationKind) == "" {
		err := fmt.Errorf("--location-kind is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"location-kind": opts.LocationKind,
	}
	if cmd.Flags().Changed("is-presence-required") {
		attributes["is-presence-required"] = opts.IsPresenceRequired
	}
	if cmd.Flags().Changed("is-present") {
		attributes["is-present"] = opts.IsPresent
	}
	if opts.LocationLatitude != "" {
		attributes["location-latitude"] = opts.LocationLatitude
	}
	if opts.LocationLongitude != "" {
		attributes["location-longitude"] = opts.LocationLongitude
	}

	relationships := map[string]any{
		"meeting": map[string]any{
			"data": map[string]any{
				"type": "meetings",
				"id":   opts.Meeting,
			},
		},
		"user": map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.User,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "meeting-attendees",
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

	body, _, err := client.Post(cmd.Context(), "/v1/meeting-attendees", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created meeting attendee %s\n", row.ID)
	return nil
}

func parseDoMeetingAttendeesCreateOptions(cmd *cobra.Command) (doMeetingAttendeesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	meeting, _ := cmd.Flags().GetString("meeting")
	user, _ := cmd.Flags().GetString("user")
	locationKind, _ := cmd.Flags().GetString("location-kind")
	isPresenceRequired, _ := cmd.Flags().GetBool("is-presence-required")
	isPresent, _ := cmd.Flags().GetBool("is-present")
	locationLatitude, _ := cmd.Flags().GetString("location-latitude")
	locationLongitude, _ := cmd.Flags().GetString("location-longitude")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMeetingAttendeesCreateOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		Meeting:            meeting,
		User:               user,
		LocationKind:       locationKind,
		IsPresenceRequired: isPresenceRequired,
		IsPresent:          isPresent,
		LocationLatitude:   locationLatitude,
		LocationLongitude:  locationLongitude,
	}, nil
}

func buildMeetingAttendeeRowFromSingle(resp jsonAPISingleResponse) meetingAttendeeRow {
	row := meetingAttendeeRow{
		ID:                 resp.Data.ID,
		LocationKind:       stringAttr(resp.Data.Attributes, "location-kind"),
		IsPresenceRequired: boolAttr(resp.Data.Attributes, "is-presence-required"),
		IsPresent:          boolAttr(resp.Data.Attributes, "is-present"),
		UserName:           strings.TrimSpace(stringAttr(resp.Data.Attributes, "user-name")),
	}

	if rel, ok := resp.Data.Relationships["meeting"]; ok && rel.Data != nil {
		row.MeetingID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["user"]; ok && rel.Data != nil {
		row.UserID = rel.Data.ID
	}

	return row
}
