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

type jobProductionPlanDuplicationWorksShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type jobProductionPlanDuplicationWorkDetails struct {
	ID                                                       string `json:"id"`
	JID                                                      string `json:"jid,omitempty"`
	ScheduledAt                                              string `json:"scheduled_at,omitempty"`
	ProcessedAt                                              string `json:"processed_at,omitempty"`
	StartOn                                                  string `json:"start_on,omitempty"`
	DerivedJobProductionPlanTemplateName                     string `json:"derived_job_production_plan_template_name,omitempty"`
	SkipTemplateDuplicationValidation                        bool   `json:"skip_template_duplication_validation"`
	DisableOverlappingCrewRequirementIsValidatingOverlapping bool   `json:"disable_overlapping_crew_requirement_is_validating_overlapping"`
	WorkResults                                              any    `json:"work_results,omitempty"`
	WorkErrors                                               any    `json:"work_errors,omitempty"`
	WorkWarnings                                             any    `json:"work_warnings,omitempty"`

	JobProductionPlanTemplateID        string `json:"job_production_plan_template_id,omitempty"`
	JobProductionPlanTemplateJobNumber string `json:"job_production_plan_template_job_number,omitempty"`
	JobProductionPlanTemplateJobName   string `json:"job_production_plan_template_job_name,omitempty"`
	DerivedJobProductionPlanID         string `json:"derived_job_production_plan_id,omitempty"`
	DerivedJobProductionPlanJobNumber  string `json:"derived_job_production_plan_job_number,omitempty"`
	DerivedJobProductionPlanJobName    string `json:"derived_job_production_plan_job_name,omitempty"`
	NewCustomerID                      string `json:"new_customer_id,omitempty"`
	NewCustomerName                    string `json:"new_customer_name,omitempty"`
	CreatedByID                        string `json:"created_by_id,omitempty"`
	CreatedByName                      string `json:"created_by_name,omitempty"`

	SkipJobProductionPlanAlarms                          bool `json:"skip_job_production_plan_alarms"`
	SkipJobProductionPlanLocations                       bool `json:"skip_job_production_plan_locations"`
	SkipJobProductionPlanSafetyRisks                     bool `json:"skip_job_production_plan_safety_risks"`
	SkipJobProductionPlanMaterialSites                   bool `json:"skip_job_production_plan_material_sites"`
	SkipEquipmentRequirements                            bool `json:"skip_equipment_requirements"`
	SkipLaborRequirements                                bool `json:"skip_labor_requirements"`
	SkipJobProductionPlanMaterialTypes                   bool `json:"skip_job_production_plan_material_types"`
	SkipJobProductionPlanServiceTypeUnitOfMeasures       bool `json:"skip_job_production_plan_service_type_unit_of_measures"`
	SkipJobProductionPlanDisplayUnitOfMeasures           bool `json:"skip_job_production_plan_display_unit_of_measures"`
	SkipJobProductionPlanServiceTypeUnitOfMeasureCohorts bool `json:"skip_job_production_plan_service_type_unit_of_measure_cohorts"`
	SkipJobProductionPlanTrailerClassifications          bool `json:"skip_job_production_plan_trailer_classifications"`
	SkipJobProductionPlanSegmentSets                     bool `json:"skip_job_production_plan_segment_sets"`
	SkipJobProductionPlanSegments                        bool `json:"skip_job_production_plan_segments"`
	SkipJobProductionPlanSubscriptions                   bool `json:"skip_job_production_plan_subscriptions"`
	SkipJobProductionPlanTimeCardApprovers               bool `json:"skip_job_production_plan_time_card_approvers"`
	SkipJobScheduleShifts                                bool `json:"skip_job_schedule_shifts"`
	SkipDeveloperReferences                              bool `json:"skip_developer_references"`
	SkipJobProductionPlanInspectors                      bool `json:"skip_job_production_plan_inspectors"`
	SkipJobProductionPlanProjectPhaseRevenueItems        bool `json:"skip_job_production_plan_project_phase_revenue_items"`
	SkipEquipmentRequirementsResource                    bool `json:"skip_equipment_requirements_resource"`
	SkipLaborRequirementsResource                        bool `json:"skip_labor_requirements_resource"`
	SkipLaborRequirementsCraftClass                      bool `json:"skip_labor_requirements_craft_class"`
	SkipJobScheduleShiftsDriverAssignmentRuleTextCached  bool `json:"skip_job_schedule_shifts_driver_assignment_rule_text_cached"`

	SkipLaborRequirementsOverlappingResource       bool `json:"skip_labor_requirements_overlapping_resource"`
	SkippedLaborRequirementsOverlappingResourceIDs any  `json:"skipped_labor_requirements_overlapping_resource_ids,omitempty"`
	SkippedLaborRequirementsNotValidToAssignIDs    any  `json:"skipped_labor_requirements_not_valid_to_assign_ids,omitempty"`
}

func newJobProductionPlanDuplicationWorksShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show job production plan duplication work details",
		Long: `Show the full details of a specific job production plan duplication work item.

Includes async processing metadata, parameters, skip flags, and related records.

Arguments:
  <id>  The duplication work ID (required)`,
		Example: `  # Show duplication work details
  xbe view job-production-plan-duplication-works show 123

  # Output as JSON
  xbe view job-production-plan-duplication-works show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runJobProductionPlanDuplicationWorksShow,
	}
	initJobProductionPlanDuplicationWorksShowFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanDuplicationWorksCmd.AddCommand(newJobProductionPlanDuplicationWorksShowCmd())
}

func initJobProductionPlanDuplicationWorksShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanDuplicationWorksShow(cmd *cobra.Command, args []string) error {
	opts, err := parseJobProductionPlanDuplicationWorksShowOptions(cmd)
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("job production plan duplication work id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("include", "job-production-plan-template,derived-job-production-plan,new-customer,created-by")
	query.Set("fields[job-production-plans]", "job-number,job-name,start-on")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[users]", "name")

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-duplication-works/"+id, query)
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

	details := buildJobProductionPlanDuplicationWorkDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderJobProductionPlanDuplicationWorkDetails(cmd, details)
}

func parseJobProductionPlanDuplicationWorksShowOptions(cmd *cobra.Command) (jobProductionPlanDuplicationWorksShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanDuplicationWorksShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildJobProductionPlanDuplicationWorkDetails(resp jsonAPISingleResponse) jobProductionPlanDuplicationWorkDetails {
	attrs := resp.Data.Attributes

	included := make(map[string]map[string]any)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc.Attributes
	}

	details := jobProductionPlanDuplicationWorkDetails{
		ID:                                   resp.Data.ID,
		JID:                                  stringAttr(attrs, "jid"),
		ScheduledAt:                          formatDateTime(stringAttr(attrs, "scheduled-at")),
		ProcessedAt:                          formatDateTime(stringAttr(attrs, "processed-at")),
		StartOn:                              formatDate(stringAttr(attrs, "start-on")),
		DerivedJobProductionPlanTemplateName: stringAttr(attrs, "derived-job-production-plan-template-name"),
		SkipTemplateDuplicationValidation:    boolAttr(attrs, "skip-template-duplication-validation"),
		DisableOverlappingCrewRequirementIsValidatingOverlapping: boolAttr(attrs, "disable-overlapping-crew-requirement-is-validating-overlapping"),
		WorkResults:                                          anyAttr(attrs, "work-results"),
		WorkErrors:                                           anyAttr(attrs, "work-errors"),
		WorkWarnings:                                         anyAttr(attrs, "work-warnings"),
		SkipJobProductionPlanAlarms:                          boolAttr(attrs, "skip-job-production-plan-alarms"),
		SkipJobProductionPlanLocations:                       boolAttr(attrs, "skip-job-production-plan-locations"),
		SkipJobProductionPlanSafetyRisks:                     boolAttr(attrs, "skip-job-production-plan-safety-risks"),
		SkipJobProductionPlanMaterialSites:                   boolAttr(attrs, "skip-job-production-plan-material-sites"),
		SkipEquipmentRequirements:                            boolAttr(attrs, "skip-equipment-requirements"),
		SkipLaborRequirements:                                boolAttr(attrs, "skip-labor-requirements"),
		SkipJobProductionPlanMaterialTypes:                   boolAttr(attrs, "skip-job-production-plan-material-types"),
		SkipJobProductionPlanServiceTypeUnitOfMeasures:       boolAttr(attrs, "skip-job-production-plan-service-type-unit-of-measures"),
		SkipJobProductionPlanDisplayUnitOfMeasures:           boolAttr(attrs, "skip-job-production-plan-display-unit-of-measures"),
		SkipJobProductionPlanServiceTypeUnitOfMeasureCohorts: boolAttr(attrs, "skip-job-production-plan-service-type-unit-of-measure-cohorts"),
		SkipJobProductionPlanTrailerClassifications:          boolAttr(attrs, "skip-job-production-plan-trailer-classifications"),
		SkipJobProductionPlanSegmentSets:                     boolAttr(attrs, "skip-job-production-plan-segment-sets"),
		SkipJobProductionPlanSegments:                        boolAttr(attrs, "skip-job-production-plan-segments"),
		SkipJobProductionPlanSubscriptions:                   boolAttr(attrs, "skip-job-production-plan-subscriptions"),
		SkipJobProductionPlanTimeCardApprovers:               boolAttr(attrs, "skip-job-production-plan-time-card-approvers"),
		SkipJobScheduleShifts:                                boolAttr(attrs, "skip-job-schedule-shifts"),
		SkipDeveloperReferences:                              boolAttr(attrs, "skip-developer-references"),
		SkipJobProductionPlanInspectors:                      boolAttr(attrs, "skip-job-production-plan-inspectors"),
		SkipJobProductionPlanProjectPhaseRevenueItems:        boolAttr(attrs, "skip-job-production-plan-project-phase-revenue-items"),
		SkipEquipmentRequirementsResource:                    boolAttr(attrs, "skip-equipment-requirements-resource"),
		SkipLaborRequirementsResource:                        boolAttr(attrs, "skip-labor-requirements-resource"),
		SkipLaborRequirementsCraftClass:                      boolAttr(attrs, "skip-labor-requirements-craft-class"),
		SkipJobScheduleShiftsDriverAssignmentRuleTextCached:  boolAttr(attrs, "skip-job-schedule-shifts-driver-assignment-rule-text-cached"),
		SkipLaborRequirementsOverlappingResource:             boolAttr(attrs, "skip-labor-requirements-overlapping-resource"),
		SkippedLaborRequirementsOverlappingResourceIDs:       anyAttr(attrs, "skipped-labor-requirements-overlapping-resource-ids"),
		SkippedLaborRequirementsNotValidToAssignIDs:          anyAttr(attrs, "skipped-labor-requirements-not-valid-to-assign-ids"),
	}

	if rel, ok := resp.Data.Relationships["job-production-plan-template"]; ok && rel.Data != nil {
		details.JobProductionPlanTemplateID = rel.Data.ID
		if attrs, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.JobProductionPlanTemplateJobNumber = stringAttr(attrs, "job-number")
			details.JobProductionPlanTemplateJobName = stringAttr(attrs, "job-name")
		}
	}
	if rel, ok := resp.Data.Relationships["derived-job-production-plan"]; ok && rel.Data != nil {
		details.DerivedJobProductionPlanID = rel.Data.ID
		if attrs, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.DerivedJobProductionPlanJobNumber = stringAttr(attrs, "job-number")
			details.DerivedJobProductionPlanJobName = stringAttr(attrs, "job-name")
		}
	}
	if rel, ok := resp.Data.Relationships["new-customer"]; ok && rel.Data != nil {
		details.NewCustomerID = rel.Data.ID
		if attrs, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.NewCustomerName = stringAttr(attrs, "company-name")
		}
	}
	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
		if attrs, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.CreatedByName = stringAttr(attrs, "name")
		}
	}

	return details
}

func renderJobProductionPlanDuplicationWorkDetails(cmd *cobra.Command, details jobProductionPlanDuplicationWorkDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.JID != "" {
		fmt.Fprintf(out, "JID: %s\n", details.JID)
	}
	if details.ScheduledAt != "" {
		fmt.Fprintf(out, "Scheduled At: %s\n", details.ScheduledAt)
	}
	if details.ProcessedAt != "" {
		fmt.Fprintf(out, "Processed At: %s\n", details.ProcessedAt)
	}
	if details.StartOn != "" {
		fmt.Fprintf(out, "Start On: %s\n", details.StartOn)
	}
	if details.DerivedJobProductionPlanTemplateName != "" {
		fmt.Fprintf(out, "Derived Template Name: %s\n", details.DerivedJobProductionPlanTemplateName)
	}
	fmt.Fprintf(out, "Skip Template Duplication Validation: %t\n", details.SkipTemplateDuplicationValidation)
	fmt.Fprintf(out, "Disable Overlapping Crew Requirement Validation: %t\n", details.DisableOverlappingCrewRequirementIsValidatingOverlapping)

	if details.JobProductionPlanTemplateID != "" {
		fmt.Fprintf(out, "Template Plan ID: %s\n", details.JobProductionPlanTemplateID)
	}
	if details.JobProductionPlanTemplateJobNumber != "" {
		fmt.Fprintf(out, "Template Job Number: %s\n", details.JobProductionPlanTemplateJobNumber)
	}
	if details.JobProductionPlanTemplateJobName != "" {
		fmt.Fprintf(out, "Template Job Name: %s\n", details.JobProductionPlanTemplateJobName)
	}
	if details.DerivedJobProductionPlanID != "" {
		fmt.Fprintf(out, "Derived Plan ID: %s\n", details.DerivedJobProductionPlanID)
	}
	if details.DerivedJobProductionPlanJobNumber != "" {
		fmt.Fprintf(out, "Derived Job Number: %s\n", details.DerivedJobProductionPlanJobNumber)
	}
	if details.DerivedJobProductionPlanJobName != "" {
		fmt.Fprintf(out, "Derived Job Name: %s\n", details.DerivedJobProductionPlanJobName)
	}
	if details.NewCustomerID != "" {
		fmt.Fprintf(out, "New Customer ID: %s\n", details.NewCustomerID)
	}
	if details.NewCustomerName != "" {
		fmt.Fprintf(out, "New Customer Name: %s\n", details.NewCustomerName)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By ID: %s\n", details.CreatedByID)
	}
	if details.CreatedByName != "" {
		fmt.Fprintf(out, "Created By Name: %s\n", details.CreatedByName)
	}

	if details.WorkResults != nil {
		fmt.Fprintf(out, "Work Results: %s\n", formatAny(details.WorkResults))
	}
	if details.WorkErrors != nil {
		fmt.Fprintf(out, "Work Errors: %s\n", formatAny(details.WorkErrors))
	}
	if details.WorkWarnings != nil {
		fmt.Fprintf(out, "Work Warnings: %s\n", formatAny(details.WorkWarnings))
	}

	fmt.Fprintln(out, "Skip Flags:")
	for _, flag := range []struct {
		Label string
		Value bool
	}{
		{"Skip Job Production Plan Alarms", details.SkipJobProductionPlanAlarms},
		{"Skip Job Production Plan Locations", details.SkipJobProductionPlanLocations},
		{"Skip Job Production Plan Safety Risks", details.SkipJobProductionPlanSafetyRisks},
		{"Skip Job Production Plan Material Sites", details.SkipJobProductionPlanMaterialSites},
		{"Skip Equipment Requirements", details.SkipEquipmentRequirements},
		{"Skip Labor Requirements", details.SkipLaborRequirements},
		{"Skip Job Production Plan Material Types", details.SkipJobProductionPlanMaterialTypes},
		{"Skip Job Production Plan Service Type Unit Of Measures", details.SkipJobProductionPlanServiceTypeUnitOfMeasures},
		{"Skip Job Production Plan Display Unit Of Measures", details.SkipJobProductionPlanDisplayUnitOfMeasures},
		{"Skip Job Production Plan Service Type Unit Of Measure Cohorts", details.SkipJobProductionPlanServiceTypeUnitOfMeasureCohorts},
		{"Skip Job Production Plan Trailer Classifications", details.SkipJobProductionPlanTrailerClassifications},
		{"Skip Job Production Plan Segment Sets", details.SkipJobProductionPlanSegmentSets},
		{"Skip Job Production Plan Segments", details.SkipJobProductionPlanSegments},
		{"Skip Job Production Plan Subscriptions", details.SkipJobProductionPlanSubscriptions},
		{"Skip Job Production Plan Time Card Approvers", details.SkipJobProductionPlanTimeCardApprovers},
		{"Skip Job Schedule Shifts", details.SkipJobScheduleShifts},
		{"Skip Developer References", details.SkipDeveloperReferences},
		{"Skip Job Production Plan Inspectors", details.SkipJobProductionPlanInspectors},
		{"Skip Job Production Plan Project Phase Revenue Items", details.SkipJobProductionPlanProjectPhaseRevenueItems},
		{"Skip Equipment Requirements Resource", details.SkipEquipmentRequirementsResource},
		{"Skip Labor Requirements Resource", details.SkipLaborRequirementsResource},
		{"Skip Labor Requirements Craft Class", details.SkipLaborRequirementsCraftClass},
		{"Skip Job Schedule Shifts Driver Assignment Rule Text Cached", details.SkipJobScheduleShiftsDriverAssignmentRuleTextCached},
	} {
		fmt.Fprintf(out, "  %s: %t\n", flag.Label, flag.Value)
	}

	fmt.Fprintln(out, "Special Attributes:")
	fmt.Fprintf(out, "  Skip Labor Requirements Overlapping Resource: %t\n", details.SkipLaborRequirementsOverlappingResource)
	fmt.Fprintf(out, "  Skipped Labor Requirements Overlapping Resource IDs: %s\n", formatAny(details.SkippedLaborRequirementsOverlappingResourceIDs))
	fmt.Fprintf(out, "  Skipped Labor Requirements Not Valid To Assign IDs: %s\n", formatAny(details.SkippedLaborRequirementsNotValidToAssignIDs))

	return nil
}

func anyAttr(attrs map[string]any, key string) any {
	if attrs == nil {
		return nil
	}
	if value, ok := attrs[key]; ok {
		return value
	}
	return nil
}

func formatAny(value any) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	default:
		pretty, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			return fmt.Sprintf("%v", value)
		}
		return string(pretty)
	}
}
