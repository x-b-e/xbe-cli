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

type tennaVehiclesListOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	NoAuth                 bool
	Limit                  int
	Offset                 int
	Sort                   string
	Broker                 string
	Trucker                string
	Tractor                string
	Trailer                string
	Equipment              string
	HasTractor             string
	HasTrailer             string
	HasEquipment           string
	AssignedAtMin          string
	EquipmentAssignedAtMin string
	IntegrationIdentifier  string
	TrailerSetAtMin        string
	TrailerSetAtMax        string
	IsTrailerSetAt         string
	TractorSetAtMin        string
	TractorSetAtMax        string
	IsTractorSetAt         string
	EquipmentSetAtMin      string
	EquipmentSetAtMax      string
	IsEquipmentSetAt       string
	CreatedAtMin           string
	CreatedAtMax           string
	IsCreatedAt            string
	UpdatedAtMin           string
	UpdatedAtMax           string
	IsUpdatedAt            string
}

type tennaVehicleRow struct {
	ID                      string `json:"id"`
	VehicleID               string `json:"vehicle_id,omitempty"`
	VehicleNumber           string `json:"vehicle_number,omitempty"`
	SerialNumber            string `json:"serial_number,omitempty"`
	IntegrationIdentifier   string `json:"integration_identifier,omitempty"`
	TrailerSetAt            string `json:"trailer_set_at,omitempty"`
	TractorSetAt            string `json:"tractor_set_at,omitempty"`
	EquipmentSetAt          string `json:"equipment_set_at,omitempty"`
	BrokerID                string `json:"broker_id,omitempty"`
	BrokerName              string `json:"broker_name,omitempty"`
	TruckerID               string `json:"trucker_id,omitempty"`
	TruckerName             string `json:"trucker_name,omitempty"`
	TrailerID               string `json:"trailer_id,omitempty"`
	TrailerNumber           string `json:"trailer_number,omitempty"`
	TractorID               string `json:"tractor_id,omitempty"`
	TractorNumber           string `json:"tractor_number,omitempty"`
	EquipmentID             string `json:"equipment_id,omitempty"`
	EquipmentNickname       string `json:"equipment_nickname,omitempty"`
	AssignedEquipmentSerial string `json:"assigned_equipment_serial_number,omitempty"`
}

func newTennaVehiclesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Tenna vehicles",
		Long: `List Tenna vehicles with filtering and pagination.

Output Columns:
  ID         Tenna vehicle identifier
  VEHICLE #  Tenna vehicle number
  VEHICLE ID Tenna vehicle source identifier
  SERIAL     Tenna vehicle serial number
  TRACTOR    Assigned tractor number or ID
  TRAILER    Assigned trailer number or ID
  EQUIPMENT  Assigned equipment nickname, serial number, or ID
  TRUCKER    Trucker name or ID
  BROKER     Broker name or ID

Filters:
  --broker                   Filter by broker ID
  --trucker                  Filter by trucker ID
  --tractor                  Filter by tractor ID
  --trailer                  Filter by trailer ID
  --equipment                Filter by equipment ID
  --has-tractor              Filter by tractor assignment (true/false)
  --has-trailer              Filter by trailer assignment (true/false)
  --has-equipment            Filter by equipment assignment (true/false)
  --assigned-at-min          Filter by assigned-at timestamp (ISO 8601)
  --equipment-assigned-at-min Filter by equipment assigned-at timestamp (ISO 8601)
  --integration-identifier   Filter by integration identifier
  --trailer-set-at-min       Filter by minimum trailer set timestamp (ISO 8601)
  --trailer-set-at-max       Filter by maximum trailer set timestamp (ISO 8601)
  --is-trailer-set-at        Filter by has trailer set timestamp (true/false)
  --tractor-set-at-min       Filter by minimum tractor set timestamp (ISO 8601)
  --tractor-set-at-max       Filter by maximum tractor set timestamp (ISO 8601)
  --is-tractor-set-at        Filter by has tractor set timestamp (true/false)
  --equipment-set-at-min     Filter by minimum equipment set timestamp (ISO 8601)
  --equipment-set-at-max     Filter by maximum equipment set timestamp (ISO 8601)
  --is-equipment-set-at      Filter by has equipment set timestamp (true/false)
  --created-at-min           Filter by created-at on/after (ISO 8601)
  --created-at-max           Filter by created-at on/before (ISO 8601)
  --is-created-at            Filter by has created-at (true/false)
  --updated-at-min           Filter by updated-at on/after (ISO 8601)
  --updated-at-max           Filter by updated-at on/before (ISO 8601)
  --is-updated-at            Filter by has updated-at (true/false)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List Tenna vehicles
  xbe view tenna-vehicles list

  # Filter by broker
  xbe view tenna-vehicles list --broker 123

  # Filter by equipment assignment
  xbe view tenna-vehicles list --has-equipment true

  # Output as JSON
  xbe view tenna-vehicles list --json`,
		Args: cobra.NoArgs,
		RunE: runTennaVehiclesList,
	}
	initTennaVehiclesListFlags(cmd)
	return cmd
}

