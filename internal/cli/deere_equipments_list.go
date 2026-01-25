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

type deereEquipmentsListOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	NoAuth                 bool
	Limit                  int
	Offset                 int
	Sort                   string
	Broker                 string
	Equipment              string
	HasEquipment           string
	EquipmentAssignedAtMin string
	IntegrationIdentifier  string
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

type deereEquipmentRow struct {
	ID                            string `json:"id"`
	EquipmentName                 string `json:"equipment_name,omitempty"`
	EquipmentSourceID             string `json:"equipment_source_id,omitempty"`
	EquipmentSerialNumber         string `json:"equipment_serial_number,omitempty"`
	IntegrationIdentifier         string `json:"integration_identifier,omitempty"`
	EquipmentSetAt                string `json:"equipment_set_at,omitempty"`
	BrokerID                      string `json:"broker_id,omitempty"`
	BrokerName                    string `json:"broker_name,omitempty"`
	EquipmentID                   string `json:"equipment_id,omitempty"`
	EquipmentNickname             string `json:"equipment_nickname,omitempty"`
	AssignedEquipmentSerialNumber string `json:"assigned_equipment_serial_number,omitempty"`
}

func newDeereEquipmentsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Deere equipment",
		Long: `List Deere equipment with filtering and pagination.

Output Columns:
  ID         Deere equipment identifier
  NAME       Equipment name
  SOURCE ID  Deere equipment source identifier
  SERIAL     Deere equipment serial number
  SET AT     Equipment set timestamp
  EQUIPMENT  Assigned equipment nickname or ID
  BROKER     Broker name or ID

Filters:
  --broker                   Filter by broker ID
  --equipment                Filter by assigned equipment ID
  --has-equipment            Filter by equipment assignment (true/false)
  --equipment-assigned-at-min Filter by equipment assignment timestamp (ISO 8601)
  --integration-identifier   Filter by integration identifier
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
		Example: `  # List Deere equipment
  xbe view deere-equipments list

  # Filter by broker
  xbe view deere-equipments list --broker 123

  # Filter by equipment assignment
  xbe view deere-equipments list --has-equipment true

  # Output as JSON
  xbe view deere-equipments list --json`,
		Args: cobra.NoArgs,
		RunE: runDeereEquipmentsList,
	}
	initDeereEquipmentsListFlags(cmd)
	return cmd
}

func init() {
	deereEquipmentsCmd.AddCommand(newDeereEquipmentsListCmd())
}

func initDeereEquipmentsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("equipment", "", "Filter by assigned equipment ID")
	cmd.Flags().String("has-equipment", "", "Filter by equipment assignment (true/false)")
	cmd.Flags().String("equipment-assigned-at-min", "", "Filter by equipment assignment timestamp (ISO 8601)")
	cmd.Flags().String("integration-identifier", "", "Filter by integration identifier")
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

func runDeereEquipmentsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseDeereEquipmentsListOptions(cmd)
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
	query.Set("fields[deere-equipments]", "equipment-name,equipment-source-id,equipment-serial-number,integration-identifier,equipment-set-at,broker,equipment")
	query.Set("include", "broker,equipment")
	query.Set("fields[brokers]", "company-name")
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
	setFilterIfPresent(query, "filter[equipment]", opts.Equipment)
	setFilterIfPresent(query, "filter[has_equipment]", opts.HasEquipment)
	setFilterIfPresent(query, "filter[equipment_assigned_at_min]", opts.EquipmentAssignedAtMin)
	setFilterIfPresent(query, "filter[integration_identifier]", opts.IntegrationIdentifier)
	setFilterIfPresent(query, "filter[equipment_set_at_min]", opts.EquipmentSetAtMin)
	setFilterIfPresent(query, "filter[equipment_set_at_max]", opts.EquipmentSetAtMax)
	setFilterIfPresent(query, "filter[is_equipment_set_at]", opts.IsEquipmentSetAt)
	setFilterIfPresent(query, "filter[created_at_min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created_at_max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[is_created_at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[updated_at_min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated_at_max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[is_updated_at]", opts.IsUpdatedAt)

	body, _, err := client.Get(cmd.Context(), "/v1/deere-equipments", query)
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

	rows := buildDeereEquipmentRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderDeereEquipmentsTable(cmd, rows)
}

func parseDeereEquipmentsListOptions(cmd *cobra.Command) (deereEquipmentsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	broker, _ := cmd.Flags().GetString("broker")
	equipment, _ := cmd.Flags().GetString("equipment")
	hasEquipment, _ := cmd.Flags().GetString("has-equipment")
	equipmentAssignedAtMin, _ := cmd.Flags().GetString("equipment-assigned-at-min")
	integrationIdentifier, _ := cmd.Flags().GetString("integration-identifier")
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

	return deereEquipmentsListOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		NoAuth:                 noAuth,
		Limit:                  limit,
		Offset:                 offset,
		Sort:                   sort,
		Broker:                 broker,
		Equipment:              equipment,
		HasEquipment:           hasEquipment,
		EquipmentAssignedAtMin: equipmentAssignedAtMin,
		IntegrationIdentifier:  integrationIdentifier,
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

func buildDeereEquipmentRows(resp jsonAPIResponse) []deereEquipmentRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]deereEquipmentRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildDeereEquipmentRow(resource, included))
	}
	return rows
}

func deereEquipmentRowFromSingle(resp jsonAPISingleResponse) deereEquipmentRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}
	return buildDeereEquipmentRow(resp.Data, included)
}

func buildDeereEquipmentRow(resource jsonAPIResource, included map[string]jsonAPIResource) deereEquipmentRow {
	attrs := resource.Attributes
	row := deereEquipmentRow{
		ID:                    resource.ID,
		EquipmentName:         stringAttr(attrs, "equipment-name"),
		EquipmentSourceID:     stringAttr(attrs, "equipment-source-id"),
		EquipmentSerialNumber: stringAttr(attrs, "equipment-serial-number"),
		IntegrationIdentifier: stringAttr(attrs, "integration-identifier"),
		EquipmentSetAt:        formatDateTime(stringAttr(attrs, "equipment-set-at")),
	}

	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
		if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.BrokerName = stringAttr(broker.Attributes, "company-name")
		}
	}

	if rel, ok := resource.Relationships["equipment"]; ok && rel.Data != nil {
		row.EquipmentID = rel.Data.ID
		if equipment, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.EquipmentNickname = stringAttr(equipment.Attributes, "nickname")
			row.AssignedEquipmentSerialNumber = stringAttr(equipment.Attributes, "serial-number")
		}
	}

	return row
}

func renderDeereEquipmentsTable(cmd *cobra.Command, rows []deereEquipmentRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No Deere equipments found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tSOURCE ID\tSERIAL\tSET AT\tEQUIPMENT\tBROKER")
	for _, row := range rows {
		equipment := firstNonEmpty(row.EquipmentNickname, row.EquipmentID)
		broker := firstNonEmpty(row.BrokerName, row.BrokerID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.EquipmentName, 24),
			truncateString(row.EquipmentSourceID, 18),
			truncateString(row.EquipmentSerialNumber, 18),
			truncateString(row.EquipmentSetAt, 20),
			truncateString(equipment, 20),
			truncateString(broker, 20),
		)
	}
	return writer.Flush()
}
