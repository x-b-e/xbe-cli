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

type driverMovementSegmentsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type driverMovementSegmentDetails struct {
	ID                         string   `json:"id"`
	SequenceIndex              int      `json:"sequence_index"`
	StartAt                    string   `json:"start_at,omitempty"`
	EndAt                      string   `json:"end_at,omitempty"`
	Latitude                   string   `json:"latitude,omitempty"`
	Longitude                  string   `json:"longitude,omitempty"`
	MetersTravelled            string   `json:"meters_travelled,omitempty"`
	IsMoving                   bool     `json:"is_moving"`
	SiteKind                   string   `json:"site_kind,omitempty"`
	DurationMinutes            string   `json:"duration_minutes,omitempty"`
	DurationHours              string   `json:"duration_hours,omitempty"`
	Location                   string   `json:"location,omitempty"`
	DriverMovementSegmentSetID string   `json:"driver_movement_segment_set_id,omitempty"`
	DriverDayID                string   `json:"driver_day_id,omitempty"`
	DriverID                   string   `json:"driver_id,omitempty"`
	SiteType                   string   `json:"site_type,omitempty"`
	SiteID                     string   `json:"site_id,omitempty"`
	TrailerID                  string   `json:"trailer_id,omitempty"`
	TractorID                  string   `json:"tractor_id,omitempty"`
	TripIDs                    []string `json:"trip_ids,omitempty"`
}

func newDriverMovementSegmentsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show driver movement segment details",
		Long: `Show the full details of a driver movement segment.

Output Fields:
  ID                          Segment identifier
  Sequence Index              Sequence index in the segment set
  Start At                    Segment start timestamp
  End At                      Segment end timestamp
  Moving                      Whether the segment is moving
  Meters Travelled            Distance travelled in meters
  Site Kind                   Site classification (if available)
  Duration Minutes            Duration in minutes
  Duration Hours              Duration in hours
  Latitude/Longitude          Segment location coordinates
  Driver Movement Segment Set Segment set ID
  Driver Day                  Driver day (trucker shift set) ID
  Driver                      Driver user ID
  Site                        Site relationship (type/id)
  Trailer                     Trailer ID
  Tractor                     Tractor ID
  Trips                       Trip IDs (if any)

Global flags (see xbe --help): --json, --base-url, --token, --no-auth

Arguments:
  <id>    The segment ID (required). You can find IDs using the list command.`,
		Example: `  # Show a movement segment
  xbe view driver-movement-segments show 123

  # Get JSON output
  xbe view driver-movement-segments show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDriverMovementSegmentsShow,
	}
	initDriverMovementSegmentsShowFlags(cmd)
	return cmd
}

func init() {
	driverMovementSegmentsCmd.AddCommand(newDriverMovementSegmentsShowCmd())
}

func initDriverMovementSegmentsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDriverMovementSegmentsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseDriverMovementSegmentsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("driver movement segment id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[driver-movement-segments]", "sequence-index,start-at,end-at,latitude,longitude,meters-travelled,is-moving,site-kind,duration-minutes,duration-hours,location")

	body, _, err := client.Get(cmd.Context(), "/v1/driver-movement-segments/"+id, query)
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

	details := buildDriverMovementSegmentDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderDriverMovementSegmentDetails(cmd, details)
}

func parseDriverMovementSegmentsShowOptions(cmd *cobra.Command) (driverMovementSegmentsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return driverMovementSegmentsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildDriverMovementSegmentDetails(resp jsonAPISingleResponse) driverMovementSegmentDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := driverMovementSegmentDetails{
		ID:              resource.ID,
		SequenceIndex:   intAttr(attrs, "sequence-index"),
		StartAt:         formatDateTime(stringAttr(attrs, "start-at")),
		EndAt:           formatDateTime(stringAttr(attrs, "end-at")),
		Latitude:        stringAttr(attrs, "latitude"),
		Longitude:       stringAttr(attrs, "longitude"),
		MetersTravelled: stringAttr(attrs, "meters-travelled"),
		IsMoving:        boolAttr(attrs, "is-moving"),
		SiteKind:        stringAttr(attrs, "site-kind"),
		DurationMinutes: stringAttr(attrs, "duration-minutes"),
		DurationHours:   stringAttr(attrs, "duration-hours"),
	}

	locationParts := stringSliceAttr(attrs, "location")
	if len(locationParts) > 0 {
		details.Location = strings.Join(locationParts, ", ")
	} else if details.Latitude != "" && details.Longitude != "" {
		details.Location = details.Latitude + ", " + details.Longitude
	}

	if rel, ok := resource.Relationships["driver-movement-segment-set"]; ok && rel.Data != nil {
		details.DriverMovementSegmentSetID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["driver-day"]; ok && rel.Data != nil {
		details.DriverDayID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["driver"]; ok && rel.Data != nil {
		details.DriverID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["site"]; ok && rel.Data != nil {
		details.SiteType = rel.Data.Type
		details.SiteID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["trailer"]; ok && rel.Data != nil {
		details.TrailerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["tractor"]; ok && rel.Data != nil {
		details.TractorID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["trips"]; ok && rel.raw != nil {
		var refs []jsonAPIResourceIdentifier
		if err := json.Unmarshal(rel.raw, &refs); err == nil {
			for _, ref := range refs {
				details.TripIDs = append(details.TripIDs, ref.ID)
			}
		}
	}

	return details
}

func renderDriverMovementSegmentDetails(cmd *cobra.Command, details driverMovementSegmentDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	fmt.Fprintf(out, "Sequence Index: %d\n", details.SequenceIndex)
	if details.StartAt != "" {
		fmt.Fprintf(out, "Start At: %s\n", details.StartAt)
	}
	if details.EndAt != "" {
		fmt.Fprintf(out, "End At: %s\n", details.EndAt)
	}
	fmt.Fprintf(out, "Moving: %s\n", formatYesNo(details.IsMoving))
	if details.MetersTravelled != "" {
		fmt.Fprintf(out, "Meters Travelled: %s\n", details.MetersTravelled)
	}
	if details.SiteKind != "" {
		fmt.Fprintf(out, "Site Kind: %s\n", details.SiteKind)
	}
	if details.DurationMinutes != "" {
		fmt.Fprintf(out, "Duration Minutes: %s\n", details.DurationMinutes)
	}
	if details.DurationHours != "" {
		fmt.Fprintf(out, "Duration Hours: %s\n", details.DurationHours)
	}
	if details.Latitude != "" {
		fmt.Fprintf(out, "Latitude: %s\n", details.Latitude)
	}
	if details.Longitude != "" {
		fmt.Fprintf(out, "Longitude: %s\n", details.Longitude)
	}
	if details.Location != "" {
		fmt.Fprintf(out, "Location: %s\n", details.Location)
	}
	if details.DriverMovementSegmentSetID != "" {
		fmt.Fprintf(out, "Driver Movement Segment Set: %s\n", details.DriverMovementSegmentSetID)
	}
	if details.DriverDayID != "" {
		fmt.Fprintf(out, "Driver Day: %s\n", details.DriverDayID)
	}
	if details.DriverID != "" {
		fmt.Fprintf(out, "Driver: %s\n", details.DriverID)
	}
	if details.SiteType != "" || details.SiteID != "" {
		site := details.SiteType
		if details.SiteID != "" {
			if site != "" {
				site += "/"
			}
			site += details.SiteID
		}
		fmt.Fprintf(out, "Site: %s\n", site)
	}
	if details.TrailerID != "" {
		fmt.Fprintf(out, "Trailer: %s\n", details.TrailerID)
	}
	if details.TractorID != "" {
		fmt.Fprintf(out, "Tractor: %s\n", details.TractorID)
	}
	if len(details.TripIDs) > 0 {
		fmt.Fprintf(out, "Trips: %s\n", strings.Join(details.TripIDs, ", "))
	}

	return nil
}
