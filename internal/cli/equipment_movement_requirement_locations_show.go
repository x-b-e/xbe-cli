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

type equipmentMovementRequirementLocationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type equipmentMovementRequirementLocationDetails struct {
	ID                        string   `json:"id"`
	Name                      string   `json:"name,omitempty"`
	Latitude                  string   `json:"latitude,omitempty"`
	Longitude                 string   `json:"longitude,omitempty"`
	DistanceMiles             string   `json:"distance_miles,omitempty"`
	BrokerID                  string   `json:"broker_id,omitempty"`
	BrokerName                string   `json:"broker,omitempty"`
	OriginRequirementIDs      []string `json:"origin_requirement_ids,omitempty"`
	DestinationRequirementIDs []string `json:"destination_requirement_ids,omitempty"`
	StopIDs                   []string `json:"stop_ids,omitempty"`
}

func newEquipmentMovementRequirementLocationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show equipment movement requirement location details",
		Long: `Show the full details of an equipment movement requirement location.

Output Fields:
  ID
  Name
  Latitude
  Longitude
  Distance (miles, when --near is used in list)
  Broker
  Origin Requirement IDs
  Destination Requirement IDs
  Stop IDs

Arguments:
  <id>    The location ID (required). You can find IDs using the list command.`,
		Example: `  # Show a location
  xbe view equipment-movement-requirement-locations show 123

  # Get JSON output
  xbe view equipment-movement-requirement-locations show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runEquipmentMovementRequirementLocationsShow,
	}
	initEquipmentMovementRequirementLocationsShowFlags(cmd)
	return cmd
}

func init() {
	equipmentMovementRequirementLocationsCmd.AddCommand(newEquipmentMovementRequirementLocationsShowCmd())
}

func initEquipmentMovementRequirementLocationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runEquipmentMovementRequirementLocationsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseEquipmentMovementRequirementLocationsShowOptions(cmd)
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
		return fmt.Errorf("equipment movement requirement location id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("include", "broker")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[equipment-movement-requirement-locations]", "name,latitude,longitude,distance-from-coordinates-miles,broker")

	body, _, err := client.Get(cmd.Context(), "/v1/equipment-movement-requirement-locations/"+id, query)
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

	details := buildEquipmentMovementRequirementLocationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderEquipmentMovementRequirementLocationDetails(cmd, details)
}

func parseEquipmentMovementRequirementLocationsShowOptions(cmd *cobra.Command) (equipmentMovementRequirementLocationsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return equipmentMovementRequirementLocationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildEquipmentMovementRequirementLocationDetails(resp jsonAPISingleResponse) equipmentMovementRequirementLocationDetails {
	details := equipmentMovementRequirementLocationDetails{
		ID:            resp.Data.ID,
		Name:          stringAttr(resp.Data.Attributes, "name"),
		Latitude:      stringAttr(resp.Data.Attributes, "latitude"),
		Longitude:     stringAttr(resp.Data.Attributes, "longitude"),
		DistanceMiles: stringAttr(resp.Data.Attributes, "distance-from-coordinates-miles"),
	}

	included := map[string]map[string]any{}
	for _, inc := range resp.Included {
		if inc.Attributes == nil {
			continue
		}
		included[resourceKey(inc.Type, inc.ID)] = inc.Attributes
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
		if attrs, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.BrokerName = stringAttr(attrs, "company-name")
		}
	}

	if rel, ok := resp.Data.Relationships["origin-equipment-movement-requirements"]; ok {
		details.OriginRequirementIDs = relationshipIDStrings(rel)
	}
	if rel, ok := resp.Data.Relationships["destination-equipment-movement-requirements"]; ok {
		details.DestinationRequirementIDs = relationshipIDStrings(rel)
	}
	if rel, ok := resp.Data.Relationships["stops"]; ok {
		details.StopIDs = relationshipIDStrings(rel)
	}

	return details
}

func relationshipIDStrings(rel jsonAPIRelationship) []string {
	refs := relationshipIDs(rel)
	if len(refs) == 0 {
		return nil
	}
	ids := make([]string, 0, len(refs))
	for _, ref := range refs {
		if ref.ID != "" {
			ids = append(ids, ref.ID)
		}
	}
	return ids
}

func renderEquipmentMovementRequirementLocationDetails(cmd *cobra.Command, details equipmentMovementRequirementLocationDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Name != "" {
		fmt.Fprintf(out, "Name: %s\n", details.Name)
	}
	if details.Latitude != "" {
		fmt.Fprintf(out, "Latitude: %s\n", details.Latitude)
	}
	if details.Longitude != "" {
		fmt.Fprintf(out, "Longitude: %s\n", details.Longitude)
	}
	if details.DistanceMiles != "" {
		fmt.Fprintf(out, "Distance (miles): %s\n", details.DistanceMiles)
	}
	if details.BrokerName != "" {
		fmt.Fprintf(out, "Broker: %s\n", details.BrokerName)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
	}
	if len(details.OriginRequirementIDs) > 0 {
		fmt.Fprintf(out, "Origin Requirement IDs: %s\n", strings.Join(details.OriginRequirementIDs, ", "))
	}
	if len(details.DestinationRequirementIDs) > 0 {
		fmt.Fprintf(out, "Destination Requirement IDs: %s\n", strings.Join(details.DestinationRequirementIDs, ", "))
	}
	if len(details.StopIDs) > 0 {
		fmt.Fprintf(out, "Stop IDs: %s\n", strings.Join(details.StopIDs, ", "))
	}

	return nil
}
