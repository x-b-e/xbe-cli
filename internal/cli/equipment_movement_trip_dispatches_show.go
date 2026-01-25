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

type equipmentMovementTripDispatchesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type equipmentMovementTripDispatchDetails struct {
	ID                             string `json:"id"`
	Status                         string `json:"status,omitempty"`
	TellClerkSynchronously         bool   `json:"tell_clerk_synchronously"`
	FulfillmentCount               int    `json:"fulfillment_count"`
	ErrorSummary                   string `json:"error_summary,omitempty"`
	FailedStep                     string `json:"failed_step,omitempty"`
	TotalErrorCount                int    `json:"total_error_count"`
	EquipmentMovementTripID        string `json:"equipment_movement_trip_id,omitempty"`
	EquipmentMovementRequirementID string `json:"equipment_movement_requirement_id,omitempty"`
	InboundEquipmentRequirementID  string `json:"inbound_equipment_requirement_id,omitempty"`
	OutboundEquipmentRequirementID string `json:"outbound_equipment_requirement_id,omitempty"`
	OriginLocationID               string `json:"origin_location_id,omitempty"`
	DestinationLocationID          string `json:"destination_location_id,omitempty"`
	TruckerID                      string `json:"trucker_id,omitempty"`
	DriverID                       string `json:"driver_id,omitempty"`
	TrailerID                      string `json:"trailer_id,omitempty"`
	CreatedByID                    string `json:"created_by_id,omitempty"`
	LineupDispatchID               string `json:"lineup_dispatch_id,omitempty"`
	CreatedAt                      string `json:"created_at,omitempty"`
	UpdatedAt                      string `json:"updated_at,omitempty"`
	Options                        any    `json:"options,omitempty"`
	StepErrors                     any    `json:"step_errors,omitempty"`
}

func newEquipmentMovementTripDispatchesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show equipment movement trip dispatch details",
		Long: `Show the full details of an equipment movement trip dispatch.

Output Fields:
  ID
  Status
  Tell Clerk Synchronously
  Fulfillment Count
  Error Summary
  Failed Step
  Total Error Count
  Equipment Movement Trip ID
  Equipment Movement Requirement ID
  Inbound Equipment Requirement ID
  Outbound Equipment Requirement ID
  Origin Location ID
  Destination Location ID
  Trucker ID
  Driver ID
  Trailer ID
  Created By ID
  Lineup Dispatch ID
  Created At
  Updated At
  Options
  Step Errors

Arguments:
  <id>    The trip dispatch ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a trip dispatch
  xbe view equipment-movement-trip-dispatches show 123

  # Get JSON output
  xbe view equipment-movement-trip-dispatches show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runEquipmentMovementTripDispatchesShow,
	}
	initEquipmentMovementTripDispatchesShowFlags(cmd)
	return cmd
}

func init() {
	equipmentMovementTripDispatchesCmd.AddCommand(newEquipmentMovementTripDispatchesShowCmd())
}

func initEquipmentMovementTripDispatchesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runEquipmentMovementTripDispatchesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseEquipmentMovementTripDispatchesShowOptions(cmd)
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
		return fmt.Errorf("equipment movement trip dispatch id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[equipment-movement-trip-dispatches]", "status,tell-clerk-synchronously,options,step-errors,fulfillment-count,error-summary,failed-step,total-error-count,created-at,updated-at,equipment-movement-trip,equipment-movement-requirement,inbound-equipment-requirement,outbound-equipment-requirement,origin-location,destination-location,trucker,driver,trailer,created-by,lineup-dispatch")

	body, _, err := client.Get(cmd.Context(), "/v1/equipment-movement-trip-dispatches/"+id, query)
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

	details := buildEquipmentMovementTripDispatchDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderEquipmentMovementTripDispatchDetails(cmd, details)
}

func parseEquipmentMovementTripDispatchesShowOptions(cmd *cobra.Command) (equipmentMovementTripDispatchesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return equipmentMovementTripDispatchesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildEquipmentMovementTripDispatchDetails(resp jsonAPISingleResponse) equipmentMovementTripDispatchDetails {
	attrs := resp.Data.Attributes
	details := equipmentMovementTripDispatchDetails{
		ID:                             resp.Data.ID,
		Status:                         stringAttr(attrs, "status"),
		TellClerkSynchronously:         boolAttr(attrs, "tell-clerk-synchronously"),
		FulfillmentCount:               intAttr(attrs, "fulfillment-count"),
		ErrorSummary:                   stringAttr(attrs, "error-summary"),
		FailedStep:                     stringAttr(attrs, "failed-step"),
		TotalErrorCount:                intAttr(attrs, "total-error-count"),
		CreatedAt:                      formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:                      formatDateTime(stringAttr(attrs, "updated-at")),
		Options:                        attrs["options"],
		StepErrors:                     attrs["step-errors"],
		EquipmentMovementTripID:        "",
		EquipmentMovementRequirementID: "",
		InboundEquipmentRequirementID:  "",
		OutboundEquipmentRequirementID: "",
		OriginLocationID:               "",
		DestinationLocationID:          "",
		TruckerID:                      "",
		DriverID:                       "",
		TrailerID:                      "",
		CreatedByID:                    "",
		LineupDispatchID:               "",
	}

	if rel, ok := resp.Data.Relationships["equipment-movement-trip"]; ok && rel.Data != nil {
		details.EquipmentMovementTripID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["equipment-movement-requirement"]; ok && rel.Data != nil {
		details.EquipmentMovementRequirementID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["inbound-equipment-requirement"]; ok && rel.Data != nil {
		details.InboundEquipmentRequirementID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["outbound-equipment-requirement"]; ok && rel.Data != nil {
		details.OutboundEquipmentRequirementID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["origin-location"]; ok && rel.Data != nil {
		details.OriginLocationID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["destination-location"]; ok && rel.Data != nil {
		details.DestinationLocationID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["trucker"]; ok && rel.Data != nil {
		details.TruckerID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["driver"]; ok && rel.Data != nil {
		details.DriverID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["trailer"]; ok && rel.Data != nil {
		details.TrailerID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["lineup-dispatch"]; ok && rel.Data != nil {
		details.LineupDispatchID = rel.Data.ID
	}

	return details
}

func renderEquipmentMovementTripDispatchDetails(cmd *cobra.Command, details equipmentMovementTripDispatchDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.TellClerkSynchronously {
		fmt.Fprintln(out, "Tell Clerk Synchronously: yes")
	}
	fmt.Fprintf(out, "Fulfillment Count: %d\n", details.FulfillmentCount)
	if details.ErrorSummary != "" {
		fmt.Fprintf(out, "Error Summary: %s\n", details.ErrorSummary)
	}
	if details.FailedStep != "" {
		fmt.Fprintf(out, "Failed Step: %s\n", details.FailedStep)
	}
	fmt.Fprintf(out, "Total Error Count: %d\n", details.TotalErrorCount)

	if details.EquipmentMovementTripID != "" {
		fmt.Fprintf(out, "Equipment Movement Trip ID: %s\n", details.EquipmentMovementTripID)
	}
	if details.EquipmentMovementRequirementID != "" {
		fmt.Fprintf(out, "Equipment Movement Requirement ID: %s\n", details.EquipmentMovementRequirementID)
	}
	if details.InboundEquipmentRequirementID != "" {
		fmt.Fprintf(out, "Inbound Equipment Requirement ID: %s\n", details.InboundEquipmentRequirementID)
	}
	if details.OutboundEquipmentRequirementID != "" {
		fmt.Fprintf(out, "Outbound Equipment Requirement ID: %s\n", details.OutboundEquipmentRequirementID)
	}
	if details.OriginLocationID != "" {
		fmt.Fprintf(out, "Origin Location ID: %s\n", details.OriginLocationID)
	}
	if details.DestinationLocationID != "" {
		fmt.Fprintf(out, "Destination Location ID: %s\n", details.DestinationLocationID)
	}
	if details.TruckerID != "" {
		fmt.Fprintf(out, "Trucker ID: %s\n", details.TruckerID)
	}
	if details.DriverID != "" {
		fmt.Fprintf(out, "Driver ID: %s\n", details.DriverID)
	}
	if details.TrailerID != "" {
		fmt.Fprintf(out, "Trailer ID: %s\n", details.TrailerID)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By ID: %s\n", details.CreatedByID)
	}
	if details.LineupDispatchID != "" {
		fmt.Fprintf(out, "Lineup Dispatch ID: %s\n", details.LineupDispatchID)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}
	if details.Options != nil {
		fmt.Fprintln(out, "Options:")
		fmt.Fprintln(out, formatEquipmentMovementTripDispatchJSON(details.Options))
	}
	if details.StepErrors != nil {
		fmt.Fprintln(out, "Step Errors:")
		fmt.Fprintln(out, formatEquipmentMovementTripDispatchJSON(details.StepErrors))
	}

	return nil
}

func formatEquipmentMovementTripDispatchJSON(value any) string {
	if value == nil {
		return ""
	}
	pretty, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return string(pretty)
}