func init() {
	tennaVehiclesCmd.AddCommand(newTennaVehiclesListCmd())
}

func initTennaVehiclesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("trucker", "", "Filter by trucker ID")
	cmd.Flags().String("tractor", "", "Filter by tractor ID")
	cmd.Flags().String("trailer", "", "Filter by trailer ID")
	cmd.Flags().String("equipment", "", "Filter by equipment ID")
	cmd.Flags().String("has-tractor", "", "Filter by tractor assignment (true/false)")
	cmd.Flags().String("has-trailer", "", "Filter by trailer assignment (true/false)")
	cmd.Flags().String("has-equipment", "", "Filter by equipment assignment (true/false)")
	cmd.Flags().String("assigned-at-min", "", "Filter by assigned-at timestamp (ISO 8601)")
	cmd.Flags().String("equipment-assigned-at-min", "", "Filter by equipment assigned-at timestamp (ISO 8601)")
	cmd.Flags().String("integration-identifier", "", "Filter by integration identifier")
	cmd.Flags().String("trailer-set-at-min", "", "Filter by minimum trailer set timestamp (ISO 8601)")
	cmd.Flags().String("trailer-set-at-max", "", "Filter by maximum trailer set timestamp (ISO 8601)")
	cmd.Flags().String("is-trailer-set-at", "", "Filter by has trailer set timestamp (true/false)")
	cmd.Flags().String("tractor-set-at-min", "", "Filter by minimum tractor set timestamp (ISO 8601)")
	cmd.Flags().String("tractor-set-at-max", "", "Filter by maximum tractor set timestamp (ISO 8601)")
	cmd.Flags().String("is-tractor-set-at", "", "Filter by has tractor set timestamp (true/false)")
	cmd.Flags().String("equipment-set-at-min", "", "Filter by minimum equipment set timestamp (ISO 8601)")
	cmd.Flags().String("equipment-set-at-max", "", "Filter by maximum equipment set timestamp (ISO 8601)")
	cmd.Flags().String("is-equipment-set-at", "", "Filter by has equipment set timestamp (true/false)")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by has created-at (true/false)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("is-updated-at", "", "Filter by has updated-at (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTennaVehiclesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTennaVehiclesListOptions(cmd)
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
	query.Set("fields[tenna-vehicles]", strings.Join([]string{
		"vehicle-id",
		"vehicle-number",
		"serial-number",
		"integration-identifier",
		"trailer-set-at",
		"tractor-set-at",
		"equipment-set-at",
		"broker",
		"trucker",
		"tractor",
		"trailer",
		"equipment",
	}, ","))
	query.Set("include", "broker,trucker,tractor,trailer,equipment")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[truckers]", "company-name")
	query.Set("fields[tractors]", "number")
	query.Set("fields[trailers]", "number")
	query.Set("fields[equipment]", "nickname,serial-number")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)
	setFilterIfPresent(query, "filter[tractor]", opts.Tractor)
	setFilterIfPresent(query, "filter[trailer]", opts.Trailer)
	setFilterIfPresent(query, "filter[equipment]", opts.Equipment)
	setFilterIfPresent(query, "filter[has_tractor]", opts.HasTractor)
	setFilterIfPresent(query, "filter[has_trailer]", opts.HasTrailer)
	setFilterIfPresent(query, "filter[has_equipment]", opts.HasEquipment)
	setFilterIfPresent(query, "filter[assigned_at_min]", opts.AssignedAtMin)
	setFilterIfPresent(query, "filter[equipment_assigned_at_min]", opts.EquipmentAssignedAtMin)
	setFilterIfPresent(query, "filter[integration_identifier]", opts.IntegrationIdentifier)
	setFilterIfPresent(query, "filter[trailer_set_at_min]", opts.TrailerSetAtMin)
	setFilterIfPresent(query, "filter[trailer_set_at_max]", opts.TrailerSetAtMax)
	setFilterIfPresent(query, "filter[is_trailer_set_at]", opts.IsTrailerSetAt)
	setFilterIfPresent(query, "filter[tractor_set_at_min]", opts.TractorSetAtMin)
	setFilterIfPresent(query, "filter[tractor_set_at_max]", opts.TractorSetAtMax)
	setFilterIfPresent(query, "filter[is_tractor_set_at]", opts.IsTractorSetAt)
	setFilterIfPresent(query, "filter[equipment_set_at_min]", opts.EquipmentSetAtMin)
	setFilterIfPresent(query, "filter[equipment_set_at_max]", opts.EquipmentSetAtMax)
	setFilterIfPresent(query, "filter[is_equipment_set_at]", opts.IsEquipmentSetAt)
	setFilterIfPresent(query, "filter[created_at_min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created_at_max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[is_created_at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[updated_at_min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated_at_max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[is_updated_at]", opts.IsUpdatedAt)

	body, _, err := client.Get(cmd.Context(), "/v1/tenna-vehicles", query)
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

	rows := buildTennaVehicleRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTennaVehiclesTable(cmd, rows)
}

