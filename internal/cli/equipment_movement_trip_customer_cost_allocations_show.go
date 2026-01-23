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

type equipmentMovementTripCustomerCostAllocationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type equipmentMovementTripCustomerCostAllocationDetails struct {
	ID                              string         `json:"id"`
	TripID                          string         `json:"trip_id,omitempty"`
	IsExplicit                      bool           `json:"is_explicit"`
	Allocation                      map[string]any `json:"allocation,omitempty"`
	EquipmentMovementRequirementIDs []string       `json:"equipment_movement_requirement_ids,omitempty"`
}

func newEquipmentMovementTripCustomerCostAllocationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show equipment movement trip customer cost allocation details",
		Long: `Show the full details of a customer cost allocation.

Output Fields:
  ID
  Trip
  Is Explicit
  Allocation
  Equipment Movement Requirements

Arguments:
  <id>    The allocation ID (required). You can find IDs using the list command.`,
		Example: `  # Show a cost allocation
  xbe view equipment-movement-trip-customer-cost-allocations show 123

  # Show as JSON
  xbe view equipment-movement-trip-customer-cost-allocations show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runEquipmentMovementTripCustomerCostAllocationsShow,
	}
	initEquipmentMovementTripCustomerCostAllocationsShowFlags(cmd)
	return cmd
}

func init() {
	equipmentMovementTripCustomerCostAllocationsCmd.AddCommand(newEquipmentMovementTripCustomerCostAllocationsShowCmd())
}

func initEquipmentMovementTripCustomerCostAllocationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runEquipmentMovementTripCustomerCostAllocationsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseEquipmentMovementTripCustomerCostAllocationsShowOptions(cmd)
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
		return fmt.Errorf("equipment movement trip customer cost allocation id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[equipment-movement-trip-customer-cost-allocations]", "is-explicit,allocation,trip,equipment-movement-requirements")

	body, _, err := client.Get(cmd.Context(), "/v1/equipment-movement-trip-customer-cost-allocations/"+id, query)
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

	details := buildEquipmentMovementTripCustomerCostAllocationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderEquipmentMovementTripCustomerCostAllocationDetails(cmd, details)
}

func parseEquipmentMovementTripCustomerCostAllocationsShowOptions(cmd *cobra.Command) (equipmentMovementTripCustomerCostAllocationsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return equipmentMovementTripCustomerCostAllocationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildEquipmentMovementTripCustomerCostAllocationDetails(resp jsonAPISingleResponse) equipmentMovementTripCustomerCostAllocationDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := equipmentMovementTripCustomerCostAllocationDetails{
		ID:         resource.ID,
		IsExplicit: boolAttr(attrs, "is-explicit"),
		Allocation: allocationAttr(attrs, "allocation"),
	}

	if rel, ok := resource.Relationships["trip"]; ok && rel.Data != nil {
		details.TripID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["equipment-movement-requirements"]; ok {
		details.EquipmentMovementRequirementIDs = relationshipIDStrings(rel)
	}

	return details
}

func renderEquipmentMovementTripCustomerCostAllocationDetails(cmd *cobra.Command, details equipmentMovementTripCustomerCostAllocationDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TripID != "" {
		fmt.Fprintf(out, "Trip: %s\n", details.TripID)
	}
	fmt.Fprintf(out, "Is Explicit: %t\n", details.IsExplicit)
	if len(details.Allocation) > 0 {
		if payload, err := json.Marshal(details.Allocation); err == nil {
			fmt.Fprintf(out, "Allocation: %s\n", string(payload))
		}
	}
	if len(details.EquipmentMovementRequirementIDs) > 0 {
		fmt.Fprintf(out, "Equipment Movement Requirements: %s\n", strings.Join(details.EquipmentMovementRequirementIDs, ", "))
	}

	return nil
}

func relationshipIDStrings(rel jsonAPIRelationship) []string {
	ids := relationshipIDs(rel)
	if len(ids) == 0 {
		return nil
	}
	values := make([]string, 0, len(ids))
	for _, id := range ids {
		if id.ID == "" {
			continue
		}
		values = append(values, id.ID)
	}
	return values
}
