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

type verizonRevealVehiclesListOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	NoAuth                 bool
	Limit                  int
	Offset                 int
	Sort                   string
	Broker                 string
	Trucker                string
	Trailer                string
	Tractor                string
	Equipment              string
	HasTrailer             string
	HasTractor             string
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

type verizonRevealVehicleRow struct {
	ID                    string `json:"id"`
	VehicleID             string `json:"vehicle_id,omitempty"`
	VehicleNumber         string `json:"vehicle_number,omitempty"`
	IntegrationIdentifier string `json:"integration_identifier,omitempty"`
	TrailerSetAt          string `json:"trailer_set_at,omitempty"`
	TractorSetAt          string `json:"tractor_set_at,omitempty"`
	EquipmentSetAt        string `json:"equipment_set_at,omitempty"`
	BrokerID              string `json:"broker_id,omitempty"`
	TruckerID             string `json:"trucker_id,omitempty"`
	TrailerID             string `json:"trailer_id,omitempty"`
	TractorID             string `json:"tractor_id,omitempty"`
	EquipmentID           string `json:"equipment_id,omitempty"`
}

func newVerizonRevealVehiclesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Verizon Reveal vehicles",
		Long: `List Verizon Reveal vehicles with filtering and pagination.

Output Columns:
  ID               Verizon Reveal vehicle identifier
  VEHICLE NUMBER   Verizon Reveal vehicle number
  VEHICLE ID       Verizon Reveal vehicle external ID
  TRUCKER          Trucker ID
  TRAILER          Trailer ID
  TRACTOR          Tractor ID
  EQUIPMENT        Equipment ID
  INTEGRATION ID   Integration identifier

Filters:
  --broker                    Filter by broker ID
  --trucker                   Filter by trucker ID
  --trailer                   Filter by trailer ID
  --tractor                   Filter by tractor ID
  --equipment                 Filter by equipment ID
  --has-trailer               Filter by presence of trailer (true/false)
  --has-tractor               Filter by presence of tractor (true/false)
  --has-equipment             Filter by presence of equipment (true/false)
  --assigned-at-min           Filter by minimum assignment time (ISO 8601)
  --equipment-assigned-at-min Filter by minimum equipment assignment time (ISO 8601)
  --integration-identifier    Filter by integration identifier
  --trailer-set-at-min        Filter by trailer set on/after (ISO 8601)
  --trailer-set-at-max        Filter by trailer set on/before (ISO 8601)
  --is-trailer-set-at         Filter by presence of trailer-set-at (true/false)
  --tractor-set-at-min        Filter by tractor set on/after (ISO 8601)
  --tractor-set-at-max        Filter by tractor set on/before (ISO 8601)
  --is-tractor-set-at         Filter by presence of tractor-set-at (true/false)
  --equipment-set-at-min      Filter by equipment set on/after (ISO 8601)
  --equipment-set-at-max      Filter by equipment set on/before (ISO 8601)
  --is-equipment-set-at       Filter by presence of equipment-set-at (true/false)
  --created-at-min            Filter by created-at on/after (ISO 8601)
  --created-at-max            Filter by created-at on/before (ISO 8601)
  --is-created-at             Filter by presence of created-at (true/false)
  --updated-at-min            Filter by updated-at on/after (ISO 8601)
  --updated-at-max            Filter by updated-at on/before (ISO 8601)
  --is-updated-at             Filter by presence of updated-at (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List Verizon Reveal vehicles
  xbe view verizon-reveal-vehicles list

  # Filter by broker
  xbe view verizon-reveal-vehicles list --broker 123

  # Filter by assignment time
  xbe view verizon-reveal-vehicles list --assigned-at-min "2024-01-01T00:00:00Z"

  # Output as JSON
  xbe view verizon-reveal-vehicles list --json`,
		Args: cobra.NoArgs,
		RunE: runVerizonRevealVehiclesList,
	}
	initVerizonRevealVehiclesListFlags(cmd)
	return cmd
}