func parseTennaVehiclesListOptions(cmd *cobra.Command) (tennaVehiclesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	broker, _ := cmd.Flags().GetString("broker")
	trucker, _ := cmd.Flags().GetString("trucker")
	tractor, _ := cmd.Flags().GetString("tractor")
	trailer, _ := cmd.Flags().GetString("trailer")
	equipment, _ := cmd.Flags().GetString("equipment")
	hasTractor, _ := cmd.Flags().GetString("has-tractor")
	hasTrailer, _ := cmd.Flags().GetString("has-trailer")
	hasEquipment, _ := cmd.Flags().GetString("has-equipment")
	assignedAtMin, _ := cmd.Flags().GetString("assigned-at-min")
	equipmentAssignedAtMin, _ := cmd.Flags().GetString("equipment-assigned-at-min")
	integrationIdentifier, _ := cmd.Flags().GetString("integration-identifier")
	trailerSetAtMin, _ := cmd.Flags().GetString("trailer-set-at-min")
	trailerSetAtMax, _ := cmd.Flags().GetString("trailer-set-at-max")
	isTrailerSetAt, _ := cmd.Flags().GetString("is-trailer-set-at")
	tractorSetAtMin, _ := cmd.Flags().GetString("tractor-set-at-min")
	tractorSetAtMax, _ := cmd.Flags().GetString("tractor-set-at-max")
	isTractorSetAt, _ := cmd.Flags().GetString("is-tractor-set-at")
	equipmentSetAtMin, _ := cmd.Flags().GetString("equipment-set-at-min")
	equipmentSetAtMax, _ := cmd.Flags().GetString("equipment-set-at-max")
	isEquipmentSetAt, _ := cmd.Flags().GetString("is-equipment-set-at")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	isCreatedAt, _ := cmd.Flags().GetString("is-created-at")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	isUpdatedAt, _ := cmd.Flags().GetString("is-updated-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return tennaVehiclesListOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		NoAuth:                 noAuth,
		Limit:                  limit,
		Offset:                 offset,
		Sort:                   sort,
		Broker:                 broker,
		Trucker:                trucker,
		Tractor:                tractor,
		Trailer:                trailer,
		Equipment:              equipment,
		HasTractor:             hasTractor,
		HasTrailer:             hasTrailer,
		HasEquipment:           hasEquipment,
		AssignedAtMin:          assignedAtMin,
		EquipmentAssignedAtMin: equipmentAssignedAtMin,
		IntegrationIdentifier:  integrationIdentifier,
		TrailerSetAtMin:        trailerSetAtMin,
		TrailerSetAtMax:        trailerSetAtMax,
		IsTrailerSetAt:         isTrailerSetAt,
		TractorSetAtMin:        tractorSetAtMin,
		TractorSetAtMax:        tractorSetAtMax,
		IsTractorSetAt:         isTractorSetAt,
		EquipmentSetAtMin:      equipmentSetAtMin,
		EquipmentSetAtMax:      equipmentSetAtMax,
		IsEquipmentSetAt:       isEquipmentSetAt,
		CreatedAtMin:           createdAtMin,
		CreatedAtMax:           createdAtMax,
		IsCreatedAt:            isCreatedAt,
		UpdatedAtMin:           updatedAtMin,
		UpdatedAtMax:           updatedAtMax,
		IsUpdatedAt:            isUpdatedAt,
	}, nil
}

