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

type t3EquipmentshareVehiclesListOptions struct {
	BaseURL               string
	Token                 string
	JSON                  bool
	NoAuth                bool
	Limit                 int
	Offset                int
	Sort                  string
	Broker                string
	Trucker               string
	Trailer               string
	Tractor               string
	HasTrailer            string
	HasTractor            string
	AssignedAtMin         string
	IntegrationIdentifier string
	TrailerSetAtMin       string
	TrailerSetAtMax       string
	IsTrailerSetAt        string
	TractorSetAtMin       string
	TractorSetAtMax       string
	IsTractorSetAt        string
	CreatedAtMin          string
	CreatedAtMax          string
	IsCreatedAt           string
	UpdatedAtMin          string
	UpdatedAtMax          string
	IsUpdatedAt           string
}

type t3EquipmentshareVehicleRow struct {
	ID                    string `json:"id"`
	VehicleID             string `json:"vehicle_id,omitempty"`
	VehicleNumber         string `json:"vehicle_number,omitempty"`
	IntegrationIdentifier string `json:"integration_identifier,omitempty"`
	TrailerSetAt          string `json:"trailer_set_at,omitempty"`
	TractorSetAt          string `json:"tractor_set_at,omitempty"`
	BrokerID              string `json:"broker_id,omitempty"`
	TruckerID             string `json:"trucker_id,omitempty"`
	TrailerID             string `json:"trailer_id,omitempty"`
	TractorID             string `json:"tractor_id,omitempty"`
}

func newT3EquipmentshareVehiclesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List T3 EquipmentShare vehicles",
		Long: `List T3 EquipmentShare vehicles with filtering and pagination.

Output Columns:
  ID               T3 EquipmentShare vehicle identifier
  VEHICLE NUMBER   T3 EquipmentShare vehicle number
  VEHICLE ID       T3 EquipmentShare vehicle external ID
  TRUCKER          Trucker ID
  TRAILER          Trailer ID
  TRACTOR          Tractor ID
  INTEGRATION ID   Integration identifier

Filters:
  --broker                 Filter by broker ID
  --trucker                Filter by trucker ID
  --trailer                Filter by trailer ID
  --tractor                Filter by tractor ID
  --has-trailer            Filter by presence of trailer (true/false)
  --has-tractor            Filter by presence of tractor (true/false)
  --assigned-at-min        Filter by minimum assignment time (ISO 8601)
  --integration-identifier Filter by integration identifier
  --trailer-set-at-min     Filter by trailer set on/after (ISO 8601)
  --trailer-set-at-max     Filter by trailer set on/before (ISO 8601)
  --is-trailer-set-at      Filter by presence of trailer-set-at (true/false)
  --tractor-set-at-min     Filter by tractor set on/after (ISO 8601)
  --tractor-set-at-max     Filter by tractor set on/before (ISO 8601)
  --is-tractor-set-at      Filter by presence of tractor-set-at (true/false)
  --created-at-min         Filter by created-at on/after (ISO 8601)
  --created-at-max         Filter by created-at on/before (ISO 8601)
  --is-created-at          Filter by presence of created-at (true/false)
  --updated-at-min         Filter by updated-at on/after (ISO 8601)
  --updated-at-max         Filter by updated-at on/before (ISO 8601)
  --is-updated-at          Filter by presence of updated-at (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List T3 EquipmentShare vehicles
  xbe view t3-equipmentshare-vehicles list

  # Filter by broker
  xbe view t3-equipmentshare-vehicles list --broker 123

  # Filter by assignment time
  xbe view t3-equipmentshare-vehicles list --assigned-at-min "2024-01-01T00:00:00Z"

  # Output as JSON
  xbe view t3-equipmentshare-vehicles list --json`,
		Args: cobra.NoArgs,
		RunE: runT3EquipmentshareVehiclesList,
	}
	initT3EquipmentshareVehiclesListFlags(cmd)
	return cmd
}

func init() {
	t3EquipmentshareVehiclesCmd.AddCommand(newT3EquipmentshareVehiclesListCmd())
}