func init() {
	verizonRevealVehiclesCmd.AddCommand(newVerizonRevealVehiclesListCmd())
}

func initVerizonRevealVehiclesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("trucker", "", "Filter by trucker ID")
	cmd.Flags().String("trailer", "", "Filter by trailer ID")
	cmd.Flags().String("tractor", "", "Filter by tractor ID")
	cmd.Flags().String("equipment", "", "Filter by equipment ID")
	cmd.Flags().String("has-trailer", "", "Filter by presence of trailer (true/false)")
	cmd.Flags().String("has-tractor", "", "Filter by presence of tractor (true/false)")
	cmd.Flags().String("has-equipment", "", "Filter by presence of equipment (true/false)")
	cmd.Flags().String("assigned-at-min", "", "Filter by minimum assignment time (ISO 8601)")
	cmd.Flags().String("equipment-assigned-at-min", "", "Filter by minimum equipment assignment time (ISO 8601)")
	cmd.Flags().String("integration-identifier", "", "Filter by integration identifier")
	cmd.Flags().String("trailer-set-at-min", "", "Filter by trailer set on/after (ISO 8601)")
	cmd.Flags().String("trailer-set-at-max", "", "Filter by trailer set on/before (ISO 8601)")
	cmd.Flags().String("is-trailer-set-at", "", "Filter by presence of trailer-set-at (true/false)")
	cmd.Flags().String("tractor-set-at-min", "", "Filter by tractor set on/after (ISO 8601)")
	cmd.Flags().String("tractor-set-at-max", "", "Filter by tractor set on/before (ISO 8601)")
	cmd.Flags().String("is-tractor-set-at", "", "Filter by presence of tractor-set-at (true/false)")
	cmd.Flags().String("equipment-set-at-min", "", "Filter by equipment set on/after (ISO 8601)")
	cmd.Flags().String("equipment-set-at-max", "", "Filter by equipment set on/before (ISO 8601)")
	cmd.Flags().String("is-equipment-set-at", "", "Filter by presence of equipment-set-at (true/false)")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by presence of created-at (true/false)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("is-updated-at", "", "Filter by presence of updated-at (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runVerizonRevealVehiclesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseVerizonRevealVehiclesListOptions(cmd)
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
	query.Set("fields[verizon-reveal-vehicles]", "vehicle-id,vehicle-number,integration-identifier,trailer-set-at,tractor-set-at,equipment-set-at,broker,trucker,trailer,tractor,equipment")
	query.Set("include", "broker,trucker,trailer,tractor,equipment")

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
	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)
	setFilterIfPresent(query, "filter[trailer]", opts.Trailer)
	setFilterIfPresent(query, "filter[tractor]", opts.Tractor)
	setFilterIfPresent(query, "filter[equipment]", opts.Equipment)
	setFilterIfPresent(query, "filter[has-trailer]", opts.HasTrailer)
	setFilterIfPresent(query, "filter[has-tractor]", opts.HasTractor)
	setFilterIfPresent(query, "filter[has-equipment]", opts.HasEquipment)
	setFilterIfPresent(query, "filter[assigned-at-min]", opts.AssignedAtMin)
	setFilterIfPresent(query, "filter[equipment-assigned-at-min]", opts.EquipmentAssignedAtMin)
	setFilterIfPresent(query, "filter[integration-identifier]", opts.IntegrationIdentifier)
	setFilterIfPresent(query, "filter[trailer-set-at-min]", opts.TrailerSetAtMin)
	setFilterIfPresent(query, "filter[trailer-set-at-max]", opts.TrailerSetAtMax)
	setFilterIfPresent(query, "filter[is-trailer-set-at]", opts.IsTrailerSetAt)
	setFilterIfPresent(query, "filter[tractor-set-at-min]", opts.TractorSetAtMin)
	setFilterIfPresent(query, "filter[tractor-set-at-max]", opts.TractorSetAtMax)
	setFilterIfPresent(query, "filter[is-tractor-set-at]", opts.IsTractorSetAt)
	setFilterIfPresent(query, "filter[equipment-set-at-min]", opts.EquipmentSetAtMin)
	setFilterIfPresent(query, "filter[equipment-set-at-max]", opts.EquipmentSetAtMax)
	setFilterIfPresent(query, "filter[is-equipment-set-at]", opts.IsEquipmentSetAt)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[is-created-at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[is-updated-at]", opts.IsUpdatedAt)

	body, _, err := client.Get(cmd.Context(), "/v1/verizon-reveal-vehicles", query)
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

	rows := buildVerizonRevealVehicleRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderVerizonRevealVehiclesTable(cmd, rows)
}

