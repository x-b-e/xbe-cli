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

type vehicleLocationEventsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type vehicleLocationEventDetails struct {
	ID             string `json:"id"`
	TractorID      string `json:"tractor_id,omitempty"`
	TrailerID      string `json:"trailer_id,omitempty"`
	EventAt        string `json:"event_at,omitempty"`
	EventLatitude  string `json:"event_latitude,omitempty"`
	EventLongitude string `json:"event_longitude,omitempty"`
	CreatedAt      string `json:"created_at,omitempty"`
	UpdatedAt      string `json:"updated_at,omitempty"`
}

func newVehicleLocationEventsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show vehicle location event details",
		Long: `Show the full details of a specific vehicle location event.

Output Fields:
  ID         Vehicle location event identifier
  TRACTOR    Tractor ID
  TRAILER    Trailer ID
  EVENT AT   Event timestamp
  EVENT LAT  Event latitude
  EVENT LON  Event longitude
  CREATED AT Record creation timestamp
  UPDATED AT Record update timestamp

Arguments:
  <id>  Vehicle location event ID (required). Find IDs using the list command.`,
		Example: `  # View a vehicle location event by ID
  xbe view vehicle-location-events show 123

  # Get JSON output
  xbe view vehicle-location-events show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runVehicleLocationEventsShow,
	}
	initVehicleLocationEventsShowFlags(cmd)
	return cmd
}

func init() {
	vehicleLocationEventsCmd.AddCommand(newVehicleLocationEventsShowCmd())
}

func initVehicleLocationEventsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runVehicleLocationEventsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseVehicleLocationEventsShowOptions(cmd)
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
		return fmt.Errorf("vehicle location event id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[vehicle-location-events]", "event-latitude,event-longitude,event-at,created-at,updated-at,tractor,trailer")

	body, _, err := client.Get(cmd.Context(), "/v1/vehicle-location-events/"+id, query)
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

	details := buildVehicleLocationEventDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderVehicleLocationEventDetails(cmd, details)
}

func parseVehicleLocationEventsShowOptions(cmd *cobra.Command) (vehicleLocationEventsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return vehicleLocationEventsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildVehicleLocationEventDetails(resp jsonAPISingleResponse) vehicleLocationEventDetails {
	attrs := resp.Data.Attributes

	details := vehicleLocationEventDetails{
		ID:             resp.Data.ID,
		EventAt:        formatDateTime(stringAttr(attrs, "event-at")),
		EventLatitude:  stringAttr(attrs, "event-latitude"),
		EventLongitude: stringAttr(attrs, "event-longitude"),
		CreatedAt:      formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:      formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resp.Data.Relationships["tractor"]; ok && rel.Data != nil {
		details.TractorID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["trailer"]; ok && rel.Data != nil {
		details.TrailerID = rel.Data.ID
	}

	return details
}

func renderVehicleLocationEventDetails(cmd *cobra.Command, details vehicleLocationEventDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TractorID != "" {
		fmt.Fprintf(out, "Tractor: %s\n", details.TractorID)
	}
	if details.TrailerID != "" {
		fmt.Fprintf(out, "Trailer: %s\n", details.TrailerID)
	}
	if details.EventAt != "" {
		fmt.Fprintf(out, "Event At: %s\n", details.EventAt)
	}
	if details.EventLatitude != "" {
		fmt.Fprintf(out, "Event Latitude: %s\n", details.EventLatitude)
	}
	if details.EventLongitude != "" {
		fmt.Fprintf(out, "Event Longitude: %s\n", details.EventLongitude)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
