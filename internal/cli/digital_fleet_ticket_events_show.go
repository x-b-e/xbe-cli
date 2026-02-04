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

type digitalFleetTicketEventsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type digitalFleetTicketEventDetails struct {
	ID                  string `json:"id"`
	EventAt             string `json:"event_at,omitempty"`
	EventLatitude       string `json:"event_latitude,omitempty"`
	EventLongitude      string `json:"event_longitude,omitempty"`
	EventName           string `json:"event_name,omitempty"`
	UniqueID            string `json:"uniqueid,omitempty"`
	EventID             string `json:"event_id,omitempty"`
	TruckID             string `json:"truck_id,omitempty"`
	BrokerID            string `json:"broker_id,omitempty"`
	DigitalFleetTruckID string `json:"digital_fleet_truck_id,omitempty"`
	TruckerID           string `json:"trucker_id,omitempty"`
	TractorID           string `json:"tractor_id,omitempty"`
	TrailerID           string `json:"trailer_id,omitempty"`
	SiteEventID         string `json:"site_event_id,omitempty"`
}

func newDigitalFleetTicketEventsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show digital fleet ticket event details",
		Long: `Show the full details of a digital fleet ticket event.

Output Fields:
  ID
  Event At
  Event Latitude
  Event Longitude
  Event Name
  Unique ID
  Event ID
  Truck ID
  Broker ID
  Digital Fleet Truck ID
  Trucker ID
  Tractor ID
  Trailer ID
  Site Event ID

Arguments:
  <id>    The ticket event ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show ticket event details
  xbe view digital-fleet-ticket-events show 123

  # Get JSON output
  xbe view digital-fleet-ticket-events show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDigitalFleetTicketEventsShow,
	}
	initDigitalFleetTicketEventsShowFlags(cmd)
	return cmd
}

func init() {
	digitalFleetTicketEventsCmd.AddCommand(newDigitalFleetTicketEventsShowCmd())
}

func initDigitalFleetTicketEventsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDigitalFleetTicketEventsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseDigitalFleetTicketEventsShowOptions(cmd)
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
		return fmt.Errorf("digital fleet ticket event id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[digital-fleet-ticket-events]", "event-at,event-latitude,event-longitude,event-name,uniqueid,event-id,truck-id,broker,digital-fleet-truck,trucker,tractor,trailer,site-event")

	body, _, err := client.Get(cmd.Context(), "/v1/digital-fleet-ticket-events/"+id, query)
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

	details := buildDigitalFleetTicketEventDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderDigitalFleetTicketEventDetails(cmd, details)
}

func parseDigitalFleetTicketEventsShowOptions(cmd *cobra.Command) (digitalFleetTicketEventsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return digitalFleetTicketEventsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildDigitalFleetTicketEventDetails(resp jsonAPISingleResponse) digitalFleetTicketEventDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := digitalFleetTicketEventDetails{
		ID:             resource.ID,
		EventAt:        formatDateTime(stringAttr(attrs, "event-at")),
		EventLatitude:  stringAttr(attrs, "event-latitude"),
		EventLongitude: stringAttr(attrs, "event-longitude"),
		EventName:      stringAttr(attrs, "event-name"),
		UniqueID:       stringAttr(attrs, "uniqueid"),
		EventID:        stringAttr(attrs, "event-id"),
		TruckID:        stringAttr(attrs, "truck-id"),
	}

	details.BrokerID = relationshipIDFromMap(resource.Relationships, "broker")
	details.DigitalFleetTruckID = relationshipIDFromMap(resource.Relationships, "digital-fleet-truck")
	details.TruckerID = relationshipIDFromMap(resource.Relationships, "trucker")
	details.TractorID = relationshipIDFromMap(resource.Relationships, "tractor")
	details.TrailerID = relationshipIDFromMap(resource.Relationships, "trailer")
	details.SiteEventID = relationshipIDFromMap(resource.Relationships, "site-event")

	return details
}

func renderDigitalFleetTicketEventDetails(cmd *cobra.Command, details digitalFleetTicketEventDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.EventAt != "" {
		fmt.Fprintf(out, "Event At: %s\n", details.EventAt)
	}
	if details.EventLatitude != "" {
		fmt.Fprintf(out, "Event Latitude: %s\n", details.EventLatitude)
	}
	if details.EventLongitude != "" {
		fmt.Fprintf(out, "Event Longitude: %s\n", details.EventLongitude)
	}
	if details.EventName != "" {
		fmt.Fprintf(out, "Event Name: %s\n", details.EventName)
	}
	if details.UniqueID != "" {
		fmt.Fprintf(out, "Unique ID: %s\n", details.UniqueID)
	}
	if details.EventID != "" {
		fmt.Fprintf(out, "Event ID: %s\n", details.EventID)
	}
	if details.TruckID != "" {
		fmt.Fprintf(out, "Truck ID: %s\n", details.TruckID)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
	}
	if details.DigitalFleetTruckID != "" {
		fmt.Fprintf(out, "Digital Fleet Truck ID: %s\n", details.DigitalFleetTruckID)
	}
	if details.TruckerID != "" {
		fmt.Fprintf(out, "Trucker ID: %s\n", details.TruckerID)
	}
	if details.TractorID != "" {
		fmt.Fprintf(out, "Tractor ID: %s\n", details.TractorID)
	}
	if details.TrailerID != "" {
		fmt.Fprintf(out, "Trailer ID: %s\n", details.TrailerID)
	}
	if details.SiteEventID != "" {
		fmt.Fprintf(out, "Site Event ID: %s\n", details.SiteEventID)
	}

	return nil
}
