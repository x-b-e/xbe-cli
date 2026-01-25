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

type laborRequirementsListOptions struct {
	BaseURL                       string
	Token                         string
	JSON                          bool
	NoAuth                        bool
	Limit                         int
	Offset                        int
	Sort                          string
	JobProductionPlan             string
	ResourceClassificationType    string
	ResourceClassificationID      string
	NotResourceClassificationType string
	ResourceType                  string
	ResourceID                    string
	NotResourceType               string
	Broker                        string
	Customer                      string
	ProjectManager                string
	Project                       string
	HasResource                   string
	StartAtMin                    string
	StartAtMax                    string
	IsStartAt                     string
	EndAtMin                      string
	EndAtMax                      string
	IsEndAt                       string
	IsOnlyForEquipmentMovement    string
	StartAtEffectiveMin           string
	StartAtEffectiveMax           string
	EndAtEffectiveMin             string
	EndAtEffectiveMax             string
	StartOnEffectiveMin           string
	StartOnEffectiveMax           string
	CalculatedMobilizationMethod  string
	JobProductionPlanStatus       string
	LaborRequirement              string
	LaborRequirementLaborer       string
	LaborRequirementLaborerID     string
	LaborRequirementUser          string
	LaborRequirementUserID        string
	RequiresInboundMovement       string
	RequiresOutboundMovement      string
	WithoutApprovedTimeSheet      string
	WithoutSubmittedTimeSheet     string
	IsExpectingTimeSheet          string
	LaborerUser                   string
	IsAssignmentCandidateFor      string
}

type laborRequirementRow struct {
	ID                     string `json:"id"`
	JobProductionPlanID    string `json:"job_production_plan_id,omitempty"`
	ResourceClassification string `json:"resource_classification_id,omitempty"`
	LaborerID              string `json:"laborer_id,omitempty"`
	StartAt                string `json:"start_at,omitempty"`
	EndAt                  string `json:"end_at,omitempty"`
}

func newLaborRequirementsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List labor requirements",
		Long: `List labor requirements.

Output Columns:
  ID        Requirement identifier
  JOB PLAN  Job production plan ID
  CLASS     Labor classification ID
  LABORER   Laborer ID
  START     Start time (effective)
  END       End time (effective)

Filters:
  --job-production-plan               Filter by job production plan ID
  --resource-classification-type      Filter by resource classification type (e.g., LaborClassification)
  --resource-classification-id        Filter by resource classification ID (use with --resource-classification-type)
  --not-resource-classification-type  Exclude resource classification type
  --resource-type                     Filter by resource type (e.g., Laborer)
  --resource-id                       Filter by resource ID (use with --resource-type)
  --not-resource-type                 Exclude resource type
  --broker                            Filter by broker ID
  --customer                          Filter by customer ID
  --project-manager                   Filter by project manager user ID
  --project                           Filter by project ID
  --has-resource                      Filter by assignment status (true/false)
  --start-at-min                      Filter by minimum start time (ISO 8601)
  --start-at-max                      Filter by maximum start time (ISO 8601)
  --is-start-at                       Filter by presence of start time (true/false)
  --end-at-min                        Filter by minimum end time (ISO 8601)
  --end-at-max                        Filter by maximum end time (ISO 8601)
  --is-end-at                         Filter by presence of end time (true/false)
  --is-only-for-equipment-movement    Filter by equipment-movement-only plans (true/false)
  --start-at-effective-min            Filter by minimum effective start time (ISO 8601)
  --start-at-effective-max            Filter by maximum effective start time (ISO 8601)
  --end-at-effective-min              Filter by minimum effective end time (ISO 8601)
  --end-at-effective-max              Filter by maximum effective end time (ISO 8601)
  --start-on-effective-min            Filter by minimum effective start date (YYYY-MM-DD)
  --start-on-effective-max            Filter by maximum effective start date (YYYY-MM-DD)
  --calculated-mobilization-method    Filter by calculated mobilization method
  --job-production-plan-status        Filter by job production plan status
  --labor-requirement                 Filter by labor requirement ID
  --labor-requirement-laborer         Filter by labor requirement laborer ID
  --labor-requirement-laborer-id      Filter by labor requirement laborer ID (via labor requirement)
  --labor-requirement-user            Filter by labor requirement user ID
  --labor-requirement-user-id         Filter by labor requirement user ID (via labor requirement)
  --requires-inbound-movement         Filter by inbound movement requirement (true/false)
  --requires-outbound-movement        Filter by outbound movement requirement (true/false)
  --without-approved-time-sheet       Filter by missing approved time sheet (true/false)
  --without-submitted-time-sheet      Filter by missing submitted time sheet (true/false)
  --is-expecting-time-sheet           Filter by time sheet expectation (true/false)
  --laborer-user                      Filter by laborer user ID
  --is-assignment-candidate-for       Filter by assignment candidate laborer ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List labor requirements
  xbe view labor-requirements list

  # Filter by job production plan
  xbe view labor-requirements list --job-production-plan 123

  # Filter by labor classification
  xbe view labor-requirements list --resource-classification-type LaborClassification --resource-classification-id 456

  # Filter by laborer assignment
  xbe view labor-requirements list --resource-type Laborer --resource-id 789

  # Filter by start window
  xbe view labor-requirements list --start-at-min 2026-01-01T00:00:00Z --start-at-max 2026-01-31T23:59:59Z

  # Output as JSON
  xbe view labor-requirements list --json`,
		Args: cobra.NoArgs,
		RunE: runLaborRequirementsList,
	}
	initLaborRequirementsListFlags(cmd)
	return cmd
}

