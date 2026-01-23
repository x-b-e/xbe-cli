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

type doJobProductionPlanChangeSetsUpdateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string

	ScopeJobProductionPlanIDs  []string
	ScopeProjectIDs            []string
	ScopePlannerIDs            []string
	ScopeMaterialSiteIDs       []string
	ScopeStatuses              []string
	ScopeStartOnMin            string
	ScopeStartOnMax            string
	ScopeUltimateMaterialTypes []string
	ScopeMaterialTypeIDs       []string
	ScopeForemanIDs            []string

	ChangeJobNumberNew    string
	ChangeRawJobNumberNew string

	ChangeNewPlannerNullify        bool
	ChangeNewProjectManagerNullify bool
	ChangeNewStatus                string
	ChangeOldIsScheduleLocked      string
	ChangeNewIsScheduleLocked      string
	ChangeNewDaysOffset            string
	ChangeNewOffsetSkipSaturdays   bool
	ChangeNewOffsetSkipSundays     bool
	ShouldPersist                  bool
	SkipInvalidPlans               bool

	Broker                                               string
	Customer                                             string
	ChangeOldMaterialType                                string
	ChangeNewMaterialType                                string
	ChangeOldMaterialSite                                string
	ChangeNewMaterialSite                                string
	ChangeOldCostCode                                    string
	ChangeNewCostCode                                    string
	ChangeOldInspector                                   string
	ChangeNewInspector                                   string
	ChangeNewPlanner                                     string
	ChangeNewProjectManager                              string
	ChangeOldJppMaterialTypeQualityControlClassification string
	ChangeNewJppMaterialTypeQualityControlClassification string
	ChangeOldJppMaterialTypeExplicitMaterialMixDesign    string
	ChangeNewJppMaterialTypeExplicitMaterialMixDesign    string
	ChangeOldLaborer                                     string
	ChangeNewLaborer                                     string
	ChangeOldJobSite                                     string
	ChangeNewJobSite                                     string
}

func newDoJobProductionPlanChangeSetsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a job production plan change set",
		Long: `Update a job production plan change set.

Note: Change sets are immutable after creation; update requests
will be rejected by the API. This command exists for completeness.`,
		Example: `  # Attempt to update a change set
  xbe do job-production-plan-change-sets update 123 --change-new-status approved`,
		Args: cobra.ExactArgs(1),
		RunE: runDoJobProductionPlanChangeSetsUpdate,
	}
	initDoJobProductionPlanChangeSetsUpdateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanChangeSetsCmd.AddCommand(newDoJobProductionPlanChangeSetsUpdateCmd())
}

func initDoJobProductionPlanChangeSetsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().StringSlice("scope-job-production-plan-ids", nil, "Job production plan IDs (comma-separated)")
	cmd.Flags().StringSlice("scope-project-ids", nil, "Project IDs (comma-separated)")
	cmd.Flags().StringSlice("scope-planner-ids", nil, "Planner user IDs (comma-separated)")
	cmd.Flags().StringSlice("scope-material-site-ids", nil, "Material site IDs (comma-separated)")
	cmd.Flags().StringSlice("scope-statuses", nil, "Job production plan statuses (comma-separated)")
	cmd.Flags().String("scope-start-on-min", "", "Minimum start-on date (YYYY-MM-DD)")
	cmd.Flags().String("scope-start-on-max", "", "Maximum start-on date (YYYY-MM-DD)")
	cmd.Flags().StringSlice("scope-ultimate-material-types", nil, "Ultimate material type names (comma-separated)")
	cmd.Flags().StringSlice("scope-material-type-ids", nil, "Material type IDs (comma-separated)")
	cmd.Flags().StringSlice("scope-foreman-ids", nil, "Foreman laborer IDs (comma-separated)")

	cmd.Flags().String("change-job-number-new", "", "New job number")
	cmd.Flags().String("change-raw-job-number-new", "", "New raw job number")
	cmd.Flags().Bool("change-new-planner-nullify", false, "Nullify planner")
	cmd.Flags().Bool("change-new-project-manager-nullify", false, "Nullify project manager")
	cmd.Flags().String("change-new-status", "", "New job production plan status")
	cmd.Flags().String("change-old-is-schedule-locked", "", "Old schedule locked value (true/false)")
	cmd.Flags().String("change-new-is-schedule-locked", "", "New schedule locked value (true/false)")
	cmd.Flags().String("change-new-days-offset", "", "Offset start-on date by days")
	cmd.Flags().Bool("change-new-offset-skip-saturdays", false, "Skip Saturdays when offsetting")
	cmd.Flags().Bool("change-new-offset-skip-sundays", false, "Skip Sundays when offsetting")
	cmd.Flags().Bool("should-persist", false, "Persist changes")
	cmd.Flags().Bool("skip-invalid-plans", false, "Skip invalid plans when persisting")

	cmd.Flags().String("broker", "", "Broker ID")
	cmd.Flags().String("customer", "", "Customer ID")
	cmd.Flags().String("change-old-material-type", "", "Old material type ID")
	cmd.Flags().String("change-new-material-type", "", "New material type ID")
	cmd.Flags().String("change-old-material-site", "", "Old material site ID")
	cmd.Flags().String("change-new-material-site", "", "New material site ID")
	cmd.Flags().String("change-old-cost-code", "", "Old cost code ID")
	cmd.Flags().String("change-new-cost-code", "", "New cost code ID")
	cmd.Flags().String("change-old-inspector", "", "Old inspector user ID")
	cmd.Flags().String("change-new-inspector", "", "New inspector user ID")
	cmd.Flags().String("change-new-planner", "", "New planner user ID")
	cmd.Flags().String("change-new-project-manager", "", "New project manager user ID")
	cmd.Flags().String("change-old-jpp-material-type-quality-control-classification", "", "Old QC classification ID")
	cmd.Flags().String("change-new-jpp-material-type-quality-control-classification", "", "New QC classification ID")
	cmd.Flags().String("change-old-jpp-material-type-explicit-material-mix-design", "", "Old material mix design ID")
	cmd.Flags().String("change-new-jpp-material-type-explicit-material-mix-design", "", "New material mix design ID")
	cmd.Flags().String("change-old-laborer", "", "Old laborer ID")
	cmd.Flags().String("change-new-laborer", "", "New laborer ID")
	cmd.Flags().String("change-old-job-site", "", "Old job site ID")
	cmd.Flags().String("change-new-job-site", "", "New job site ID")

	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlanChangeSetsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoJobProductionPlanChangeSetsUpdateOptions(cmd, args)
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

	if !hasUpdateChanges(cmd, opts) {
		err := fmt.Errorf("no update fields specified")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}

	if cmd.Flags().Changed("scope-job-production-plan-ids") {
		attributes["scope-job-production-plan-ids"] = cleanStringSlice(opts.ScopeJobProductionPlanIDs)
	}
	if cmd.Flags().Changed("scope-project-ids") {
		attributes["scope-project-ids"] = cleanStringSlice(opts.ScopeProjectIDs)
	}
	if cmd.Flags().Changed("scope-planner-ids") {
		attributes["scope-planner-ids"] = cleanStringSlice(opts.ScopePlannerIDs)
	}
	if cmd.Flags().Changed("scope-material-site-ids") {
		attributes["scope-material-site-ids"] = cleanStringSlice(opts.ScopeMaterialSiteIDs)
	}
	if cmd.Flags().Changed("scope-statuses") {
		attributes["scope-statuses"] = cleanStringSlice(opts.ScopeStatuses)
	}
	if cmd.Flags().Changed("scope-start-on-min") {
		attributes["scope-start-on-min"] = opts.ScopeStartOnMin
	}
	if cmd.Flags().Changed("scope-start-on-max") {
		attributes["scope-start-on-max"] = opts.ScopeStartOnMax
	}
	if cmd.Flags().Changed("scope-ultimate-material-types") {
		attributes["scope-ultimate-material-types"] = cleanStringSlice(opts.ScopeUltimateMaterialTypes)
	}
	if cmd.Flags().Changed("scope-material-type-ids") {
		attributes["scope-material-type-ids"] = cleanStringSlice(opts.ScopeMaterialTypeIDs)
	}
	if cmd.Flags().Changed("scope-foreman-ids") {
		attributes["scope-foreman-ids"] = cleanStringSlice(opts.ScopeForemanIDs)
	}

	if cmd.Flags().Changed("change-job-number-new") {
		attributes["change-job-number-new"] = opts.ChangeJobNumberNew
	}
	if cmd.Flags().Changed("change-raw-job-number-new") {
		attributes["change-raw-job-number-new"] = opts.ChangeRawJobNumberNew
	}
	if cmd.Flags().Changed("change-new-planner-nullify") {
		attributes["change-new-planner-nullify"] = opts.ChangeNewPlannerNullify
	}
	if cmd.Flags().Changed("change-new-project-manager-nullify") {
		attributes["change-new-project-manager-nullify"] = opts.ChangeNewProjectManagerNullify
	}
	if cmd.Flags().Changed("change-new-status") {
		attributes["change-new-status"] = opts.ChangeNewStatus
	}
	if cmd.Flags().Changed("change-old-is-schedule-locked") {
		setBoolAttrIfPresent(attributes, "change-old-is-schedule-locked", opts.ChangeOldIsScheduleLocked)
	}
	if cmd.Flags().Changed("change-new-is-schedule-locked") {
		setBoolAttrIfPresent(attributes, "change-new-is-schedule-locked", opts.ChangeNewIsScheduleLocked)
	}
	if cmd.Flags().Changed("change-new-days-offset") {
		setIntAttrIfPresent(attributes, "change-new-days-offset", opts.ChangeNewDaysOffset)
	}
	if cmd.Flags().Changed("change-new-offset-skip-saturdays") {
		attributes["change-new-offset-skip-saturdays"] = opts.ChangeNewOffsetSkipSaturdays
	}
	if cmd.Flags().Changed("change-new-offset-skip-sundays") {
		attributes["change-new-offset-skip-sundays"] = opts.ChangeNewOffsetSkipSundays
	}
	if cmd.Flags().Changed("should-persist") {
		attributes["should-persist"] = opts.ShouldPersist
	}
	if cmd.Flags().Changed("skip-invalid-plans") {
		attributes["skip-invalid-plans"] = opts.SkipInvalidPlans
	}

	relationships := map[string]any{}
	setRelationshipIfChanged(cmd, relationships, "broker", "broker", "brokers", opts.Broker)
	setRelationshipIfChanged(cmd, relationships, "customer", "customer", "customers", opts.Customer)
	setRelationshipIfChanged(cmd, relationships, "change-old-material-type", "change-old-material-type", "material-types", opts.ChangeOldMaterialType)
	setRelationshipIfChanged(cmd, relationships, "change-new-material-type", "change-new-material-type", "material-types", opts.ChangeNewMaterialType)
	setRelationshipIfChanged(cmd, relationships, "change-old-material-site", "change-old-material-site", "material-sites", opts.ChangeOldMaterialSite)
	setRelationshipIfChanged(cmd, relationships, "change-new-material-site", "change-new-material-site", "material-sites", opts.ChangeNewMaterialSite)
	setRelationshipIfChanged(cmd, relationships, "change-old-cost-code", "change-old-cost-code", "cost-codes", opts.ChangeOldCostCode)
	setRelationshipIfChanged(cmd, relationships, "change-new-cost-code", "change-new-cost-code", "cost-codes", opts.ChangeNewCostCode)
	setRelationshipIfChanged(cmd, relationships, "change-old-inspector", "change-old-inspector", "users", opts.ChangeOldInspector)
	setRelationshipIfChanged(cmd, relationships, "change-new-inspector", "change-new-inspector", "users", opts.ChangeNewInspector)
	setRelationshipIfChanged(cmd, relationships, "change-new-planner", "change-new-planner", "users", opts.ChangeNewPlanner)
	setRelationshipIfChanged(cmd, relationships, "change-new-project-manager", "change-new-project-manager", "users", opts.ChangeNewProjectManager)
	setRelationshipIfChanged(cmd, relationships, "change-old-jpp-material-type-quality-control-classification", "change-old-jpp-material-type-quality-control-classification", "quality-control-classifications", opts.ChangeOldJppMaterialTypeQualityControlClassification)
	setRelationshipIfChanged(cmd, relationships, "change-new-jpp-material-type-quality-control-classification", "change-new-jpp-material-type-quality-control-classification", "quality-control-classifications", opts.ChangeNewJppMaterialTypeQualityControlClassification)
	setRelationshipIfChanged(cmd, relationships, "change-old-jpp-material-type-explicit-material-mix-design", "change-old-jpp-material-type-explicit-material-mix-design", "material-mix-designs", opts.ChangeOldJppMaterialTypeExplicitMaterialMixDesign)
	setRelationshipIfChanged(cmd, relationships, "change-new-jpp-material-type-explicit-material-mix-design", "change-new-jpp-material-type-explicit-material-mix-design", "material-mix-designs", opts.ChangeNewJppMaterialTypeExplicitMaterialMixDesign)
	setRelationshipIfChanged(cmd, relationships, "change-old-laborer", "change-old-laborer", "laborers", opts.ChangeOldLaborer)
	setRelationshipIfChanged(cmd, relationships, "change-new-laborer", "change-new-laborer", "laborers", opts.ChangeNewLaborer)
	setRelationshipIfChanged(cmd, relationships, "change-old-job-site", "change-old-job-site", "job-sites", opts.ChangeOldJobSite)
	setRelationshipIfChanged(cmd, relationships, "change-new-job-site", "change-new-job-site", "job-sites", opts.ChangeNewJobSite)

	data := map[string]any{
		"type": "job-production-plan-change-sets",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/job-production-plan-change-sets/"+opts.ID, jsonBody)
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

	row := jobProductionPlanChangeSetRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated job production plan change set %s\n", row.ID)
	return nil
}

