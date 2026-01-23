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

type jobProductionPlanChangeSetsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type jobProductionPlanChangeSetDetails struct {
	ID           string `json:"id"`
	BrokerID     string `json:"broker_id,omitempty"`
	BrokerName   string `json:"broker_name,omitempty"`
	CustomerID   string `json:"customer_id,omitempty"`
	CustomerName string `json:"customer_name,omitempty"`
	CreatedByID  string `json:"created_by_id,omitempty"`
	CreatedBy    string `json:"created_by_name,omitempty"`

	ScopeJobProductionPlanIDs  []string `json:"scope_job_production_plan_ids,omitempty"`
	ScopeProjectIDs            []string `json:"scope_project_ids,omitempty"`
	ScopePlannerIDs            []string `json:"scope_planner_ids,omitempty"`
	ScopeMaterialSiteIDs       []string `json:"scope_material_site_ids,omitempty"`
	ScopeStatuses              []string `json:"scope_statuses,omitempty"`
	ScopeStartOnMin            string   `json:"scope_start_on_min,omitempty"`
	ScopeStartOnMax            string   `json:"scope_start_on_max,omitempty"`
	ScopeUltimateMaterialTypes []string `json:"scope_ultimate_material_types,omitempty"`
	ScopeMaterialTypeIDs       []string `json:"scope_material_type_ids,omitempty"`
	ScopeForemanIDs            []string `json:"scope_foreman_ids,omitempty"`

	ChangeJobNumberNew    string `json:"change_job_number_new,omitempty"`
	ChangeRawJobNumberNew string `json:"change_raw_job_number_new,omitempty"`

	ChangeOldMaterialTypeID   string `json:"change_old_material_type_id,omitempty"`
	ChangeOldMaterialType     string `json:"change_old_material_type,omitempty"`
	ChangeNewMaterialTypeID   string `json:"change_new_material_type_id,omitempty"`
	ChangeNewMaterialType     string `json:"change_new_material_type,omitempty"`
	ChangeOldMaterialSiteID   string `json:"change_old_material_site_id,omitempty"`
	ChangeOldMaterialSite     string `json:"change_old_material_site,omitempty"`
	ChangeNewMaterialSiteID   string `json:"change_new_material_site_id,omitempty"`
	ChangeNewMaterialSite     string `json:"change_new_material_site,omitempty"`
	ChangeOldCostCodeID       string `json:"change_old_cost_code_id,omitempty"`
	ChangeOldCostCode         string `json:"change_old_cost_code,omitempty"`
	ChangeNewCostCodeID       string `json:"change_new_cost_code_id,omitempty"`
	ChangeNewCostCode         string `json:"change_new_cost_code,omitempty"`
	ChangeOldInspectorID      string `json:"change_old_inspector_id,omitempty"`
	ChangeOldInspector        string `json:"change_old_inspector,omitempty"`
	ChangeNewInspectorID      string `json:"change_new_inspector_id,omitempty"`
	ChangeNewInspector        string `json:"change_new_inspector,omitempty"`
	ChangeNewPlannerID        string `json:"change_new_planner_id,omitempty"`
	ChangeNewPlanner          string `json:"change_new_planner,omitempty"`
	ChangeNewProjectManagerID string `json:"change_new_project_manager_id,omitempty"`
	ChangeNewProjectManager   string `json:"change_new_project_manager,omitempty"`

	ChangeOldJppMaterialTypeQualityControlClassificationID string `json:"change_old_jpp_material_type_quality_control_classification_id,omitempty"`
	ChangeOldJppMaterialTypeQualityControlClassification   string `json:"change_old_jpp_material_type_quality_control_classification,omitempty"`
	ChangeNewJppMaterialTypeQualityControlClassificationID string `json:"change_new_jpp_material_type_quality_control_classification_id,omitempty"`
	ChangeNewJppMaterialTypeQualityControlClassification   string `json:"change_new_jpp_material_type_quality_control_classification,omitempty"`
	ChangeOldJppMaterialTypeExplicitMaterialMixDesignID    string `json:"change_old_jpp_material_type_explicit_material_mix_design_id,omitempty"`
	ChangeOldJppMaterialTypeExplicitMaterialMixDesign      string `json:"change_old_jpp_material_type_explicit_material_mix_design,omitempty"`
	ChangeNewJppMaterialTypeExplicitMaterialMixDesignID    string `json:"change_new_jpp_material_type_explicit_material_mix_design_id,omitempty"`
	ChangeNewJppMaterialTypeExplicitMaterialMixDesign      string `json:"change_new_jpp_material_type_explicit_material_mix_design,omitempty"`
	ChangeOldLaborerID                                     string `json:"change_old_laborer_id,omitempty"`
	ChangeOldLaborer                                       string `json:"change_old_laborer,omitempty"`
	ChangeNewLaborerID                                     string `json:"change_new_laborer_id,omitempty"`
	ChangeNewLaborer                                       string `json:"change_new_laborer,omitempty"`
	ChangeOldJobSiteID                                     string `json:"change_old_job_site_id,omitempty"`
	ChangeOldJobSite                                       string `json:"change_old_job_site,omitempty"`
	ChangeNewJobSiteID                                     string `json:"change_new_job_site_id,omitempty"`
	ChangeNewJobSite                                       string `json:"change_new_job_site,omitempty"`

	ChangeNewPlannerNullify        bool   `json:"change_new_planner_nullify"`
	ChangeNewProjectManagerNullify bool   `json:"change_new_project_manager_nullify"`
	ChangeNewStatus                string `json:"change_new_status,omitempty"`
	ChangeOldIsScheduleLocked      string `json:"change_old_is_schedule_locked,omitempty"`
	ChangeNewIsScheduleLocked      string `json:"change_new_is_schedule_locked,omitempty"`
	ChangeNewDaysOffset            string `json:"change_new_days_offset,omitempty"`
	ChangeNewOffsetSkipSaturdays   string `json:"change_new_offset_skip_saturdays,omitempty"`
	ChangeNewOffsetSkipSundays     string `json:"change_new_offset_skip_sundays,omitempty"`
	ShouldPersist                  bool   `json:"should_persist"`
	SkipInvalidPlans               bool   `json:"skip_invalid_plans"`

	Results     []map[string]any `json:"results,omitempty"`
	MatchIDs    []string         `json:"match_ids,omitempty"`
	ProcessedAt string           `json:"processed_at,omitempty"`
}

func newJobProductionPlanChangeSetsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show job production plan change set details",
		Long: `Show the full details of a specific job production plan change set.

Includes scope filters, change instructions, processing results, and
relationships.

Arguments:
  <id>  Change set ID (required). Find IDs using the list command.`,
		Example: `  # View a change set by ID
  xbe view job-production-plan-change-sets show 123

  # Get JSON output
  xbe view job-production-plan-change-sets show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runJobProductionPlanChangeSetsShow,
	}
	initJobProductionPlanChangeSetsShowFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanChangeSetsCmd.AddCommand(newJobProductionPlanChangeSetsShowCmd())
}

func initJobProductionPlanChangeSetsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanChangeSetsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseJobProductionPlanChangeSetsShowOptions(cmd)
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
		return fmt.Errorf("job production plan change set id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[job-production-plan-change-sets]", strings.Join([]string{
		"broker",
		"customer",
		"created-by",
		"scope-job-production-plan-ids",
		"scope-project-ids",
		"scope-planner-ids",
		"scope-material-site-ids",
		"scope-statuses",
		"scope-start-on-min",
		"scope-start-on-max",
		"scope-ultimate-material-types",
		"scope-material-type-ids",
		"scope-foreman-ids",
		"change-job-number-new",
		"change-raw-job-number-new",
		"change-old-material-type",
		"change-new-material-type",
		"change-old-material-site",
		"change-new-material-site",
		"change-old-cost-code",
		"change-new-cost-code",
		"change-old-inspector",
		"change-new-inspector",
		"change-new-planner",
		"change-new-project-manager",
		"change-old-jpp-material-type-quality-control-classification",
		"change-new-jpp-material-type-quality-control-classification",
		"change-old-jpp-material-type-explicit-material-mix-design",
		"change-new-jpp-material-type-explicit-material-mix-design",
		"change-old-laborer",
		"change-new-laborer",
		"change-old-job-site",
		"change-new-job-site",
		"change-new-planner-nullify",
		"change-new-project-manager-nullify",
		"change-new-status",
		"change-old-is-schedule-locked",
		"change-new-is-schedule-locked",
		"change-new-days-offset",
		"change-new-offset-skip-saturdays",
		"change-new-offset-skip-sundays",
		"should-persist",
		"skip-invalid-plans",
		"results",
		"match-ids",
		"processed-at",
	}, ","))
	query.Set("include", strings.Join([]string{
		"broker",
		"customer",
		"created-by",
		"change-old-material-type",
		"change-new-material-type",
		"change-old-material-site",
		"change-new-material-site",
		"change-old-cost-code",
		"change-new-cost-code",
		"change-old-inspector",
		"change-new-inspector",
		"change-new-planner",
		"change-new-project-manager",
		"change-old-jpp-material-type-quality-control-classification",
		"change-new-jpp-material-type-quality-control-classification",
		"change-old-jpp-material-type-explicit-material-mix-design",
		"change-new-jpp-material-type-explicit-material-mix-design",
		"change-old-laborer",
		"change-new-laborer",
		"change-old-job-site",
		"change-new-job-site",
	}, ","))
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[users]", "name")
	query.Set("fields[material-types]", "name")
	query.Set("fields[material-sites]", "name")
	query.Set("fields[cost-codes]", "code")
	query.Set("fields[quality-control-classifications]", "name")
	query.Set("fields[material-mix-designs]", "description")
	query.Set("fields[laborers]", "nickname")
	query.Set("fields[job-sites]", "name")

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-change-sets/"+id, query)
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

	details := buildJobProductionPlanChangeSetDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderJobProductionPlanChangeSetDetails(cmd, details)
}

func parseJobProductionPlanChangeSetsShowOptions(cmd *cobra.Command) (jobProductionPlanChangeSetsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanChangeSetsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildJobProductionPlanChangeSetDetails(resp jsonAPISingleResponse) jobProductionPlanChangeSetDetails {
	attrs := resp.Data.Attributes

	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := jobProductionPlanChangeSetDetails{
		ID:                             resp.Data.ID,
		ScopeJobProductionPlanIDs:      stringSliceAttr(attrs, "scope-job-production-plan-ids"),
		ScopeProjectIDs:                stringSliceAttr(attrs, "scope-project-ids"),
		ScopePlannerIDs:                stringSliceAttr(attrs, "scope-planner-ids"),
		ScopeMaterialSiteIDs:           stringSliceAttr(attrs, "scope-material-site-ids"),
		ScopeStatuses:                  stringSliceAttr(attrs, "scope-statuses"),
		ScopeStartOnMin:                formatDate(stringAttr(attrs, "scope-start-on-min")),
		ScopeStartOnMax:                formatDate(stringAttr(attrs, "scope-start-on-max")),
		ScopeUltimateMaterialTypes:     stringSliceAttr(attrs, "scope-ultimate-material-types"),
		ScopeMaterialTypeIDs:           stringSliceAttr(attrs, "scope-material-type-ids"),
		ScopeForemanIDs:                stringSliceAttr(attrs, "scope-foreman-ids"),
		ChangeJobNumberNew:             stringAttr(attrs, "change-job-number-new"),
		ChangeRawJobNumberNew:          stringAttr(attrs, "change-raw-job-number-new"),
		ChangeNewPlannerNullify:        boolAttr(attrs, "change-new-planner-nullify"),
		ChangeNewProjectManagerNullify: boolAttr(attrs, "change-new-project-manager-nullify"),
		ChangeNewStatus:                stringAttr(attrs, "change-new-status"),
		ChangeOldIsScheduleLocked:      stringAttr(attrs, "change-old-is-schedule-locked"),
		ChangeNewIsScheduleLocked:      stringAttr(attrs, "change-new-is-schedule-locked"),
		ChangeNewDaysOffset:            stringAttr(attrs, "change-new-days-offset"),
		ChangeNewOffsetSkipSaturdays:   stringAttr(attrs, "change-new-offset-skip-saturdays"),
		ChangeNewOffsetSkipSundays:     stringAttr(attrs, "change-new-offset-skip-sundays"),
		ShouldPersist:                  boolAttr(attrs, "should-persist"),
		SkipInvalidPlans:               boolAttr(attrs, "skip-invalid-plans"),
		Results:                        mapSliceAttr(attrs, "results"),
		MatchIDs:                       stringSliceAttr(attrs, "match-ids"),
		ProcessedAt:                    formatDateTime(stringAttr(attrs, "processed-at")),
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID, details.BrokerName = resolveIncludedName(included, rel, "company-name")
	}
	if rel, ok := resp.Data.Relationships["customer"]; ok && rel.Data != nil {
		details.CustomerID, details.CustomerName = resolveIncludedName(included, rel, "company-name")
	}
	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID, details.CreatedBy = resolveIncludedName(included, rel, "name")
	}

	if rel, ok := resp.Data.Relationships["change-old-material-type"]; ok && rel.Data != nil {
		details.ChangeOldMaterialTypeID, details.ChangeOldMaterialType = resolveIncludedName(included, rel, "name")
	}
	if rel, ok := resp.Data.Relationships["change-new-material-type"]; ok && rel.Data != nil {
		details.ChangeNewMaterialTypeID, details.ChangeNewMaterialType = resolveIncludedName(included, rel, "name")
	}
	if rel, ok := resp.Data.Relationships["change-old-material-site"]; ok && rel.Data != nil {
		details.ChangeOldMaterialSiteID, details.ChangeOldMaterialSite = resolveIncludedName(included, rel, "name")
	}
	if rel, ok := resp.Data.Relationships["change-new-material-site"]; ok && rel.Data != nil {
		details.ChangeNewMaterialSiteID, details.ChangeNewMaterialSite = resolveIncludedName(included, rel, "name")
	}
	if rel, ok := resp.Data.Relationships["change-old-cost-code"]; ok && rel.Data != nil {
		details.ChangeOldCostCodeID, details.ChangeOldCostCode = resolveIncludedName(included, rel, "code")
	}
	if rel, ok := resp.Data.Relationships["change-new-cost-code"]; ok && rel.Data != nil {
		details.ChangeNewCostCodeID, details.ChangeNewCostCode = resolveIncludedName(included, rel, "code")
	}
	if rel, ok := resp.Data.Relationships["change-old-inspector"]; ok && rel.Data != nil {
		details.ChangeOldInspectorID, details.ChangeOldInspector = resolveIncludedName(included, rel, "name")
	}
	if rel, ok := resp.Data.Relationships["change-new-inspector"]; ok && rel.Data != nil {
		details.ChangeNewInspectorID, details.ChangeNewInspector = resolveIncludedName(included, rel, "name")
	}
	if rel, ok := resp.Data.Relationships["change-new-planner"]; ok && rel.Data != nil {
		details.ChangeNewPlannerID, details.ChangeNewPlanner = resolveIncludedName(included, rel, "name")
	}
	if rel, ok := resp.Data.Relationships["change-new-project-manager"]; ok && rel.Data != nil {
		details.ChangeNewProjectManagerID, details.ChangeNewProjectManager = resolveIncludedName(included, rel, "name")
	}
	if rel, ok := resp.Data.Relationships["change-old-jpp-material-type-quality-control-classification"]; ok && rel.Data != nil {
		details.ChangeOldJppMaterialTypeQualityControlClassificationID, details.ChangeOldJppMaterialTypeQualityControlClassification = resolveIncludedName(included, rel, "name")
	}
	if rel, ok := resp.Data.Relationships["change-new-jpp-material-type-quality-control-classification"]; ok && rel.Data != nil {
		details.ChangeNewJppMaterialTypeQualityControlClassificationID, details.ChangeNewJppMaterialTypeQualityControlClassification = resolveIncludedName(included, rel, "name")
	}
	if rel, ok := resp.Data.Relationships["change-old-jpp-material-type-explicit-material-mix-design"]; ok && rel.Data != nil {
		details.ChangeOldJppMaterialTypeExplicitMaterialMixDesignID, details.ChangeOldJppMaterialTypeExplicitMaterialMixDesign = resolveIncludedName(included, rel, "description")
	}
	if rel, ok := resp.Data.Relationships["change-new-jpp-material-type-explicit-material-mix-design"]; ok && rel.Data != nil {
		details.ChangeNewJppMaterialTypeExplicitMaterialMixDesignID, details.ChangeNewJppMaterialTypeExplicitMaterialMixDesign = resolveIncludedName(included, rel, "description")
	}
	if rel, ok := resp.Data.Relationships["change-old-laborer"]; ok && rel.Data != nil {
		details.ChangeOldLaborerID, details.ChangeOldLaborer = resolveIncludedName(included, rel, "nickname")
	}
	if rel, ok := resp.Data.Relationships["change-new-laborer"]; ok && rel.Data != nil {
		details.ChangeNewLaborerID, details.ChangeNewLaborer = resolveIncludedName(included, rel, "nickname")
	}
	if rel, ok := resp.Data.Relationships["change-old-job-site"]; ok && rel.Data != nil {
		details.ChangeOldJobSiteID, details.ChangeOldJobSite = resolveIncludedName(included, rel, "name")
	}
	if rel, ok := resp.Data.Relationships["change-new-job-site"]; ok && rel.Data != nil {
		details.ChangeNewJobSiteID, details.ChangeNewJobSite = resolveIncludedName(included, rel, "name")
	}

	return details
}

func renderJobProductionPlanChangeSetDetails(cmd *cobra.Command, details jobProductionPlanChangeSetDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)

	if details.BrokerID != "" || details.BrokerName != "" {
		fmt.Fprintf(out, "Broker: %s\n", formatRelated(details.BrokerName, details.BrokerID))
	}
	if details.CustomerID != "" || details.CustomerName != "" {
		fmt.Fprintf(out, "Customer: %s\n", formatRelated(details.CustomerName, details.CustomerID))
	}
	if details.CreatedByID != "" || details.CreatedBy != "" {
		fmt.Fprintf(out, "Created By: %s\n", formatRelated(details.CreatedBy, details.CreatedByID))
	}

	fmt.Fprintf(out, "Should Persist: %t\n", details.ShouldPersist)
	fmt.Fprintf(out, "Skip Invalid Plans: %t\n", details.SkipInvalidPlans)
	if details.ProcessedAt != "" {
		fmt.Fprintf(out, "Processed At: %s\n", details.ProcessedAt)
	}
	if len(details.MatchIDs) > 0 {
		fmt.Fprintf(out, "Match IDs: %s\n", strings.Join(details.MatchIDs, ", "))
	}

	if hasScopeFields(details) {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Scope:")
		if len(details.ScopeJobProductionPlanIDs) > 0 {
			fmt.Fprintf(out, "  Job Production Plan IDs: %s\n", strings.Join(details.ScopeJobProductionPlanIDs, ", "))
		}
		if len(details.ScopeProjectIDs) > 0 {
			fmt.Fprintf(out, "  Project IDs: %s\n", strings.Join(details.ScopeProjectIDs, ", "))
		}
		if len(details.ScopePlannerIDs) > 0 {
			fmt.Fprintf(out, "  Planner IDs: %s\n", strings.Join(details.ScopePlannerIDs, ", "))
		}
		if len(details.ScopeMaterialSiteIDs) > 0 {
			fmt.Fprintf(out, "  Material Site IDs: %s\n", strings.Join(details.ScopeMaterialSiteIDs, ", "))
		}
		if len(details.ScopeStatuses) > 0 {
			fmt.Fprintf(out, "  Statuses: %s\n", strings.Join(details.ScopeStatuses, ", "))
		}
		if details.ScopeStartOnMin != "" || details.ScopeStartOnMax != "" {
			fmt.Fprintf(out, "  Start On Range: %s - %s\n", details.ScopeStartOnMin, details.ScopeStartOnMax)
		}
		if len(details.ScopeUltimateMaterialTypes) > 0 {
			fmt.Fprintf(out, "  Ultimate Material Types: %s\n", strings.Join(details.ScopeUltimateMaterialTypes, ", "))
		}
		if len(details.ScopeMaterialTypeIDs) > 0 {
			fmt.Fprintf(out, "  Material Type IDs: %s\n", strings.Join(details.ScopeMaterialTypeIDs, ", "))
		}
		if len(details.ScopeForemanIDs) > 0 {
			fmt.Fprintf(out, "  Foreman IDs: %s\n", strings.Join(details.ScopeForemanIDs, ", "))
		}
	}

	if hasChangeFields(details) {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Changes:")
		if details.ChangeJobNumberNew != "" {
			fmt.Fprintf(out, "  Job Number New: %s\n", details.ChangeJobNumberNew)
		}
		if details.ChangeRawJobNumberNew != "" {
			fmt.Fprintf(out, "  Raw Job Number New: %s\n", details.ChangeRawJobNumberNew)
		}
		if details.ChangeOldMaterialTypeID != "" || details.ChangeNewMaterialTypeID != "" {
			fmt.Fprintf(out, "  Material Type: %s -> %s\n",
				formatRelated(details.ChangeOldMaterialType, details.ChangeOldMaterialTypeID),
				formatRelated(details.ChangeNewMaterialType, details.ChangeNewMaterialTypeID),
			)
		}
		if details.ChangeOldMaterialSiteID != "" || details.ChangeNewMaterialSiteID != "" {
			fmt.Fprintf(out, "  Material Site: %s -> %s\n",
				formatRelated(details.ChangeOldMaterialSite, details.ChangeOldMaterialSiteID),
				formatRelated(details.ChangeNewMaterialSite, details.ChangeNewMaterialSiteID),
			)
		}
		if details.ChangeOldCostCodeID != "" || details.ChangeNewCostCodeID != "" {
			fmt.Fprintf(out, "  Cost Code: %s -> %s\n",
				formatRelated(details.ChangeOldCostCode, details.ChangeOldCostCodeID),
				formatRelated(details.ChangeNewCostCode, details.ChangeNewCostCodeID),
			)
		}
		if details.ChangeOldInspectorID != "" || details.ChangeNewInspectorID != "" {
			fmt.Fprintf(out, "  Inspector: %s -> %s\n",
				formatRelated(details.ChangeOldInspector, details.ChangeOldInspectorID),
				formatRelated(details.ChangeNewInspector, details.ChangeNewInspectorID),
			)
		}
		if details.ChangeNewPlannerID != "" || details.ChangeNewPlannerNullify {
			fmt.Fprintf(out, "  New Planner: %s\n", formatRelated(details.ChangeNewPlanner, details.ChangeNewPlannerID))
			if details.ChangeNewPlannerNullify {
				fmt.Fprintln(out, "  New Planner Nullify: true")
			}
		}
		if details.ChangeNewProjectManagerID != "" || details.ChangeNewProjectManagerNullify {
			fmt.Fprintf(out, "  New Project Manager: %s\n", formatRelated(details.ChangeNewProjectManager, details.ChangeNewProjectManagerID))
			if details.ChangeNewProjectManagerNullify {
				fmt.Fprintln(out, "  New Project Manager Nullify: true")
			}
		}
		if details.ChangeOldJppMaterialTypeQualityControlClassificationID != "" || details.ChangeNewJppMaterialTypeQualityControlClassificationID != "" {
			fmt.Fprintf(out, "  QC Classification: %s -> %s\n",
				formatRelated(details.ChangeOldJppMaterialTypeQualityControlClassification, details.ChangeOldJppMaterialTypeQualityControlClassificationID),
				formatRelated(details.ChangeNewJppMaterialTypeQualityControlClassification, details.ChangeNewJppMaterialTypeQualityControlClassificationID),
			)
		}
		if details.ChangeOldJppMaterialTypeExplicitMaterialMixDesignID != "" || details.ChangeNewJppMaterialTypeExplicitMaterialMixDesignID != "" {
			fmt.Fprintf(out, "  Mix Design: %s -> %s\n",
				formatRelated(details.ChangeOldJppMaterialTypeExplicitMaterialMixDesign, details.ChangeOldJppMaterialTypeExplicitMaterialMixDesignID),
				formatRelated(details.ChangeNewJppMaterialTypeExplicitMaterialMixDesign, details.ChangeNewJppMaterialTypeExplicitMaterialMixDesignID),
			)
		}
		if details.ChangeOldLaborerID != "" || details.ChangeNewLaborerID != "" {
			fmt.Fprintf(out, "  Laborer: %s -> %s\n",
				formatRelated(details.ChangeOldLaborer, details.ChangeOldLaborerID),
				formatRelated(details.ChangeNewLaborer, details.ChangeNewLaborerID),
			)
		}
		if details.ChangeOldJobSiteID != "" || details.ChangeNewJobSiteID != "" {
			fmt.Fprintf(out, "  Job Site: %s -> %s\n",
				formatRelated(details.ChangeOldJobSite, details.ChangeOldJobSiteID),
				formatRelated(details.ChangeNewJobSite, details.ChangeNewJobSiteID),
			)
		}
		if details.ChangeNewStatus != "" {
			fmt.Fprintf(out, "  New Status: %s\n", details.ChangeNewStatus)
		}
		if details.ChangeOldIsScheduleLocked != "" || details.ChangeNewIsScheduleLocked != "" {
			fmt.Fprintf(out, "  Schedule Locked: %s -> %s\n", details.ChangeOldIsScheduleLocked, details.ChangeNewIsScheduleLocked)
		}
		if details.ChangeNewDaysOffset != "" {
			fmt.Fprintf(out, "  Days Offset: %s\n", details.ChangeNewDaysOffset)
		}
		if details.ChangeNewOffsetSkipSaturdays != "" {
			fmt.Fprintf(out, "  Skip Saturdays: %s\n", details.ChangeNewOffsetSkipSaturdays)
		}
		if details.ChangeNewOffsetSkipSundays != "" {
			fmt.Fprintf(out, "  Skip Sundays: %s\n", details.ChangeNewOffsetSkipSundays)
		}
	}

	if len(details.Results) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Results:")
		if pretty, err := json.MarshalIndent(details.Results, "", "  "); err == nil {
			fmt.Fprintln(out, string(pretty))
		} else {
			fmt.Fprintln(out, "  (unable to format results)")
		}
	}

	return nil
}

func resolveIncludedName(included map[string]jsonAPIResource, rel jsonAPIRelationship, attr string) (string, string) {
	if rel.Data == nil {
		return "", ""
	}
	id := rel.Data.ID
	if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
		return id, stringAttr(inc.Attributes, attr)
	}
	return id, ""
}

func formatRelated(name, id string) string {
	if name == "" {
		return id
	}
	if id == "" {
		return name
	}
	return fmt.Sprintf("%s (%s)", name, id)
}

func mapSliceAttr(attrs map[string]any, key string) []map[string]any {
	if attrs == nil {
		return nil
	}
	value, ok := attrs[key]
	if !ok || value == nil {
		return nil
	}
	switch typed := value.(type) {
	case []map[string]any:
		return typed
	case []any:
		results := make([]map[string]any, 0, len(typed))
		for _, item := range typed {
			if item == nil {
				continue
			}
			if asMap, ok := item.(map[string]any); ok {
				results = append(results, asMap)
				continue
			}
			results = append(results, map[string]any{"value": item})
		}
		return results
	default:
		return []map[string]any{{"value": typed}}
	}
}

func hasScopeFields(details jobProductionPlanChangeSetDetails) bool {
	return len(details.ScopeJobProductionPlanIDs) > 0 ||
		len(details.ScopeProjectIDs) > 0 ||
		len(details.ScopePlannerIDs) > 0 ||
		len(details.ScopeMaterialSiteIDs) > 0 ||
		len(details.ScopeStatuses) > 0 ||
		details.ScopeStartOnMin != "" ||
		details.ScopeStartOnMax != "" ||
		len(details.ScopeUltimateMaterialTypes) > 0 ||
		len(details.ScopeMaterialTypeIDs) > 0 ||
		len(details.ScopeForemanIDs) > 0
}

func hasChangeFields(details jobProductionPlanChangeSetDetails) bool {
	return details.ChangeJobNumberNew != "" ||
		details.ChangeRawJobNumberNew != "" ||
		details.ChangeOldMaterialTypeID != "" ||
		details.ChangeNewMaterialTypeID != "" ||
		details.ChangeOldMaterialSiteID != "" ||
		details.ChangeNewMaterialSiteID != "" ||
		details.ChangeOldCostCodeID != "" ||
		details.ChangeNewCostCodeID != "" ||
		details.ChangeOldInspectorID != "" ||
		details.ChangeNewInspectorID != "" ||
		details.ChangeNewPlannerID != "" ||
		details.ChangeNewPlannerNullify ||
		details.ChangeNewProjectManagerID != "" ||
		details.ChangeNewProjectManagerNullify ||
		details.ChangeOldJppMaterialTypeQualityControlClassificationID != "" ||
		details.ChangeNewJppMaterialTypeQualityControlClassificationID != "" ||
		details.ChangeOldJppMaterialTypeExplicitMaterialMixDesignID != "" ||
		details.ChangeNewJppMaterialTypeExplicitMaterialMixDesignID != "" ||
		details.ChangeOldLaborerID != "" ||
		details.ChangeNewLaborerID != "" ||
		details.ChangeOldJobSiteID != "" ||
		details.ChangeNewJobSiteID != "" ||
		details.ChangeNewStatus != "" ||
		details.ChangeOldIsScheduleLocked != "" ||
		details.ChangeNewIsScheduleLocked != "" ||
		details.ChangeNewDaysOffset != "" ||
		details.ChangeNewOffsetSkipSaturdays != "" ||
		details.ChangeNewOffsetSkipSundays != ""
}
