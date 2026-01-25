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

type crewRequirementsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type crewRequirementDetails struct {
	ID                                         string   `json:"id"`
	PolymorphicType                            string   `json:"polymorphic_type,omitempty"`
	TimeZoneID                                 string   `json:"time_zone_id,omitempty"`
	AssignmentConfirmationUUID                 string   `json:"assignment_confirmation_uuid,omitempty"`
	JobProductionPlanID                        string   `json:"job_production_plan_id,omitempty"`
	ResourceClassificationType                 string   `json:"resource_classification_type,omitempty"`
	ResourceClassificationID                   string   `json:"resource_classification_id,omitempty"`
	ResourceType                               string   `json:"resource_type,omitempty"`
	ResourceID                                 string   `json:"resource_id,omitempty"`
	OriginMaterialSiteID                       string   `json:"origin_material_site_id,omitempty"`
	JobSiteID                                  string   `json:"job_site_id,omitempty"`
	LaborRequirementID                         string   `json:"labor_requirement_id,omitempty"`
	CraftClassID                               string   `json:"craft_class_id,omitempty"`
	ProjectCostClassificationID                string   `json:"project_cost_classification_id,omitempty"`
	CraftClassEffectiveID                      string   `json:"craft_class_effective_id,omitempty"`
	LaborRequirementLaborerID                  string   `json:"labor_requirement_laborer_id,omitempty"`
	LaborRequirementUserID                     string   `json:"labor_requirement_user_id,omitempty"`
	EquipmentRequirementIDs                    []string `json:"equipment_requirement_ids,omitempty"`
	CrewAssignmentConfirmationID               string   `json:"crew_assignment_confirmation_id,omitempty"`
	CrewAssignmentConfirmationIDs              []string `json:"crew_assignment_confirmation_ids,omitempty"`
	CrewRequirementCredentialClassificationIDs []string `json:"crew_requirement_credential_classification_ids,omitempty"`
	TimeSheetID                                string   `json:"time_sheet_id,omitempty"`
	StartAt                                    string   `json:"start_at,omitempty"`
	EndAt                                      string   `json:"end_at,omitempty"`
	StartAtEffective                           string   `json:"start_at_effective,omitempty"`
	EndAtEffective                             string   `json:"end_at_effective,omitempty"`
	MobilizationMethod                         string   `json:"mobilization_method,omitempty"`
	CalculatedMobilizationMethod               string   `json:"calculated_mobilization_method,omitempty"`
	Note                                       string   `json:"note,omitempty"`
	RequiresInboundMovement                    bool     `json:"requires_inbound_movement"`
	RequiresOutboundMovement                   bool     `json:"requires_outbound_movement"`
	IsValidatingOverlapping                    bool     `json:"is_validating_overlapping"`
	IsOnlyForEquipmentMovement                 bool     `json:"is_only_for_equipment_movement"`
	InboundLatitude                            string   `json:"inbound_latitude,omitempty"`
	InboundLongitude                           string   `json:"inbound_longitude,omitempty"`
	OutboundLatitude                           string   `json:"outbound_latitude,omitempty"`
	OutboundLongitude                          string   `json:"outbound_longitude,omitempty"`
	ExplicitInboundLatitude                    string   `json:"explicit_inbound_latitude,omitempty"`
	ExplicitInboundLongitude                   string   `json:"explicit_inbound_longitude,omitempty"`
	ExplicitOutboundLatitude                   string   `json:"explicit_outbound_latitude,omitempty"`
	ExplicitOutboundLongitude                  string   `json:"explicit_outbound_longitude,omitempty"`
}

func newCrewRequirementsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show crew requirement details",
		Long: `Show the full details of a crew requirement.

Output Fields:
  ID
  Type
  Time Zone ID
  Assignment Confirmation UUID
  Job Production Plan
  Resource Classification
  Resource
  Origin Material Site
  Job Site
  Labor Requirement
  Craft Class
  Project Cost Classification
  Craft Class Effective
  Labor Requirement Laborer
  Labor Requirement User
  Equipment Requirements
  Crew Assignment Confirmation(s)
  Crew Requirement Credential Classifications
  Time Sheet
  Start At / End At
  Start At Effective / End At Effective
  Mobilization Method / Calculated Mobilization Method
  Note
  Requires Inbound Movement / Requires Outbound Movement
  Validating Overlapping
  Only For Equipment Movement
  Inbound/Outbound Coordinates (effective + explicit)

Arguments:
  <id>    The crew requirement ID (required). You can find IDs using the list command.`,
		Example: `  # Show a crew requirement
  xbe view crew-requirements show 123

  # Show as JSON
  xbe view crew-requirements show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runCrewRequirementsShow,
	}
	initCrewRequirementsShowFlags(cmd)
	return cmd
}

func init() {
	crewRequirementsCmd.AddCommand(newCrewRequirementsShowCmd())
}

func initCrewRequirementsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCrewRequirementsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseCrewRequirementsShowOptions(cmd)
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
		return fmt.Errorf("crew requirement id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[crew-requirements]", "polymorphic-type,time-zone-id,assignment-confirmation-uuid,start-at,end-at,start-at-effective,end-at-effective,mobilization-method,calculated-mobilization-method,note,requires-inbound-movement,requires-outbound-movement,is-validating-overlapping,is-only-for-equipment-movement,inbound-latitude,inbound-longitude,outbound-latitude,outbound-longitude,explicit-inbound-latitude,explicit-inbound-longitude,explicit-outbound-latitude,explicit-outbound-longitude,job-production-plan,resource-classification,resource,origin-material-site,job-site,labor-requirement,craft-class,project-cost-classification,craft-class-effective,labor-requirement-laborer,labor-requirement-user,equipment-requirements,crew-assignment-confirmation,crew-assignment-confirmations,crew-requirement-credential-classifications,time-sheet")

	body, _, err := client.Get(cmd.Context(), "/v1/crew-requirements/"+id, query)
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

	details := buildCrewRequirementDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderCrewRequirementDetails(cmd, details)
}

func parseCrewRequirementsShowOptions(cmd *cobra.Command) (crewRequirementsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return crewRequirementsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildCrewRequirementDetails(resp jsonAPISingleResponse) crewRequirementDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := crewRequirementDetails{
		ID:                           resource.ID,
		PolymorphicType:              stringAttr(attrs, "polymorphic-type"),
		TimeZoneID:                   stringAttr(attrs, "time-zone-id"),
		AssignmentConfirmationUUID:   stringAttr(attrs, "assignment-confirmation-uuid"),
		StartAt:                      formatDateTime(stringAttr(attrs, "start-at")),
		EndAt:                        formatDateTime(stringAttr(attrs, "end-at")),
		StartAtEffective:             formatDateTime(stringAttr(attrs, "start-at-effective")),
		EndAtEffective:               formatDateTime(stringAttr(attrs, "end-at-effective")),
		MobilizationMethod:           stringAttr(attrs, "mobilization-method"),
		CalculatedMobilizationMethod: stringAttr(attrs, "calculated-mobilization-method"),
		Note:                         stringAttr(attrs, "note"),
		RequiresInboundMovement:      boolAttr(attrs, "requires-inbound-movement"),
		RequiresOutboundMovement:     boolAttr(attrs, "requires-outbound-movement"),
		IsValidatingOverlapping:      boolAttr(attrs, "is-validating-overlapping"),
		IsOnlyForEquipmentMovement:   boolAttr(attrs, "is-only-for-equipment-movement"),
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
	if rel, ok := resource.Relationships["origin-material-site"]; ok && rel.Data != nil {
		details.OriginMaterialSiteID = rel.Data.ID
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
	if rel, ok := resource.Relationships["project-cost-classification"]; ok && rel.Data != nil {
		details.ProjectCostClassificationID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["craft-class-effective"]; ok && rel.Data != nil {
		details.CraftClassEffectiveID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["labor-requirement-laborer"]; ok && rel.Data != nil {
		details.LaborRequirementLaborerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["labor-requirement-user"]; ok && rel.Data != nil {
		details.LaborRequirementUserID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["crew-assignment-confirmation"]; ok && rel.Data != nil {
		details.CrewAssignmentConfirmationID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["time-sheet"]; ok && rel.Data != nil {
		details.TimeSheetID = rel.Data.ID
	}

	if rel, ok := resource.Relationships["equipment-requirements"]; ok {
		details.EquipmentRequirementIDs = relationshipIDList(rel)
	}
	if rel, ok := resource.Relationships["crew-assignment-confirmations"]; ok {
		details.CrewAssignmentConfirmationIDs = relationshipIDList(rel)
	}
	if rel, ok := resource.Relationships["crew-requirement-credential-classifications"]; ok {
		details.CrewRequirementCredentialClassificationIDs = relationshipIDList(rel)
	}

	return details
}

func renderCrewRequirementDetails(cmd *cobra.Command, details crewRequirementDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.PolymorphicType != "" {
		fmt.Fprintf(out, "Type: %s\n", details.PolymorphicType)
	}
	if details.TimeZoneID != "" {
		fmt.Fprintf(out, "Time Zone ID: %s\n", details.TimeZoneID)
	}
	if details.AssignmentConfirmationUUID != "" {
		fmt.Fprintf(out, "Assignment Confirmation UUID: %s\n", details.AssignmentConfirmationUUID)
	}
	if details.JobProductionPlanID != "" {
		fmt.Fprintf(out, "Job Production Plan: %s\n", details.JobProductionPlanID)
	}
	if details.ResourceClassificationType != "" && details.ResourceClassificationID != "" {
		fmt.Fprintf(out, "Resource Classification: %s/%s\n", details.ResourceClassificationType, details.ResourceClassificationID)
	}
	if details.ResourceType != "" && details.ResourceID != "" {
		fmt.Fprintf(out, "Resource: %s/%s\n", details.ResourceType, details.ResourceID)
	}
	if details.OriginMaterialSiteID != "" {
		fmt.Fprintf(out, "Origin Material Site: %s\n", details.OriginMaterialSiteID)
	}
	if details.JobSiteID != "" {
		fmt.Fprintf(out, "Job Site: %s\n", details.JobSiteID)
	}
	if details.LaborRequirementID != "" {
		fmt.Fprintf(out, "Labor Requirement: %s\n", details.LaborRequirementID)
	}
	if details.CraftClassID != "" {
		fmt.Fprintf(out, "Craft Class: %s\n", details.CraftClassID)
	}
	if details.ProjectCostClassificationID != "" {
		fmt.Fprintf(out, "Project Cost Classification: %s\n", details.ProjectCostClassificationID)
	}
	if details.CraftClassEffectiveID != "" {
		fmt.Fprintf(out, "Craft Class Effective: %s\n", details.CraftClassEffectiveID)
	}
	if details.LaborRequirementLaborerID != "" {
		fmt.Fprintf(out, "Labor Requirement Laborer: %s\n", details.LaborRequirementLaborerID)
	}
	if details.LaborRequirementUserID != "" {
		fmt.Fprintf(out, "Labor Requirement User: %s\n", details.LaborRequirementUserID)
	}
	if len(details.EquipmentRequirementIDs) > 0 {
		fmt.Fprintf(out, "Equipment Requirements: %s\n", strings.Join(details.EquipmentRequirementIDs, ", "))
	}
	if details.CrewAssignmentConfirmationID != "" {
		fmt.Fprintf(out, "Crew Assignment Confirmation: %s\n", details.CrewAssignmentConfirmationID)
	}
	if len(details.CrewAssignmentConfirmationIDs) > 0 {
		fmt.Fprintf(out, "Crew Assignment Confirmations: %s\n", strings.Join(details.CrewAssignmentConfirmationIDs, ", "))
	}
	if len(details.CrewRequirementCredentialClassificationIDs) > 0 {
		fmt.Fprintf(out, "Crew Requirement Credential Classifications: %s\n", strings.Join(details.CrewRequirementCredentialClassificationIDs, ", "))
	}
	if details.TimeSheetID != "" {
		fmt.Fprintf(out, "Time Sheet: %s\n", details.TimeSheetID)
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
	fmt.Fprintf(out, "Validating Overlapping: %t\n", details.IsValidatingOverlapping)
	fmt.Fprintf(out, "Only For Equipment Movement: %t\n", details.IsOnlyForEquipmentMovement)
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

	return nil
}

func relationshipIDList(rel jsonAPIRelationship) []string {
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