func init() {
	laborRequirementsCmd.AddCommand(newLaborRequirementsListCmd())
}

func initLaborRequirementsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("job-production-plan", "", "Filter by job production plan ID")
	cmd.Flags().String("resource-classification-type", "", "Filter by resource classification type (e.g., LaborClassification)")
	cmd.Flags().String("resource-classification-id", "", "Filter by resource classification ID (use with --resource-classification-type)")
	cmd.Flags().String("not-resource-classification-type", "", "Exclude resource classification type")
	cmd.Flags().String("resource-type", "", "Filter by resource type (e.g., Laborer)")
	cmd.Flags().String("resource-id", "", "Filter by resource ID (use with --resource-type)")
	cmd.Flags().String("not-resource-type", "", "Exclude resource type")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("customer", "", "Filter by customer ID")
	cmd.Flags().String("project-manager", "", "Filter by project manager user ID")
	cmd.Flags().String("project", "", "Filter by project ID")
	cmd.Flags().String("has-resource", "", "Filter by assignment status (true/false)")
	cmd.Flags().String("start-at-min", "", "Filter by minimum start time (ISO 8601)")
	cmd.Flags().String("start-at-max", "", "Filter by maximum start time (ISO 8601)")
	cmd.Flags().String("is-start-at", "", "Filter by presence of start time (true/false)")
	cmd.Flags().String("end-at-min", "", "Filter by minimum end time (ISO 8601)")
	cmd.Flags().String("end-at-max", "", "Filter by maximum end time (ISO 8601)")
	cmd.Flags().String("is-end-at", "", "Filter by presence of end time (true/false)")
	cmd.Flags().String("is-only-for-equipment-movement", "", "Filter by equipment-movement-only plans (true/false)")
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
	cmd.Flags().String("labor-requirement-laborer-id", "", "Filter by labor requirement laborer ID (via labor requirement)")
	cmd.Flags().String("labor-requirement-user", "", "Filter by labor requirement user ID")
	cmd.Flags().String("labor-requirement-user-id", "", "Filter by labor requirement user ID (via labor requirement)")
	cmd.Flags().String("requires-inbound-movement", "", "Filter by inbound movement requirement (true/false)")
	cmd.Flags().String("requires-outbound-movement", "", "Filter by outbound movement requirement (true/false)")
	cmd.Flags().String("without-approved-time-sheet", "", "Filter by missing approved time sheet (true/false)")
	cmd.Flags().String("without-submitted-time-sheet", "", "Filter by missing submitted time sheet (true/false)")
	cmd.Flags().String("is-expecting-time-sheet", "", "Filter by time sheet expectation (true/false)")
	cmd.Flags().String("laborer-user", "", "Filter by laborer user ID")
	cmd.Flags().String("is-assignment-candidate-for", "", "Filter by assignment candidate laborer ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runLaborRequirementsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseLaborRequirementsListOptions(cmd)
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
	query.Set("include", "job-production-plan,resource-classification,resource,laborer")
	query.Set("fields[labor-requirements]", "start-at,end-at,start-at-effective,end-at-effective")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
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
	setFilterIfPresent(query, "filter[is-only-for-equipment-movement]", opts.IsOnlyForEquipmentMovement)
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
	setFilterIfPresent(query, "filter[without-approved-time-sheet]", opts.WithoutApprovedTimeSheet)
	setFilterIfPresent(query, "filter[without-submitted-time-sheet]", opts.WithoutSubmittedTimeSheet)
	setFilterIfPresent(query, "filter[is-expecting-time-sheet]", opts.IsExpectingTimeSheet)
	setFilterIfPresent(query, "filter[laborer-user]", opts.LaborerUser)
	setFilterIfPresent(query, "filter[is-assignment-candidate-for]", opts.IsAssignmentCandidateFor)

	if opts.ResourceClassificationType != "" && opts.ResourceClassificationID != "" {
		resourceClassificationType := normalizeResourceTypeForFilter(opts.ResourceClassificationType)
		query.Set("filter[resource-classification]", resourceClassificationType+"|"+opts.ResourceClassificationID)
	} else if opts.ResourceClassificationType != "" {
		resourceClassificationType := normalizeResourceTypeForFilter(opts.ResourceClassificationType)
		query.Set("filter[resource-classification-type]", resourceClassificationType)
	}
	if opts.NotResourceClassificationType != "" {
		resourceClassificationType := normalizeResourceTypeForFilter(opts.NotResourceClassificationType)
		query.Set("filter[not-resource-classification-type]", resourceClassificationType)
	}

	if opts.ResourceType != "" && opts.ResourceID != "" {
		resourceType := normalizeResourceTypeForFilter(opts.ResourceType)
		query.Set("filter[resource]", resourceType+"|"+opts.ResourceID)
	} else if opts.ResourceType != "" {
		resourceType := normalizeResourceTypeForFilter(opts.ResourceType)
		query.Set("filter[resource-type]", resourceType)
	}
	if opts.NotResourceType != "" {
		resourceType := normalizeResourceTypeForFilter(opts.NotResourceType)
		query.Set("filter[not-resource-type]", resourceType)
	}

	body, _, err := client.Get(cmd.Context(), "/v1/labor-requirements", query)
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

	rows := buildLaborRequirementRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderLaborRequirementsTable(cmd, rows)
}