func initT3EquipmentshareVehiclesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("trucker", "", "Filter by trucker ID")
	cmd.Flags().String("trailer", "", "Filter by trailer ID")
	cmd.Flags().String("tractor", "", "Filter by tractor ID")
	cmd.Flags().String("has-trailer", "", "Filter by presence of trailer (true/false)")
	cmd.Flags().String("has-tractor", "", "Filter by presence of tractor (true/false)")
	cmd.Flags().String("assigned-at-min", "", "Filter by minimum assignment time (ISO 8601)")
	cmd.Flags().String("integration-identifier", "", "Filter by integration identifier")
	cmd.Flags().String("trailer-set-at-min", "", "Filter by trailer set on/after (ISO 8601)")
	cmd.Flags().String("trailer-set-at-max", "", "Filter by trailer set on/before (ISO 8601)")
	cmd.Flags().String("is-trailer-set-at", "", "Filter by presence of trailer-set-at (true/false)")
	cmd.Flags().String("tractor-set-at-min", "", "Filter by tractor set on/after (ISO 8601)")
	cmd.Flags().String("tractor-set-at-max", "", "Filter by tractor set on/before (ISO 8601)")
	cmd.Flags().String("is-tractor-set-at", "", "Filter by presence of tractor-set-at (true/false)")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by presence of created-at (true/false)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("is-updated-at", "", "Filter by presence of updated-at (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runT3EquipmentshareVehiclesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseT3EquipmentshareVehiclesListOptions(cmd)
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
	query.Set("fields[t3-equipmentshare-vehicles]", "vehicle-id,vehicle-number,integration-identifier,trailer-set-at,tractor-set-at,broker,trucker,trailer,tractor")
	query.Set("include", "broker,trucker,trailer,tractor")

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
	setFilterIfPresent(query, "filter[has-trailer]", opts.HasTrailer)
	setFilterIfPresent(query, "filter[has-tractor]", opts.HasTractor)
	setFilterIfPresent(query, "filter[assigned-at-min]", opts.AssignedAtMin)
	setFilterIfPresent(query, "filter[integration-identifier]", opts.IntegrationIdentifier)
	setFilterIfPresent(query, "filter[trailer-set-at-min]", opts.TrailerSetAtMin)
	setFilterIfPresent(query, "filter[trailer-set-at-max]", opts.TrailerSetAtMax)
	setFilterIfPresent(query, "filter[is-trailer-set-at]", opts.IsTrailerSetAt)
	setFilterIfPresent(query, "filter[tractor-set-at-min]", opts.TractorSetAtMin)
	setFilterIfPresent(query, "filter[tractor-set-at-max]", opts.TractorSetAtMax)
	setFilterIfPresent(query, "filter[is-tractor-set-at]", opts.IsTractorSetAt)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[is-created-at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[is-updated-at]", opts.IsUpdatedAt)

	body, _, err := client.Get(cmd.Context(), "/v1/t3-equipmentshare-vehicles", query)
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

	rows := buildT3EquipmentshareVehicleRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderT3EquipmentshareVehiclesTable(cmd, rows)
}

func parseT3EquipmentshareVehiclesListOptions(cmd *cobra.Command) (t3EquipmentshareVehiclesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	broker, _ := cmd.Flags().GetString("broker")
	trucker, _ := cmd.Flags().GetString("trucker")
	trailer, _ := cmd.Flags().GetString("trailer")
	tractor, _ := cmd.Flags().GetString("tractor")
	hasTrailer, _ := cmd.Flags().GetString("has-trailer")
	hasTractor, _ := cmd.Flags().GetString("has-tractor")
	assignedAtMin, _ := cmd.Flags().GetString("assigned-at-min")
	integrationIdentifier, _ := cmd.Flags().GetString("integration-identifier")
	trailerSetAtMin, _ := cmd.Flags().GetString("trailer-set-at-min")
	trailerSetAtMax, _ := cmd.Flags().GetString("trailer-set-at-max")
	isTrailerSetAt, _ := cmd.Flags().GetString("is-trailer-set-at")
	tractorSetAtMin, _ := cmd.Flags().GetString("tractor-set-at-min")
	tractorSetAtMax, _ := cmd.Flags().GetString("tractor-set-at-max")
	isTractorSetAt, _ := cmd.Flags().GetString("is-tractor-set-at")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	isCreatedAt, _ := cmd.Flags().GetString("is-created-at")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	isUpdatedAt, _ := cmd.Flags().GetString("is-updated-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return t3EquipmentshareVehiclesListOptions{
		BaseURL:               baseURL,
		Token:                 token,
		JSON:                  jsonOut,
		NoAuth:                noAuth,
		Limit:                 limit,
		Offset:                offset,
		Sort:                  sort,
		Broker:                broker,
		Trucker:               trucker,
		Trailer:               trailer,
		Tractor:               tractor,
		HasTrailer:            hasTrailer,
		HasTractor:            hasTractor,
		AssignedAtMin:         assignedAtMin,
		IntegrationIdentifier: integrationIdentifier,
		TrailerSetAtMin:       trailerSetAtMin,
		TrailerSetAtMax:       trailerSetAtMax,
		IsTrailerSetAt:        isTrailerSetAt,
		TractorSetAtMin:       tractorSetAtMin,
		TractorSetAtMax:       tractorSetAtMax,
		IsTractorSetAt:        isTractorSetAt,
		CreatedAtMin:          createdAtMin,
		CreatedAtMax:          createdAtMax,
		IsCreatedAt:           isCreatedAt,
		UpdatedAtMin:          updatedAtMin,
		UpdatedAtMax:          updatedAtMax,
		IsUpdatedAt:           isUpdatedAt,
	}, nil
}

func buildT3EquipmentshareVehicleRows(resp jsonAPIResponse) []t3EquipmentshareVehicleRow {
	rows := make([]t3EquipmentshareVehicleRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := t3EquipmentshareVehicleRow{
			ID:                    resource.ID,
			VehicleID:             stringAttr(resource.Attributes, "vehicle-id"),
			VehicleNumber:         stringAttr(resource.Attributes, "vehicle-number"),
			IntegrationIdentifier: stringAttr(resource.Attributes, "integration-identifier"),
			TrailerSetAt:          formatDateTime(stringAttr(resource.Attributes, "trailer-set-at")),
			TractorSetAt:          formatDateTime(stringAttr(resource.Attributes, "tractor-set-at")),
			BrokerID:              relationshipIDFromMap(resource.Relationships, "broker"),
			TruckerID:             relationshipIDFromMap(resource.Relationships, "trucker"),
			TrailerID:             relationshipIDFromMap(resource.Relationships, "trailer"),
			TractorID:             relationshipIDFromMap(resource.Relationships, "tractor"),
		}

		rows = append(rows, row)
	}
	return rows
}

func renderT3EquipmentshareVehiclesTable(cmd *cobra.Command, rows []t3EquipmentshareVehicleRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No T3 EquipmentShare vehicles found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tVEHICLE NUMBER\tVEHICLE ID\tTRUCKER\tTRAILER\tTRACTOR\tINTEGRATION ID")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.VehicleNumber,
			row.VehicleID,
			row.TruckerID,
			row.TrailerID,
			row.TractorID,
			row.IntegrationIdentifier,
		)
	}
	return writer.Flush()
}
