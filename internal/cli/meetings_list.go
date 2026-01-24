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

type meetingsListOptions struct {
	BaseURL         string
	Token           string
	JSON            bool
	NoAuth          bool
	Limit           int
	Offset          int
	Sort            string
	Organization    string
	Organizer       string
	Broker          string
	Attendee        string
	StartAtMin      string
	StartAtMax      string
	EndAtMin        string
	EndAtMax        string
	IsStartAt       string
	IsEndAt         string
	IsSafetyMeeting string
	CreatedAtMin    string
	CreatedAtMax    string
	UpdatedAtMin    string
	UpdatedAtMax    string
	IsCreatedAt     string
	IsUpdatedAt     string
}

type meetingRow struct {
	ID               string `json:"id"`
	Subject          string `json:"subject,omitempty"`
	StartAt          string `json:"start_at,omitempty"`
	EndAt            string `json:"end_at,omitempty"`
	OrganizationType string `json:"organization_type,omitempty"`
	OrganizationID   string `json:"organization_id,omitempty"`
	OrganizationName string `json:"organization_name,omitempty"`
	OrganizerID      string `json:"organizer_id,omitempty"`
	OrganizerName    string `json:"organizer_name,omitempty"`
}

func newMeetingsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List meetings",
		Long: `List meetings with filtering and pagination.

Output Columns:
  ID            Meeting identifier
  SUBJECT       Meeting subject
  START AT      Start timestamp
  END AT        End timestamp
  ORGANIZATION  Organization name (or type/id)
  ORGANIZER     Organizer name (or ID)

Filters:
  --organization       Filter by organization (Type|ID, e.g. Broker|123)
  --organizer          Filter by organizer user ID
  --broker             Filter by broker ID
  --attendee           Filter by attendee user ID
  --start-at-min        Filter by start-at on/after (ISO 8601)
  --start-at-max        Filter by start-at on/before (ISO 8601)
  --end-at-min          Filter by end-at on/after (ISO 8601)
  --end-at-max          Filter by end-at on/before (ISO 8601)
  --is-start-at         Filter by presence of start-at (true/false)
  --is-end-at           Filter by presence of end-at (true/false)
  --is-safety-meeting   Filter safety meetings (true/false)
  --created-at-min      Filter by created-at on/after (ISO 8601)
  --created-at-max      Filter by created-at on/before (ISO 8601)
  --updated-at-min      Filter by updated-at on/after (ISO 8601)
  --updated-at-max      Filter by updated-at on/before (ISO 8601)
  --is-created-at       Filter by presence of created-at (true/false)
  --is-updated-at       Filter by presence of updated-at (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List meetings
  xbe view meetings list

  # Filter by organization
  xbe view meetings list --organization "Broker|123"

  # Filter by organizer
  xbe view meetings list --organizer 456

  # Filter by time range
  xbe view meetings list --start-at-min 2025-01-01T00:00:00Z --end-at-max 2025-01-31T23:59:59Z

  # Output as JSON
  xbe view meetings list --json`,
		Args: cobra.NoArgs,
		RunE: runMeetingsList,
	}
	initMeetingsListFlags(cmd)
	return cmd
}

func init() {
	meetingsCmd.AddCommand(newMeetingsListCmd())
}

func initMeetingsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort order")
	cmd.Flags().String("organization", "", "Filter by organization (Type|ID, e.g. Broker|123)")
	cmd.Flags().String("organizer", "", "Filter by organizer user ID")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("attendee", "", "Filter by attendee user ID")
	cmd.Flags().String("start-at-min", "", "Filter by start-at on/after (ISO 8601)")
	cmd.Flags().String("start-at-max", "", "Filter by start-at on/before (ISO 8601)")
	cmd.Flags().String("end-at-min", "", "Filter by end-at on/after (ISO 8601)")
	cmd.Flags().String("end-at-max", "", "Filter by end-at on/before (ISO 8601)")
	cmd.Flags().String("is-start-at", "", "Filter by presence of start-at (true/false)")
	cmd.Flags().String("is-end-at", "", "Filter by presence of end-at (true/false)")
	cmd.Flags().String("is-safety-meeting", "", "Filter safety meetings (true/false)")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by presence of created-at (true/false)")
	cmd.Flags().String("is-updated-at", "", "Filter by presence of updated-at (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMeetingsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMeetingsListOptions(cmd)
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
	query.Set("fields[meetings]", "subject,start-at,end-at,organization,organizer")
	query.Set("include", "organization,organizer")
	query.Set("fields[users]", "name")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[truckers]", "company-name")
	query.Set("fields[material-suppliers]", "name")
	query.Set("fields[developers]", "name")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[organization]", opts.Organization)
	setFilterIfPresent(query, "filter[organizer]", opts.Organizer)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[attendee]", opts.Attendee)
	setFilterIfPresent(query, "filter[start_at_min]", opts.StartAtMin)
	setFilterIfPresent(query, "filter[start_at_max]", opts.StartAtMax)
	setFilterIfPresent(query, "filter[end_at_min]", opts.EndAtMin)
	setFilterIfPresent(query, "filter[end_at_max]", opts.EndAtMax)
	setFilterIfPresent(query, "filter[is_start_at]", opts.IsStartAt)
	setFilterIfPresent(query, "filter[is_end_at]", opts.IsEndAt)
	setFilterIfPresent(query, "filter[is_safety_meeting]", opts.IsSafetyMeeting)
	setFilterIfPresent(query, "filter[created_at_min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created_at_max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated_at_min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated_at_max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[is_created_at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[is_updated_at]", opts.IsUpdatedAt)

	body, _, err := client.Get(cmd.Context(), "/v1/meetings", query)
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

	rows := buildMeetingRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderMeetingsTable(cmd, rows)
}

