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

type geofenceRestrictionViolationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type geofenceRestrictionViolationDetails struct {
	ID                     string `json:"id"`
	GeofenceID             string `json:"geofence_id,omitempty"`
	TrailerID              string `json:"trailer_id,omitempty"`
	TractorID              string `json:"tractor_id,omitempty"`
	DriverID               string `json:"driver_id,omitempty"`
	TenderJobScheduleShift string `json:"tender_job_schedule_shift_id,omitempty"`
	Latitude               string `json:"latitude,omitempty"`
	Longitude              string `json:"longitude,omitempty"`
	EventAt                string `json:"event_at,omitempty"`
	ShouldNotify           bool   `json:"should_notify"`
	NotificationSentAt     string `json:"notification_sent_at,omitempty"`
	EventSourceType        string `json:"event_source_type,omitempty"`
	EventSourceID          string `json:"event_source_id,omitempty"`
	CreatedAt              string `json:"created_at,omitempty"`
	UpdatedAt              string `json:"updated_at,omitempty"`
}

func newGeofenceRestrictionViolationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show geofence restriction violation details",
		Long: `Show the full details of a specific geofence restriction violation.

Output Fields:
  ID                 Violation identifier
  Geofence            Geofence ID
  Trailer             Trailer ID (if present)
  Tractor             Tractor ID (if present)
  Driver              Driver ID (if present)
  Shift               Tender job schedule shift ID (if present)
  Coordinates         Latitude/longitude coordinates
  Event At            Event timestamp
  Should Notify       Whether notifications should be sent
  Notification At     Notification timestamp (if sent)
  Event Source        Event source type and ID (if present)
  Created At          Record creation timestamp
  Updated At          Record update timestamp

Arguments:
  <id>    The violation ID (required). You can find IDs using the list command.`,
		Example: `  # Show a violation
  xbe view geofence-restriction-violations show 123

  # Output as JSON
  xbe view geofence-restriction-violations show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runGeofenceRestrictionViolationsShow,
	}
	initGeofenceRestrictionViolationsShowFlags(cmd)
	return cmd
}

func init() {
	geofenceRestrictionViolationsCmd.AddCommand(newGeofenceRestrictionViolationsShowCmd())
}

func initGeofenceRestrictionViolationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runGeofenceRestrictionViolationsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseGeofenceRestrictionViolationsShowOptions(cmd)
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
		return fmt.Errorf("geofence restriction violation id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[geofence-restriction-violations]", "latitude,longitude,event-at,should-notify,notification-sent-at,event-source-type,event-source-id,created-at,updated-at,geofence,trailer,tractor,driver,tender-job-schedule-shift")

	body, _, err := client.Get(cmd.Context(), "/v1/geofence-restriction-violations/"+id, query)
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

	details := buildGeofenceRestrictionViolationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderGeofenceRestrictionViolationDetails(cmd, details)
}

func parseGeofenceRestrictionViolationsShowOptions(cmd *cobra.Command) (geofenceRestrictionViolationsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return geofenceRestrictionViolationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildGeofenceRestrictionViolationDetails(resp jsonAPISingleResponse) geofenceRestrictionViolationDetails {
	attrs := resp.Data.Attributes
	row := buildGeofenceRestrictionViolationRowFromSingle(resp)

	details := geofenceRestrictionViolationDetails{
		ID:                     resp.Data.ID,
		GeofenceID:             row.GeofenceID,
		TrailerID:              row.TrailerID,
		TractorID:              row.TractorID,
		DriverID:               row.DriverID,
		TenderJobScheduleShift: row.TenderJobScheduleShift,
		Latitude:               strings.TrimSpace(stringAttr(attrs, "latitude")),
		Longitude:              strings.TrimSpace(stringAttr(attrs, "longitude")),
		EventAt:                formatDateTime(stringAttr(attrs, "event-at")),
		ShouldNotify:           boolAttr(attrs, "should-notify"),
		NotificationSentAt:     formatDateTime(stringAttr(attrs, "notification-sent-at")),
		EventSourceType:        strings.TrimSpace(stringAttr(attrs, "event-source-type")),
		EventSourceID:          strings.TrimSpace(stringAttr(attrs, "event-source-id")),
		CreatedAt:              formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:              formatDateTime(stringAttr(attrs, "updated-at")),
	}

	return details
}

func renderGeofenceRestrictionViolationDetails(cmd *cobra.Command, details geofenceRestrictionViolationDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.GeofenceID != "" {
		fmt.Fprintf(out, "Geofence: %s\n", details.GeofenceID)
	}
	if details.TrailerID != "" {
		fmt.Fprintf(out, "Trailer: %s\n", details.TrailerID)
	}
	if details.TractorID != "" {
		fmt.Fprintf(out, "Tractor: %s\n", details.TractorID)
	}
	if details.DriverID != "" {
		fmt.Fprintf(out, "Driver: %s\n", details.DriverID)
	}
	if details.TenderJobScheduleShift != "" {
		fmt.Fprintf(out, "Shift: %s\n", details.TenderJobScheduleShift)
	}
	if details.Latitude != "" || details.Longitude != "" {
		fmt.Fprintf(out, "Coordinates: %s, %s\n", details.Latitude, details.Longitude)
	}
	if details.EventAt != "" {
		fmt.Fprintf(out, "Event At: %s\n", details.EventAt)
	}
	fmt.Fprintf(out, "Should Notify: %t\n", details.ShouldNotify)
	if details.NotificationSentAt != "" {
		fmt.Fprintf(out, "Notification At: %s\n", details.NotificationSentAt)
	}
	if details.EventSourceType != "" || details.EventSourceID != "" {
		source := details.EventSourceType
		if details.EventSourceID != "" {
			if source != "" {
				source += "/"
			}
			source += details.EventSourceID
		}
		fmt.Fprintf(out, "Event Source: %s\n", source)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
