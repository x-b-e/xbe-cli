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

type equipmentRequirementsListOptions struct {
	BaseURL                      string
	Token                        string
	JSON                         bool
	NoAuth                       bool
	Limit                        int
	Offset                       int
	Sort                         string
	JobProductionPlan            string
	ResourceClassificationType   string
	ResourceClassificationID     string
	ResourceType                 string
	ResourceID                   string
	Broker                       string
	Customer                     string
	ProjectManager               string
	Project                      string
	HasResource                  string
	StartAtMin                   string
	StartAtMax                   string
	IsStartAt                    string
	EndAtMin                     string
	EndAtMax                     string
	IsEndAt                      string
	StartAtEffectiveMin          string
	StartAtEffectiveMax          string
	EndAtEffectiveMin            string
	EndAtEffectiveMax            string
	StartOnEffectiveMin          string
	StartOnEffectiveMax          string
	CalculatedMobilizationMethod string
	JobProductionPlanStatus      string
	LaborRequirement             string
	LaborRequirementLaborer      string
	LaborRequirementLaborerID    string
	LaborRequirementUser         string
	LaborRequirementUserID       string
	RequiresInboundMovement      string
	RequiresOutboundMovement     string
	IsOnlyForEquipmentMovement   string
	WithoutApprovedTimeSheet     string
	WithoutSubmittedTimeSheet    string
	IsExpectingTimeSheet         string
	CreatedAtMin                 string
	CreatedAtMax                 string
	IsCreatedAt                  string
	UpdatedAtMin                 string
	UpdatedAtMax                 string
	IsUpdatedAt                  string
	IsAssignmentCandidateFor     string
}

type equipmentRequirementRow struct {
	ID                         string `json:"id"`
	JobProductionPlanID        string `json:"job_production_plan_id,omitempty"`
	ResourceClassificationType string `json:"resource_classification_type,omitempty"`
	ResourceClassificationID   string `json:"resource_classification_id,omitempty"`
	ResourceType               string `json:"resource_type,omitempty"`
	ResourceID                 string `json:"resource_id,omitempty"`
	StartAt                    string `json:"start_at,omitempty"`
	EndAt                      string `json:"end_at,omitempty"`
	MobilizationMethod         string `json:"mobilization_method,omitempty"`
	RequiresInboundMovement    bool   `json:"requires_inbound_movement,omitempty"`
	RequiresOutboundMovement   bool   `json:"requires_outbound_movement,omitempty"`
}

func newEquipmentRequirementsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List equipment requirements",
		Long: `List equipment requirements.

Output Columns:
  ID           Equipment requirement identifier
  JPP          Job production plan ID
  CLASS        Resource classification type/ID
  EQUIPMENT    Assigned equipment type/ID
  START AT     Requirement start time
  END AT       Requirement end time
  MOBILIZATION Mobilization method
  INBOUND      Requires inbound movement
  OUTBOUND     Requires outbound movement

Filters:
  --job-production-plan              Filter by job production plan ID
  --resource-classification-type     Filter by resource classification type (e.g., EquipmentClassification)
  --resource-classification-id       Filter by resource classification ID (requires --resource-classification-type)
  --resource-type                    Filter by resource type (e.g., Equipment)
  --resource-id                      Filter by resource ID (requires --resource-type)
  --broker                           Filter by broker ID
  --customer                         Filter by customer ID
  --project-manager                  Filter by project manager user ID
  --project                          Filter by project ID
  --has-resource                     Filter by has resource (true/false)
  --start-at-min                     Filter by minimum start time (ISO 8601)
  --start-at-max                     Filter by maximum start time (ISO 8601)
  --is-start-at                      Filter by has start time (true/false)
  --end-at-min                       Filter by minimum end time (ISO 8601)
  --end-at-max                       Filter by maximum end time (ISO 8601)
  --is-end-at                        Filter by has end time (true/false)
  --start-at-effective-min           Filter by minimum effective start time (ISO 8601)
  --start-at-effective-max           Filter by maximum effective start time (ISO 8601)
  --end-at-effective-min             Filter by minimum effective end time (ISO 8601)
  --end-at-effective-max             Filter by maximum effective end time (ISO 8601)
  --start-on-effective-min           Filter by minimum effective start date (YYYY-MM-DD)
  --start-on-effective-max           Filter by maximum effective start date (YYYY-MM-DD)
  --calculated-mobilization-method   Filter by calculated mobilization method
  --job-production-plan-status       Filter by job production plan status
  --labor-requirement                Filter by labor requirement ID
  --labor-requirement-laborer        Filter by labor requirement laborer ID
  --labor-requirement-laborer-id     Filter by laborer resource ID (via labor requirement)
  --labor-requirement-user           Filter by labor requirement user ID
  --labor-requirement-user-id        Filter by user ID (via labor requirement)
  --requires-inbound-movement        Filter by requires inbound movement (true/false)
  --requires-outbound-movement       Filter by requires outbound movement (true/false)
  --is-only-for-equipment-movement   Filter by only-for-equipment-movement (true/false)
  --without-approved-time-sheet      Filter by without approved time sheet (true/false)
  --without-submitted-time-sheet     Filter by without submitted time sheet (true/false)
  --is-expecting-time-sheet          Filter by expecting time sheet (true/false)
  --created-at-min                   Filter by created-at on/after (ISO 8601)
  --created-at-max                   Filter by created-at on/before (ISO 8601)
  --is-created-at                    Filter by has created-at (true/false)
  --updated-at-min                   Filter by updated-at on/after (ISO 8601)
  --updated-at-max                   Filter by updated-at on/before (ISO 8601)
  --is-updated-at                    Filter by has updated-at (true/false)
  --is-assignment-candidate-for      Filter by equipment ID for assignment candidates

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List equipment requirements
  xbe view equipment-requirements list

  # Filter by job production plan
  xbe view equipment-requirements list --job-production-plan 123

  # Filter by resource classification
  xbe view equipment-requirements list --resource-classification-type EquipmentClassification --resource-classification-id 456

  # Filter by start time range
  xbe view equipment-requirements list --start-at-min 2025-01-01T00:00:00Z --start-at-max 2025-01-31T23:59:59Z

  # Output as JSON
  xbe view equipment-requirements list --json`,
		Args: cobra.NoArgs,
		RunE: runEquipmentRequirementsList,
	}
	initEquipmentRequirementsListFlags(cmd)
	return cmd
}

