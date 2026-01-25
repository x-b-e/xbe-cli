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

type siteEventsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type siteEventDetails struct {
	ID                     string `json:"id"`
	EventType              string `json:"event_type,omitempty"`
	EventKind              string `json:"event_kind,omitempty"`
	EventDetails           string `json:"event_details,omitempty"`
	EventAt                string `json:"event_at,omitempty"`
	EventLatitude          string `json:"event_latitude,omitempty"`
	EventLongitude         string `json:"event_longitude,omitempty"`
	EventTimeZoneID        string `json:"event_time_zone_id,omitempty"`
	EventSiteType          string `json:"event_site_type,omitempty"`
	EventSiteID            string `json:"event_site_id,omitempty"`
	EventSourceType        string `json:"event_source_type,omitempty"`
	EventSourceID          string `json:"event_source_id,omitempty"`
	TenderJobScheduleShift string `json:"tender_job_schedule_shift_id,omitempty"`
	MaterialTransaction    string `json:"material_transaction_id,omitempty"`
	BrokerID               string `json:"broker_id,omitempty"`
	TruckerID              string `json:"trucker_id,omitempty"`
}

func newSiteEventsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show site event details",
		Long: `Show the full details of a site event.

Output Fields:
  ID           Site event identifier
  Type         Event type
  Kind         Event kind
  Details      Event details
  Event At     Event timestamp
  Latitude     Event latitude
  Longitude    Event longitude
  Time Zone    Event time zone ID
  Event Site   Event site (type/id)
  Event Source Event source (type/id)
  Shift        Tender job schedule shift ID
  Transaction  Material transaction ID
  Trucker      Trucker ID
  Broker       Broker ID

Arguments:
  <id>   The site event ID (required). You can find IDs using the list command.`,
		Example: `  # Show a site event
  xbe view site-events show 123

  # JSON output
  xbe view site-events show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runSiteEventsShow,
	}
	initSiteEventsShowFlags(cmd)
	return cmd
}

func init() {
	siteEventsCmd.AddCommand(newSiteEventsShowCmd())
}

func initSiteEventsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runSiteEventsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseSiteEventsShowOptions(cmd)
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
		return fmt.Errorf("site event id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[site-events]", "event-type,event-kind,event-details,event-at,event-latitude,event-longitude,event-time-zone-id,event-site,event-source,tender-job-schedule-shift,material-transaction,broker,trucker")

	body, _, err := client.Get(cmd.Context(), "/v1/site-events/"+id, query)
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

	details := buildSiteEventDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderSiteEventDetails(cmd, details)
}

func parseSiteEventsShowOptions(cmd *cobra.Command) (siteEventsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return siteEventsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildSiteEventDetails(resp jsonAPISingleResponse) siteEventDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := siteEventDetails{
		ID:              resource.ID,
		EventType:       stringAttr(attrs, "event-type"),
		EventKind:       stringAttr(attrs, "event-kind"),
		EventDetails:    stringAttr(attrs, "event-details"),
		EventAt:         formatDateTime(stringAttr(attrs, "event-at")),
		EventLatitude:   stringAttr(attrs, "event-latitude"),
		EventLongitude:  stringAttr(attrs, "event-longitude"),
		EventTimeZoneID: stringAttr(attrs, "event-time-zone-id"),
	}

	if rel, ok := resource.Relationships["event-site"]; ok && rel.Data != nil {
		details.EventSiteType = rel.Data.Type
		details.EventSiteID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["event-source"]; ok && rel.Data != nil {
		details.EventSourceType = rel.Data.Type
		details.EventSourceID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["tender-job-schedule-shift"]; ok && rel.Data != nil {
		details.TenderJobScheduleShift = rel.Data.ID
	}
	if rel, ok := resource.Relationships["material-transaction"]; ok && rel.Data != nil {
		details.MaterialTransaction = rel.Data.ID
	}
	if rel, ok := resource.Relationships["trucker"]; ok && rel.Data != nil {
		details.TruckerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
	}

	return details
}

func renderSiteEventDetails(cmd *cobra.Command, details siteEventDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.EventType != "" {
		fmt.Fprintf(out, "Type: %s\n", details.EventType)
	}
	if details.EventKind != "" {
		fmt.Fprintf(out, "Kind: %s\n", details.EventKind)
	}
	if details.EventDetails != "" {
		fmt.Fprintf(out, "Details: %s\n", details.EventDetails)
	}
	if details.EventAt != "" {
		fmt.Fprintf(out, "Event At: %s\n", details.EventAt)
	}
	if details.EventLatitude != "" {
		fmt.Fprintf(out, "Latitude: %s\n", details.EventLatitude)
	}
	if details.EventLongitude != "" {
		fmt.Fprintf(out, "Longitude: %s\n", details.EventLongitude)
	}
	if details.EventTimeZoneID != "" {
		fmt.Fprintf(out, "Time Zone: %s\n", details.EventTimeZoneID)
	}
	if details.EventSiteType != "" || details.EventSiteID != "" {
		fmt.Fprintf(out, "Event Site: %s\n", formatResourceRef(details.EventSiteType, details.EventSiteID))
	}
	if details.EventSourceType != "" || details.EventSourceID != "" {
		fmt.Fprintf(out, "Event Source: %s\n", formatResourceRef(details.EventSourceType, details.EventSourceID))
	}
	if details.TenderJobScheduleShift != "" {
		fmt.Fprintf(out, "Shift: %s\n", details.TenderJobScheduleShift)
	}
	if details.MaterialTransaction != "" {
		fmt.Fprintf(out, "Transaction: %s\n", details.MaterialTransaction)
	}
	if details.TruckerID != "" {
		fmt.Fprintf(out, "Trucker: %s\n", details.TruckerID)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker: %s\n", details.BrokerID)
	}

	return nil
}
