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

type laborRequirementsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type laborRequirementDetails struct {
	ID                                string   `json:"id"`
	JobProductionPlanID               string   `json:"job_production_plan_id,omitempty"`
	ResourceClassificationType        string   `json:"resource_classification_type,omitempty"`
	ResourceClassificationID          string   `json:"resource_classification_id,omitempty"`
	ResourceType                      string   `json:"resource_type,omitempty"`
	ResourceID                        string   `json:"resource_id,omitempty"`
	LaborerID                         string   `json:"laborer_id,omitempty"`
	UserID                            string   `json:"user_id,omitempty"`
	JobSiteID                         string   `json:"job_site_id,omitempty"`
	LaborRequirementID                string   `json:"labor_requirement_id,omitempty"`
	CraftClassID                      string   `json:"craft_class_id,omitempty"`
	CraftClassEffectiveID             string   `json:"craft_class_effective_id,omitempty"`
	ProjectCostClassificationID       string   `json:"project_cost_classification_id,omitempty"`
	OriginMaterialSiteID              string   `json:"origin_material_site_id,omitempty"`
	LaborRequirementLaborerID         string   `json:"labor_requirement_laborer_id,omitempty"`
	LaborRequirementUserID            string   `json:"labor_requirement_user_id,omitempty"`
	TimeSheetID                       string   `json:"time_sheet_id,omitempty"`
	AssignmentConfirmationUUID        string   `json:"assignment_confirmation_uuid,omitempty"`
	TimeZoneID                        string   `json:"time_zone_id,omitempty"`
	StartAt                           string   `json:"start_at,omitempty"`
	EndAt                             string   `json:"end_at,omitempty"`
	StartAtEffective                  string   `json:"start_at_effective,omitempty"`
	EndAtEffective                    string   `json:"end_at_effective,omitempty"`
	MobilizationMethod                string   `json:"mobilization_method,omitempty"`
	CalculatedMobilizationMethod      string   `json:"calculated_mobilization_method,omitempty"`
	Note                              string   `json:"note,omitempty"`
	RequiresInboundMovement           bool     `json:"requires_inbound_movement,omitempty"`
	RequiresOutboundMovement          bool     `json:"requires_outbound_movement,omitempty"`
	IsOnlyForEquipmentMovement        bool     `json:"is_only_for_equipment_movement,omitempty"`
	IsValidatingOverlapping           bool     `json:"is_validating_overlapping,omitempty"`
	InboundLatitude                   string   `json:"inbound_latitude,omitempty"`
	InboundLongitude                  string   `json:"inbound_longitude,omitempty"`
	OutboundLatitude                  string   `json:"outbound_latitude,omitempty"`
	OutboundLongitude                 string   `json:"outbound_longitude,omitempty"`
	ExplicitInboundLatitude           string   `json:"explicit_inbound_latitude,omitempty"`
	ExplicitInboundLongitude          string   `json:"explicit_inbound_longitude,omitempty"`
	ExplicitOutboundLatitude          string   `json:"explicit_outbound_latitude,omitempty"`
	ExplicitOutboundLongitude         string   `json:"explicit_outbound_longitude,omitempty"`
	EquipmentRequirementIDs           []string `json:"equipment_requirement_ids,omitempty"`
	CrewAssignmentConfirmationIDs     []string `json:"crew_assignment_confirmation_ids,omitempty"`
	CrewAssignmentConfirmationID      string   `json:"crew_assignment_confirmation_id,omitempty"`
	CrewRequirementCredentialClassIDs []string `json:"crew_requirement_credential_classification_ids,omitempty"`
}

func newLaborRequirementsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show labor requirement details",
		Long: `Show the full details of a labor requirement.

Output Fields:
  ID
  Job Production Plan ID
  Resource Classification (type + ID)
  Resource (type + ID)
  Laborer ID
  User ID
  Job Site ID
  Labor Requirement ID
  Craft Class ID
  Craft Class Effective ID
  Project Cost Classification ID
  Origin Material Site ID
  Time Sheet ID
  Assignment Confirmation UUID
  Time Zone ID
  Start At
  End At
  Start At Effective
  End At Effective
  Mobilization Method
  Calculated Mobilization Method
  Note
  Requires Inbound Movement
  Requires Outbound Movement
  Is Only For Equipment Movement
  Is Validating Overlapping
  Inbound/Outbound Coordinates (explicit + calculated)
  Equipment Requirement IDs
  Crew Assignment Confirmation IDs
  Crew Assignment Confirmation ID
  Crew Requirement Credential Classification IDs

Arguments:
  <id>    The labor requirement ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a labor requirement
  xbe view labor-requirements show 123

  # Output as JSON
  xbe view labor-requirements show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runLaborRequirementsShow,
	}
	initLaborRequirementsShowFlags(cmd)
	return cmd
}

func init() {
	laborRequirementsCmd.AddCommand(newLaborRequirementsShowCmd())
}

func initLaborRequirementsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runLaborRequirementsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseLaborRequirementsShowOptions(cmd)
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
		return fmt.Errorf("labor requirement id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("include", "job-production-plan,resource-classification,resource,laborer,user,job-site,labor-requirement,craft-class,project-cost-classification,origin-material-site,craft-class-effective,labor-requirement-laborer,labor-requirement-user,equipment-requirements,crew-assignment-confirmations,crew-assignment-confirmation,crew-requirement-credential-classifications,time-sheet")
	query.Set("fields[labor-requirements]", "start-at,end-at,start-at-effective,end-at-effective,mobilization-method,calculated-mobilization-method,note,requires-inbound-movement,requires-outbound-movement,is-only-for-equipment-movement,is-validating-overlapping,time-zone-id,assignment-confirmation-uuid,inbound-latitude,inbound-longitude,outbound-latitude,outbound-longitude,explicit-inbound-latitude,explicit-inbound-longitude,explicit-outbound-latitude,explicit-outbound-longitude")

	body, _, err := client.Get(cmd.Context(), "/v1/labor-requirements/"+id, query)
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

	details := buildLaborRequirementDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderLaborRequirementDetails(cmd, details)
}

func parseLaborRequirementsShowOptions(cmd *cobra.Command) (laborRequirementsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return laborRequirementsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildLaborRequirementDetails(resp jsonAPISingleResponse) laborRequirementDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := laborRequirementDetails{
		ID:                           resource.ID,
		StartAt:                      formatDateTime(stringAttr(attrs, "start-at")),
		EndAt:                        formatDateTime(stringAttr(attrs, "end-at")),
		StartAtEffective:             formatDateTime(stringAttr(attrs, "start-at-effective")),
		EndAtEffective:               formatDateTime(stringAttr(attrs, "end-at-effective")),
		MobilizationMethod:           stringAttr(attrs, "mobilization-method"),
		CalculatedMobilizationMethod: stringAttr(attrs, "calculated-mobilization-method"),
		Note:                         stringAttr(attrs, "note"),
		RequiresInboundMovement:      boolAttr(attrs, "requires-inbound-movement"),
		RequiresOutboundMovement:     boolAttr(attrs, "requires-outbound-movement"),
		IsOnlyForEquipmentMovement:   boolAttr(attrs, "is-only-for-equipment-movement"),
		IsValidatingOverlapping:      boolAttr(attrs, "is-validating-overlapping"),
		TimeZoneID:                   stringAttr(attrs, "time-zone-id"),
		AssignmentConfirmationUUID:   stringAttr(attrs, "assignment-confirmation-uuid"),
		InboundLatitude:              stringAttr(attrs, "inbound-latitude"),
		InboundLongitude:             stringAttr(attrs, "inbound-longitude"),
		OutboundLatitude:             stringAttr(attrs, "outbound-latitude"),
		OutboundLongitude:            stringAttr(attrs, "outbound-longitude"),
		ExplicitInboundLatitude:      stringAttr(attrs, "explicit-inbound-latitude"),
		ExplicitInboundLongitude:     stringAttr(attrs, "explicit-inbound-longitude"),
		ExplicitOutboundLatitude:     stringAttr(attrs, "explicit-outbound-latitude"),
		ExplicitOutboundLongitude:    stringAttr(attrs, "explicit-outbound-longitude"),
	}

	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		details.JobProductionPlanID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["resource-classification"]; ok && rel.Data != nil {
		details.ResourceClassificationType = rel.Data.Type
		details.ResourceClassificationID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["resource"]; ok && rel.Data != nil {
		details.ResourceType = rel.Data.Type
		details.ResourceID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["laborer"]; ok && rel.Data != nil {
		details.LaborerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
		details.UserID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["job-site"]; ok && rel.Data != nil {
		details.JobSiteID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["labor-requirement"]; ok && rel.Data != nil {
		details.LaborRequirementID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["craft-class"]; ok && rel.Data != nil {
		details.CraftClassID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["craft-class-effective"]; ok && rel.Data != nil {
		details.CraftClassEffectiveID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["project-cost-classification"]; ok && rel.Data != nil {
		details.ProjectCostClassificationID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["origin-material-site"]; ok && rel.Data != nil {
		details.OriginMaterialSiteID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["labor-requirement-laborer"]; ok && rel.Data != nil {
		details.LaborRequirementLaborerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["labor-requirement-user"]; ok && rel.Data != nil {
		details.LaborRequirementUserID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["time-sheet"]; ok && rel.Data != nil {
		details.TimeSheetID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["crew-assignment-confirmation"]; ok && rel.Data != nil {
		details.CrewAssignmentConfirmationID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["equipment-requirements"]; ok {
		details.EquipmentRequirementIDs = relationshipIDList(rel)
	}
	if rel, ok := resource.Relationships["crew-assignment-confirmations"]; ok {
		details.CrewAssignmentConfirmationIDs = relationshipIDList(rel)
	}
	if rel, ok := resource.Relationships["crew-requirement-credential-classifications"]; ok {
		details.CrewRequirementCredentialClassIDs = relationshipIDList(rel)
	}

	return details
}

func renderLaborRequirementDetails(cmd *cobra.Command, details laborRequirementDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.JobProductionPlanID != "" {
		fmt.Fprintf(out, "Job Production Plan ID: %s\n", details.JobProductionPlanID)
	}
	if details.ResourceClassificationID != "" {
		label := details.ResourceClassificationID
		if details.ResourceClassificationType != "" {
			label = details.ResourceClassificationType + "/" + details.ResourceClassificationID
		}
		fmt.Fprintf(out, "Resource Classification: %s\n", label)
	}
	if details.ResourceID != "" {
		label := details.ResourceID
		if details.ResourceType != "" {
			label = details.ResourceType + "/" + details.ResourceID
		}
		fmt.Fprintf(out, "Resource: %s\n", label)
	}
	if details.LaborerID != "" {
		fmt.Fprintf(out, "Laborer ID: %s\n", details.LaborerID)
	}
	if details.UserID != "" {
		fmt.Fprintf(out, "User ID: %s\n", details.UserID)
	}
	if details.JobSiteID != "" {
		fmt.Fprintf(out, "Job Site ID: %s\n", details.JobSiteID)
	}
	if details.LaborRequirementID != "" {
		fmt.Fprintf(out, "Labor Requirement ID: %s\n", details.LaborRequirementID)
	}
	if details.CraftClassID != "" {
		fmt.Fprintf(out, "Craft Class ID: %s\n", details.CraftClassID)
	}
	if details.CraftClassEffectiveID != "" {
		fmt.Fprintf(out, "Craft Class Effective ID: %s\n", details.CraftClassEffectiveID)
	}
	if details.ProjectCostClassificationID != "" {
		fmt.Fprintf(out, "Project Cost Classification ID: %s\n", details.ProjectCostClassificationID)
	}
	if details.OriginMaterialSiteID != "" {
		fmt.Fprintf(out, "Origin Material Site ID: %s\n", details.OriginMaterialSiteID)
	}
	if details.TimeSheetID != "" {
		fmt.Fprintf(out, "Time Sheet ID: %s\n", details.TimeSheetID)
	}
	if details.AssignmentConfirmationUUID != "" {
		fmt.Fprintf(out, "Assignment Confirmation UUID: %s\n", details.AssignmentConfirmationUUID)
	}
	if details.TimeZoneID != "" {
		fmt.Fprintf(out, "Time Zone ID: %s\n", details.TimeZoneID)
	}
	if details.StartAt != "" {
		fmt.Fprintf(out, "Start At: %s\n", details.StartAt)
	}
	if details.EndAt != "" {
		fmt.Fprintf(out, "End At: %s\n", details.EndAt)
	}
	if details.StartAtEffective != "" {
		fmt.Fprintf(out, "Start At Effective: %s\n", details.StartAtEffective)
	}
	if details.EndAtEffective != "" {
		fmt.Fprintf(out, "End At Effective: %s\n", details.EndAtEffective)
	}
	if details.MobilizationMethod != "" {
		fmt.Fprintf(out, "Mobilization Method: %s\n", details.MobilizationMethod)
	}
	if details.CalculatedMobilizationMethod != "" {
		fmt.Fprintf(out, "Calculated Mobilization Method: %s\n", details.CalculatedMobilizationMethod)
	}
	if details.Note != "" {
		fmt.Fprintf(out, "Note: %s\n", details.Note)
	}

	fmt.Fprintf(out, "Requires Inbound Movement: %t\n", details.RequiresInboundMovement)
	fmt.Fprintf(out, "Requires Outbound Movement: %t\n", details.RequiresOutboundMovement)
	fmt.Fprintf(out, "Is Only For Equipment Movement: %t\n", details.IsOnlyForEquipmentMovement)
	fmt.Fprintf(out, "Is Validating Overlapping: %t\n", details.IsValidatingOverlapping)

	if details.InboundLatitude != "" || details.InboundLongitude != "" {
		fmt.Fprintf(out, "Inbound Coordinates: %s, %s\n", details.InboundLatitude, details.InboundLongitude)
	}
	if details.OutboundLatitude != "" || details.OutboundLongitude != "" {
		fmt.Fprintf(out, "Outbound Coordinates: %s, %s\n", details.OutboundLatitude, details.OutboundLongitude)
	}
	if details.ExplicitInboundLatitude != "" || details.ExplicitInboundLongitude != "" {
		fmt.Fprintf(out, "Explicit Inbound Coordinates: %s, %s\n", details.ExplicitInboundLatitude, details.ExplicitInboundLongitude)
	}
	if details.ExplicitOutboundLatitude != "" || details.ExplicitOutboundLongitude != "" {
		fmt.Fprintf(out, "Explicit Outbound Coordinates: %s, %s\n", details.ExplicitOutboundLatitude, details.ExplicitOutboundLongitude)
	}
	if details.LaborRequirementLaborerID != "" {
		fmt.Fprintf(out, "Labor Requirement Laborer ID: %s\n", details.LaborRequirementLaborerID)
	}
	if details.LaborRequirementUserID != "" {
		fmt.Fprintf(out, "Labor Requirement User ID: %s\n", details.LaborRequirementUserID)
	}
	if details.CrewAssignmentConfirmationID != "" {
		fmt.Fprintf(out, "Crew Assignment Confirmation ID: %s\n", details.CrewAssignmentConfirmationID)
	}

	if len(details.EquipmentRequirementIDs) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintf(out, "Equipment Requirements (%d):\n", len(details.EquipmentRequirementIDs))
		fmt.Fprintln(out, strings.Repeat("-", 40))
		for _, id := range details.EquipmentRequirementIDs {
			fmt.Fprintf(out, "  - %s\n", id)
		}
	}
	if len(details.CrewAssignmentConfirmationIDs) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintf(out, "Crew Assignment Confirmations (%d):\n", len(details.CrewAssignmentConfirmationIDs))
		fmt.Fprintln(out, strings.Repeat("-", 40))
		for _, id := range details.CrewAssignmentConfirmationIDs {
			fmt.Fprintf(out, "  - %s\n", id)
		}
	}
	if len(details.CrewRequirementCredentialClassIDs) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintf(out, "Crew Requirement Credential Classifications (%d):\n", len(details.CrewRequirementCredentialClassIDs))
		fmt.Fprintln(out, strings.Repeat("-", 40))
		for _, id := range details.CrewRequirementCredentialClassIDs {
			fmt.Fprintf(out, "  - %s\n", id)
		}
	}

	return nil
}
