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

type userLocationEstimatesListOptions struct {
	BaseURL              string
	Token                string
	JSON                 bool
	NoAuth               bool
	Limit                int
	Offset               int
	Sort                 string
	User                 string
	AsOf                 string
	EarliestEventAt      string
	LatestEventAt        string
	MaxAbsLatencySeconds string
	MaxLatestSeconds     string
}

type userLocationEstimateRow struct {
	ID                       string   `json:"id"`
	UserID                   string   `json:"user_id,omitempty"`
	AsOf                     string   `json:"as_of,omitempty"`
	LastKnownAt              string   `json:"last_known_at,omitempty"`
	LastKnownLatitude        *float64 `json:"last_known_latitude,omitempty"`
	LastKnownLongitude       *float64 `json:"last_known_longitude,omitempty"`
	LastKnownHeading         *float64 `json:"last_known_heading,omitempty"`
	LastKnownLatencySeconds  *int     `json:"last_known_latency_seconds,omitempty"`
	LastKnownTimeZoneID      string   `json:"last_known_time_zone_id,omitempty"`
	LastKnownSourceClassName string   `json:"last_known_source_class_name,omitempty"`
}

func newUserLocationEstimatesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List user location estimates",
		Long: `List user location estimates for one or more users.

User location estimates derive the most recent known location for a user based
on device, vehicle, and activity events. A user filter is required.

Output Columns:
  ID            Estimate identifier
  USER          User ID
  AS OF         Reference timestamp used for estimation
  LAST KNOWN    Timestamp of the last known location
  LAT           Estimated latitude
  LON           Estimated longitude
  LATENCY       Age of last known event in seconds
  SOURCE        Source class name for the last known event

Filters:
  --user                      Filter by user ID (required)
  --as-of                     Reference time for estimation (ISO8601)
  --earliest-event-at          Override earliest event timestamp (ISO8601)
  --latest-event-at            Override latest event timestamp (ISO8601)
  --max-abs-latency-seconds    Override maximum absolute latency window (seconds)
  --max-latest-seconds         Override maximum latest event window (seconds)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # Estimate a user's location
  xbe view user-location-estimates list --user 123

  # Use a custom as-of time
  xbe view user-location-estimates list --user 123 --as-of 2025-01-01T12:00:00Z

  # Override the event window
  xbe view user-location-estimates list --user 123 \
    --earliest-event-at 2025-01-01T00:00:00Z \
    --latest-event-at 2025-01-02T00:00:00Z

  # Output as JSON
  xbe view user-location-estimates list --user 123 --json`,
		Args: cobra.NoArgs,
		RunE: runUserLocationEstimatesList,
	}
	initUserLocationEstimatesListFlags(cmd)
	return cmd
}

func init() {
	userLocationEstimatesCmd.AddCommand(newUserLocationEstimatesListCmd())
}

func initUserLocationEstimatesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("user", "", "Filter by user ID (required)")
	cmd.Flags().String("as-of", "", "Reference time for estimation (ISO8601)")
	cmd.Flags().String("earliest-event-at", "", "Override earliest event timestamp (ISO8601)")
	cmd.Flags().String("latest-event-at", "", "Override latest event timestamp (ISO8601)")
	cmd.Flags().String("max-abs-latency-seconds", "", "Override maximum absolute latency window (seconds)")
	cmd.Flags().String("max-latest-seconds", "", "Override maximum latest event window (seconds)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runUserLocationEstimatesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseUserLocationEstimatesListOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.User) == "" {
		err := fmt.Errorf("--user is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[user-location-estimates]", "as-of,last-known-latitude,last-known-longitude,last-known-heading,last-known-time-zone-id,last-known-at,last-known-latency-seconds,last-known-source-class-name")
	query.Set("filter[user]", opts.User)

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[as-of]", opts.AsOf)
	setFilterIfPresent(query, "filter[earliest-event-at]", opts.EarliestEventAt)
	setFilterIfPresent(query, "filter[latest-event-at]", opts.LatestEventAt)
	setFilterIfPresent(query, "filter[max-abs-latency-seconds]", opts.MaxAbsLatencySeconds)
	setFilterIfPresent(query, "filter[max-latest-seconds]", opts.MaxLatestSeconds)

	body, _, err := client.Get(cmd.Context(), "/v1/user-location-estimates", query)
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

	rows := buildUserLocationEstimateRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderUserLocationEstimatesTable(cmd, rows)
}

func parseUserLocationEstimatesListOptions(cmd *cobra.Command) (userLocationEstimatesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	user, _ := cmd.Flags().GetString("user")
	asOf, _ := cmd.Flags().GetString("as-of")
	earliestEventAt, _ := cmd.Flags().GetString("earliest-event-at")
	latestEventAt, _ := cmd.Flags().GetString("latest-event-at")
	maxAbsLatencySeconds, _ := cmd.Flags().GetString("max-abs-latency-seconds")
	maxLatestSeconds, _ := cmd.Flags().GetString("max-latest-seconds")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return userLocationEstimatesListOptions{
		BaseURL:              baseURL,
		Token:                token,
		JSON:                 jsonOut,
		NoAuth:               noAuth,
		Limit:                limit,
		Offset:               offset,
		Sort:                 sort,
		User:                 user,
		AsOf:                 asOf,
		EarliestEventAt:      earliestEventAt,
		LatestEventAt:        latestEventAt,
		MaxAbsLatencySeconds: maxAbsLatencySeconds,
		MaxLatestSeconds:     maxLatestSeconds,
	}, nil
}

func buildUserLocationEstimateRows(resp jsonAPIResponse) []userLocationEstimateRow {
	rows := make([]userLocationEstimateRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := userLocationEstimateRow{
			ID:                       resource.ID,
			AsOf:                     formatDateTime(stringAttr(attrs, "as-of")),
			LastKnownAt:              formatDateTime(stringAttr(attrs, "last-known-at")),
			LastKnownLatitude:        floatAttrPointer(attrs, "last-known-latitude"),
			LastKnownLongitude:       floatAttrPointer(attrs, "last-known-longitude"),
			LastKnownHeading:         floatAttrPointer(attrs, "last-known-heading"),
			LastKnownLatencySeconds:  intAttrPointer(attrs, "last-known-latency-seconds"),
			LastKnownTimeZoneID:      stringAttr(attrs, "last-known-time-zone-id"),
			LastKnownSourceClassName: stringAttr(attrs, "last-known-source-class-name"),
		}

		if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
			row.UserID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderUserLocationEstimatesTable(cmd *cobra.Command, rows []userLocationEstimateRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No user location estimates found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tUSER\tAS OF\tLAST KNOWN AT\tLAT\tLON\tLATENCY\tSOURCE")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.UserID,
			row.AsOf,
			row.LastKnownAt,
			formatFloatValue(row.LastKnownLatitude, 6),
			formatFloatValue(row.LastKnownLongitude, 6),
			formatIntPointer(row.LastKnownLatencySeconds),
			row.LastKnownSourceClassName,
		)
	}
	return writer.Flush()
}

func formatFloatValue(value *float64, decimals int) string {
	if value == nil {
		return ""
	}
	format := fmt.Sprintf("%%.%df", decimals)
	return fmt.Sprintf(format, *value)
}

func formatIntPointer(value *int) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%d", *value)
}
