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

type equipmentLocationEventsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type equipmentLocationEventDetails struct {
	ID             string `json:"id"`
	EquipmentID    string `json:"equipment_id,omitempty"`
	UpdatedByID    string `json:"updated_by_id,omitempty"`
	EventAt        string `json:"event_at,omitempty"`
	EventLatitude  string `json:"event_latitude,omitempty"`
	EventLongitude string `json:"event_longitude,omitempty"`
	Provenance     string `json:"provenance,omitempty"`
	CreatedAt      string `json:"created_at,omitempty"`
	UpdatedAt      string `json:"updated_at,omitempty"`
}

func newEquipmentLocationEventsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show equipment location event details",
		Long: `Show the full details of a specific equipment location event.

Output Fields:
  ID             Equipment location event identifier
  EQUIPMENT      Equipment ID
  EVENT AT       Event timestamp
  EVENT LAT      Event latitude
  EVENT LON      Event longitude
  PROVENANCE     Event provenance (gps, map)
  UPDATED BY     Updated by user ID
  CREATED AT     Record creation timestamp
  UPDATED AT     Record update timestamp

Arguments:
  <id>  Equipment location event ID (required). Find IDs using the list command.`,
		Example: `  # View an equipment location event by ID
  xbe view equipment-location-events show 123

  # Get JSON output
  xbe view equipment-location-events show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runEquipmentLocationEventsShow,
	}
	initEquipmentLocationEventsShowFlags(cmd)
	return cmd
}

func init() {
	equipmentLocationEventsCmd.AddCommand(newEquipmentLocationEventsShowCmd())
}

func initEquipmentLocationEventsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runEquipmentLocationEventsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseEquipmentLocationEventsShowOptions(cmd)
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
		return fmt.Errorf("equipment location event id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[equipment-location-events]", "event-latitude,event-longitude,event-at,provenance,created-at,updated-at,updated-by,equipment")

	body, _, err := client.Get(cmd.Context(), "/v1/equipment-location-events/"+id, query)
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

	details := buildEquipmentLocationEventDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderEquipmentLocationEventDetails(cmd, details)
}

func parseEquipmentLocationEventsShowOptions(cmd *cobra.Command) (equipmentLocationEventsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return equipmentLocationEventsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildEquipmentLocationEventDetails(resp jsonAPISingleResponse) equipmentLocationEventDetails {
	attrs := resp.Data.Attributes

	details := equipmentLocationEventDetails{
		ID:             resp.Data.ID,
		EventAt:        formatDateTime(stringAttr(attrs, "event-at")),
		EventLatitude:  stringAttr(attrs, "event-latitude"),
		EventLongitude: stringAttr(attrs, "event-longitude"),
		Provenance:     stringAttr(attrs, "provenance"),
		CreatedAt:      formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:      formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resp.Data.Relationships["equipment"]; ok && rel.Data != nil {
		details.EquipmentID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["updated-by"]; ok && rel.Data != nil {
		details.UpdatedByID = rel.Data.ID
	}

	return details
}

func renderEquipmentLocationEventDetails(cmd *cobra.Command, details equipmentLocationEventDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.EquipmentID != "" {
		fmt.Fprintf(out, "Equipment: %s\n", details.EquipmentID)
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
	if details.Provenance != "" {
		fmt.Fprintf(out, "Provenance: %s\n", details.Provenance)
	}
	if details.UpdatedByID != "" {
		fmt.Fprintf(out, "Updated By: %s\n", details.UpdatedByID)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
