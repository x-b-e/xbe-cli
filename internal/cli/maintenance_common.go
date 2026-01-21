package cli

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
)

// BUEquipmentContext holds the equipment and classification IDs for a set of business units.
// This is used for client-side filtering of rules and requirement sets.
type BUEquipmentContext struct {
	BusinessUnitIDs   []string
	EquipmentIDs      []string
	ClassificationIDs []string
}

// getCurrentUserBusinessUnitIDs returns the BU IDs for the current user.
// This is used to implement the --me flag across maintenance commands.
func getCurrentUserBusinessUnitIDs(cmd *cobra.Command, client *api.Client) ([]string, error) {
	// 1. GET /v1/users/me → user_id
	userBody, _, err := client.Get(cmd.Context(), "/v1/users/me", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	var userResp jsonAPISingleResponse
	if err := json.Unmarshal(userBody, &userResp); err != nil {
		return nil, fmt.Errorf("failed to parse user response: %w", err)
	}

	if userResp.Data.ID == "" {
		return nil, fmt.Errorf("no user ID in response")
	}

	// 2. GET /v1/memberships?filter[user]={user_id} → membership_ids
	membershipQuery := url.Values{}
	membershipQuery.Set("filter[user]", userResp.Data.ID)
	membershipQuery.Set("fields[memberships]", "id")
	membershipQuery.Set("page[limit]", "500")

	membershipsBody, _, err := client.Get(cmd.Context(), "/v1/memberships", membershipQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get memberships: %w", err)
	}

	var membershipsResp jsonAPIResponse
	if err := json.Unmarshal(membershipsBody, &membershipsResp); err != nil {
		return nil, fmt.Errorf("failed to parse memberships response: %w", err)
	}

	if len(membershipsResp.Data) == 0 {
		return nil, fmt.Errorf("no memberships found for current user")
	}

	membershipIDs := make([]string, 0, len(membershipsResp.Data))
	for _, m := range membershipsResp.Data {
		membershipIDs = append(membershipIDs, m.ID)
	}

	// 3. GET /v1/business-unit-memberships?filter[membership]={membership_ids} → bu_ids
	buMembershipQuery := url.Values{}
	buMembershipQuery.Set("filter[membership]", strings.Join(membershipIDs, ","))
	buMembershipQuery.Set("include", "business-unit")
	buMembershipQuery.Set("page[limit]", "500")

	buMembershipsBody, _, err := client.Get(cmd.Context(), "/v1/business-unit-memberships", buMembershipQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get business unit memberships: %w", err)
	}

	var buMembershipsResp jsonAPIResponse
	if err := json.Unmarshal(buMembershipsBody, &buMembershipsResp); err != nil {
		return nil, fmt.Errorf("failed to parse business unit memberships response: %w", err)
	}

	buIDSet := make(map[string]struct{})
	for _, bum := range buMembershipsResp.Data {
		if rel, ok := bum.Relationships["business-unit"]; ok && rel.Data != nil {
			buIDSet[rel.Data.ID] = struct{}{}
		}
	}

	buIDs := make([]string, 0, len(buIDSet))
	for id := range buIDSet {
		buIDs = append(buIDs, id)
	}

	if len(buIDs) == 0 {
		return nil, fmt.Errorf("no business units found for current user")
	}

	return buIDs, nil
}

// getBUEquipmentContext fetches equipment for BUs and extracts IDs + classification IDs.
// This context is used for client-side filtering of rules and requirement sets.
func getBUEquipmentContext(cmd *cobra.Command, client *api.Client, buIDs []string) (*BUEquipmentContext, error) {
	// GET /v1/equipment?filter[business-unit]={bu_ids}&filter[is_active]=true
	//     include=equipment-classification
	query := url.Values{}
	query.Set("filter[business-unit]", strings.Join(buIDs, ","))
	query.Set("filter[is_active]", "true")
	query.Set("include", "equipment-classification")
	query.Set("page[limit]", "1000") // High limit to get all equipment

	body, _, err := client.Get(cmd.Context(), "/v1/equipment", query)
	if err != nil {
		return nil, fmt.Errorf("failed to get equipment: %w", err)
	}

	var resp jsonAPIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse equipment response: %w", err)
	}

	equipmentIDs := make([]string, 0, len(resp.Data))
	classificationIDSet := make(map[string]struct{})

	for _, eq := range resp.Data {
		equipmentIDs = append(equipmentIDs, eq.ID)
		// Get classification ID from relationship
		if rel, ok := eq.Relationships["equipment-classification"]; ok && rel.Data != nil {
			classificationIDSet[rel.Data.ID] = struct{}{}
		}
	}

	classificationIDs := make([]string, 0, len(classificationIDSet))
	for id := range classificationIDSet {
		classificationIDs = append(classificationIDs, id)
	}

	return &BUEquipmentContext{
		BusinessUnitIDs:   buIDs,
		EquipmentIDs:      equipmentIDs,
		ClassificationIDs: classificationIDs,
	}, nil
}

