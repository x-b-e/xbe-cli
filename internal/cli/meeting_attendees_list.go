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

type meetingAttendeesListOptions struct {
	BaseURL            string
	Token              string
	JSON               bool
	NoAuth             bool
	Limit              int
	Offset             int
	Sort               string
	Meeting            string
	User               string
	LocationKind       string
	IsPresenceRequired string
	IsPresent          string
}

type meetingAttendeeRow struct {
	ID                 string `json:"id"`
	MeetingID          string `json:"meeting_id,omitempty"`
	UserID             string `json:"user_id,omitempty"`
	UserName           string `json:"user_name,omitempty"`
	LocationKind       string `json:"location_kind,omitempty"`
	IsPresenceRequired bool   `json:"is_presence_required"`
	IsPresent          bool   `json:"is_present"`
}

func newMeetingAttendeesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List meeting attendees",
		Long: `List meeting attendees with filtering and pagination.

Output Columns:
  ID        Meeting attendee identifier
  MEETING   Meeting ID
  USER      User name (falls back to ID)
  LOCATION  Location kind (on_site, remote)
  REQUIRED  Presence required
  PRESENT   Present status

Filters:
  --meeting              Filter by meeting ID
  --user                 Filter by user ID
  --location-kind        Filter by location kind (on_site, remote)
  --is-presence-required Filter by presence requirement (true/false)
  --is-present           Filter by present status (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List meeting attendees
  xbe view meeting-attendees list

  # Filter by meeting
  xbe view meeting-attendees list --meeting 123

  # Filter by user
  xbe view meeting-attendees list --user 456

  # Filter by location kind
  xbe view meeting-attendees list --location-kind on_site

  # JSON output
  xbe view meeting-attendees list --json`,
		Args: cobra.NoArgs,
		RunE: runMeetingAttendeesList,
	}
	initMeetingAttendeesListFlags(cmd)
	return cmd
}

func init() {
	meetingAttendeesCmd.AddCommand(newMeetingAttendeesListCmd())
}

func initMeetingAttendeesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("meeting", "", "Filter by meeting ID")
	cmd.Flags().String("user", "", "Filter by user ID")
	cmd.Flags().String("location-kind", "", "Filter by location kind (on_site, remote)")
	cmd.Flags().String("is-presence-required", "", "Filter by presence requirement (true/false)")
	cmd.Flags().String("is-present", "", "Filter by present status (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMeetingAttendeesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMeetingAttendeesListOptions(cmd)
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
	query.Set("fields[meeting-attendees]", "meeting,user,location-kind,is-presence-required,is-present,user-name")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[meeting]", opts.Meeting)
	setFilterIfPresent(query, "filter[user]", opts.User)
	setFilterIfPresent(query, "filter[location-kind]", opts.LocationKind)
	setFilterIfPresent(query, "filter[is-presence-required]", opts.IsPresenceRequired)
	setFilterIfPresent(query, "filter[is-present]", opts.IsPresent)

	body, _, err := client.Get(cmd.Context(), "/v1/meeting-attendees", query)
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

	rows := buildMeetingAttendeeRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderMeetingAttendeesTable(cmd, rows)
}

func parseMeetingAttendeesListOptions(cmd *cobra.Command) (meetingAttendeesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	meeting, _ := cmd.Flags().GetString("meeting")
	user, _ := cmd.Flags().GetString("user")
	locationKind, _ := cmd.Flags().GetString("location-kind")
	isPresenceRequired, _ := cmd.Flags().GetString("is-presence-required")
	isPresent, _ := cmd.Flags().GetString("is-present")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return meetingAttendeesListOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		NoAuth:             noAuth,
		Limit:              limit,
		Offset:             offset,
		Sort:               sort,
		Meeting:            meeting,
		User:               user,
		LocationKind:       locationKind,
		IsPresenceRequired: isPresenceRequired,
		IsPresent:          isPresent,
	}, nil
}

func buildMeetingAttendeeRows(resp jsonAPIResponse) []meetingAttendeeRow {
	rows := make([]meetingAttendeeRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := buildMeetingAttendeeRow(resource)
		rows = append(rows, row)
	}
	return rows
}

func buildMeetingAttendeeRow(resource jsonAPIResource) meetingAttendeeRow {
	row := meetingAttendeeRow{
		ID:                 resource.ID,
		LocationKind:       stringAttr(resource.Attributes, "location-kind"),
		IsPresenceRequired: boolAttr(resource.Attributes, "is-presence-required"),
		IsPresent:          boolAttr(resource.Attributes, "is-present"),
		UserName:           strings.TrimSpace(stringAttr(resource.Attributes, "user-name")),
	}

	if rel, ok := resource.Relationships["meeting"]; ok && rel.Data != nil {
		row.MeetingID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
		row.UserID = rel.Data.ID
	}

	return row
}

func renderMeetingAttendeesTable(cmd *cobra.Command, rows []meetingAttendeeRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No meeting attendees found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tMEETING\tUSER\tLOCATION\tREQUIRED\tPRESENT")
	for _, row := range rows {
		userDisplay := firstNonEmpty(row.UserName, row.UserID)
		required := "No"
		if row.IsPresenceRequired {
			required = "Yes"
		}
		present := "No"
		if row.IsPresent {
			present = "Yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.MeetingID, 20),
			truncateString(userDisplay, 25),
			truncateString(row.LocationKind, 12),
			required,
			present,
		)
	}
	return writer.Flush()
}