func parseLaborRequirementsListOptions(cmd *cobra.Command) (laborRequirementsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	resourceClassificationType, _ := cmd.Flags().GetString("resource-classification-type")
	resourceClassificationID, _ := cmd.Flags().GetString("resource-classification-id")
	notResourceClassificationType, _ := cmd.Flags().GetString("not-resource-classification-type")
	resourceType, _ := cmd.Flags().GetString("resource-type")
	resourceID, _ := cmd.Flags().GetString("resource-id")
	notResourceType, _ := cmd.Flags().GetString("not-resource-type")
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
	isOnlyForEquipmentMovement, _ := cmd.Flags().GetString("is-only-for-equipment-movement")
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
	withoutApprovedTimeSheet, _ := cmd.Flags().GetString("without-approved-time-sheet")
	withoutSubmittedTimeSheet, _ := cmd.Flags().GetString("without-submitted-time-sheet")
	isExpectingTimeSheet, _ := cmd.Flags().GetString("is-expecting-time-sheet")
	laborerUser, _ := cmd.Flags().GetString("laborer-user")
	isAssignmentCandidateFor, _ := cmd.Flags().GetString("is-assignment-candidate-for")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return laborRequirementsListOptions{
		BaseURL:                       baseURL,
		Token:                         token,
		JSON:                          jsonOut,
		NoAuth:                        noAuth,
		Limit:                         limit,
		Offset:                        offset,
		Sort:                          sort,
		JobProductionPlan:             jobProductionPlan,
		ResourceClassificationType:    resourceClassificationType,
		ResourceClassificationID:      resourceClassificationID,
		NotResourceClassificationType: notResourceClassificationType,
		ResourceType:                  resourceType,
		ResourceID:                    resourceID,
		NotResourceType:               notResourceType,
		Broker:                        broker,
		Customer:                      customer,
		ProjectManager:                projectManager,
		Project:                       project,
		HasResource:                   hasResource,
		StartAtMin:                    startAtMin,
		StartAtMax:                    startAtMax,
		IsStartAt:                     isStartAt,
		EndAtMin:                      endAtMin,
		EndAtMax:                      endAtMax,
		IsEndAt:                       isEndAt,
		IsOnlyForEquipmentMovement:    isOnlyForEquipmentMovement,
		StartAtEffectiveMin:           startAtEffectiveMin,
		StartAtEffectiveMax:           startAtEffectiveMax,
		EndAtEffectiveMin:             endAtEffectiveMin,
		EndAtEffectiveMax:             endAtEffectiveMax,
		StartOnEffectiveMin:           startOnEffectiveMin,
		StartOnEffectiveMax:           startOnEffectiveMax,
		CalculatedMobilizationMethod:  calculatedMobilizationMethod,
		JobProductionPlanStatus:       jobProductionPlanStatus,
		LaborRequirement:              laborRequirement,
		LaborRequirementLaborer:       laborRequirementLaborer,
		LaborRequirementLaborerID:     laborRequirementLaborerID,
		LaborRequirementUser:          laborRequirementUser,
		LaborRequirementUserID:        laborRequirementUserID,
		RequiresInboundMovement:       requiresInboundMovement,
		RequiresOutboundMovement:      requiresOutboundMovement,
		WithoutApprovedTimeSheet:      withoutApprovedTimeSheet,
		WithoutSubmittedTimeSheet:     withoutSubmittedTimeSheet,
		IsExpectingTimeSheet:          isExpectingTimeSheet,
		LaborerUser:                   laborerUser,
		IsAssignmentCandidateFor:      isAssignmentCandidateFor,
	}, nil
}