func parseMeetingsListOptions(cmd *cobra.Command) (meetingsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	organization, _ := cmd.Flags().GetString("organization")
	organizer, _ := cmd.Flags().GetString("organizer")
	broker, _ := cmd.Flags().GetString("broker")
	attendee, _ := cmd.Flags().GetString("attendee")
	startAtMin, _ := cmd.Flags().GetString("start-at-min")
	startAtMax, _ := cmd.Flags().GetString("start-at-max")
	endAtMin, _ := cmd.Flags().GetString("end-at-min")
	endAtMax, _ := cmd.Flags().GetString("end-at-max")
	isStartAt, _ := cmd.Flags().GetString("is-start-at")
	isEndAt, _ := cmd.Flags().GetString("is-end-at")
	isSafetyMeeting, _ := cmd.Flags().GetString("is-safety-meeting")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	isCreatedAt, _ := cmd.Flags().GetString("is-created-at")
	isUpdatedAt, _ := cmd.Flags().GetString("is-updated-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return meetingsListOptions{
		BaseURL:         baseURL,
		Token:           token,
		JSON:            jsonOut,
		NoAuth:          noAuth,
		Limit:           limit,
		Offset:          offset,
		Sort:            sort,
		Organization:    organization,
		Organizer:       organizer,
		Broker:          broker,
		Attendee:        attendee,
		StartAtMin:      startAtMin,
		StartAtMax:      startAtMax,
		EndAtMin:        endAtMin,
		EndAtMax:        endAtMax,
		IsStartAt:       isStartAt,
		IsEndAt:         isEndAt,
		IsSafetyMeeting: isSafetyMeeting,
		CreatedAtMin:    createdAtMin,
		CreatedAtMax:    createdAtMax,
		UpdatedAtMin:    updatedAtMin,
		UpdatedAtMax:    updatedAtMax,
		IsCreatedAt:     isCreatedAt,
		IsUpdatedAt:     isUpdatedAt,
	}, nil
}

func buildMeetingRows(resp jsonAPIResponse) []meetingRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]meetingRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := meetingRow{
			ID:      resource.ID,
			Subject: strings.TrimSpace(stringAttr(resource.Attributes, "subject")),
			StartAt: formatDateTime(stringAttr(resource.Attributes, "start-at")),
			EndAt:   formatDateTime(stringAttr(resource.Attributes, "end-at")),
		}

		if rel, ok := resource.Relationships["organization"]; ok && rel.Data != nil {
			row.OrganizationType = rel.Data.Type
			row.OrganizationID = rel.Data.ID
			if org, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.OrganizationName = firstNonEmpty(
					stringAttr(org.Attributes, "company-name"),
					stringAttr(org.Attributes, "name"),
				)
			}
		}

		if rel, ok := resource.Relationships["organizer"]; ok && rel.Data != nil {
			row.OrganizerID = rel.Data.ID
			if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.OrganizerName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
			}
		}

		rows = append(rows, row)
	}
	return rows
}

func renderMeetingsTable(cmd *cobra.Command, rows []meetingRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No meetings found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSUBJECT\tSTART\tEND\tORGANIZATION\tORGANIZER")
	for _, row := range rows {
		orgLabel := row.OrganizationName
		if orgLabel == "" {
			if row.OrganizationID != "" {
				orgType := strings.TrimSuffix(row.OrganizationType, "s")
				if orgType != "" {
					orgType = strings.ToUpper(orgType[:1]) + orgType[1:]
				}
				orgLabel = fmt.Sprintf("%s:%s", orgType, row.OrganizationID)
			}
		}

		organizer := row.OrganizerName
		if organizer == "" {
			organizer = row.OrganizerID
		}

		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Subject, 30),
			truncateString(row.StartAt, 20),
			truncateString(row.EndAt, 20),
			truncateString(orgLabel, 25),
			truncateString(organizer, 20),
		)
	}

	return writer.Flush()
}
