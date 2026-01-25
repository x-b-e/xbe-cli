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

type digitalFleetTicketEventsListOptions struct {
	BaseURL    string
	Token      string
	JSON       bool
	NoAuth     bool
	Limit      int
	Offset     int
	Sort       string
	Broker     string
	EventAtMin string
	EventAtMax string
	IsEventAt  string
}

type digitalFleetTicketEventRow struct {
	ID        string `json:"id"`
	EventAt   string `json:"event_at,omitempty"`
	EventName string `json:"event_name,omitempty"`
	EventID   string `json:"event_id,omitempty"`
	UniqueID  string `json:"uniqueid,omitempty"`
	TruckID   string `json:"truck_id,omitempty"`
	BrokerID  string `json:"broker_id,omitempty"`
}

func newDigitalFleetTicketEventsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List digital fleet ticket events",
		Long: `List digital fleet ticket events with filtering and pagination.

Digital fleet ticket events capture telematics ticket events ingested from
Digital Fleet.

Output Columns:
  ID         Ticket event identifier
  EVENT AT   When the event occurred
  EVENT NAME Event name from Digital Fleet
  EVENT ID   Digital Fleet event ID
  UNIQUE ID  Unique event identifier
  TRUCK ID   Truck ID from Digital Fleet
  BROKER     Broker ID

Filters:
  --broker        Filter by broker ID
  --event-at-min  Filter by event-at on/after (ISO 8601)
  --event-at-max  Filter by event-at on/before (ISO 8601)
  --is-event-at   Filter by presence of event-at (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List ticket events
  xbe view digital-fleet-ticket-events list

  # Filter by broker
  xbe view digital-fleet-ticket-events list --broker 123

  # Filter by event time range
  xbe view digital-fleet-ticket-events list \
    --event-at-min 2025-01-01T00:00:00Z --event-at-max 2025-01-31T23:59:59Z

  # Output as JSON
  xbe view digital-fleet-ticket-events list --json`,
		Args: cobra.NoArgs,
		RunE: runDigitalFleetTicketEventsList,
	}
	initDigitalFleetTicketEventsListFlags(cmd)
	return cmd
}

func init() {
	digitalFleetTicketEventsCmd.AddCommand(newDigitalFleetTicketEventsListCmd())
}

func initDigitalFleetTicketEventsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("event-at-min", "", "Filter by event-at on/after (ISO 8601)")
	cmd.Flags().String("event-at-max", "", "Filter by event-at on/before (ISO 8601)")
	cmd.Flags().String("is-event-at", "", "Filter by presence of event-at (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDigitalFleetTicketEventsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseDigitalFleetTicketEventsListOptions(cmd)
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
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[digital-fleet-ticket-events]", "event-at,event-name,event-id,uniqueid,truck-id,broker")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[event-at-min]", opts.EventAtMin)
	setFilterIfPresent(query, "filter[event-at-max]", opts.EventAtMax)
	setFilterIfPresent(query, "filter[is-event-at]", opts.IsEventAt)

	body, _, err := client.Get(cmd.Context(), "/v1/digital-fleet-ticket-events", query)
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

	rows := buildDigitalFleetTicketEventRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderDigitalFleetTicketEventsTable(cmd, rows)
}

func parseDigitalFleetTicketEventsListOptions(cmd *cobra.Command) (digitalFleetTicketEventsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	broker, _ := cmd.Flags().GetString("broker")
	eventAtMin, _ := cmd.Flags().GetString("event-at-min")
	eventAtMax, _ := cmd.Flags().GetString("event-at-max")
	isEventAt, _ := cmd.Flags().GetString("is-event-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return digitalFleetTicketEventsListOptions{
		BaseURL:    baseURL,
		Token:      token,
		JSON:       jsonOut,
		NoAuth:     noAuth,
		Limit:      limit,
		Offset:     offset,
		Sort:       sort,
		Broker:     broker,
		EventAtMin: eventAtMin,
		EventAtMax: eventAtMax,
		IsEventAt:  isEventAt,
	}, nil
}

func buildDigitalFleetTicketEventRows(resp jsonAPIResponse) []digitalFleetTicketEventRow {
	rows := make([]digitalFleetTicketEventRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildDigitalFleetTicketEventRow(resource))
	}
	return rows
}

func buildDigitalFleetTicketEventRow(resource jsonAPIResource) digitalFleetTicketEventRow {
	attrs := resource.Attributes
	return digitalFleetTicketEventRow{
		ID:        resource.ID,
		EventAt:   formatDateTime(stringAttr(attrs, "event-at")),
		EventName: stringAttr(attrs, "event-name"),
		EventID:   stringAttr(attrs, "event-id"),
		UniqueID:  stringAttr(attrs, "uniqueid"),
		TruckID:   stringAttr(attrs, "truck-id"),
		BrokerID:  relationshipIDFromMap(resource.Relationships, "broker"),
	}
}

func renderDigitalFleetTicketEventsTable(cmd *cobra.Command, rows []digitalFleetTicketEventRow) error {
	out := cmd.OutOrStdout()
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)

	fmt.Fprintln(w, "ID\tEVENT AT\tEVENT NAME\tEVENT ID\tUNIQUE ID\tTRUCK ID\tBROKER")
	for _, row := range rows {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.EventAt,
			row.EventName,
			row.EventID,
			row.UniqueID,
			row.TruckID,
			row.BrokerID,
		)
	}

	return w.Flush()
}