func parseDoJobProductionPlanChangeSetsUpdateOptions(cmd *cobra.Command, args []string) (doJobProductionPlanChangeSetsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	scopeJobProductionPlanIDs, _ := cmd.Flags().GetStringSlice("scope-job-production-plan-ids")
	scopeProjectIDs, _ := cmd.Flags().GetStringSlice("scope-project-ids")
	scopePlannerIDs, _ := cmd.Flags().GetStringSlice("scope-planner-ids")
	scopeMaterialSiteIDs, _ := cmd.Flags().GetStringSlice("scope-material-site-ids")
	scopeStatuses, _ := cmd.Flags().GetStringSlice("scope-statuses")
	scopeStartOnMin, _ := cmd.Flags().GetString("scope-start-on-min")
	scopeStartOnMax, _ := cmd.Flags().GetString("scope-start-on-max")
	scopeUltimateMaterialTypes, _ := cmd.Flags().GetStringSlice("scope-ultimate-material-types")
	scopeMaterialTypeIDs, _ := cmd.Flags().GetStringSlice("scope-material-type-ids")
	scopeForemanIDs, _ := cmd.Flags().GetStringSlice("scope-foreman-ids")

	changeJobNumberNew, _ := cmd.Flags().GetString("change-job-number-new")
	changeRawJobNumberNew, _ := cmd.Flags().GetString("change-raw-job-number-new")
	changeNewPlannerNullify, _ := cmd.Flags().GetBool("change-new-planner-nullify")
	changeNewProjectManagerNullify, _ := cmd.Flags().GetBool("change-new-project-manager-nullify")
	changeNewStatus, _ := cmd.Flags().GetString("change-new-status")
	changeOldIsScheduleLocked, _ := cmd.Flags().GetString("change-old-is-schedule-locked")
	changeNewIsScheduleLocked, _ := cmd.Flags().GetString("change-new-is-schedule-locked")
	changeNewDaysOffset, _ := cmd.Flags().GetString("change-new-days-offset")
	changeNewOffsetSkipSaturdays, _ := cmd.Flags().GetBool("change-new-offset-skip-saturdays")
	changeNewOffsetSkipSundays, _ := cmd.Flags().GetBool("change-new-offset-skip-sundays")
	shouldPersist, _ := cmd.Flags().GetBool("should-persist")
	skipInvalidPlans, _ := cmd.Flags().GetBool("skip-invalid-plans")

	broker, _ := cmd.Flags().GetString("broker")
	customer, _ := cmd.Flags().GetString("customer")
	changeOldMaterialType, _ := cmd.Flags().GetString("change-old-material-type")
	changeNewMaterialType, _ := cmd.Flags().GetString("change-new-material-type")
	changeOldMaterialSite, _ := cmd.Flags().GetString("change-old-material-site")
	changeNewMaterialSite, _ := cmd.Flags().GetString("change-new-material-site")
	changeOldCostCode, _ := cmd.Flags().GetString("change-old-cost-code")
	changeNewCostCode, _ := cmd.Flags().GetString("change-new-cost-code")
	changeOldInspector, _ := cmd.Flags().GetString("change-old-inspector")
	changeNewInspector, _ := cmd.Flags().GetString("change-new-inspector")
	changeNewPlanner, _ := cmd.Flags().GetString("change-new-planner")
	changeNewProjectManager, _ := cmd.Flags().GetString("change-new-project-manager")
	changeOldJppMaterialTypeQualityControlClassification, _ := cmd.Flags().GetString("change-old-jpp-material-type-quality-control-classification")
	changeNewJppMaterialTypeQualityControlClassification, _ := cmd.Flags().GetString("change-new-jpp-material-type-quality-control-classification")
	changeOldJppMaterialTypeExplicitMaterialMixDesign, _ := cmd.Flags().GetString("change-old-jpp-material-type-explicit-material-mix-design")
	changeNewJppMaterialTypeExplicitMaterialMixDesign, _ := cmd.Flags().GetString("change-new-jpp-material-type-explicit-material-mix-design")
	changeOldLaborer, _ := cmd.Flags().GetString("change-old-laborer")
	changeNewLaborer, _ := cmd.Flags().GetString("change-new-laborer")
	changeOldJobSite, _ := cmd.Flags().GetString("change-old-job-site")
	changeNewJobSite, _ := cmd.Flags().GetString("change-new-job-site")

	return doJobProductionPlanChangeSetsUpdateOptions{
		BaseURL:                        baseURL,
		Token:                          token,
		JSON:                           jsonOut,
		ID:                             args[0],
		ScopeJobProductionPlanIDs:      scopeJobProductionPlanIDs,
		ScopeProjectIDs:                scopeProjectIDs,
		ScopePlannerIDs:                scopePlannerIDs,
		ScopeMaterialSiteIDs:           scopeMaterialSiteIDs,
		ScopeStatuses:                  scopeStatuses,
		ScopeStartOnMin:                scopeStartOnMin,
		ScopeStartOnMax:                scopeStartOnMax,
		ScopeUltimateMaterialTypes:     scopeUltimateMaterialTypes,
		ScopeMaterialTypeIDs:           scopeMaterialTypeIDs,
		ScopeForemanIDs:                scopeForemanIDs,
		ChangeJobNumberNew:             changeJobNumberNew,
		ChangeRawJobNumberNew:          changeRawJobNumberNew,
		ChangeNewPlannerNullify:        changeNewPlannerNullify,
		ChangeNewProjectManagerNullify: changeNewProjectManagerNullify,
		ChangeNewStatus:                changeNewStatus,
		ChangeOldIsScheduleLocked:      changeOldIsScheduleLocked,
		ChangeNewIsScheduleLocked:      changeNewIsScheduleLocked,
		ChangeNewDaysOffset:            changeNewDaysOffset,
		ChangeNewOffsetSkipSaturdays:   changeNewOffsetSkipSaturdays,
		ChangeNewOffsetSkipSundays:     changeNewOffsetSkipSundays,
		ShouldPersist:                  shouldPersist,
		SkipInvalidPlans:               skipInvalidPlans,
		Broker:                         broker,
		Customer:                       customer,
		ChangeOldMaterialType:          changeOldMaterialType,
		ChangeNewMaterialType:          changeNewMaterialType,
		ChangeOldMaterialSite:          changeOldMaterialSite,
		ChangeNewMaterialSite:          changeNewMaterialSite,
		ChangeOldCostCode:              changeOldCostCode,
		ChangeNewCostCode:              changeNewCostCode,
		ChangeOldInspector:             changeOldInspector,
		ChangeNewInspector:             changeNewInspector,
		ChangeNewPlanner:               changeNewPlanner,
		ChangeNewProjectManager:        changeNewProjectManager,
		ChangeOldJppMaterialTypeQualityControlClassification: changeOldJppMaterialTypeQualityControlClassification,
		ChangeNewJppMaterialTypeQualityControlClassification: changeNewJppMaterialTypeQualityControlClassification,
		ChangeOldJppMaterialTypeExplicitMaterialMixDesign:    changeOldJppMaterialTypeExplicitMaterialMixDesign,
		ChangeNewJppMaterialTypeExplicitMaterialMixDesign:    changeNewJppMaterialTypeExplicitMaterialMixDesign,
		ChangeOldLaborer: changeOldLaborer,
		ChangeNewLaborer: changeNewLaborer,
		ChangeOldJobSite: changeOldJobSite,
		ChangeNewJobSite: changeNewJobSite,
	}, nil
}

func hasUpdateChanges(cmd *cobra.Command, opts doJobProductionPlanChangeSetsUpdateOptions) bool {
	for _, flag := range []string{
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
		"broker",
		"customer",
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
	} {
		if cmd.Flags().Changed(flag) {
			return true
		}
	}
	return false
}

func setRelationshipIfChanged(cmd *cobra.Command, relationships map[string]any, flag, key, typ, id string) {
	if !cmd.Flags().Changed(flag) {
		return
	}
	id = strings.TrimSpace(id)
	if id == "" {
		relationships[key] = map[string]any{"data": nil}
		return
	}
	relationships[key] = map[string]any{
		"data": map[string]any{
			"type": typ,
			"id":   id,
		},
	}
}