func parseVerizonRevealVehiclesListOptions(cmd *cobra.Command) (verizonRevealVehiclesListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return verizonRevealVehiclesListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return verizonRevealVehiclesListOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return verizonRevealVehiclesListOptions{}, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return verizonRevealVehiclesListOptions{}, err
	}
	sort, err := cmd.Flags().GetString("sort")
	if err != nil {
		return verizonRevealVehiclesListOptions{}, err
	}
	broker, err := cmd.Flags().GetString("broker")
	if err != nil {
		return verizonRevealVehiclesListOptions{}, err
	}
	trucker, err := cmd.Flags().GetString("trucker")
	if err != nil {
		return verizonRevealVehiclesListOptions{}, err
	}
	trailer, err := cmd.Flags().GetString("trailer")
	if err != nil {
		return verizonRevealVehiclesListOptions{}, err
	}
	tractor, err := cmd.Flags().GetString("tractor")
	if err != nil {
		return verizonRevealVehiclesListOptions{}, err
	}
	equipment, err := cmd.Flags().GetString("equipment")
	if err != nil {
		return verizonRevealVehiclesListOptions{}, err
	}
	hasTrailer, err := cmd.Flags().GetString("has-trailer")
	if err != nil {
		return verizonRevealVehiclesListOptions{}, err
	}
	hasTractor, err := cmd.Flags().GetString("has-tractor")
	if err != nil {
		return verizonRevealVehiclesListOptions{}, err
	}
	hasEquipment, err := cmd.Flags().GetString("has-equipment")
	if err != nil {
		return verizonRevealVehiclesListOptions{}, err
	}
	assignedAtMin, err := cmd.Flags().GetString("assigned-at-min")
	if err != nil {
		return verizonRevealVehiclesListOptions{}, err
	}
	equipmentAssignedAtMin, err := cmd.Flags().GetString("equipment-assigned-at-min")
	if err != nil {
		return verizonRevealVehiclesListOptions{}, err
	}
	integrationIdentifier, err := cmd.Flags().GetString("integration-identifier")
	if err != nil {
		return verizonRevealVehiclesListOptions{}, err
	}
	trailerSetAtMin, err := cmd.Flags().GetString("trailer-set-at-min")
	if err != nil {
		return verizonRevealVehiclesListOptions{}, err
	}
	trailerSetAtMax, err := cmd.Flags().GetString("trailer-set-at-max")
	if err != nil {
		return verizonRevealVehiclesListOptions{}, err
	}
	isTrailerSetAt, err := cmd.Flags().GetString("is-trailer-set-at")
	if err != nil {
		return verizonRevealVehiclesListOptions{}, err
	}
	tractorSetAtMin, err := cmd.Flags().GetString("tractor-set-at-min")
	if err != nil {
		return verizonRevealVehiclesListOptions{}, err
	}
	tractorSetAtMax, err := cmd.Flags().GetString("tractor-set-at-max")
	if err != nil {
		return verizonRevealVehiclesListOptions{}, err
	}
	isTractorSetAt, err := cmd.Flags().GetString("is-tractor-set-at")
	if err != nil {
		return verizonRevealVehiclesListOptions{}, err
	}
	equipmentSetAtMin, err := cmd.Flags().GetString("equipment-set-at-min")
	if err != nil {
		return verizonRevealVehiclesListOptions{}, err
	}
	equipmentSetAtMax, err := cmd.Flags().GetString("equipment-set-at-max")
	if err != nil {
		return verizonRevealVehiclesListOptions{}, err
	}
	isEquipmentSetAt, err := cmd.Flags().GetString("is-equipment-set-at")
	if err != nil {
		return verizonRevealVehiclesListOptions{}, err
	}
	createdAtMin, err := cmd.Flags().GetString("created-at-min")
	if err != nil {
		return verizonRevealVehiclesListOptions{}, err
	}
	createdAtMax, err := cmd.Flags().GetString("created-at-max")
	if err != nil {
		return verizonRevealVehiclesListOptions{}, err
	}
	isCreatedAt, err := cmd.Flags().GetString("is-created-at")
	if err != nil {
		return verizonRevealVehiclesListOptions{}, err
	}
	updatedAtMin, err := cmd.Flags().GetString("updated-at-min")
	if err != nil {
		return verizonRevealVehiclesListOptions{}, err
	}
	updatedAtMax, err := cmd.Flags().GetString("updated-at-max")
	if err != nil {
		return verizonRevealVehiclesListOptions{}, err
	}
	isUpdatedAt, err := cmd.Flags().GetString("is-updated-at")
	if err != nil {
		return verizonRevealVehiclesListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return verizonRevealVehiclesListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return verizonRevealVehiclesListOptions{}, err
	}

	return verizonRevealVehiclesListOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		NoAuth:                 noAuth,
		Limit:                  limit,
		Offset:                 offset,
		Sort:                   sort,
		Broker:                 broker,
		Trucker:                trucker,
		Trailer:                trailer,
		Tractor:                tractor,
		Equipment:              equipment,
		HasTrailer:             hasTrailer,
		HasTractor:             hasTractor,
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

