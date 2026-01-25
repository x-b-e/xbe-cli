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

type equipmentMovementTripDispatchesListOptions struct {
	BaseURL                      string
	Token                        string
	JSON                         bool
	NoAuth                       bool
	Limit                        int
	Offset                       int
	Sort                         string
	Status                       string
	CreatedBy                    string
	EquipmentMovementTrip        string
	EquipmentMovementRequirement string
	InboundEquipmentRequirement  string
	OutboundEquipmentRequirement string
	Trucker                      string
	Driver                       string
	LineupDispatch               string
	CreatedAtMin                 string
	CreatedAtMax                 string
}

type equipmentMovementTripDispatchRow struct {
	ID                             string `json:"id"`
	Status                         string `json:"status,omitempty"`
	EquipmentMovementTripID        string `json:"equipment_movement_trip_id,omitempty"`
	EquipmentMovementRequirementID string `json:"equipment_movement_requirement_id,omitempty"`
	InboundEquipmentRequirementID  string `json:"inbound_equipment_requirement_id,omitempty"`
	OutboundEquipmentRequirementID string `json:"outbound_equipment_requirement_id,omitempty"`
	TruckerID                      string `json:"trucker_id,omitempty"`
	DriverID                       string `json:"driver_id,omitempty"`
	TrailerID                      string `json:"trailer_id,omitempty"`
	CreatedByID                    string `json:"created_by_id,omitempty"`
	LineupDispatchID               string `json:"lineup_dispatch_id,omitempty"`
	CreatedAt                      string `json:"created_at,omitempty"`
}

func newEquipmentMovementTripDispatchesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List equipment movement trip dispatches",
		Long: `List equipment movement trip dispatches with filtering and pagination.

Output Columns:
  ID                Dispatch identifier
  STATUS            Dispatch status
  TRIP ID           Equipment movement trip ID
  MOVEMENT REQ ID   Equipment movement requirement ID
  INBOUND REQ ID    Inbound equipment requirement ID
  OUTBOUND REQ ID   Outbound equipment requirement ID
  TRUCKER ID        Trucker ID
  DRIVER ID         Driver ID
  TRAILER ID        Trailer ID
  CREATED BY        Creator user ID
  LINEUP DISPATCH   Lineup dispatch ID
  CREATED AT        Creation timestamp

Filters:
  --status                         Filter by status
  --created-by                     Filter by created-by user ID
  --equipment-movement-trip        Filter by equipment movement trip ID
  --equipment-movement-requirement Filter by equipment movement requirement ID
  --inbound-equipment-requirement  Filter by inbound equipment requirement ID
  --outbound-equipment-requirement Filter by outbound equipment requirement ID
  --trucker                        Filter by trucker ID
  --driver                         Filter by driver ID
  --lineup-dispatch                Filter by lineup dispatch ID
  --created-at-min                 Filter by created-at on/after (ISO 8601)
  --created-at-max                 Filter by created-at on/before (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List trip dispatches
  xbe view equipment-movement-trip-dispatches list

  # Filter by status
  xbe view equipment-movement-trip-dispatches list --status pending

  # Filter by trip
  xbe view equipment-movement-trip-dispatches list --equipment-movement-trip 123

  # Filter by created-at range
  xbe view equipment-movement-trip-dispatches list --created-at-min 2025-01-01T00:00:00Z

  # Output as JSON
  xbe view equipment-movement-trip-dispatches list --json`,
		RunE: runEquipmentMovementTripDispatchesList,
	}
	initEquipmentMovementTripDispatchesListFlags(cmd)
	return cmd
}

func init() {
	equipmentMovementTripDispatchesCmd.AddCommand(newEquipmentMovementTripDispatchesListCmd())
}

func initEquipmentMovementTripDispatchesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("status", "", "Filter by status")
	cmd.Flags().String("created-by", "", "Filter by created-by user ID")
	cmd.Flags().String("equipment-movement-trip", "", "Filter by equipment movement trip ID")
	cmd.Flags().String("equipment-movement-requirement", "", "Filter by equipment movement requirement ID")
	cmd.Flags().String("inbound-equipment-requirement", "", "Filter by inbound equipment requirement ID")
	cmd.Flags().String("outbound-equipment-requirement", "", "Filter by outbound equipment requirement ID")
	cmd.Flags().String("trucker", "", "Filter by trucker ID")
	cmd.Flags().String("driver", "", "Filter by driver ID")
	cmd.Flags().String("lineup-dispatch", "", "Filter by lineup dispatch ID")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runEquipmentMovementTripDispatchesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseEquipmentMovementTripDispatchesListOptions(cmd)
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
	query.Set("fields[equipment-movement-trip-dispatches]", "status,created-at,equipment-movement-trip,equipment-movement-requirement,inbound-equipment-requirement,outbound-equipment-requirement,trucker,driver,trailer,created-by,lineup-dispatch")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[created_by]", opts.CreatedBy)
	setFilterIfPresent(query, "filter[equipment_movement_trip]", opts.EquipmentMovementTrip)
	setFilterIfPresent(query, "filter[equipment_movement_requirement]", opts.EquipmentMovementRequirement)
	setFilterIfPresent(query, "filter[inbound_equipment_requirement]", opts.InboundEquipmentRequirement)
	setFilterIfPresent(query, "filter[outbound_equipment_requirement]", opts.OutboundEquipmentRequirement)
	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)
	setFilterIfPresent(query, "filter[driver]", opts.Driver)
	setFilterIfPresent(query, "filter[lineup_dispatch]", opts.LineupDispatch)
	setFilterIfPresent(query, "filter[created_at_min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created_at_max]", opts.CreatedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/equipment-movement-trip-dispatches", query)
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

	rows := buildEquipmentMovementTripDispatchRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderEquipmentMovementTripDispatchesTable(cmd, rows)
}

func parseEquipmentMovementTripDispatchesListOptions(cmd *cobra.Command) (equipmentMovementTripDispatchesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	status, _ := cmd.Flags().GetString("status")
	createdBy, _ := cmd.Flags().GetString("created-by")
	equipmentMovementTrip, _ := cmd.Flags().GetString("equipment-movement-trip")
	equipmentMovementRequirement, _ := cmd.Flags().GetString("equipment-movement-requirement")
	inboundEquipmentRequirement, _ := cmd.Flags().GetString("inbound-equipment-requirement")
	outboundEquipmentRequirement, _ := cmd.Flags().GetString("outbound-equipment-requirement")
	trucker, _ := cmd.Flags().GetString("trucker")
	driver, _ := cmd.Flags().GetString("driver")
	lineupDispatch, _ := cmd.Flags().GetString("lineup-dispatch")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return equipmentMovementTripDispatchesListOptions{
		BaseURL:                      baseURL,
		Token:                        token,
		JSON:                         jsonOut,
		NoAuth:                       noAuth,
		Limit:                        limit,
		Offset:                       offset,
		Sort:                         sort,
		Status:                       status,
		CreatedBy:                    createdBy,
		EquipmentMovementTrip:        equipmentMovementTrip,
		EquipmentMovementRequirement: equipmentMovementRequirement,
		InboundEquipmentRequirement:  inboundEquipmentRequirement,
		OutboundEquipmentRequirement: outboundEquipmentRequirement,
		Trucker:                      trucker,
		Driver:                       driver,
		LineupDispatch:               lineupDispatch,
		CreatedAtMin:                 createdAtMin,
		CreatedAtMax:                 createdAtMax,
	}, nil
}

func buildEquipmentMovementTripDispatchRows(resp jsonAPIResponse) []equipmentMovementTripDispatchRow {
	rows := make([]equipmentMovementTripDispatchRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := equipmentMovementTripDispatchRow{
			ID:        resource.ID,
			Status:    stringAttr(attrs, "status"),
			CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
		}

		if rel, ok := resource.Relationships["equipment-movement-trip"]; ok && rel.Data != nil {
			row.EquipmentMovementTripID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["equipment-movement-requirement"]; ok && rel.Data != nil {
			row.EquipmentMovementRequirementID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["inbound-equipment-requirement"]; ok && rel.Data != nil {
			row.InboundEquipmentRequirementID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["outbound-equipment-requirement"]; ok && rel.Data != nil {
			row.OutboundEquipmentRequirementID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["trucker"]; ok && rel.Data != nil {
			row.TruckerID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["driver"]; ok && rel.Data != nil {
			row.DriverID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["trailer"]; ok && rel.Data != nil {
			row.TrailerID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
			row.CreatedByID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["lineup-dispatch"]; ok && rel.Data != nil {
			row.LineupDispatchID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderEquipmentMovementTripDispatchesTable(cmd *cobra.Command, rows []equipmentMovementTripDispatchRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No equipment movement trip dispatches found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tTRIP ID\tMOVEMENT REQ ID\tINBOUND REQ ID\tOUTBOUND REQ ID\tTRUCKER ID\tDRIVER ID\tTRAILER ID\tCREATED BY\tLINEUP DISPATCH\tCREATED AT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Status,
			row.EquipmentMovementTripID,
			row.EquipmentMovementRequirementID,
			row.InboundEquipmentRequirementID,
			row.OutboundEquipmentRequirementID,
			row.TruckerID,
			row.DriverID,
			row.TrailerID,
			row.CreatedByID,
			row.LineupDispatchID,
			row.CreatedAt,
		)
	}
	return writer.Flush()
}
