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

type driverMovementSegmentSetsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type driverMovementSegmentSetDetails struct {
	ID                       string   `json:"id"`
	Date                     string   `json:"date,omitempty"`
	DriverName               string   `json:"driver_name,omitempty"`
	DriverID                 string   `json:"driver_id,omitempty"`
	DriverDayID              string   `json:"driver_day_id,omitempty"`
	SegmentsCount            int      `json:"segments_count"`
	MovingSegmentsCount      int      `json:"moving_segments_count"`
	TotalMetersTravelled     int      `json:"total_meters_travelled"`
	ShiftSettings            any      `json:"shift_settings,omitempty"`
	DriverMovementSegmentIDs []string `json:"driver_movement_segment_ids,omitempty"`
}

func newDriverMovementSegmentSetsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show driver movement segment set details",
		Long: `Show the full details of a driver movement segment set.

Output Fields:
  ID
  Date
  Driver Name
  Driver ID
  Driver Day ID
  Segments Count
  Moving Segments Count
  Total Meters Travelled
  Shift Settings
  Driver Movement Segment IDs

Arguments:
  <id>    The driver movement segment set ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a driver movement segment set
  xbe view driver-movement-segment-sets show 123

  # Output as JSON
  xbe view driver-movement-segment-sets show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDriverMovementSegmentSetsShow,
	}
	initDriverMovementSegmentSetsShowFlags(cmd)
	return cmd
}

func init() {
	driverMovementSegmentSetsCmd.AddCommand(newDriverMovementSegmentSetsShowCmd())
}

func initDriverMovementSegmentSetsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDriverMovementSegmentSetsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseDriverMovementSegmentSetsShowOptions(cmd)
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
		return fmt.Errorf("driver movement segment set id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[driver-movement-segment-sets]", "shift-settings,date,driver-name,segments-count,moving-segments-count,total-meters-travelled,driver-day,driver,driver-movement-segments")
	query.Set("include", "driver-day,driver")

	body, _, err := client.Get(cmd.Context(), "/v1/driver-movement-segment-sets/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildDriverMovementSegmentSetDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderDriverMovementSegmentSetDetails(cmd, details)
}

func parseDriverMovementSegmentSetsShowOptions(cmd *cobra.Command) (driverMovementSegmentSetsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return driverMovementSegmentSetsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildDriverMovementSegmentSetDetails(resp jsonAPISingleResponse) driverMovementSegmentSetDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := driverMovementSegmentSetDetails{
		ID:                   resource.ID,
		Date:                 formatDate(stringAttr(attrs, "date")),
		DriverName:           stringAttr(attrs, "driver-name"),
		SegmentsCount:        intAttr(attrs, "segments-count"),
		MovingSegmentsCount:  intAttr(attrs, "moving-segments-count"),
		TotalMetersTravelled: intAttr(attrs, "total-meters-travelled"),
	}

	if value, ok := attrs["shift-settings"]; ok && value != nil {
		details.ShiftSettings = value
	}

	if rel, ok := resource.Relationships["driver-day"]; ok && rel.Data != nil {
		details.DriverDayID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["driver"]; ok && rel.Data != nil {
		details.DriverID = rel.Data.ID
	}

	if rel, ok := resource.Relationships["driver-movement-segments"]; ok && rel.raw != nil {
		var refs []jsonAPIResourceIdentifier
		if err := json.Unmarshal(rel.raw, &refs); err == nil {
			details.DriverMovementSegmentIDs = make([]string, 0, len(refs))
			for _, ref := range refs {
				details.DriverMovementSegmentIDs = append(details.DriverMovementSegmentIDs, ref.ID)
			}
		}
	}

	return details
}

func renderDriverMovementSegmentSetDetails(cmd *cobra.Command, details driverMovementSegmentSetDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Date != "" {
		fmt.Fprintf(out, "Date: %s\n", details.Date)
	}
	if details.DriverName != "" {
		fmt.Fprintf(out, "Driver Name: %s\n", details.DriverName)
	}
	if details.DriverID != "" {
		fmt.Fprintf(out, "Driver ID: %s\n", details.DriverID)
	}
	if details.DriverDayID != "" {
		fmt.Fprintf(out, "Driver Day ID: %s\n", details.DriverDayID)
	}
	fmt.Fprintf(out, "Segments Count: %d\n", details.SegmentsCount)
	fmt.Fprintf(out, "Moving Segments Count: %d\n", details.MovingSegmentsCount)
	fmt.Fprintf(out, "Total Meters Travelled: %d\n", details.TotalMetersTravelled)

	if details.ShiftSettings != nil {
		shiftSettings := formatJSONValue(details.ShiftSettings)
		if shiftSettings != "" {
			fmt.Fprintln(out, "")
			fmt.Fprintln(out, "Shift Settings:")
			fmt.Fprintln(out, strings.Repeat("-", 40))
			fmt.Fprintln(out, shiftSettings)
		}
	}

	if len(details.DriverMovementSegmentIDs) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintf(out, "Driver Movement Segments (%d):\n", len(details.DriverMovementSegmentIDs))
		fmt.Fprintln(out, strings.Repeat("-", 40))
		for _, id := range details.DriverMovementSegmentIDs {
			fmt.Fprintf(out, "  - %s\n", id)
		}
	}

	return nil
}

func formatJSONValue(value any) string {
	if value == nil {
		return ""
	}
	pretty, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return string(pretty)
}
