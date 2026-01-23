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

type doJobProductionPlanChangeSetsCreateOptions struct {
	BaseURL string
	Token   string
	JSON    bool

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

func newDoJobProductionPlanChangeSetsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a job production plan change set",
		Long: `Create a job production plan change set.

Scope flags define which job production plans are targeted. Change flags define
what updates should be applied. At least one change flag is required.

Scope flags:
  --scope-job-production-plan-ids     Job production plan IDs (comma-separated)
  --scope-project-ids                 Project IDs (comma-separated)
  --scope-planner-ids                 Planner user IDs (comma-separated)
  --scope-material-site-ids           Material site IDs (comma-separated)
  --scope-statuses                    Job production plan statuses (comma-separated)
  --scope-start-on-min                Minimum start-on date (YYYY-MM-DD)
  --scope-start-on-max                Maximum start-on date (YYYY-MM-DD)
  --scope-ultimate-material-types     Ultimate material type names (comma-separated)
  --scope-material-type-ids           Material type IDs (comma-separated)
  --scope-foreman-ids                 Foreman laborer IDs (comma-separated)

Change flags:
  --change-job-number-new             New job number
  --change-raw-job-number-new         New raw job number
  --change-new-planner-nullify        Nullify planner
  --change-new-project-manager-nullify Nullify project manager
  --change-new-status                 New job production plan status
  --change-old-is-schedule-locked     Old schedule locked value (true/false)
  --change-new-is-schedule-locked     New schedule locked value (true/false)
  --change-new-days-offset            Offset start-on date by days
  --change-new-offset-skip-saturdays  Skip Saturdays when offsetting
  --change-new-offset-skip-sundays    Skip Sundays when offsetting
  --should-persist                    Persist changes (true/false)
  --skip-invalid-plans                Skip invalid plans when persisting (true/false)

Relationships:
  --broker                                        Broker ID
  --customer                                      Customer ID
  --change-old-material-type                      Old material type ID
  --change-new-material-type                      New material type ID
  --change-old-material-site                      Old material site ID
  --change-new-material-site                      New material site ID
  --change-old-cost-code                          Old cost code ID
  --change-new-cost-code                          New cost code ID
  --change-old-inspector                          Old inspector user ID
  --change-new-inspector                          New inspector user ID
  --change-new-planner                            New planner user ID
  --change-new-project-manager                    New project manager user ID
  --change-old-jpp-material-type-quality-control-classification Old QC classification ID
  --change-new-jpp-material-type-quality-control-classification New QC classification ID
  --change-old-jpp-material-type-explicit-material-mix-design   Old material mix design ID
  --change-new-jpp-material-type-explicit-material-mix-design   New material mix design ID
  --change-old-laborer                            Old laborer ID
  --change-new-laborer                            New laborer ID
  --change-old-job-site                           Old job site ID
  --change-new-job-site                           New job site ID`,
		Example: `  # Preview status change for a scope
  xbe do job-production-plan-change-sets create \
    --customer 123 \
    --scope-statuses editing,submitted \
    --change-new-status approved

  # Update material type and job site for specific plans
  xbe do job-production-plan-change-sets create \
    --customer 123 \
    --scope-job-production-plan-ids 456,789 \
    --change-old-material-type 111 \
    --change-new-material-type 222 \
    --change-old-job-site 333 \
    --change-new-job-site 444`,
		Args: cobra.NoArgs,
		RunE: runDoJobProductionPlanChangeSetsCreate,
	}
	initDoJobProductionPlanChangeSetsCreateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanChangeSetsCmd.AddCommand(newDoJobProductionPlanChangeSetsCreateCmd())
}

