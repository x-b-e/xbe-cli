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

type tripsListOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	NoAuth                 bool
	Limit                  int
	Offset                 int
	Tender                 string
	TenderJobScheduleShift string
	DriverDay              string
	UnloadedSubjectType    string
	UnloadedSubjectID      string
	MaterialSites          string
	JobSites               string
	ParkingSites           string
}

type tripRow struct {
	ID               string `json:"id"`
	OriginAt         string `json:"origin_at,omitempty"`
	OriginNotes      string `json:"origin_notes,omitempty"`
	DestinationAt    string `json:"destination_at,omitempty"`
	DestinationNotes string `json:"destination_notes,omitempty"`
	Mileage          string `json:"mileage,omitempty"`
	Minutes          string `json:"minutes,omitempty"`
	IsUnloaded       bool   `json:"is_unloaded"`
	OriginType       string `json:"origin_type,omitempty"`
	OriginID         string `json:"origin_id,omitempty"`
	DestinationType  string `json:"destination_type,omitempty"`
	DestinationID    string `json:"destination_id,omitempty"`
	SourceType       string `json:"source_type,omitempty"`
	SourceID         string `json:"source_id,omitempty"`
	DriverDayID      string `json:"driver_day_id,omitempty"`
}

func newTripsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List trips",
		Long: `List trips.

Output Columns:
  ID              Trip identifier
  ORIGIN          Origin type and time
  DESTINATION     Destination type and time
  MILEAGE         Trip mileage
  MINUTES         Trip duration in minutes

Filters:
  --tender                    Filter by tender ID
  --tender-job-schedule-shift Filter by tender job schedule shift ID
  --driver-day                Filter by driver day ID
  --unloaded-subject-type     Filter by unloaded subject type
  --unloaded-subject-id       Filter by unloaded subject ID
  --material-sites            Filter by material site IDs (comma-separated)
  --job-sites                 Filter by job site IDs (comma-separated)
  --parking-sites             Filter by parking site IDs (comma-separated)`,
		Example: `  # List all trips
  xbe view trips list

  # Filter by tender
  xbe view trips list --tender 123

  # Filter by driver day
  xbe view trips list --driver-day 456

  # Filter by job sites
  xbe view trips list --job-sites 789,101

  # Output as JSON
  xbe view trips list --json`,
		RunE: runTripsList,
	}
	initTripsListFlags(cmd)
	return cmd
}

func init() {
	tripsCmd.AddCommand(newTripsListCmd())
}

func initTripsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("tender", "", "Filter by tender ID")
	cmd.Flags().String("tender-job-schedule-shift", "", "Filter by tender job schedule shift ID")
	cmd.Flags().String("driver-day", "", "Filter by driver day ID")
	cmd.Flags().String("unloaded-subject-type", "", "Filter by unloaded subject type")
	cmd.Flags().String("unloaded-subject-id", "", "Filter by unloaded subject ID")
	cmd.Flags().String("material-sites", "", "Filter by material site IDs (comma-separated)")
	cmd.Flags().String("job-sites", "", "Filter by job site IDs (comma-separated)")
	cmd.Flags().String("parking-sites", "", "Filter by parking site IDs (comma-separated)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTripsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTripsListOptions(cmd)
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
	query.Set("include", "origin,destination,driver-day")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	setFilterIfPresent(query, "filter[tender]", opts.Tender)
	setFilterIfPresent(query, "filter[tender_job_schedule_shift]", opts.TenderJobScheduleShift)
	setFilterIfPresent(query, "filter[driver_day]", opts.DriverDay)
	setFilterIfPresent(query, "filter[material_sites]", opts.MaterialSites)
	setFilterIfPresent(query, "filter[job_sites]", opts.JobSites)
	setFilterIfPresent(query, "filter[parking_sites]", opts.ParkingSites)

	// Handle polymorphic unloaded_subject filter
	if opts.UnloadedSubjectType != "" && opts.UnloadedSubjectID != "" {
		query.Set("filter[unloaded_subject]", opts.UnloadedSubjectType+"|"+opts.UnloadedSubjectID)
	}

	body, _, err := client.Get(cmd.Context(), "/v1/trips", query)
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

	rows := buildTripRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTripsTable(cmd, rows)
}

func parseTripsListOptions(cmd *cobra.Command) (tripsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	tender, _ := cmd.Flags().GetString("tender")
	tenderJobScheduleShift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	driverDay, _ := cmd.Flags().GetString("driver-day")
	unloadedSubjectType, _ := cmd.Flags().GetString("unloaded-subject-type")
	unloadedSubjectID, _ := cmd.Flags().GetString("unloaded-subject-id")
	materialSites, _ := cmd.Flags().GetString("material-sites")
	jobSites, _ := cmd.Flags().GetString("job-sites")
	parkingSites, _ := cmd.Flags().GetString("parking-sites")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return tripsListOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		NoAuth:                 noAuth,
		Limit:                  limit,
		Offset:                 offset,
		Tender:                 tender,
		TenderJobScheduleShift: tenderJobScheduleShift,
		DriverDay:              driverDay,
		UnloadedSubjectType:    unloadedSubjectType,
		UnloadedSubjectID:      unloadedSubjectID,
		MaterialSites:          materialSites,
		JobSites:               jobSites,
		ParkingSites:           parkingSites,
	}, nil
}

func buildTripRows(resp jsonAPIResponse) []tripRow {
	rows := make([]tripRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := tripRow{
			ID:               resource.ID,
			OriginAt:         stringAttr(resource.Attributes, "origin-at"),
			OriginNotes:      stringAttr(resource.Attributes, "origin-notes"),
			DestinationAt:    stringAttr(resource.Attributes, "destination-at"),
			DestinationNotes: stringAttr(resource.Attributes, "destination-notes"),
			Mileage:          stringAttr(resource.Attributes, "mileage"),
			Minutes:          stringAttr(resource.Attributes, "minutes"),
			IsUnloaded:       boolAttr(resource.Attributes, "is-unloaded"),
		}

		if rel, ok := resource.Relationships["origin"]; ok && rel.Data != nil {
			row.OriginType = rel.Data.Type
			row.OriginID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["destination"]; ok && rel.Data != nil {
			row.DestinationType = rel.Data.Type
			row.DestinationID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["source"]; ok && rel.Data != nil {
			row.SourceType = rel.Data.Type
			row.SourceID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["driver-day"]; ok && rel.Data != nil {
			row.DriverDayID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderTripsTable(cmd *cobra.Command, rows []tripRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No trips found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tORIGIN\tDESTINATION\tMILEAGE\tMINUTES")
	for _, row := range rows {
		origin := ""
		if row.OriginType != "" {
			origin = row.OriginType + "/" + row.OriginID
		}
		destination := ""
		if row.DestinationType != "" {
			destination = row.DestinationType + "/" + row.DestinationID
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(origin, 25),
			truncateString(destination, 25),
			row.Mileage,
			row.Minutes,
		)
	}
	return writer.Flush()
}