// canAccessRule determines if a rule belongs to the given BU context.
// A rule belongs to a BU if:
//   - Case 1: rule.business_unit_id is in bu_ids (direct BU ownership)
//   - Case 2: rule.equipment_id is in bu_equipment_ids (I own the equipment)
//   - Case 3: rule.classification_id is in bu_classification_ids AND rule has NO BU AND NO equipment
//     (broad classification rule that applies to everyone with that classification)
func canAccessRule(rule jsonAPIResource, ctx *BUEquipmentContext) bool {
	buID := getRelationshipID(rule, "business-unit")
	equipmentID := getRelationshipID(rule, "equipment")
	classificationID := getRelationshipID(rule, "equipment-classification")

	// Case 1: Direct BU match
	if buID != "" && containsString(ctx.BusinessUnitIDs, buID) {
		return true
	}
	// Case 2: Equipment match - if I own the equipment, the rule applies to me
	if equipmentID != "" && containsString(ctx.EquipmentIDs, equipmentID) {
		return true
	}
	// Case 3: Classification match (ONLY if rule has NO BU AND NO equipment)
	// If a rule has a BU assigned, it's scoped to that BU only, not a broad classification rule
	if classificationID != "" && equipmentID == "" && buID == "" && containsString(ctx.ClassificationIDs, classificationID) {
		return true
	}
	return false
}

// canAccessRequirementSet determines if a set belongs to the given BU context.
// Logic:
//   - If set.classification is in bu_classification_ids:
//   - If no requirements have equipment → INCLUDE
//   - Else: at least one req.equipment must be in bu_equipment_ids → INCLUDE
//   - Else: at least one req.equipment must be in bu_equipment_ids → INCLUDE
func canAccessRequirementSet(set jsonAPIResource, included map[string]jsonAPIResource, ctx *BUEquipmentContext) bool {
	classificationID := getRelationshipID(set, "equipment-classification")
	requirements := getSetRequirements(set, included)

	// If classification is in our BU's classifications
	if classificationID != "" && containsString(ctx.ClassificationIDs, classificationID) {
		// If no requirements have equipment, include the set
		if len(requirements) == 0 {
			return true
		}
		// Check if any requirement has equipment in our BUs
		hasEquipmentInBU := false
		for _, req := range requirements {
			equipID := getRelationshipID(req, "equipment")
			if equipID != "" && containsString(ctx.EquipmentIDs, equipID) {
				hasEquipmentInBU = true
				break
			}
		}
		// If some requirements exist, at least one must have equipment in our BU
		// If none have equipment at all, we still include (classification match is enough)
		anyHasEquipment := false
		for _, req := range requirements {
			if getRelationshipID(req, "equipment") != "" {
				anyHasEquipment = true
				break
			}
		}
		if !anyHasEquipment {
			return true
		}
		return hasEquipmentInBU
	}

	// Classification not in our BU - check if any requirement's equipment is in our BU
	for _, req := range requirements {
		equipID := getRelationshipID(req, "equipment")
		if equipID != "" && containsString(ctx.EquipmentIDs, equipID) {
			return true
		}
	}
	return false
}

// getSetRequirements extracts requirements from a set's relationships and included data
func getSetRequirements(set jsonAPIResource, included map[string]jsonAPIResource) []jsonAPIResource {
	rel, ok := set.Relationships["maintenance-requirements"]
	if !ok || rel.raw == nil {
		return nil
	}

	var refs []jsonAPIResourceIdentifier
	if err := json.Unmarshal(rel.raw, &refs); err != nil {
		return nil
	}

	requirements := make([]jsonAPIResource, 0, len(refs))
	for _, ref := range refs {
		key := resourceKey(ref.Type, ref.ID)
		if req, ok := included[key]; ok {
			requirements = append(requirements, req)
		}
	}
	return requirements
}