func initDoJobProductionPlanChangeSetsCreateFlags(cmd *cobra.Command) {
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

func runDoJobProductionPlanChangeSetsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoJobProductionPlanChangeSetsCreateOptions(cmd)
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

	if !hasChangeSetChanges(cmd, opts) {
		err := fmt.Errorf("at least one change flag is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}

	if values := cleanStringSlice(opts.ScopeJobProductionPlanIDs); len(values) > 0 {
		attributes["scope-job-production-plan-ids"] = values
	}
	if values := cleanStringSlice(opts.ScopeProjectIDs); len(values) > 0 {
		attributes["scope-project-ids"] = values
	}
	if values := cleanStringSlice(opts.ScopePlannerIDs); len(values) > 0 {
		attributes["scope-planner-ids"] = values
	}
	if values := cleanStringSlice(opts.ScopeMaterialSiteIDs); len(values) > 0 {
		attributes["scope-material-site-ids"] = values
	}
	if values := cleanStringSlice(opts.ScopeStatuses); len(values) > 0 {
		attributes["scope-statuses"] = values
	}
	if opts.ScopeStartOnMin != "" {
		attributes["scope-start-on-min"] = opts.ScopeStartOnMin
	}
	if opts.ScopeStartOnMax != "" {
		attributes["scope-start-on-max"] = opts.ScopeStartOnMax
	}
	if values := cleanStringSlice(opts.ScopeUltimateMaterialTypes); len(values) > 0 {
		attributes["scope-ultimate-material-types"] = values
	}
	if values := cleanStringSlice(opts.ScopeMaterialTypeIDs); len(values) > 0 {
		attributes["scope-material-type-ids"] = values
	}
	if values := cleanStringSlice(opts.ScopeForemanIDs); len(values) > 0 {
		attributes["scope-foreman-ids"] = values
	}

	if opts.ChangeJobNumberNew != "" {
		attributes["change-job-number-new"] = opts.ChangeJobNumberNew
	}
	if opts.ChangeRawJobNumberNew != "" {
		attributes["change-raw-job-number-new"] = opts.ChangeRawJobNumberNew
	}
	if cmd.Flags().Changed("change-new-planner-nullify") {
		attributes["change-new-planner-nullify"] = opts.ChangeNewPlannerNullify
	}
	if cmd.Flags().Changed("change-new-project-manager-nullify") {
		attributes["change-new-project-manager-nullify"] = opts.ChangeNewProjectManagerNullify
	}
	if opts.ChangeNewStatus != "" {
		attributes["change-new-status"] = opts.ChangeNewStatus
	}
	setBoolAttrIfPresent(attributes, "change-old-is-schedule-locked", opts.ChangeOldIsScheduleLocked)
	setBoolAttrIfPresent(attributes, "change-new-is-schedule-locked", opts.ChangeNewIsScheduleLocked)
	setIntAttrIfPresent(attributes, "change-new-days-offset", opts.ChangeNewDaysOffset)
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
	setRelationshipIfPresent(relationships, "broker", "brokers", opts.Broker)
	setRelationshipIfPresent(relationships, "customer", "customers", opts.Customer)
	setRelationshipIfPresent(relationships, "change-old-material-type", "material-types", opts.ChangeOldMaterialType)
	setRelationshipIfPresent(relationships, "change-new-material-type", "material-types", opts.ChangeNewMaterialType)
	setRelationshipIfPresent(relationships, "change-old-material-site", "material-sites", opts.ChangeOldMaterialSite)
	setRelationshipIfPresent(relationships, "change-new-material-site", "material-sites", opts.ChangeNewMaterialSite)
	setRelationshipIfPresent(relationships, "change-old-cost-code", "cost-codes", opts.ChangeOldCostCode)
	setRelationshipIfPresent(relationships, "change-new-cost-code", "cost-codes", opts.ChangeNewCostCode)
	setRelationshipIfPresent(relationships, "change-old-inspector", "users", opts.ChangeOldInspector)
	setRelationshipIfPresent(relationships, "change-new-inspector", "users", opts.ChangeNewInspector)
	setRelationshipIfPresent(relationships, "change-new-planner", "users", opts.ChangeNewPlanner)
	setRelationshipIfPresent(relationships, "change-new-project-manager", "users", opts.ChangeNewProjectManager)
	setRelationshipIfPresent(relationships, "change-old-jpp-material-type-quality-control-classification", "quality-control-classifications", opts.ChangeOldJppMaterialTypeQualityControlClassification)
	setRelationshipIfPresent(relationships, "change-new-jpp-material-type-quality-control-classification", "quality-control-classifications", opts.ChangeNewJppMaterialTypeQualityControlClassification)
	setRelationshipIfPresent(relationships, "change-old-jpp-material-type-explicit-material-mix-design", "material-mix-designs", opts.ChangeOldJppMaterialTypeExplicitMaterialMixDesign)
	setRelationshipIfPresent(relationships, "change-new-jpp-material-type-explicit-material-mix-design", "material-mix-designs", opts.ChangeNewJppMaterialTypeExplicitMaterialMixDesign)
	setRelationshipIfPresent(relationships, "change-old-laborer", "laborers", opts.ChangeOldLaborer)
	setRelationshipIfPresent(relationships, "change-new-laborer", "laborers", opts.ChangeNewLaborer)
	setRelationshipIfPresent(relationships, "change-old-job-site", "job-sites", opts.ChangeOldJobSite)
	setRelationshipIfPresent(relationships, "change-new-job-site", "job-sites", opts.ChangeNewJobSite)

	data := map[string]any{
		"type":       "job-production-plan-change-sets",
		"attributes": attributes,
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

	body, _, err := client.Post(cmd.Context(), "/v1/job-production-plan-change-sets", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created job production plan change set %s\n", row.ID)
	return nil
}

func parseDoJobProductionPlanChangeSetsCreateOptions(cmd *cobra.Command) (doJobProductionPlanChangeSetsCreateOptions, error) {
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

	return doJobProductionPlanChangeSetsCreateOptions{
		BaseURL:                        baseURL,
		Token:                          token,
		JSON:                           jsonOut,
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

func hasChangeSetChanges(cmd *cobra.Command, opts doJobProductionPlanChangeSetsCreateOptions) bool {
	if opts.ChangeJobNumberNew != "" || opts.ChangeRawJobNumberNew != "" || opts.ChangeNewStatus != "" {
		return true
	}
	if opts.ChangeOldMaterialType != "" || opts.ChangeNewMaterialType != "" {
		return true
	}
	if opts.ChangeOldMaterialSite != "" || opts.ChangeNewMaterialSite != "" {
		return true
	}
	if opts.ChangeOldCostCode != "" || opts.ChangeNewCostCode != "" {
		return true
	}
	if opts.ChangeOldInspector != "" || opts.ChangeNewInspector != "" {
		return true
	}
	if opts.ChangeNewPlanner != "" || opts.ChangeNewProjectManager != "" {
		return true
	}
	if opts.ChangeOldJppMaterialTypeQualityControlClassification != "" || opts.ChangeNewJppMaterialTypeQualityControlClassification != "" {
		return true
	}
	if opts.ChangeOldJppMaterialTypeExplicitMaterialMixDesign != "" || opts.ChangeNewJppMaterialTypeExplicitMaterialMixDesign != "" {
		return true
	}
	if opts.ChangeOldLaborer != "" || opts.ChangeNewLaborer != "" {
		return true
	}
	if opts.ChangeOldJobSite != "" || opts.ChangeNewJobSite != "" {
		return true
	}
	if opts.ChangeOldIsScheduleLocked != "" || opts.ChangeNewIsScheduleLocked != "" {
		return true
	}
	if opts.ChangeNewDaysOffset != "" || cmd.Flags().Changed("change-new-offset-skip-saturdays") || cmd.Flags().Changed("change-new-offset-skip-sundays") {
		return true
	}
	if opts.ChangeNewPlannerNullify || opts.ChangeNewProjectManagerNullify {
		return true
	}
	return false
}

func setRelationshipIfPresent(relationships map[string]any, key, typ, id string) {
	id = strings.TrimSpace(id)
	if id == "" {
		return
	}
	relationships[key] = map[string]any{
		"data": map[string]any{
			"type": typ,
			"id":   id,
		},
	}
}

func cleanStringSlice(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	cleaned := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		cleaned = append(cleaned, value)
	}
	return cleaned
}
