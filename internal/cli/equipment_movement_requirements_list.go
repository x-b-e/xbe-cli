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

type equipmentMovementRequirementsListOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	NoAuth              bool
	Limit               int
	Offset              int
	Sort                string
	Broker              string
	Equipment           string
	InboundRequirement  string
	OutboundRequirement string
	OriginAtMinMin      string
	OriginAtMinMax      string
	DestinationAtMaxMin string
	DestinationAtMaxMax string
	CreatedAtMin        string
	CreatedAtMax        string
	UpdatedAtMin        string
	UpdatedAtMax        string
}

type equipmentMovementRequirementRow struct {
	ID                    string `json:"id"`
	BrokerID              string `json:"broker_id,omitempty"`
	EquipmentID           string `json:"equipment_id,omitempty"`
	InboundRequirementID  string `json:"inbound_requirement_id,omitempty"`
	OutboundRequirementID string `json:"outbound_requirement_id,omitempty"`
	OriginID              string `json:"origin_id,omitempty"`
	DestinationID         string `json:"destination_id,omitempty"`
	CustomerExplicitID    string `json:"customer_explicit_id,omitempty"`
	CustomerID            string `json:"customer_id,omitempty"`
	OriginAtMin           string `json:"origin_at_min,omitempty"`
	DestinationAtMax      string `json:"destination_at_max,omitempty"`
	Note                  string `json:"note,omitempty"`
}

func newEquipmentMovementRequirementsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List equipment movement requirements",
		Long: `List equipment movement requirements.

Output Columns:
  ID           Requirement identifier
  EQUIPMENT    Equipment ID
  ORIGIN       Origin location ID
  DESTINATION  Destination location ID
  ORIGIN MIN   Earliest origin time
  DEST MAX     Latest destination time
  BROKER       Broker ID

Filters:
  --broker                 Filter by broker ID
  --equipment              Filter by equipment ID
  --inbound-requirement    Filter by inbound equipment requirement ID
  --outbound-requirement   Filter by outbound equipment requirement ID
  --origin-at-min-min      Filter by origin-at-min on/after (ISO 8601)
  --origin-at-min-max      Filter by origin-at-min on/before (ISO 8601)
  --destination-at-max-min Filter by destination-at-max on/after (ISO 8601)
  --destination-at-max-max Filter by destination-at-max on/before (ISO 8601)
  --created-at-min         Filter by created-at on/after (ISO 8601)
  --created-at-max         Filter by created-at on/before (ISO 8601)
  --updated-at-min         Filter by updated-at on/after (ISO 8601)
  --updated-at-max         Filter by updated-at on/before (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List requirements
  xbe view equipment-movement-requirements list

  # Filter by broker
  xbe view equipment-movement-requirements list --broker 123

  # Filter by equipment
  xbe view equipment-movement-requirements list --equipment 456

  # Filter by origin window
  xbe view equipment-movement-requirements list --origin-at-min-min 2025-01-01T00:00:00Z --origin-at-min-max 2025-01-31T23:59:59Z

  # Output as JSON
  xbe view equipment-movement-requirements list --json`,
		Args: cobra.NoArgs,
		RunE: runEquipmentMovementRequirementsList,
	}
	initEquipmentMovementRequirementsListFlags(cmd)
	return cmd
}

func init() {
	equipmentMovementRequirementsCmd.AddCommand(newEquipmentMovementRequirementsListCmd())
}

func initEquipmentMovementRequirementsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("equipment", "", "Filter by equipment ID")
	cmd.Flags().String("inbound-requirement", "", "Filter by inbound equipment requirement ID")
	cmd.Flags().String("outbound-requirement", "", "Filter by outbound equipment requirement ID")
	cmd.Flags().String("origin-at-min-min", "", "Filter by origin-at-min on/after (ISO 8601)")
	cmd.Flags().String("origin-at-min-max", "", "Filter by origin-at-min on/before (ISO 8601)")
	cmd.Flags().String("destination-at-max-min", "", "Filter by destination-at-max on/after (ISO 8601)")
	cmd.Flags().String("destination-at-max-max", "", "Filter by destination-at-max on/before (ISO 8601)")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runEquipmentMovementRequirementsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseEquipmentMovementRequirementsListOptions(cmd)
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
	query.Set("include", "broker,equipment,origin,destination,inbound-requirement,outbound-requirement,customer-explicit,customer")

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
	setFilterIfPresent(query, "filter[equipment]", opts.Equipment)
	setFilterIfPresent(query, "filter[inbound_requirement]", opts.InboundRequirement)
	setFilterIfPresent(query, "filter[outbound_requirement]", opts.OutboundRequirement)
	setFilterIfPresent(query, "filter[origin_at_min_min]", opts.OriginAtMinMin)
	setFilterIfPresent(query, "filter[origin_at_min_max]", opts.OriginAtMinMax)
	setFilterIfPresent(query, "filter[destination_at_max_min]", opts.DestinationAtMaxMin)
	setFilterIfPresent(query, "filter[destination_at_max_max]", opts.DestinationAtMaxMax)
	setFilterIfPresent(query, "filter[created_at_min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created_at_max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated_at_min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated_at_max]", opts.UpdatedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/equipment-movement-requirements", query)
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

	rows := buildEquipmentMovementRequirementRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderEquipmentMovementRequirementsTable(cmd, rows)
}

func parseEquipmentMovementRequirementsListOptions(cmd *cobra.Command) (equipmentMovementRequirementsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	broker, _ := cmd.Flags().GetString("broker")
	equipment, _ := cmd.Flags().GetString("equipment")
	inboundRequirement, _ := cmd.Flags().GetString("inbound-requirement")
	outboundRequirement, _ := cmd.Flags().GetString("outbound-requirement")
	originAtMinMin, _ := cmd.Flags().GetString("origin-at-min-min")
	originAtMinMax, _ := cmd.Flags().GetString("origin-at-min-max")
	destinationAtMaxMin, _ := cmd.Flags().GetString("destination-at-max-min")
	destinationAtMaxMax, _ := cmd.Flags().GetString("destination-at-max-max")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return equipmentMovementRequirementsListOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		NoAuth:              noAuth,
		Limit:               limit,
		Offset:              offset,
		Sort:                sort,
		Broker:              broker,
		Equipment:           equipment,
		InboundRequirement:  inboundRequirement,
		OutboundRequirement: outboundRequirement,
		OriginAtMinMin:      originAtMinMin,
		OriginAtMinMax:      originAtMinMax,
		DestinationAtMaxMin: destinationAtMaxMin,
		DestinationAtMaxMax: destinationAtMaxMax,
		CreatedAtMin:        createdAtMin,
		CreatedAtMax:        createdAtMax,
		UpdatedAtMin:        updatedAtMin,
		UpdatedAtMax:        updatedAtMax,
	}, nil
}

func buildEquipmentMovementRequirementRows(resp jsonAPIResponse) []equipmentMovementRequirementRow {
	rows := make([]equipmentMovementRequirementRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildEquipmentMovementRequirementRow(resource))
	}
	return rows
}

func buildEquipmentMovementRequirementRow(resource jsonAPIResource) equipmentMovementRequirementRow {
	row := equipmentMovementRequirementRow{
		ID:               resource.ID,
		OriginAtMin:      formatDateTime(stringAttr(resource.Attributes, "origin-at-min")),
		DestinationAtMax: formatDateTime(stringAttr(resource.Attributes, "destination-at-max")),
		Note:             stringAttr(resource.Attributes, "note"),
	}

	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["equipment"]; ok && rel.Data != nil {
		row.EquipmentID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["inbound-requirement"]; ok && rel.Data != nil {
		row.InboundRequirementID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["outbound-requirement"]; ok && rel.Data != nil {
		row.OutboundRequirementID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["origin"]; ok && rel.Data != nil {
		row.OriginID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["destination"]; ok && rel.Data != nil {
		row.DestinationID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["customer-explicit"]; ok && rel.Data != nil {
		row.CustomerExplicitID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["customer"]; ok && rel.Data != nil {
		row.CustomerID = rel.Data.ID
	}

	return row
}

func renderEquipmentMovementRequirementsTable(cmd *cobra.Command, rows []equipmentMovementRequirementRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No equipment movement requirements found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tEQUIPMENT\tORIGIN\tDESTINATION\tORIGIN MIN\tDEST MAX\tBROKER")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.EquipmentID,
			row.OriginID,
			row.DestinationID,
			row.OriginAtMin,
			row.DestinationAtMax,
			row.BrokerID,
		)
	}
	return writer.Flush()
}