func init() {
	equipmentRequirementsCmd.AddCommand(newEquipmentRequirementsListCmd())
}

func initEquipmentRequirementsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("job-production-plan", "", "Filter by job production plan ID")
	cmd.Flags().String("resource-classification-type", "", "Filter by resource classification type (e.g., EquipmentClassification)")
	cmd.Flags().String("resource-classification-id", "", "Filter by resource classification ID (requires --resource-classification-type)")
	cmd.Flags().String("resource-type", "", "Filter by resource type (e.g., Equipment)")
	cmd.Flags().String("resource-id", "", "Filter by resource ID (requires --resource-type)")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("customer", "", "Filter by customer ID")
	cmd.Flags().String("project-manager", "", "Filter by project manager user ID")
	cmd.Flags().String("project", "", "Filter by project ID")
	cmd.Flags().String("has-resource", "", "Filter by has resource (true/false)")
	cmd.Flags().String("start-at-min", "", "Filter by minimum start time (ISO 8601)")
	cmd.Flags().String("start-at-max", "", "Filter by maximum start time (ISO 8601)")
	cmd.Flags().String("is-start-at", "", "Filter by has start time (true/false)")
	cmd.Flags().String("end-at-min", "", "Filter by minimum end time (ISO 8601)")
	cmd.Flags().String("end-at-max", "", "Filter by maximum end time (ISO 8601)")
	cmd.Flags().String("is-end-at", "", "Filter by has end time (true/false)")
	cmd.Flags().String("start-at-effective-min", "", "Filter by minimum effective start time (ISO 8601)")
	cmd.Flags().String("start-at-effective-max", "", "Filter by maximum effective start time (ISO 8601)")
	cmd.Flags().String("end-at-effective-min", "", "Filter by minimum effective end time (ISO 8601)")
	cmd.Flags().String("end-at-effective-max", "", "Filter by maximum effective end time (ISO 8601)")
	cmd.Flags().String("start-on-effective-min", "", "Filter by minimum effective start date (YYYY-MM-DD)")
	cmd.Flags().String("start-on-effective-max", "", "Filter by maximum effective start date (YYYY-MM-DD)")
	cmd.Flags().String("calculated-mobilization-method", "", "Filter by calculated mobilization method")
	cmd.Flags().String("job-production-plan-status", "", "Filter by job production plan status")
	cmd.Flags().String("labor-requirement", "", "Filter by labor requirement ID")
	cmd.Flags().String("labor-requirement-laborer", "", "Filter by labor requirement laborer ID")
	cmd.Flags().String("labor-requirement-laborer-id", "", "Filter by laborer resource ID (via labor requirement)")
	cmd.Flags().String("labor-requirement-user", "", "Filter by labor requirement user ID")
	cmd.Flags().String("labor-requirement-user-id", "", "Filter by user ID (via labor requirement)")
	cmd.Flags().String("requires-inbound-movement", "", "Filter by requires inbound movement (true/false)")
	cmd.Flags().String("requires-outbound-movement", "", "Filter by requires outbound movement (true/false)")
	cmd.Flags().String("is-only-for-equipment-movement", "", "Filter by only-for-equipment-movement (true/false)")
	cmd.Flags().String("without-approved-time-sheet", "", "Filter by without approved time sheet (true/false)")
	cmd.Flags().String("without-submitted-time-sheet", "", "Filter by without submitted time sheet (true/false)")
	cmd.Flags().String("is-expecting-time-sheet", "", "Filter by expecting time sheet (true/false)")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by has created-at (true/false)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("is-updated-at", "", "Filter by has updated-at (true/false)")
	cmd.Flags().String("is-assignment-candidate-for", "", "Filter by equipment ID for assignment candidates")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runEquipmentRequirementsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseEquipmentRequirementsListOptions(cmd)
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
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run xbe auth login first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	if opts.ResourceClassificationID != "" && opts.ResourceClassificationType == "" {
		return fmt.Errorf("--resource-classification-type is required when --resource-classification-id is set")
	}
	if opts.ResourceID != "" && opts.ResourceType == "" {
		return fmt.Errorf("--resource-type is required when --resource-id is set")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[equipment-requirements]", "start-at,end-at,mobilization-method,requires-inbound-movement,requires-outbound-movement,job-production-plan,resource-classification,resource")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	resourceClassType := normalizeResourceTypeForFilter(opts.ResourceClassificationType)
	if resourceClassType != "" && opts.ResourceClassificationID != "" {
		query.Set("filter[resource_classification]", resourceClassType+"|"+opts.ResourceClassificationID)
	} else if resourceClassType != "" {
		query.Set("filter[resource_classification_type]", resourceClassType)
	}

	resourceType := normalizeResourceTypeForFilter(opts.ResourceType)
	if resourceType != "" && opts.ResourceID != "" {
		query.Set("filter[resource]", resourceType+"|"+opts.ResourceID)
	} else if resourceType != "" {
		query.Set("filter[resource_type]", resourceType)
	}

	setFilterIfPresent(query, "filter[job-production-plan]", opts.JobProductionPlan)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[customer]", opts.Customer)
	setFilterIfPresent(query, "filter[project-manager]", opts.ProjectManager)
	setFilterIfPresent(query, "filter[project]", opts.Project)
	setFilterIfPresent(query, "filter[has-resource]", opts.HasResource)
	setFilterIfPresent(query, "filter[start-at-min]", opts.StartAtMin)
	setFilterIfPresent(query, "filter[start-at-max]", opts.StartAtMax)
	setFilterIfPresent(query, "filter[is-start-at]", opts.IsStartAt)
	setFilterIfPresent(query, "filter[end-at-min]", opts.EndAtMin)
	setFilterIfPresent(query, "filter[end-at-max]", opts.EndAtMax)
	setFilterIfPresent(query, "filter[is-end-at]", opts.IsEndAt)
	setFilterIfPresent(query, "filter[start-at-effective-min]", opts.StartAtEffectiveMin)
	setFilterIfPresent(query, "filter[start-at-effective-max]", opts.StartAtEffectiveMax)
	setFilterIfPresent(query, "filter[end-at-effective-min]", opts.EndAtEffectiveMin)
	setFilterIfPresent(query, "filter[end-at-effective-max]", opts.EndAtEffectiveMax)
	setFilterIfPresent(query, "filter[start-on-effective-min]", opts.StartOnEffectiveMin)
	setFilterIfPresent(query, "filter[start-on-effective-max]", opts.StartOnEffectiveMax)
	setFilterIfPresent(query, "filter[calculated-mobilization-method]", opts.CalculatedMobilizationMethod)
	setFilterIfPresent(query, "filter[job-production-plan-status]", opts.JobProductionPlanStatus)
	setFilterIfPresent(query, "filter[labor-requirement]", opts.LaborRequirement)
	setFilterIfPresent(query, "filter[labor-requirement-laborer]", opts.LaborRequirementLaborer)
	setFilterIfPresent(query, "filter[labor-requirement-laborer-id]", opts.LaborRequirementLaborerID)
	setFilterIfPresent(query, "filter[labor-requirement-user]", opts.LaborRequirementUser)
	setFilterIfPresent(query, "filter[labor-requirement-user-id]", opts.LaborRequirementUserID)
	setFilterIfPresent(query, "filter[requires-inbound-movement]", opts.RequiresInboundMovement)
	setFilterIfPresent(query, "filter[requires-outbound-movement]", opts.RequiresOutboundMovement)
	setFilterIfPresent(query, "filter[is-only-for-equipment-movement]", opts.IsOnlyForEquipmentMovement)
	setFilterIfPresent(query, "filter[without-approved-time-sheet]", opts.WithoutApprovedTimeSheet)
	setFilterIfPresent(query, "filter[without-submitted-time-sheet]", opts.WithoutSubmittedTimeSheet)
	setFilterIfPresent(query, "filter[is-expecting-time-sheet]", opts.IsExpectingTimeSheet)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[is-created-at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[is-updated-at]", opts.IsUpdatedAt)
	setFilterIfPresent(query, "filter[is-assignment-candidate-for]", opts.IsAssignmentCandidateFor)

	body, _, err := client.Get(cmd.Context(), "/v1/equipment-requirements", query)
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

	rows := buildEquipmentRequirementRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderEquipmentRequirementsTable(cmd, rows)
}

func parseEquipmentRequirementsListOptions(cmd *cobra.Command) (equipmentRequirementsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	resourceClassificationType, _ := cmd.Flags().GetString("resource-classification-type")
	resourceClassificationID, _ := cmd.Flags().GetString("resource-classification-id")
	resourceType, _ := cmd.Flags().GetString("resource-type")
	resourceID, _ := cmd.Flags().GetString("resource-id")
	broker, _ := cmd.Flags().GetString("broker")
	customer, _ := cmd.Flags().GetString("customer")
	projectManager, _ := cmd.Flags().GetString("project-manager")
	project, _ := cmd.Flags().GetString("project")
	hasResource, _ := cmd.Flags().GetString("has-resource")
	startAtMin, _ := cmd.Flags().GetString("start-at-min")
	startAtMax, _ := cmd.Flags().GetString("start-at-max")
	isStartAt, _ := cmd.Flags().GetString("is-start-at")
	endAtMin, _ := cmd.Flags().GetString("end-at-min")
	endAtMax, _ := cmd.Flags().GetString("end-at-max")
	isEndAt, _ := cmd.Flags().GetString("is-end-at")
	startAtEffectiveMin, _ := cmd.Flags().GetString("start-at-effective-min")
	startAtEffectiveMax, _ := cmd.Flags().GetString("start-at-effective-max")
	endAtEffectiveMin, _ := cmd.Flags().GetString("end-at-effective-min")
	endAtEffectiveMax, _ := cmd.Flags().GetString("end-at-effective-max")
	startOnEffectiveMin, _ := cmd.Flags().GetString("start-on-effective-min")
	startOnEffectiveMax, _ := cmd.Flags().GetString("start-on-effective-max")
	calculatedMobilizationMethod, _ := cmd.Flags().GetString("calculated-mobilization-method")
	jobProductionPlanStatus, _ := cmd.Flags().GetString("job-production-plan-status")
	laborRequirement, _ := cmd.Flags().GetString("labor-requirement")
	laborRequirementLaborer, _ := cmd.Flags().GetString("labor-requirement-laborer")
	laborRequirementLaborerID, _ := cmd.Flags().GetString("labor-requirement-laborer-id")
	laborRequirementUser, _ := cmd.Flags().GetString("labor-requirement-user")
	laborRequirementUserID, _ := cmd.Flags().GetString("labor-requirement-user-id")
	requiresInboundMovement, _ := cmd.Flags().GetString("requires-inbound-movement")
	requiresOutboundMovement, _ := cmd.Flags().GetString("requires-outbound-movement")
	isOnlyForEquipmentMovement, _ := cmd.Flags().GetString("is-only-for-equipment-movement")
	withoutApprovedTimeSheet, _ := cmd.Flags().GetString("without-approved-time-sheet")
	withoutSubmittedTimeSheet, _ := cmd.Flags().GetString("without-submitted-time-sheet")
	isExpectingTimeSheet, _ := cmd.Flags().GetString("is-expecting-time-sheet")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	isCreatedAt, _ := cmd.Flags().GetString("is-created-at")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	isUpdatedAt, _ := cmd.Flags().GetString("is-updated-at")
	isAssignmentCandidateFor, _ := cmd.Flags().GetString("is-assignment-candidate-for")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return equipmentRequirementsListOptions{
		BaseURL:                      baseURL,
		Token:                        token,
		JSON:                         jsonOut,
		NoAuth:                       noAuth,
		Limit:                        limit,
		Offset:                       offset,
		Sort:                         sort,
		JobProductionPlan:            jobProductionPlan,
		ResourceClassificationType:   resourceClassificationType,
		ResourceClassificationID:     resourceClassificationID,
		ResourceType:                 resourceType,
		ResourceID:                   resourceID,
		Broker:                       broker,
		Customer:                     customer,
		ProjectManager:               projectManager,
		Project:                      project,
		HasResource:                  hasResource,
		StartAtMin:                   startAtMin,
		StartAtMax:                   startAtMax,
		IsStartAt:                    isStartAt,
		EndAtMin:                     endAtMin,
		EndAtMax:                     endAtMax,
		IsEndAt:                      isEndAt,
		StartAtEffectiveMin:          startAtEffectiveMin,
		StartAtEffectiveMax:          startAtEffectiveMax,
		EndAtEffectiveMin:            endAtEffectiveMin,
		EndAtEffectiveMax:            endAtEffectiveMax,
		StartOnEffectiveMin:          startOnEffectiveMin,
		StartOnEffectiveMax:          startOnEffectiveMax,
		CalculatedMobilizationMethod: calculatedMobilizationMethod,
		JobProductionPlanStatus:      jobProductionPlanStatus,
		LaborRequirement:             laborRequirement,
		LaborRequirementLaborer:      laborRequirementLaborer,
		LaborRequirementLaborerID:    laborRequirementLaborerID,
		LaborRequirementUser:         laborRequirementUser,
		LaborRequirementUserID:       laborRequirementUserID,
		RequiresInboundMovement:      requiresInboundMovement,
		RequiresOutboundMovement:     requiresOutboundMovement,
		IsOnlyForEquipmentMovement:   isOnlyForEquipmentMovement,
		WithoutApprovedTimeSheet:     withoutApprovedTimeSheet,
		WithoutSubmittedTimeSheet:    withoutSubmittedTimeSheet,
		IsExpectingTimeSheet:         isExpectingTimeSheet,
		CreatedAtMin:                 createdAtMin,
		CreatedAtMax:                 createdAtMax,
		IsCreatedAt:                  isCreatedAt,
		UpdatedAtMin:                 updatedAtMin,
		UpdatedAtMax:                 updatedAtMax,
		IsUpdatedAt:                  isUpdatedAt,
		IsAssignmentCandidateFor:     isAssignmentCandidateFor,
	}, nil
}

func buildEquipmentRequirementRows(resp jsonAPIResponse) []equipmentRequirementRow {
	rows := make([]equipmentRequirementRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildEquipmentRequirementRow(resource))
	}
	return rows
}