func buildLaborRequirementRows(resp jsonAPIResponse) []laborRequirementRow {
	rows := make([]laborRequirementRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildLaborRequirementRow(resource))
	}
	return rows
}

func buildLaborRequirementRow(resource jsonAPIResource) laborRequirementRow {
	attrs := resource.Attributes
	row := laborRequirementRow{
		ID:      resource.ID,
		StartAt: formatDateTime(firstNonEmpty(stringAttr(attrs, "start-at-effective"), stringAttr(attrs, "start-at"))),
		EndAt:   formatDateTime(firstNonEmpty(stringAttr(attrs, "end-at-effective"), stringAttr(attrs, "end-at"))),
	}

	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		row.JobProductionPlanID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["resource-classification"]; ok && rel.Data != nil {
		row.ResourceClassification = rel.Data.ID
	}
	if rel, ok := resource.Relationships["laborer"]; ok && rel.Data != nil {
		row.LaborerID = rel.Data.ID
	} else if rel, ok := resource.Relationships["resource"]; ok && rel.Data != nil {
		row.LaborerID = rel.Data.ID
	}

	return row
}

func renderLaborRequirementsTable(cmd *cobra.Command, rows []laborRequirementRow) error {
	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tJOB PLAN\tCLASS\tLABORER\tSTART\tEND")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.JobProductionPlanID,
			row.ResourceClassification,
			row.LaborerID,
			row.StartAt,
			row.EndAt,
		)
	}
	return writer.Flush()
}
