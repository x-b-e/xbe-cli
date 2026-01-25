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

type userLocationEventsListOptions struct {
	BaseURL    string
	Token      string
	JSON       bool
	NoAuth     bool
	Limit      int
	Offset     int
	Sort       string
	User       string
	UpdatedBy  string
	EventAtMin string
	EventAtMax string
	IsEventAt  string
	Provenance string
}

type userLocationEventRow struct {
	ID             string `json:"id"`
	EventAt        string `json:"event_at,omitempty"`
	EventLatitude  string `json:"event_latitude,omitempty"`
	EventLongitude string `json:"event_longitude,omitempty"`
	Provenance     string `json:"provenance,omitempty"`
	UserID         string `json:"user_id,omitempty"`
	UpdatedByID    string `json:"updated_by_id,omitempty"`
}

func newUserLocationEventsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List user location events",
		Long: `List user location events.

Output Columns:
  ID          User location event identifier
  EVENT AT    Event timestamp
  LAT         Event latitude
  LONG        Event longitude
  PROVENANCE  Event provenance
  USER        User ID
  UPDATED BY  Updated-by user ID

Filters:
  --user          Filter by user ID
  --updated-by    Filter by updated-by user ID
  --event-at-min  Filter by event-at on/after (ISO 8601)
  --event-at-max  Filter by event-at on/before (ISO 8601)
  --is-event-at   Filter by whether event-at is set (true/false)
  --provenance    Filter by provenance (gps,map)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List user location events
  xbe view user-location-events list

  # Filter by user
  xbe view user-location-events list --user 123

  # Filter by provenance and time range
  xbe view user-location-events list --provenance gps --event-at-min 2025-01-01T00:00:00Z

  # JSON output
  xbe view user-location-events list --json`,
		Args: cobra.NoArgs,
		RunE: runUserLocationEventsList,
	}
	initUserLocationEventsListFlags(cmd)
	return cmd
}

func init() {
	userLocationEventsCmd.AddCommand(newUserLocationEventsListCmd())
}

func initUserLocationEventsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("user", "", "Filter by user ID")
	cmd.Flags().String("updated-by", "", "Filter by updated-by user ID")
	cmd.Flags().String("event-at-min", "", "Filter by event-at on/after (ISO 8601)")
	cmd.Flags().String("event-at-max", "", "Filter by event-at on/before (ISO 8601)")
	cmd.Flags().String("is-event-at", "", "Filter by whether event-at is set (true/false)")
	cmd.Flags().String("provenance", "", "Filter by provenance (gps,map)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runUserLocationEventsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseUserLocationEventsListOptions(cmd)
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
	query.Set("fields[user-location-events]", "event-at,event-latitude,event-longitude,provenance,user,updated-by")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[user]", opts.User)
	setFilterIfPresent(query, "filter[updated-by]", opts.UpdatedBy)
	setFilterIfPresent(query, "filter[event-at-min]", opts.EventAtMin)
	setFilterIfPresent(query, "filter[event-at-max]", opts.EventAtMax)
	setFilterIfPresent(query, "filter[is-event-at]", opts.IsEventAt)
	setFilterIfPresent(query, "filter[provenance]", opts.Provenance)

	body, _, err := client.Get(cmd.Context(), "/v1/user-location-events", query)
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

	rows := buildUserLocationEventRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderUserLocationEventsTable(cmd, rows)
}

func parseUserLocationEventsListOptions(cmd *cobra.Command) (userLocationEventsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	user, _ := cmd.Flags().GetString("user")
	updatedBy, _ := cmd.Flags().GetString("updated-by")
	eventAtMin, _ := cmd.Flags().GetString("event-at-min")
	eventAtMax, _ := cmd.Flags().GetString("event-at-max")
	isEventAt, _ := cmd.Flags().GetString("is-event-at")
	provenance, _ := cmd.Flags().GetString("provenance")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return userLocationEventsListOptions{
		BaseURL:    baseURL,
		Token:      token,
		JSON:       jsonOut,
		NoAuth:     noAuth,
		Limit:      limit,
		Offset:     offset,
		Sort:       sort,
		User:       user,
		UpdatedBy:  updatedBy,
		EventAtMin: eventAtMin,
		EventAtMax: eventAtMax,
		IsEventAt:  isEventAt,
		Provenance: provenance,
	}, nil
}

func buildUserLocationEventRows(resp jsonAPIResponse) []userLocationEventRow {
	rows := make([]userLocationEventRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildUserLocationEventRow(resource))
	}
	return rows
}

func buildUserLocationEventRow(resource jsonAPIResource) userLocationEventRow {
	row := userLocationEventRow{
		ID:             resource.ID,
		EventAt:        formatDateTime(stringAttr(resource.Attributes, "event-at")),
		EventLatitude:  stringAttr(resource.Attributes, "event-latitude"),
		EventLongitude: stringAttr(resource.Attributes, "event-longitude"),
		Provenance:     stringAttr(resource.Attributes, "provenance"),
	}

	if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
		row.UserID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["updated-by"]; ok && rel.Data != nil {
		row.UpdatedByID = rel.Data.ID
	}

	return row
}

func renderUserLocationEventsTable(cmd *cobra.Command, rows []userLocationEventRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No user location events found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, 32, 0)
	fmt.Fprintln(writer, "ID\tEVENT AT\tLAT\tLONG\tPROVENANCE\tUSER\tUPDATED BY")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.EventAt,
			row.EventLatitude,
			row.EventLongitude,
			row.Provenance,
			row.UserID,
			row.UpdatedByID,
		)
	}
	return writer.Flush()
}