// getRelationshipID extracts the ID from a relationship, or empty string if not found
func getRelationshipID(resource jsonAPIResource, relationshipName string) string {
	rel, ok := resource.Relationships[relationshipName]
	if !ok || rel.Data == nil {
		return ""
	}
	return rel.Data.ID
}

// containsString checks if a slice contains a specific string
func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// getRuleScopeInfo returns the scope level and display string for a rule.
// Matches server's rule_level_description and client's getBUTagDisplay.
func getRuleScopeInfo(equipmentID, equipmentName, classificationID, classificationName, buID, buName string) (scopeLevel, scopeDisplay string) {
	if equipmentID != "" {
		displayName := equipmentName
		if displayName == "" {
			displayName = equipmentID
		}
		return "equipment", fmt.Sprintf("Equipment: %s", truncateString(displayName, 12))
	}
	if classificationID != "" {
		displayName := classificationName
		if displayName == "" {
			displayName = "Classification"
		}
		return "classification", displayName
	}
	if buID != "" {
		displayName := buName
		if displayName == "" {
			displayName = buID
		}
		return "business_unit", fmt.Sprintf("BU: %s", truncateString(displayName, 12))
	}
	return "broker", "Branch Level"
}

// Row types for list commands

type requirementRow struct {
	ID          string `json:"id"`
	Status      string `json:"status,omitempty"`
	Description string `json:"description,omitempty"`
	EquipmentID string `json:"equipment_id,omitempty"`
	Equipment   string `json:"equipment,omitempty"`
	SetID       string `json:"set_id,omitempty"`
	SetName     string `json:"set_name,omitempty"`
	DueDate     string `json:"due_date,omitempty"`
	IsTemplate  bool   `json:"is_template,omitempty"`
}

type setRow struct {
	ID              string `json:"id"`
	Status          string `json:"status,omitempty"`
	Type            string `json:"type,omitempty"`
	Name            string `json:"name,omitempty"`
	EquipmentID     string `json:"equipment_id,omitempty"`
	Equipment       string `json:"equipment,omitempty"`
	RequirementIDs  string `json:"requirement_ids,omitempty"`
	CompletionCount int    `json:"completion_count,omitempty"`
	TotalCount      int    `json:"total_count,omitempty"`
}

type ruleRow struct {
	ID                        string `json:"id"`
	Name                      string `json:"name,omitempty"`
	MaintenanceType           string `json:"maintenance_type,omitempty"`
	ScopeLevel                string `json:"scope_level,omitempty"` // "equipment", "classification", "business_unit", "broker"
	Scope                     string `json:"scope,omitempty"`       // Display string
	IsActive                  bool   `json:"is_active"`
	BusinessUnitID            string `json:"business_unit_id,omitempty"`
	BusinessUnit              string `json:"business_unit,omitempty"`
	EquipmentID               string `json:"equipment_id,omitempty"`
	Equipment                 string `json:"equipment,omitempty"`
	EquipmentClassificationID string `json:"equipment_classification_id,omitempty"`
	EquipmentClassification   string `json:"equipment_classification,omitempty"`
}

type partRow struct {
	ID            string  `json:"id"`
	PartNumber    string  `json:"part_number,omitempty"`
	Name          string  `json:"name,omitempty"`
	Description   string  `json:"description,omitempty"`
	Manufacturer  string  `json:"manufacturer,omitempty"`
	UnitCost      float64 `json:"unit_cost,omitempty"`
	UnitOfMeasure string  `json:"unit_of_measure,omitempty"`
}

// Helper for maintenance commands

func formatMaintenanceDateTime(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed.Format("2006-01-02 15:04")
	}
	return value
}

func intAttr(attrs map[string]any, key string) int {
	if attrs == nil {
		return 0
	}
	value, ok := attrs[key]
	if !ok || value == nil {
		return 0
	}
	switch typed := value.(type) {
	case float64:
		return int(typed)
	case int:
		return typed
	case int64:
		return int(typed)
	default:
		return 0
	}
}

func float64Attr(attrs map[string]any, key string) float64 {
	if attrs == nil {
		return 0
	}
	value, ok := attrs[key]
	if !ok || value == nil {
		return 0
	}
	switch typed := value.(type) {
	case float64:
		return typed
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	default:
		return 0
	}
}
