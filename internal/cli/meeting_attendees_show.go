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

type meetingAttendeesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type meetingAttendeeDetails struct {
	ID                   string `json:"id"`
	MeetingID            string `json:"meeting_id,omitempty"`
	UserID               string `json:"user_id,omitempty"`
	UserName             string `json:"user_name,omitempty"`
	LocationKind         string `json:"location_kind,omitempty"`
	IsPresenceRequired   bool   `json:"is_presence_required"`
	IsPresent            bool   `json:"is_present"`
	LocationLatitude     string `json:"location_latitude,omitempty"`
	LocationLongitude    string `json:"location_longitude,omitempty"`
	LocationAt           string `json:"location_at,omitempty"`
	IsMeetingOrganizer   bool   `json:"is_meeting_organizer"`
	IsSelfCheckIn        bool   `json:"is_self_check_in"`
	CreatedByType        string `json:"created_by_type,omitempty"`
	CreatedByID          string `json:"created_by_id,omitempty"`
	CurrentUserCanUpdate bool   `json:"current_user_can_update"`
}

func newMeetingAttendeesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show meeting attendee details",
		Long: `Show the full details of a meeting attendee.

Output Fields:
  ID                   Meeting attendee identifier
  Meeting              Meeting ID
  User                 User name/ID
  Location Kind        Location kind (on_site, remote)
  Presence Required    Whether presence is required
  Present              Whether attendee is present
  Location Latitude    Latitude coordinate
  Location Longitude   Longitude coordinate
  Location At          Timestamp when location was updated
  Is Meeting Organizer Whether attendee is the meeting organizer
  Self Check-In        Whether attendee checked themselves in
  Created By           Creator (type/id)
  Current User Can Update  Whether current user can update this record

Arguments:
  <id>    Meeting attendee ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a meeting attendee
  xbe view meeting-attendees show 123

  # JSON output
  xbe view meeting-attendees show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runMeetingAttendeesShow,
	}
	initMeetingAttendeesShowFlags(cmd)
	return cmd
}

func init() {
	meetingAttendeesCmd.AddCommand(newMeetingAttendeesShowCmd())
}

func initMeetingAttendeesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMeetingAttendeesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseMeetingAttendeesShowOptions(cmd)
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
		return fmt.Errorf("meeting attendee id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[meeting-attendees]", "meeting,user,created-by,location-kind,is-presence-required,is-present,location-latitude,location-longitude,location-at,is-meeting-organizer,is-self-check-in,user-name,current-user-can-update")

	body, _, err := client.Get(cmd.Context(), "/v1/meeting-attendees/"+id, query)
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

	details := buildMeetingAttendeeDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderMeetingAttendeeDetails(cmd, details)
}

func parseMeetingAttendeesShowOptions(cmd *cobra.Command) (meetingAttendeesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return meetingAttendeesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildMeetingAttendeeDetails(resp jsonAPISingleResponse) meetingAttendeeDetails {
	attrs := resp.Data.Attributes

	details := meetingAttendeeDetails{
		ID:                   resp.Data.ID,
		LocationKind:         stringAttr(attrs, "location-kind"),
		IsPresenceRequired:   boolAttr(attrs, "is-presence-required"),
		IsPresent:            boolAttr(attrs, "is-present"),
		LocationLatitude:     stringAttr(attrs, "location-latitude"),
		LocationLongitude:    stringAttr(attrs, "location-longitude"),
		LocationAt:           formatDateTime(stringAttr(attrs, "location-at")),
		IsMeetingOrganizer:   boolAttr(attrs, "is-meeting-organizer"),
		IsSelfCheckIn:        boolAttr(attrs, "is-self-check-in"),
		UserName:             strings.TrimSpace(stringAttr(attrs, "user-name")),
		CurrentUserCanUpdate: boolAttr(attrs, "current-user-can-update"),
	}

	if rel, ok := resp.Data.Relationships["meeting"]; ok && rel.Data != nil {
		details.MeetingID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["user"]; ok && rel.Data != nil {
		details.UserID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByType = rel.Data.Type
		details.CreatedByID = rel.Data.ID
	}

	return details
}

func renderMeetingAttendeeDetails(cmd *cobra.Command, details meetingAttendeeDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	writeLabelWithID(out, "Meeting", "", details.MeetingID)
	writeLabelWithID(out, "User", details.UserName, details.UserID)
	if details.LocationKind != "" {
		fmt.Fprintf(out, "Location Kind: %s\n", details.LocationKind)
	}
	fmt.Fprintf(out, "Presence Required: %s\n", yesNo(details.IsPresenceRequired))
	fmt.Fprintf(out, "Present: %s\n", yesNo(details.IsPresent))
	if details.LocationLatitude != "" {
		fmt.Fprintf(out, "Location Latitude: %s\n", details.LocationLatitude)
	}
	if details.LocationLongitude != "" {
		fmt.Fprintf(out, "Location Longitude: %s\n", details.LocationLongitude)
	}
	if details.LocationAt != "" {
		fmt.Fprintf(out, "Location At: %s\n", details.LocationAt)
	}
	fmt.Fprintf(out, "Is Meeting Organizer: %s\n", yesNo(details.IsMeetingOrganizer))
	fmt.Fprintf(out, "Self Check-In: %s\n", yesNo(details.IsSelfCheckIn))
	if details.CreatedByType != "" && details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By: %s/%s\n", details.CreatedByType, details.CreatedByID)
	} else if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By: %s\n", details.CreatedByID)
	}
	fmt.Fprintf(out, "Current User Can Update: %s\n", yesNo(details.CurrentUserCanUpdate))

	return nil
}
