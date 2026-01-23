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

type equipmentMovementRequirementsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type equipmentMovementRequirementDetails struct {
	ID                    string   `json:"id"`
	BrokerID              string   `json:"broker_id,omitempty"`
	EquipmentID           string   `json:"equipment_id,omitempty"`
	InboundRequirementID  string   `json:"inbound_requirement_id,omitempty"`
	OutboundRequirementID string   `json:"outbound_requirement_id,omitempty"`
	OriginID              string   `json:"origin_id,omitempty"`
	DestinationID         string   `json:"destination_id,omitempty"`
	CustomerExplicitID    string   `json:"customer_explicit_id,omitempty"`
	CustomerID            string   `json:"customer_id,omitempty"`
	OriginAtMin           string   `json:"origin_at_min,omitempty"`
	DestinationAtMax      string   `json:"destination_at_max,omitempty"`
	Note                  string   `json:"note,omitempty"`
	StopRequirementIDs    []string `json:"stop_requirement_ids,omitempty"`
}

func newEquipmentMovementRequirementsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show equipment movement requirement details",
		Long: `Show the full details of an equipment movement requirement.

Output Fields:
  ID
  Broker ID
  Equipment ID
  Inbound Requirement ID
  Outbound Requirement ID
  Origin Location ID
  Destination Location ID
  Customer Explicit ID
  Customer ID
  Origin At Min
  Destination At Max
  Note
  Stop Requirement IDs

Arguments:
  <id>    The requirement ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a requirement
  xbe view equipment-movement-requirements show 123

  # Output as JSON
  xbe view equipment-movement-requirements show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runEquipmentMovementRequirementsShow,
	}
	initEquipmentMovementRequirementsShowFlags(cmd)
	return cmd
}

func init() {
	equipmentMovementRequirementsCmd.AddCommand(newEquipmentMovementRequirementsShowCmd())
}

func initEquipmentMovementRequirementsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runEquipmentMovementRequirementsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseEquipmentMovementRequirementsShowOptions(cmd)
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
		return fmt.Errorf("equipment movement requirement id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("include", "broker,equipment,origin,destination,inbound-requirement,outbound-requirement,customer-explicit,customer,stop-requirements")

	body, _, err := client.Get(cmd.Context(), "/v1/equipment-movement-requirements/"+id, query)
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

	details := buildEquipmentMovementRequirementDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderEquipmentMovementRequirementDetails(cmd, details)
}

func parseEquipmentMovementRequirementsShowOptions(cmd *cobra.Command) (equipmentMovementRequirementsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return equipmentMovementRequirementsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildEquipmentMovementRequirementDetails(resp jsonAPISingleResponse) equipmentMovementRequirementDetails {
	resource := resp.Data
	details := equipmentMovementRequirementDetails{
		ID:               resource.ID,
		OriginAtMin:      formatDateTime(stringAttr(resource.Attributes, "origin-at-min")),
		DestinationAtMax: formatDateTime(stringAttr(resource.Attributes, "destination-at-max")),
		Note:             stringAttr(resource.Attributes, "note"),
	}

	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["equipment"]; ok && rel.Data != nil {
		details.EquipmentID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["inbound-requirement"]; ok && rel.Data != nil {
		details.InboundRequirementID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["outbound-requirement"]; ok && rel.Data != nil {
		details.OutboundRequirementID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["origin"]; ok && rel.Data != nil {
		details.OriginID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["destination"]; ok && rel.Data != nil {
		details.DestinationID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["customer-explicit"]; ok && rel.Data != nil {
		details.CustomerExplicitID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["customer"]; ok && rel.Data != nil {
		details.CustomerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["stop-requirements"]; ok && rel.raw != nil {
		var refs []jsonAPIResourceIdentifier
		if err := json.Unmarshal(rel.raw, &refs); err == nil {
			details.StopRequirementIDs = make([]string, 0, len(refs))
			for _, ref := range refs {
				details.StopRequirementIDs = append(details.StopRequirementIDs, ref.ID)
			}
		}
	}

	return details
}

func renderEquipmentMovementRequirementDetails(cmd *cobra.Command, details equipmentMovementRequirementDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
	}
	if details.EquipmentID != "" {
		fmt.Fprintf(out, "Equipment ID: %s\n", details.EquipmentID)
	}
	if details.InboundRequirementID != "" {
		fmt.Fprintf(out, "Inbound Requirement ID: %s\n", details.InboundRequirementID)
	}
	if details.OutboundRequirementID != "" {
		fmt.Fprintf(out, "Outbound Requirement ID: %s\n", details.OutboundRequirementID)
	}
	if details.OriginID != "" {
		fmt.Fprintf(out, "Origin Location ID: %s\n", details.OriginID)
	}
	if details.DestinationID != "" {
		fmt.Fprintf(out, "Destination Location ID: %s\n", details.DestinationID)
	}
	if details.CustomerExplicitID != "" {
		fmt.Fprintf(out, "Customer Explicit ID: %s\n", details.CustomerExplicitID)
	}
	if details.CustomerID != "" {
		fmt.Fprintf(out, "Customer ID: %s\n", details.CustomerID)
	}
	if details.OriginAtMin != "" {
		fmt.Fprintf(out, "Origin At Min: %s\n", details.OriginAtMin)
	}
	if details.DestinationAtMax != "" {
		fmt.Fprintf(out, "Destination At Max: %s\n", details.DestinationAtMax)
	}
	if details.Note != "" {
		fmt.Fprintf(out, "Note: %s\n", details.Note)
	}

	if len(details.StopRequirementIDs) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintf(out, "Stop Requirements (%d):\n", len(details.StopRequirementIDs))
		fmt.Fprintln(out, strings.Repeat("-", 40))
		for _, id := range details.StopRequirementIDs {
			fmt.Fprintf(out, "  - %s\n", id)
		}
	}

	return nil
}