func buildVerizonRevealVehicleRows(resp jsonAPIResponse) []verizonRevealVehicleRow {
	rows := make([]verizonRevealVehicleRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := verizonRevealVehicleRow{
			ID:                    resource.ID,
			VehicleID:             stringAttr(resource.Attributes, "vehicle-id"),
			VehicleNumber:         stringAttr(resource.Attributes, "vehicle-number"),
			IntegrationIdentifier: stringAttr(resource.Attributes, "integration-identifier"),
			TrailerSetAt:          formatDateTime(stringAttr(resource.Attributes, "trailer-set-at")),
			TractorSetAt:          formatDateTime(stringAttr(resource.Attributes, "tractor-set-at")),
			EquipmentSetAt:        formatDateTime(stringAttr(resource.Attributes, "equipment-set-at")),
			BrokerID:              relationshipIDFromMap(resource.Relationships, "broker"),
			TruckerID:             relationshipIDFromMap(resource.Relationships, "trucker"),
			TrailerID:             relationshipIDFromMap(resource.Relationships, "trailer"),
			TractorID:             relationshipIDFromMap(resource.Relationships, "tractor"),
			EquipmentID:           relationshipIDFromMap(resource.Relationships, "equipment"),
		}

		rows = append(rows, row)
	}
	return rows
}

func renderVerizonRevealVehiclesTable(cmd *cobra.Command, rows []verizonRevealVehicleRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No Verizon Reveal vehicles found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tVEHICLE NUMBER\tVEHICLE ID\tTRUCKER\tTRAILER\tTRACTOR\tEQUIPMENT\tINTEGRATION ID")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.VehicleNumber,
			row.VehicleID,
			row.TruckerID,
			row.TrailerID,
			row.TractorID,
			row.EquipmentID,
			row.IntegrationIdentifier,
		)
	}
	return writer.Flush()
}