func buildTennaVehicleRows(resp jsonAPIResponse) []tennaVehicleRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]tennaVehicleRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildTennaVehicleRow(resource, included))
	}
	return rows
}

func tennaVehicleRowFromSingle(resp jsonAPISingleResponse) tennaVehicleRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}
	return buildTennaVehicleRow(resp.Data, included)
}

func buildTennaVehicleRow(resource jsonAPIResource, included map[string]jsonAPIResource) tennaVehicleRow {
	attrs := resource.Attributes
	row := tennaVehicleRow{
		ID:                    resource.ID,
		VehicleID:             stringAttr(attrs, "vehicle-id"),
		VehicleNumber:         stringAttr(attrs, "vehicle-number"),
		SerialNumber:          stringAttr(attrs, "serial-number"),
		IntegrationIdentifier: stringAttr(attrs, "integration-identifier"),
		TrailerSetAt:          formatDateTime(stringAttr(attrs, "trailer-set-at")),
		TractorSetAt:          formatDateTime(stringAttr(attrs, "tractor-set-at")),
		EquipmentSetAt:        formatDateTime(stringAttr(attrs, "equipment-set-at")),
	}

	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
		if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.BrokerName = stringAttr(broker.Attributes, "company-name")
		}
	}

	if rel, ok := resource.Relationships["trucker"]; ok && rel.Data != nil {
		row.TruckerID = rel.Data.ID
		if trucker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.TruckerName = stringAttr(trucker.Attributes, "company-name")
		}
	}

	if rel, ok := resource.Relationships["trailer"]; ok && rel.Data != nil {
		row.TrailerID = rel.Data.ID
		if trailer, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.TrailerNumber = stringAttr(trailer.Attributes, "number")
		}
	}

	if rel, ok := resource.Relationships["tractor"]; ok && rel.Data != nil {
		row.TractorID = rel.Data.ID
		if tractor, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.TractorNumber = stringAttr(tractor.Attributes, "number")
		}
	}

	if rel, ok := resource.Relationships["equipment"]; ok && rel.Data != nil {
		row.EquipmentID = rel.Data.ID
		if equipment, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.EquipmentNickname = stringAttr(equipment.Attributes, "nickname")
			row.AssignedEquipmentSerial = stringAttr(equipment.Attributes, "serial-number")
		}
	}

	return row
}

func renderTennaVehiclesTable(cmd *cobra.Command, rows []tennaVehicleRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No Tenna vehicles found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tVEHICLE #\tVEHICLE ID\tSERIAL\tTRACTOR\tTRAILER\tEQUIPMENT\tTRUCKER\tBROKER")
	for _, row := range rows {
		tractor := firstNonEmpty(row.TractorNumber, row.TractorID)
		trailer := firstNonEmpty(row.TrailerNumber, row.TrailerID)
		equipment := firstNonEmpty(row.EquipmentNickname, row.AssignedEquipmentSerial, row.EquipmentID)
		trucker := firstNonEmpty(row.TruckerName, row.TruckerID)
		broker := firstNonEmpty(row.BrokerName, row.BrokerID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.VehicleNumber, 18),
			truncateString(row.VehicleID, 14),
			truncateString(row.SerialNumber, 16),
			truncateString(tractor, 18),
			truncateString(trailer, 18),
			truncateString(equipment, 20),
			truncateString(trucker, 20),
			truncateString(broker, 20),
		)
	}
	return writer.Flush()
}
