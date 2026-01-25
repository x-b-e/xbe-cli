package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doJobProductionPlanDuplicationsCreateOptions struct {
	BaseURL                                                  string
	Token                                                    string
	JSON                                                     bool
	JobProductionPlanTemplate                                string
	DerivedJobProductionPlan                                 string
	NewCustomer                                              string
	DerivedJobProductionPlanTemplateName                     string
	StartOn                                                  string
	SkipTemplateDuplicationValidation                        bool
	DisableOverlappingCrewRequirementIsValidatingOverlapping bool
	IsAsync                                                  bool
	SkipJobProductionPlanAlarms                              bool
	SkipJobProductionPlanLocations                           bool
	SkipJobProductionPlanSafetyRisks                         bool
	SkipJobProductionPlanMaterialSites                       bool
	SkipEquipmentRequirements                                bool
	SkipLaborRequirements                                    bool
	SkipJobProductionPlanMaterialTypes                       bool
	SkipJobProductionPlanServiceTypeUnitOfMeasures           bool
	SkipJobProductionPlanDisplayUnitOfMeasures               bool
	SkipJobProductionPlanServiceTypeUnitOfMeasureCohorts     bool
	SkipJobProductionPlanTrailerClassifications              bool
	SkipJobProductionPlanSegmentSets                         bool
	SkipJobProductionPlanSegments                            bool
	SkipJobProductionPlanSubscriptions                       bool
	SkipJobProductionPlanTimeCardApprovers                   bool
	SkipJobScheduleShifts                                    bool
	SkipDeveloperReferences                                  bool
	SkipJobProductionPlanInspectors                          bool
	SkipJobProductionPlanProjectPhaseRevenueItems            bool
	SkipEquipmentRequirementsResource                        bool
	SkipLaborRequirementsResource                            bool
	SkipLaborRequirementsCraftClass                          bool
	SkipJobScheduleShiftsDriverAssignmentRuleTextCached      bool
	SkipLaborRequirementsOverlappingResource                 bool
	SkippedLaborRequirementsOverlappingResourceIDs           string
	SkippedLaborRequirementsNotValidToAssignIDs              string
}

type jobProductionPlanDuplicationRow struct {
	ID                                                       string `json:"id"`
	JobProductionPlanTemplateID                              string `json:"job_production_plan_template_id,omitempty"`
	DerivedJobProductionPlanID                               string `json:"derived_job_production_plan_id,omitempty"`
	NewCustomerID                                            string `json:"new_customer_id,omitempty"`
	WorkID                                                   string `json:"work_id,omitempty"`
	DuplicationToken                                         string `json:"duplication_token,omitempty"`
	DerivedJobProductionPlanTemplateName                     string `json:"derived_job_production_plan_template_name,omitempty"`
	StartOn                                                  string `json:"start_on,omitempty"`
	IsAsync                                                  bool   `json:"is_async,omitempty"`
	SkipTemplateDuplicationValidation                        bool   `json:"skip_template_duplication_validation,omitempty"`
	DisableOverlappingCrewRequirementIsValidatingOverlapping bool   `json:"disable_overlapping_crew_requirement_is_validating_overlapping,omitempty"`
	UnsavedRelations                                         any    `json:"unsaved_relations,omitempty"`
	InvalidTemplateRelations                                 any    `json:"invalid_template_relations,omitempty"`
}

func newDoJobProductionPlanDuplicationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Duplicate a job production plan template",
		Long: `Duplicate a job production plan template into a new plan (or template).

Required flags:
  --job-production-plan-template  Template job production plan ID (required)

Conditionally required:
  --start-on  Start date (YYYY-MM-DD). Required when the template has shifts and for async duplications
             unless --derived-job-production-plan-template-name is set.

Optional flags:
  --derived-job-production-plan           Existing derived plan ID to update
  --derived-job-production-plan-template-name  Template name for the derived plan
  --new-customer                          Customer ID for cross-customer duplication
  --is-async                              Create a duplication work item instead of duplicating synchronously
  --skip-template-duplication-validation  Skip template duplication validation checks
  --disable-overlapping-crew-requirement-is-validating-overlapping  Relax overlapping crew requirement validation

Skip copy flags (all default false; see --help for full list):
  --skip-job-production-plan-alarms, --skip-job-production-plan-locations, --skip-job-schedule-shifts, etc.

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Duplicate a template into a new plan
  xbe do job-production-plan-duplications create \
    --job-production-plan-template 123 \
    --start-on 2026-01-23

  # Duplicate into another customer and skip shifts
  xbe do job-production-plan-duplications create \
    --job-production-plan-template 123 \
    --start-on 2026-01-23 \
    --new-customer 456 \
    --skip-job-schedule-shifts

  # Async duplication
  xbe do job-production-plan-duplications create \
    --job-production-plan-template 123 \
    --start-on 2026-01-23 \
    --is-async`,
		Args: cobra.NoArgs,
		RunE: runDoJobProductionPlanDuplicationsCreate,
	}
	initDoJobProductionPlanDuplicationsCreateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanDuplicationsCmd.AddCommand(newDoJobProductionPlanDuplicationsCreateCmd())
}

func initDoJobProductionPlanDuplicationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan-template", "", "Template job production plan ID (required)")
	cmd.Flags().String("derived-job-production-plan", "", "Derived job production plan ID (optional)")
	cmd.Flags().String("new-customer", "", "New customer ID (optional)")
	cmd.Flags().String("derived-job-production-plan-template-name", "", "Derived plan template name (optional)")
	cmd.Flags().String("start-on", "", "Start date (YYYY-MM-DD) (optional)")
	cmd.Flags().Bool("skip-template-duplication-validation", false, "Skip template duplication validation checks")
	cmd.Flags().Bool("disable-overlapping-crew-requirement-is-validating-overlapping", false, "Relax overlapping crew requirement validation")
	cmd.Flags().Bool("is-async", false, "Create a duplication work item asynchronously")
	cmd.Flags().Bool("skip-job-production-plan-alarms", false, "Skip copying job production plan alarms")
	cmd.Flags().Bool("skip-job-production-plan-locations", false, "Skip copying job production plan locations")
	cmd.Flags().Bool("skip-job-production-plan-safety-risks", false, "Skip copying job production plan safety risks")
	cmd.Flags().Bool("skip-job-production-plan-material-sites", false, "Skip copying job production plan material sites")
	cmd.Flags().Bool("skip-equipment-requirements", false, "Skip copying equipment requirements")
	cmd.Flags().Bool("skip-labor-requirements", false, "Skip copying labor requirements")
	cmd.Flags().Bool("skip-job-production-plan-material-types", false, "Skip copying job production plan material types")
	cmd.Flags().Bool("skip-job-production-plan-service-type-unit-of-measures", false, "Skip copying job production plan service type unit of measures")
	cmd.Flags().Bool("skip-job-production-plan-display-unit-of-measures", false, "Skip copying job production plan display unit of measures")
	cmd.Flags().Bool("skip-job-production-plan-service-type-unit-of-measure-cohorts", false, "Skip copying job production plan service type unit of measure cohorts")
	cmd.Flags().Bool("skip-job-production-plan-trailer-classifications", false, "Skip copying job production plan trailer classifications")
	cmd.Flags().Bool("skip-job-production-plan-segment-sets", false, "Skip copying job production plan segment sets")
	cmd.Flags().Bool("skip-job-production-plan-segments", false, "Skip copying job production plan segments")
	cmd.Flags().Bool("skip-job-production-plan-subscriptions", false, "Skip copying job production plan subscriptions")
	cmd.Flags().Bool("skip-job-production-plan-time-card-approvers", false, "Skip copying job production plan time card approvers")
	cmd.Flags().Bool("skip-job-schedule-shifts", false, "Skip copying job schedule shifts")
	cmd.Flags().Bool("skip-developer-references", false, "Skip copying developer references")
	cmd.Flags().Bool("skip-job-production-plan-inspectors", false, "Skip copying job production plan inspectors")
	cmd.Flags().Bool("skip-job-production-plan-project-phase-revenue-items", false, "Skip copying job production plan project phase revenue items")
	cmd.Flags().Bool("skip-equipment-requirements-resource", false, "Skip copying equipment requirement resources")
	cmd.Flags().Bool("skip-labor-requirements-resource", false, "Skip copying labor requirement resources")
	cmd.Flags().Bool("skip-labor-requirements-craft-class", false, "Skip copying labor requirement craft class")
	cmd.Flags().Bool("skip-job-schedule-shifts-driver-assignment-rule-text-cached", false, "Skip copying job schedule shift driver assignment rules")
	cmd.Flags().Bool("skip-labor-requirements-overlapping-resource", false, "Skip labor requirements with overlapping resources")
	cmd.Flags().String("skipped-labor-requirements-overlapping-resource-ids", "", "JSON array of [labor_requirement_id, resource_id] pairs")
	cmd.Flags().String("skipped-labor-requirements-not-valid-to-assign-ids", "", "JSON array of [labor_requirement_id, resource_id] pairs")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlanDuplicationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoJobProductionPlanDuplicationsCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
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

	if strings.TrimSpace(opts.JobProductionPlanTemplate) == "" {
		err := fmt.Errorf("--job-production-plan-template is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}

	if opts.DerivedJobProductionPlanTemplateName != "" {
		attributes["derived-job-production-plan-template-name"] = opts.DerivedJobProductionPlanTemplateName
	}
	if opts.StartOn != "" {
		attributes["start-on"] = opts.StartOn
	}

	setBoolAttributeIfChanged(cmd, attributes, "skip-template-duplication-validation", "skip-template-duplication-validation", opts.SkipTemplateDuplicationValidation)
	setBoolAttributeIfChanged(cmd, attributes, "disable-overlapping-crew-requirement-is-validating-overlapping", "disable-overlapping-crew-requirement-is-validating-overlapping", opts.DisableOverlappingCrewRequirementIsValidatingOverlapping)
	setBoolAttributeIfChanged(cmd, attributes, "is-async", "is-async", opts.IsAsync)

	setBoolAttributeIfChanged(cmd, attributes, "skip-job-production-plan-alarms", "skip-job-production-plan-alarms", opts.SkipJobProductionPlanAlarms)
	setBoolAttributeIfChanged(cmd, attributes, "skip-job-production-plan-locations", "skip-job-production-plan-locations", opts.SkipJobProductionPlanLocations)
	setBoolAttributeIfChanged(cmd, attributes, "skip-job-production-plan-safety-risks", "skip-job-production-plan-safety-risks", opts.SkipJobProductionPlanSafetyRisks)
	setBoolAttributeIfChanged(cmd, attributes, "skip-job-production-plan-material-sites", "skip-job-production-plan-material-sites", opts.SkipJobProductionPlanMaterialSites)
	setBoolAttributeIfChanged(cmd, attributes, "skip-equipment-requirements", "skip-equipment-requirements", opts.SkipEquipmentRequirements)
	setBoolAttributeIfChanged(cmd, attributes, "skip-labor-requirements", "skip-labor-requirements", opts.SkipLaborRequirements)
	setBoolAttributeIfChanged(cmd, attributes, "skip-job-production-plan-material-types", "skip-job-production-plan-material-types", opts.SkipJobProductionPlanMaterialTypes)
	setBoolAttributeIfChanged(cmd, attributes, "skip-job-production-plan-service-type-unit-of-measures", "skip-job-production-plan-service-type-unit-of-measures", opts.SkipJobProductionPlanServiceTypeUnitOfMeasures)
	setBoolAttributeIfChanged(cmd, attributes, "skip-job-production-plan-display-unit-of-measures", "skip-job-production-plan-display-unit-of-measures", opts.SkipJobProductionPlanDisplayUnitOfMeasures)
	setBoolAttributeIfChanged(cmd, attributes, "skip-job-production-plan-service-type-unit-of-measure-cohorts", "skip-job-production-plan-service-type-unit-of-measure-cohorts", opts.SkipJobProductionPlanServiceTypeUnitOfMeasureCohorts)
	setBoolAttributeIfChanged(cmd, attributes, "skip-job-production-plan-trailer-classifications", "skip-job-production-plan-trailer-classifications", opts.SkipJobProductionPlanTrailerClassifications)
	setBoolAttributeIfChanged(cmd, attributes, "skip-job-production-plan-segment-sets", "skip-job-production-plan-segment-sets", opts.SkipJobProductionPlanSegmentSets)
	setBoolAttributeIfChanged(cmd, attributes, "skip-job-production-plan-segments", "skip-job-production-plan-segments", opts.SkipJobProductionPlanSegments)
	setBoolAttributeIfChanged(cmd, attributes, "skip-job-production-plan-subscriptions", "skip-job-production-plan-subscriptions", opts.SkipJobProductionPlanSubscriptions)
	setBoolAttributeIfChanged(cmd, attributes, "skip-job-production-plan-time-card-approvers", "skip-job-production-plan-time-card-approvers", opts.SkipJobProductionPlanTimeCardApprovers)
	setBoolAttributeIfChanged(cmd, attributes, "skip-job-schedule-shifts", "skip-job-schedule-shifts", opts.SkipJobScheduleShifts)
	setBoolAttributeIfChanged(cmd, attributes, "skip-developer-references", "skip-developer-references", opts.SkipDeveloperReferences)
	setBoolAttributeIfChanged(cmd, attributes, "skip-job-production-plan-inspectors", "skip-job-production-plan-inspectors", opts.SkipJobProductionPlanInspectors)
	setBoolAttributeIfChanged(cmd, attributes, "skip-job-production-plan-project-phase-revenue-items", "skip-job-production-plan-project-phase-revenue-items", opts.SkipJobProductionPlanProjectPhaseRevenueItems)
	setBoolAttributeIfChanged(cmd, attributes, "skip-equipment-requirements-resource", "skip-equipment-requirements-resource", opts.SkipEquipmentRequirementsResource)
	setBoolAttributeIfChanged(cmd, attributes, "skip-labor-requirements-resource", "skip-labor-requirements-resource", opts.SkipLaborRequirementsResource)
	setBoolAttributeIfChanged(cmd, attributes, "skip-labor-requirements-craft-class", "skip-labor-requirements-craft-class", opts.SkipLaborRequirementsCraftClass)
	setBoolAttributeIfChanged(cmd, attributes, "skip-job-schedule-shifts-driver-assignment-rule-text-cached", "skip-job-schedule-shifts-driver-assignment-rule-text-cached", opts.SkipJobScheduleShiftsDriverAssignmentRuleTextCached)
	setBoolAttributeIfChanged(cmd, attributes, "skip-labor-requirements-overlapping-resource", "skip-labor-requirements-overlapping-resource", opts.SkipLaborRequirementsOverlappingResource)

	if opts.SkippedLaborRequirementsOverlappingResourceIDs != "" {
		var parsed any
		if err := json.Unmarshal([]byte(opts.SkippedLaborRequirementsOverlappingResourceIDs), &parsed); err != nil {
			return fmt.Errorf("invalid skipped-labor-requirements-overlapping-resource-ids JSON: %w", err)
		}
		if _, ok := parsed.([]any); !ok {
			return fmt.Errorf("--skipped-labor-requirements-overlapping-resource-ids must be a JSON array")
		}
		attributes["skipped-labor-requirements-overlapping-resource-ids"] = parsed
	}

	if opts.SkippedLaborRequirementsNotValidToAssignIDs != "" {
		var parsed any
		if err := json.Unmarshal([]byte(opts.SkippedLaborRequirementsNotValidToAssignIDs), &parsed); err != nil {
			return fmt.Errorf("invalid skipped-labor-requirements-not-valid-to-assign-ids JSON: %w", err)
		}
		if _, ok := parsed.([]any); !ok {
			return fmt.Errorf("--skipped-labor-requirements-not-valid-to-assign-ids must be a JSON array")
		}
		attributes["skipped-labor-requirements-not-valid-to-assign-ids"] = parsed
	}

	relationships := map[string]any{
		"job-production-plan-template": map[string]any{
			"data": map[string]any{
				"type": "job-production-plans",
				"id":   opts.JobProductionPlanTemplate,
			},
		},
	}

	if opts.DerivedJobProductionPlan != "" {
		relationships["derived-job-production-plan"] = map[string]any{
			"data": map[string]any{
				"type": "job-production-plans",
				"id":   opts.DerivedJobProductionPlan,
			},
		}
	}

	if opts.NewCustomer != "" {
		relationships["new-customer"] = map[string]any{
			"data": map[string]any{
				"type": "customers",
				"id":   opts.NewCustomer,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "job-production-plan-duplications",
			"relationships": relationships,
		},
	}

	if len(attributes) > 0 {
		requestBody["data"].(map[string]any)["attributes"] = attributes
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/job-production-plan-duplications", jsonBody)
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

	row := buildJobProductionPlanDuplicationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	if row.DerivedJobProductionPlanID != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Created job production plan duplication %s (derived plan %s)\n", row.ID, row.DerivedJobProductionPlanID)
		return nil
	}
	if row.WorkID != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Created job production plan duplication %s (work %s)\n", row.ID, row.WorkID)
		return nil
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Created job production plan duplication %s\n", row.ID)
	return nil
}

func parseDoJobProductionPlanDuplicationsCreateOptions(cmd *cobra.Command) (doJobProductionPlanDuplicationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobProductionPlanTemplate, _ := cmd.Flags().GetString("job-production-plan-template")
	derivedJobProductionPlan, _ := cmd.Flags().GetString("derived-job-production-plan")
	newCustomer, _ := cmd.Flags().GetString("new-customer")
	derivedJobProductionPlanTemplateName, _ := cmd.Flags().GetString("derived-job-production-plan-template-name")
	startOn, _ := cmd.Flags().GetString("start-on")
	skipTemplateDuplicationValidation, _ := cmd.Flags().GetBool("skip-template-duplication-validation")
	disableOverlappingCrewRequirementIsValidatingOverlapping, _ := cmd.Flags().GetBool("disable-overlapping-crew-requirement-is-validating-overlapping")
	isAsync, _ := cmd.Flags().GetBool("is-async")
	skipJobProductionPlanAlarms, _ := cmd.Flags().GetBool("skip-job-production-plan-alarms")
	skipJobProductionPlanLocations, _ := cmd.Flags().GetBool("skip-job-production-plan-locations")
	skipJobProductionPlanSafetyRisks, _ := cmd.Flags().GetBool("skip-job-production-plan-safety-risks")
	skipJobProductionPlanMaterialSites, _ := cmd.Flags().GetBool("skip-job-production-plan-material-sites")
	skipEquipmentRequirements, _ := cmd.Flags().GetBool("skip-equipment-requirements")
	skipLaborRequirements, _ := cmd.Flags().GetBool("skip-labor-requirements")
	skipJobProductionPlanMaterialTypes, _ := cmd.Flags().GetBool("skip-job-production-plan-material-types")
	skipJobProductionPlanServiceTypeUnitOfMeasures, _ := cmd.Flags().GetBool("skip-job-production-plan-service-type-unit-of-measures")
	skipJobProductionPlanDisplayUnitOfMeasures, _ := cmd.Flags().GetBool("skip-job-production-plan-display-unit-of-measures")
	skipJobProductionPlanServiceTypeUnitOfMeasureCohorts, _ := cmd.Flags().GetBool("skip-job-production-plan-service-type-unit-of-measure-cohorts")
	skipJobProductionPlanTrailerClassifications, _ := cmd.Flags().GetBool("skip-job-production-plan-trailer-classifications")
	skipJobProductionPlanSegmentSets, _ := cmd.Flags().GetBool("skip-job-production-plan-segment-sets")
	skipJobProductionPlanSegments, _ := cmd.Flags().GetBool("skip-job-production-plan-segments")
	skipJobProductionPlanSubscriptions, _ := cmd.Flags().GetBool("skip-job-production-plan-subscriptions")
	skipJobProductionPlanTimeCardApprovers, _ := cmd.Flags().GetBool("skip-job-production-plan-time-card-approvers")
	skipJobScheduleShifts, _ := cmd.Flags().GetBool("skip-job-schedule-shifts")
	skipDeveloperReferences, _ := cmd.Flags().GetBool("skip-developer-references")
	skipJobProductionPlanInspectors, _ := cmd.Flags().GetBool("skip-job-production-plan-inspectors")
	skipJobProductionPlanProjectPhaseRevenueItems, _ := cmd.Flags().GetBool("skip-job-production-plan-project-phase-revenue-items")
	skipEquipmentRequirementsResource, _ := cmd.Flags().GetBool("skip-equipment-requirements-resource")
	skipLaborRequirementsResource, _ := cmd.Flags().GetBool("skip-labor-requirements-resource")
	skipLaborRequirementsCraftClass, _ := cmd.Flags().GetBool("skip-labor-requirements-craft-class")
	skipJobScheduleShiftsDriverAssignmentRuleTextCached, _ := cmd.Flags().GetBool("skip-job-schedule-shifts-driver-assignment-rule-text-cached")
	skipLaborRequirementsOverlappingResource, _ := cmd.Flags().GetBool("skip-labor-requirements-overlapping-resource")
	skippedLaborRequirementsOverlappingResourceIDs, _ := cmd.Flags().GetString("skipped-labor-requirements-overlapping-resource-ids")
	skippedLaborRequirementsNotValidToAssignIDs, _ := cmd.Flags().GetString("skipped-labor-requirements-not-valid-to-assign-ids")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanDuplicationsCreateOptions{
		BaseURL:                              baseURL,
		Token:                                token,
		JSON:                                 jsonOut,
		JobProductionPlanTemplate:            jobProductionPlanTemplate,
		DerivedJobProductionPlan:             derivedJobProductionPlan,
		NewCustomer:                          newCustomer,
		DerivedJobProductionPlanTemplateName: derivedJobProductionPlanTemplateName,
		StartOn:                              startOn,
		SkipTemplateDuplicationValidation:    skipTemplateDuplicationValidation,
		DisableOverlappingCrewRequirementIsValidatingOverlapping: disableOverlappingCrewRequirementIsValidatingOverlapping,
		IsAsync:                                              isAsync,
		SkipJobProductionPlanAlarms:                          skipJobProductionPlanAlarms,
		SkipJobProductionPlanLocations:                       skipJobProductionPlanLocations,
		SkipJobProductionPlanSafetyRisks:                     skipJobProductionPlanSafetyRisks,
		SkipJobProductionPlanMaterialSites:                   skipJobProductionPlanMaterialSites,
		SkipEquipmentRequirements:                            skipEquipmentRequirements,
		SkipLaborRequirements:                                skipLaborRequirements,
		SkipJobProductionPlanMaterialTypes:                   skipJobProductionPlanMaterialTypes,
		SkipJobProductionPlanServiceTypeUnitOfMeasures:       skipJobProductionPlanServiceTypeUnitOfMeasures,
		SkipJobProductionPlanDisplayUnitOfMeasures:           skipJobProductionPlanDisplayUnitOfMeasures,
		SkipJobProductionPlanServiceTypeUnitOfMeasureCohorts: skipJobProductionPlanServiceTypeUnitOfMeasureCohorts,
		SkipJobProductionPlanTrailerClassifications:          skipJobProductionPlanTrailerClassifications,
		SkipJobProductionPlanSegmentSets:                     skipJobProductionPlanSegmentSets,
		SkipJobProductionPlanSegments:                        skipJobProductionPlanSegments,
		SkipJobProductionPlanSubscriptions:                   skipJobProductionPlanSubscriptions,
		SkipJobProductionPlanTimeCardApprovers:               skipJobProductionPlanTimeCardApprovers,
		SkipJobScheduleShifts:                                skipJobScheduleShifts,
		SkipDeveloperReferences:                              skipDeveloperReferences,
		SkipJobProductionPlanInspectors:                      skipJobProductionPlanInspectors,
		SkipJobProductionPlanProjectPhaseRevenueItems:        skipJobProductionPlanProjectPhaseRevenueItems,
		SkipEquipmentRequirementsResource:                    skipEquipmentRequirementsResource,
		SkipLaborRequirementsResource:                        skipLaborRequirementsResource,
		SkipLaborRequirementsCraftClass:                      skipLaborRequirementsCraftClass,
		SkipJobScheduleShiftsDriverAssignmentRuleTextCached:  skipJobScheduleShiftsDriverAssignmentRuleTextCached,
		SkipLaborRequirementsOverlappingResource:             skipLaborRequirementsOverlappingResource,
		SkippedLaborRequirementsOverlappingResourceIDs:       skippedLaborRequirementsOverlappingResourceIDs,
		SkippedLaborRequirementsNotValidToAssignIDs:          skippedLaborRequirementsNotValidToAssignIDs,
	}, nil
}

func buildJobProductionPlanDuplicationRowFromSingle(resp jsonAPISingleResponse) jobProductionPlanDuplicationRow {
	attrs := resp.Data.Attributes
	row := jobProductionPlanDuplicationRow{
		ID:                                   resp.Data.ID,
		DuplicationToken:                     stringAttr(attrs, "duplication-token"),
		DerivedJobProductionPlanTemplateName: stringAttr(attrs, "derived-job-production-plan-template-name"),
		StartOn:                              stringAttr(attrs, "start-on"),
		IsAsync:                              boolAttr(attrs, "is-async"),
		SkipTemplateDuplicationValidation:    boolAttr(attrs, "skip-template-duplication-validation"),
		DisableOverlappingCrewRequirementIsValidatingOverlapping: boolAttr(attrs, "disable-overlapping-crew-requirement-is-validating-overlapping"),
		UnsavedRelations:         attrs["unsaved-relations"],
		InvalidTemplateRelations: attrs["invalid-template-relations"],
	}

	if rel, ok := resp.Data.Relationships["job-production-plan-template"]; ok && rel.Data != nil {
		row.JobProductionPlanTemplateID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["derived-job-production-plan"]; ok && rel.Data != nil {
		row.DerivedJobProductionPlanID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["new-customer"]; ok && rel.Data != nil {
		row.NewCustomerID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["work"]; ok && rel.Data != nil {
		row.WorkID = rel.Data.ID
	}

	return row
}

func setBoolAttributeIfChanged(cmd *cobra.Command, attributes map[string]any, flagName, attrName string, value bool) {
	if cmd.Flags().Changed(flagName) {
		attributes[attrName] = value
	}
}