func buildEquipmentRequirementRow(resource jsonAPIResource) equipmentRequirementRow {
	attrs := resource.Attributes
	row := equipmentRequirementRow{
		ID:                       resource.ID,
		StartAt:                  formatDateTime(stringAttr(attrs, "start-at")),
		EndAt:                    formatDateTime(stringAttr(attrs, "end-at")),
		MobilizationMethod:       stringAttr(attrs, "mobilization-method"),
		RequiresInboundMovement:  boolAttr(attrs, "requires-inbound-movement"),
		RequiresOutboundMovement: boolAttr(attrs, "requires-outbound-movement"),
	}

	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		row.JobProductionPlanID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["resource-classification"]; ok && rel.Data != nil {
		row.ResourceClassificationType = rel.Data.Type
		row.ResourceClassificationID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["resource"]; ok && rel.Data != nil {
		row.ResourceType = rel.Data.Type
		row.ResourceID = rel.Data.ID
	}

	return row
}

func renderEquipmentRequirementsTable(cmd *cobra.Command, rows []equipmentRequirementRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No equipment requirements found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tJPP\tCLASS\tEQUIPMENT\tSTART AT\tEND AT\tMOBILIZATION\tINBOUND\tOUTBOUND")
	for _, row := range rows {
		resourceClass := ""
		if row.ResourceClassificationType != "" && row.ResourceClassificationID != "" {
			resourceClass = row.ResourceClassificationType + "/" + row.ResourceClassificationID
		}
		resource := ""
		if row.ResourceType != "" && row.ResourceID != "" {
			resource = row.ResourceType + "/" + row.ResourceID
		}

		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%t\t%t\n",
			row.ID,
			row.JobProductionPlanID,
			truncateString(resourceClass, 32),
			truncateString(resource, 32),
			row.StartAt,
			row.EndAt,
			row.MobilizationMethod,
			row.RequiresInboundMovement,
			row.RequiresOutboundMovement,
		)
	}
	return writer.Flush()
}

func buildEquipmentRequirementRowFromSingle(resp jsonAPISingleResponse) equipmentRequirementRow {
	return buildEquipmentRequirementRow(resp.Data)
}
