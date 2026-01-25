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

type doJobProductionPlansUpdateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	// String attributes
	JobNumber                                             string
	JobName                                               string
	PhaseName                                             string
	RawJobNumber                                          string
	StartOn                                               string
	StartTime                                             string
	EndTime                                               string
	MaterialSiteStartOn                                   string
	MaterialSiteStartTime                                 string
	ExplicitJobSiteStartOn                                string
	ExplicitJobSiteStartTime                              string
	Notes                                                 string
	GoalHours                                             string
	GoalQuantity                                          string
	RemainingQuantity                                     string
	DispatchInstructions                                  string
	TemplateName                                          string
	TemplateStartOnMin                                    string
	TemplateStartOnMax                                    string
	ExplicitLoadedMiles                                   string
	ExplicitPlanValidMaterialTransactionUnitOfMeasures    string
	DefaultMaterialTransactionTonsMax                     string
	CostPerTruckHour                                      string
	CostPerCrewHour                                       string
	DefaultTimeCardApprovalProcess                        string
	ParallelProductionCount                               string
	PlannedNonProductionTruckCount                        string
	ExplicitTimeZoneID                                    string
	ExplicitColorHex                                      string
	ExplicitJobSiteProximityMeters                        string
	ExplicitMaterialSiteProximityMeters                   string
	DefaultCrewRequirementStartAtOffsetMinutes            string
	ObservedPossibleCycleMinutes                          string
	ExplicitDriverDayMobilizationBeforeMinutes            string
	ExplicitExcessiveJobSiteWaitTimeThresholdMinutes      string
	ExplicitExcessiveMaterialSiteWaitTimeThresholdMinutes string
	ReferenceData                                         string
	ExplicitCrewRequirementInboundLatitude                string
	ExplicitCrewRequirementInboundLongitude               string
	ExplicitCrewRequirementOutboundLatitude               string
	ExplicitCrewRequirementOutboundLongitude              string
	BenchmarkTonsPerTruckHour                             string
	DefaultTransactionsStartBeforeSeconds                 string
	DefaultTransactionsEndAfterSeconds                    string
	ExplicitJobSitePhoneNumber                            string
	OnHoldComment                                         string
	NotOnHoldComment                                      string
	// Boolean attributes
	IsTemplate                                                bool
	IsOnHold                                                  bool
	IsScheduleLocked                                          bool
	IsRawJobNumberLocked                                      bool
	IsCostCodeRequiredPerSegment                              bool
	IsCostCodeAllocationRequiredPerTimeCard                   bool
	IsCostCodeAllocationRequiredPerTimeSheet                  bool
	EnableRecapNotifications                                  bool
	CreateDetectedProductionIncidents                         bool
	ApprovalRequiresJobNumber                                 bool
	IsMaintainingServiceTypeUnitOfMeasureCohort               bool
	IsManagingCrewRequirements                                bool
	IsMaterialTransactionInspectionEnabled                    bool
	IsNotifyingCrew                                           bool
	RequiresTrucking                                          bool
	RequiresMaterials                                         bool
	LockObservedPossibleCycleMinutes                          bool
	AutoCheckInDriverOnArrivalAtStartSite                     bool
	RequiresDrivingMinutes                                    bool
	RequiresMaterialSiteMinutes                               bool
	ExplicitNotifyJobProductionPlanTimeCardApprover           bool
	ExplicitEstimatesCostCodesVia                             bool
	ExplicitIsDriverExpectingMaterialTransactionInspection    bool
	ExplicitRequiresBusinessUnit                              bool
	ExplicitNotifyDriverWhenGPSNotAvailable                   bool
	ExplicitNotifyWhenAllPlanTimeSheetsSubmitted              bool
	ExplicitNotifyWhenAllPlanTimeSheetsApproved               bool
	ExplicitNotifyWhenPlanScheduleChanges                     bool
	ExplicitNotifyOnExcessiveSiteWaitTime                     bool
	ExplicitAutoApproveAutoTimeCardsWithNonMaterialQuantities bool
	ExplicitRequireAdminApprovalTimeCardAttachments           bool
	ExplicitIsTimeCardExpectingMtxnsAcceptedBeforeApproval    bool
	ExplicitIsUpdatingChecksumRangeFromSegments               bool
	EnableImplicitTimeCardApproval                            bool
	IsUsingVolumetricMeasurements                             bool
	ExplicitIsAuditingTimeCardApprovals                       bool
	ExplicitSubmissionRequiresDifferentJobSite                bool
	IsPrevailingWageExplicit                                  string // tri-state: true/false/null
	IsCertificationRequiredExplicit                           string // tri-state
	IsManagingMaterialSiteStartTimingExplicit                 string // tri-state
	IsTimeCardPayrollCertificationRequiredExplicit            string // tri-state
	IsOneWayJobExplicit                                       string // tri-state
	IsExpectingSafetyMeeting                                  bool
	AllowsUnmanagedShift                                      bool
	IsJobSiteMaterialSiteMaterialTransactionSource            bool
	ShowLoadoutPositionToDrivers                              bool
	IsExpectingDriverFieldApprovalTimeCard                    bool
	AreShiftsExpectingTimeCards                               bool
	ExplicitRequiresInspector                                 bool
	ExplicitRequiresCertifiedWeigher                          bool
	ExplicitRequiresProject                                   bool
	ExplicitIsMaterialTypeDefaultCostCodeRequired             bool
	ExplicitIsValidatingProjectCostCodes                      bool
	ExplicitAutomaticallyCreateProjectPhaseCostItems          bool
	ExplicitPlanDisallowsMtxnsImplicitMixDesignMatch          bool
	AreGoalsSyncedFromSegments                                bool
	IsValidatingProjectMaterialTypes                          bool
	IsManagingJobSiteTimesExplicit                            bool
	IsJobSiteTimesCreationAutomatedExplicit                   bool
	IsTruckerIncidentCreationAutomatedExplicit                bool
	// Relationships
	Customer               string
	JobSite                string
	BusinessUnit           string
	Planner                string
	ProjectManager         string
	Project                string
	DefaultTrucker         string
	Contractor             string
	Developer              string
	UnitOfMeasure          string
	Template               string
	ExplicitDispatchUser   string
	CertifiedWeigher       string
	SafetyMeeting          string
	EquipmentMovementTrip  string
	TrailerClassifications string // comma-separated IDs
	CostCodes              string // comma-separated IDs
}

func newDoJobProductionPlansUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a job production plan",
		Long: `Update a job production plan.

Key attributes:
  --job-number              Job number identifier
  --job-name                Job name/description
  --start-on                Start date (YYYY-MM-DD)
  --start-time              Start time (HH:MM)
  --end-time                End time (HH:MM)
  --notes                   Notes about the plan
  --goal-hours              Goal hours
  --goal-quantity           Goal quantity (e.g., tons)
  --remaining-quantity      Remaining quantity
  --dispatch-instructions   Instructions for dispatch
  --is-schedule-locked      Lock the schedule
  --is-template             Mark as template
  --template-name           Template name (for templates)
  --is-on-hold              Put plan on hold
  --on-hold-comment         Comment for hold status
  --not-on-hold-comment     Comment when removing hold

Relationships:
  --customer                Customer ID
  --job-site                Job site ID
  --business-unit           Business unit ID
  --planner                 Planner user ID
  --project-manager         Project manager user ID
  --project                 Project ID
  --template                Template job production plan ID
  --trailer-classifications Trailer classification IDs (comma-separated)
  --cost-codes              Cost code IDs (comma-separated)`,
		Example: `  # Update job name
  xbe do job-production-plans update 123 --job-name "Updated Name"

  # Update goal quantity
  xbe do job-production-plans update 123 --goal-quantity 1000

  # Lock the schedule
  xbe do job-production-plans update 123 --is-schedule-locked

  # Put plan on hold
  xbe do job-production-plans update 123 --is-on-hold --on-hold-comment "Weather delay"

  # Update cost tracking
  xbe do job-production-plans update 123 --cost-per-truck-hour 150 --cost-per-crew-hour 75`,
		Args: cobra.ExactArgs(1),
		RunE: runDoJobProductionPlansUpdate,
	}
	initDoJobProductionPlansUpdateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlansCmd.AddCommand(newDoJobProductionPlansUpdateCmd())
}

func initDoJobProductionPlansUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	// String attributes
	cmd.Flags().String("job-number", "", "Job number identifier")
	cmd.Flags().String("job-name", "", "Job name/description")
	cmd.Flags().String("phase-name", "", "Phase name")
	cmd.Flags().String("raw-job-number", "", "Raw job number")
	cmd.Flags().String("start-on", "", "Start date (YYYY-MM-DD)")
	cmd.Flags().String("start-time", "", "Start time (HH:MM)")
	cmd.Flags().String("end-time", "", "End time (HH:MM)")
	cmd.Flags().String("material-site-start-on", "", "Material site start date")
	cmd.Flags().String("material-site-start-time", "", "Material site start time")
	cmd.Flags().String("explicit-job-site-start-on", "", "Explicit job site start date")
	cmd.Flags().String("explicit-job-site-start-time", "", "Explicit job site start time (HH:MM)")
	cmd.Flags().String("notes", "", "Notes about the plan")
	cmd.Flags().String("goal-hours", "", "Goal hours")
	cmd.Flags().String("goal-quantity", "", "Goal quantity (e.g., tons)")
	cmd.Flags().String("remaining-quantity", "", "Remaining quantity")
	cmd.Flags().String("dispatch-instructions", "", "Instructions for dispatch")
	cmd.Flags().String("template-name", "", "Template name (for templates)")
	cmd.Flags().String("template-start-on-min", "", "Template start on minimum date")
	cmd.Flags().String("template-start-on-max", "", "Template start on maximum date")
	cmd.Flags().String("explicit-loaded-miles", "", "Explicit loaded miles")
	cmd.Flags().String("explicit-plan-valid-material-transaction-unit-of-measures", "", "Explicit plan valid material transaction unit of measures")
	cmd.Flags().String("default-material-transaction-tons-max", "", "Default material transaction tons max")
	cmd.Flags().String("cost-per-truck-hour", "", "Cost per truck hour")
	cmd.Flags().String("cost-per-crew-hour", "", "Cost per crew hour")
	cmd.Flags().String("default-time-card-approval-process", "", "Default time card approval process (admin/field)")
	cmd.Flags().String("parallel-production-count", "", "Parallel production count")
	cmd.Flags().String("planned-non-production-truck-count", "", "Planned non-production truck count")
	cmd.Flags().String("explicit-time-zone-id", "", "Explicit time zone ID")
	cmd.Flags().String("explicit-color-hex", "", "Explicit color hex code")
	cmd.Flags().String("explicit-job-site-proximity-meters", "", "Explicit job site proximity in meters")
	cmd.Flags().String("explicit-material-site-proximity-meters", "", "Explicit material site proximity in meters")
	cmd.Flags().String("default-crew-requirement-start-at-offset-minutes", "", "Default crew requirement start offset in minutes")
	cmd.Flags().String("observed-possible-cycle-minutes", "", "Observed possible cycle minutes")
	cmd.Flags().String("explicit-driver-day-mobilization-before-minutes", "", "Explicit driver day mobilization before minutes")
	cmd.Flags().String("explicit-excessive-job-site-wait-time-threshold-minutes", "", "Explicit excessive job site wait time threshold minutes")
	cmd.Flags().String("explicit-excessive-material-site-wait-time-threshold-minutes", "", "Explicit excessive material site wait time threshold minutes")
	cmd.Flags().String("reference-data", "", "Reference data (JSON object)")
	cmd.Flags().String("explicit-crew-requirement-inbound-latitude", "", "Explicit crew requirement inbound latitude")
	cmd.Flags().String("explicit-crew-requirement-inbound-longitude", "", "Explicit crew requirement inbound longitude")
	cmd.Flags().String("explicit-crew-requirement-outbound-latitude", "", "Explicit crew requirement outbound latitude")
	cmd.Flags().String("explicit-crew-requirement-outbound-longitude", "", "Explicit crew requirement outbound longitude")
	cmd.Flags().String("benchmark-tons-per-truck-hour", "", "Benchmark tons per truck hour")
	cmd.Flags().String("default-transactions-start-before-seconds", "", "Default transactions start before seconds")
	cmd.Flags().String("default-transactions-end-after-seconds", "", "Default transactions end after seconds")
	cmd.Flags().String("explicit-job-site-phone-number", "", "Explicit job site phone number")
	cmd.Flags().String("on-hold-comment", "", "Comment for hold status")
	cmd.Flags().String("not-on-hold-comment", "", "Comment when removing hold")
	// Boolean attributes
	cmd.Flags().Bool("is-template", false, "Mark as template")
	cmd.Flags().Bool("is-on-hold", false, "Put plan on hold")
	cmd.Flags().Bool("is-schedule-locked", false, "Lock the schedule")
	cmd.Flags().Bool("is-raw-job-number-locked", false, "Lock raw job number")
	cmd.Flags().Bool("is-cost-code-required-per-segment", false, "Require cost code per segment")
	cmd.Flags().Bool("is-cost-code-allocation-required-per-time-card", false, "Require cost code allocation per time card")
	cmd.Flags().Bool("is-cost-code-allocation-required-per-time-sheet", false, "Require cost code allocation per time sheet")
	cmd.Flags().Bool("enable-recap-notifications", false, "Enable recap notifications")
	cmd.Flags().Bool("create-detected-production-incidents", false, "Create detected production incidents")
	cmd.Flags().Bool("approval-requires-job-number", false, "Approval requires job number")
	cmd.Flags().Bool("is-maintaining-service-type-unit-of-measure-cohort", false, "Maintain service type unit of measure cohort")
	cmd.Flags().Bool("is-managing-crew-requirements", false, "Manage crew requirements")
	cmd.Flags().Bool("is-material-transaction-inspection-enabled", false, "Enable material transaction inspection")
	cmd.Flags().Bool("is-notifying-crew", false, "Notify crew")
	cmd.Flags().Bool("requires-trucking", false, "Requires trucking")
	cmd.Flags().Bool("requires-materials", false, "Requires materials")
	cmd.Flags().Bool("lock-observed-possible-cycle-minutes", false, "Lock observed possible cycle minutes")
	cmd.Flags().Bool("auto-check-in-driver-on-arrival-at-start-site", false, "Auto check-in driver on arrival at start site")
	cmd.Flags().Bool("requires-driving-minutes", false, "Requires driving minutes")
	cmd.Flags().Bool("requires-material-site-minutes", false, "Requires material site minutes")
	cmd.Flags().Bool("explicit-notify-job-production-plan-time-card-approver", false, "Explicit notify JPP time card approver")
	cmd.Flags().Bool("explicit-estimates-cost-codes-via", false, "Explicit estimates cost codes via")
	cmd.Flags().Bool("explicit-is-driver-expecting-material-transaction-inspection", false, "Explicit is driver expecting material transaction inspection")
	cmd.Flags().Bool("explicit-requires-business-unit", false, "Explicit requires business unit")
	cmd.Flags().Bool("explicit-notify-driver-when-gps-not-available", false, "Explicit notify driver when GPS not available")
	cmd.Flags().Bool("explicit-notify-when-all-plan-time-sheets-submitted", false, "Explicit notify when all plan time sheets submitted")
	cmd.Flags().Bool("explicit-notify-when-all-plan-time-sheets-approved", false, "Explicit notify when all plan time sheets approved")
	cmd.Flags().Bool("explicit-notify-when-plan-schedule-changes", false, "Explicit notify when plan schedule changes")
	cmd.Flags().Bool("explicit-notify-on-excessive-site-wait-time", false, "Explicit notify on excessive site wait time")
	cmd.Flags().Bool("explicit-auto-approve-auto-time-cards-with-non-material-quantities", false, "Explicit auto approve auto time cards with non-material quantities")
	cmd.Flags().Bool("explicit-require-admin-approval-time-card-attachments", false, "Explicit require admin approval time card attachments")
	cmd.Flags().Bool("explicit-is-time-card-expecting-mtxns-accepted-before-approval", false, "Explicit is time card expecting mtxns accepted before approval")
	cmd.Flags().Bool("explicit-is-updating-checksum-range-from-segments", false, "Explicit is updating checksum range from segments")
	cmd.Flags().Bool("enable-implicit-time-card-approval", false, "Enable implicit time card approval")
	cmd.Flags().Bool("is-using-volumetric-measurements", false, "Use volumetric measurements")
	cmd.Flags().Bool("explicit-is-auditing-time-card-approvals", false, "Explicit is auditing time card approvals")
	cmd.Flags().Bool("explicit-submission-requires-different-job-site", false, "Explicit submission requires different job site")
	cmd.Flags().String("is-prevailing-wage-explicit", "", "Is prevailing wage explicit (true/false/null)")
	cmd.Flags().String("is-certification-required-explicit", "", "Is certification required explicit (true/false/null)")
	cmd.Flags().String("is-managing-material-site-start-timing-explicit", "", "Is managing material site start timing explicit (true/false/null)")
	cmd.Flags().String("is-time-card-payroll-certification-required-explicit", "", "Is time card payroll certification required explicit (true/false/null)")
	cmd.Flags().String("is-one-way-job-explicit", "", "Is one way job explicit (true/false/null)")
	cmd.Flags().Bool("is-expecting-safety-meeting", false, "Expecting safety meeting")
	cmd.Flags().Bool("allows-unmanaged-shift", false, "Allow unmanaged shift")
	cmd.Flags().Bool("is-job-site-material-site-material-transaction-source", false, "Job site is material site material transaction source")
	cmd.Flags().Bool("show-loadout-position-to-drivers", false, "Show loadout position to drivers")
	cmd.Flags().Bool("is-expecting-driver-field-approval-time-card", false, "Expecting driver field approval time card")
	cmd.Flags().Bool("are-shifts-expecting-time-cards", false, "Shifts expecting time cards")
	cmd.Flags().Bool("explicit-requires-inspector", false, "Explicit requires inspector")
	cmd.Flags().Bool("explicit-requires-certified-weigher", false, "Explicit requires certified weigher")
	cmd.Flags().Bool("explicit-requires-project", false, "Explicit requires project")
	cmd.Flags().Bool("explicit-is-material-type-default-cost-code-required", false, "Explicit is material type default cost code required")
	cmd.Flags().Bool("explicit-is-validating-project-cost-codes", false, "Explicit is validating project cost codes")
	cmd.Flags().Bool("explicit-automatically-create-project-phase-cost-items", false, "Explicit automatically create project phase cost items")
	cmd.Flags().Bool("explicit-plan-disallows-mtxns-implicit-mix-design-match", false, "Explicit plan disallows mtxns implicit mix design match")
	cmd.Flags().Bool("are-goals-synced-from-segments", false, "Goals synced from segments")
	cmd.Flags().Bool("is-validating-project-material-types", false, "Validating project material types")
	cmd.Flags().Bool("is-managing-job-site-times-explicit", false, "Is managing job site times explicit")
	cmd.Flags().Bool("is-job-site-times-creation-automated-explicit", false, "Is job site times creation automated explicit")
	cmd.Flags().Bool("is-trucker-incident-creation-automated-explicit", false, "Is trucker incident creation automated explicit")
	// Relationships
	cmd.Flags().String("customer", "", "Customer ID")
	cmd.Flags().String("job-site", "", "Job site ID")
	cmd.Flags().String("business-unit", "", "Business unit ID")
	cmd.Flags().String("planner", "", "Planner user ID")
	cmd.Flags().String("project-manager", "", "Project manager user ID")
	cmd.Flags().String("project", "", "Project ID")
	cmd.Flags().String("default-trucker", "", "Default trucker ID")
	cmd.Flags().String("contractor", "", "Contractor ID")
	cmd.Flags().String("developer", "", "Developer ID")
	cmd.Flags().String("unit-of-measure", "", "Unit of measure ID")
	cmd.Flags().String("template", "", "Template job production plan ID")
	cmd.Flags().String("explicit-dispatch-user", "", "Explicit dispatch user ID")
	cmd.Flags().String("certified-weigher", "", "Certified weigher ID")
	cmd.Flags().String("safety-meeting", "", "Safety meeting ID")
	cmd.Flags().String("equipment-movement-trip", "", "Equipment movement trip ID")
	cmd.Flags().String("trailer-classifications", "", "Trailer classification IDs (comma-separated)")
	cmd.Flags().String("cost-codes", "", "Cost code IDs (comma-separated)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlansUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoJobProductionPlansUpdateOptions(cmd, args)
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

	attributes := map[string]any{}
	hasChanges := false

	// String attributes
	if cmd.Flags().Changed("job-number") {
		attributes["job-number"] = opts.JobNumber
		hasChanges = true
	}
	if cmd.Flags().Changed("job-name") {
		attributes["job-name"] = opts.JobName
		hasChanges = true
	}
	if cmd.Flags().Changed("phase-name") {
		attributes["phase-name"] = opts.PhaseName
		hasChanges = true
	}
	if cmd.Flags().Changed("raw-job-number") {
		attributes["raw-job-number"] = opts.RawJobNumber
		hasChanges = true
	}
	if cmd.Flags().Changed("start-on") {
		attributes["start-on"] = opts.StartOn
		hasChanges = true
	}
	if cmd.Flags().Changed("start-time") {
		attributes["start-time"] = opts.StartTime
		hasChanges = true
	}
	if cmd.Flags().Changed("end-time") {
		attributes["end-time"] = opts.EndTime
		hasChanges = true
	}
	if cmd.Flags().Changed("material-site-start-on") {
		attributes["material-site-start-on"] = opts.MaterialSiteStartOn
		hasChanges = true
	}
	if cmd.Flags().Changed("material-site-start-time") {
		attributes["material-site-start-time"] = opts.MaterialSiteStartTime
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-job-site-start-on") {
		attributes["explicit-job-site-start-on"] = opts.ExplicitJobSiteStartOn
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-job-site-start-time") {
		attributes["explicit-job-site-start-time"] = opts.ExplicitJobSiteStartTime
		hasChanges = true
	}
	if cmd.Flags().Changed("notes") {
		attributes["notes"] = opts.Notes
		hasChanges = true
	}
	if cmd.Flags().Changed("goal-hours") {
		attributes["goal-hours"] = opts.GoalHours
		hasChanges = true
	}
	if cmd.Flags().Changed("goal-quantity") {
		attributes["goal-quantity"] = opts.GoalQuantity
		hasChanges = true
	}
	if cmd.Flags().Changed("remaining-quantity") {
		attributes["remaining-quantity"] = opts.RemainingQuantity
		hasChanges = true
	}
	if cmd.Flags().Changed("dispatch-instructions") {
		attributes["dispatch-instructions"] = opts.DispatchInstructions
		hasChanges = true
	}
	if cmd.Flags().Changed("template-name") {
		attributes["template-name"] = opts.TemplateName
		hasChanges = true
	}
	if cmd.Flags().Changed("template-start-on-min") {
		attributes["template-start-on-min"] = opts.TemplateStartOnMin
		hasChanges = true
	}
	if cmd.Flags().Changed("template-start-on-max") {
		attributes["template-start-on-max"] = opts.TemplateStartOnMax
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-loaded-miles") {
		attributes["explicit-loaded-miles"] = opts.ExplicitLoadedMiles
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-plan-valid-material-transaction-unit-of-measures") {
		attributes["explicit-plan-valid-material-transaction-unit-of-measures"] = opts.ExplicitPlanValidMaterialTransactionUnitOfMeasures
		hasChanges = true
	}
	if cmd.Flags().Changed("default-material-transaction-tons-max") {
		attributes["default-material-transaction-tons-max"] = opts.DefaultMaterialTransactionTonsMax
		hasChanges = true
	}
	if cmd.Flags().Changed("cost-per-truck-hour") {
		attributes["cost-per-truck-hour"] = opts.CostPerTruckHour
		hasChanges = true
	}
	if cmd.Flags().Changed("cost-per-crew-hour") {
		attributes["cost-per-crew-hour"] = opts.CostPerCrewHour
		hasChanges = true
	}
	if cmd.Flags().Changed("default-time-card-approval-process") {
		attributes["default-time-card-approval-process"] = opts.DefaultTimeCardApprovalProcess
		hasChanges = true
	}
	if cmd.Flags().Changed("parallel-production-count") {
		attributes["parallel-production-count"] = opts.ParallelProductionCount
		hasChanges = true
	}
	if cmd.Flags().Changed("planned-non-production-truck-count") {
		attributes["planned-non-production-truck-count"] = opts.PlannedNonProductionTruckCount
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-time-zone-id") {
		attributes["explicit-time-zone-id"] = opts.ExplicitTimeZoneID
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-color-hex") {
		attributes["explicit-color-hex"] = opts.ExplicitColorHex
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-job-site-proximity-meters") {
		attributes["explicit-job-site-proximity-meters"] = opts.ExplicitJobSiteProximityMeters
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-material-site-proximity-meters") {
		attributes["explicit-material-site-proximity-meters"] = opts.ExplicitMaterialSiteProximityMeters
		hasChanges = true
	}
	if cmd.Flags().Changed("default-crew-requirement-start-at-offset-minutes") {
		attributes["default-crew-requirement-start-at-offset-minutes"] = opts.DefaultCrewRequirementStartAtOffsetMinutes
		hasChanges = true
	}
	if cmd.Flags().Changed("observed-possible-cycle-minutes") {
		attributes["observed-possible-cycle-minutes"] = opts.ObservedPossibleCycleMinutes
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-driver-day-mobilization-before-minutes") {
		attributes["explicit-driver-day-mobilization-before-minutes"] = opts.ExplicitDriverDayMobilizationBeforeMinutes
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-excessive-job-site-wait-time-threshold-minutes") {
		attributes["explicit-excessive-job-site-wait-time-threshold-minutes"] = opts.ExplicitExcessiveJobSiteWaitTimeThresholdMinutes
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-excessive-material-site-wait-time-threshold-minutes") {
		attributes["explicit-excessive-material-site-wait-time-threshold-minutes"] = opts.ExplicitExcessiveMaterialSiteWaitTimeThresholdMinutes
		hasChanges = true
	}
	if cmd.Flags().Changed("reference-data") {
		// Parse as JSON object
		var refData map[string]any
		if err := json.Unmarshal([]byte(opts.ReferenceData), &refData); err != nil {
			err = fmt.Errorf("invalid JSON for reference-data: %w", err)
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["reference-data"] = refData
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-crew-requirement-inbound-latitude") {
		attributes["explicit-crew-requirement-inbound-latitude"] = opts.ExplicitCrewRequirementInboundLatitude
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-crew-requirement-inbound-longitude") {
		attributes["explicit-crew-requirement-inbound-longitude"] = opts.ExplicitCrewRequirementInboundLongitude
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-crew-requirement-outbound-latitude") {
		attributes["explicit-crew-requirement-outbound-latitude"] = opts.ExplicitCrewRequirementOutboundLatitude
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-crew-requirement-outbound-longitude") {
		attributes["explicit-crew-requirement-outbound-longitude"] = opts.ExplicitCrewRequirementOutboundLongitude
		hasChanges = true
	}
	if cmd.Flags().Changed("benchmark-tons-per-truck-hour") {
		attributes["benchmark-tons-per-truck-hour"] = opts.BenchmarkTonsPerTruckHour
		hasChanges = true
	}
	if cmd.Flags().Changed("default-transactions-start-before-seconds") {
		attributes["default-transactions-start-before-seconds"] = opts.DefaultTransactionsStartBeforeSeconds
		hasChanges = true
	}
	if cmd.Flags().Changed("default-transactions-end-after-seconds") {
		attributes["default-transactions-end-after-seconds"] = opts.DefaultTransactionsEndAfterSeconds
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-job-site-phone-number") {
		attributes["explicit-job-site-phone-number"] = opts.ExplicitJobSitePhoneNumber
		hasChanges = true
	}
	if cmd.Flags().Changed("on-hold-comment") {
		attributes["on-hold-comment"] = opts.OnHoldComment
		hasChanges = true
	}
	if cmd.Flags().Changed("not-on-hold-comment") {
		attributes["not-on-hold-comment"] = opts.NotOnHoldComment
		hasChanges = true
	}

	// Boolean attributes
	if cmd.Flags().Changed("is-template") {
		attributes["is-template"] = opts.IsTemplate
		hasChanges = true
	}
	if cmd.Flags().Changed("is-on-hold") {
		attributes["is-on-hold"] = opts.IsOnHold
		hasChanges = true
	}
	if cmd.Flags().Changed("is-schedule-locked") {
		attributes["is-schedule-locked"] = opts.IsScheduleLocked
		hasChanges = true
	}
	if cmd.Flags().Changed("is-raw-job-number-locked") {
		attributes["is-raw-job-number-locked"] = opts.IsRawJobNumberLocked
		hasChanges = true
	}
	if cmd.Flags().Changed("is-cost-code-required-per-segment") {
		attributes["is-cost-code-required-per-segment"] = opts.IsCostCodeRequiredPerSegment
		hasChanges = true
	}
	if cmd.Flags().Changed("is-cost-code-allocation-required-per-time-card") {
		attributes["is-cost-code-allocation-required-per-time-card"] = opts.IsCostCodeAllocationRequiredPerTimeCard
		hasChanges = true
	}
	if cmd.Flags().Changed("is-cost-code-allocation-required-per-time-sheet") {
		attributes["is-cost-code-allocation-required-per-time-sheet"] = opts.IsCostCodeAllocationRequiredPerTimeSheet
		hasChanges = true
	}
	if cmd.Flags().Changed("enable-recap-notifications") {
		attributes["enable-recap-notifications"] = opts.EnableRecapNotifications
		hasChanges = true
	}
	if cmd.Flags().Changed("create-detected-production-incidents") {
		attributes["create-detected-production-incidents"] = opts.CreateDetectedProductionIncidents
		hasChanges = true
	}
	if cmd.Flags().Changed("approval-requires-job-number") {
		attributes["approval-requires-job-number"] = opts.ApprovalRequiresJobNumber
		hasChanges = true
	}
	if cmd.Flags().Changed("is-maintaining-service-type-unit-of-measure-cohort") {
		attributes["is-maintaining-service-type-unit-of-measure-cohort"] = opts.IsMaintainingServiceTypeUnitOfMeasureCohort
		hasChanges = true
	}
	if cmd.Flags().Changed("is-managing-crew-requirements") {
		attributes["is-managing-crew-requirements"] = opts.IsManagingCrewRequirements
		hasChanges = true
	}
	if cmd.Flags().Changed("is-material-transaction-inspection-enabled") {
		attributes["is-material-transaction-inspection-enabled"] = opts.IsMaterialTransactionInspectionEnabled
		hasChanges = true
	}
	if cmd.Flags().Changed("is-notifying-crew") {
		attributes["is-notifying-crew"] = opts.IsNotifyingCrew
		hasChanges = true
	}
	if cmd.Flags().Changed("requires-trucking") {
		attributes["requires-trucking"] = opts.RequiresTrucking
		hasChanges = true
	}
	if cmd.Flags().Changed("requires-materials") {
		attributes["requires-materials"] = opts.RequiresMaterials
		hasChanges = true
	}
	if cmd.Flags().Changed("lock-observed-possible-cycle-minutes") {
		attributes["lock-observed-possible-cycle-minutes"] = opts.LockObservedPossibleCycleMinutes
		hasChanges = true
	}
	if cmd.Flags().Changed("auto-check-in-driver-on-arrival-at-start-site") {
		attributes["auto-check-in-driver-on-arrival-at-start-site"] = opts.AutoCheckInDriverOnArrivalAtStartSite
		hasChanges = true
	}
	if cmd.Flags().Changed("requires-driving-minutes") {
		attributes["requires-driving-minutes"] = opts.RequiresDrivingMinutes
		hasChanges = true
	}
	if cmd.Flags().Changed("requires-material-site-minutes") {
		attributes["requires-material-site-minutes"] = opts.RequiresMaterialSiteMinutes
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-notify-job-production-plan-time-card-approver") {
		attributes["explicit-notify-job-production-plan-time-card-approver"] = opts.ExplicitNotifyJobProductionPlanTimeCardApprover
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-estimates-cost-codes-via") {
		attributes["explicit-estimates-cost-codes-via"] = opts.ExplicitEstimatesCostCodesVia
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-is-driver-expecting-material-transaction-inspection") {
		attributes["explicit-is-driver-expecting-material-transaction-inspection"] = opts.ExplicitIsDriverExpectingMaterialTransactionInspection
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-requires-business-unit") {
		attributes["explicit-requires-business-unit"] = opts.ExplicitRequiresBusinessUnit
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-notify-driver-when-gps-not-available") {
		attributes["explicit-notify-driver-when-gps-not-available"] = opts.ExplicitNotifyDriverWhenGPSNotAvailable
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-notify-when-all-plan-time-sheets-submitted") {
		attributes["explicit-notify-when-all-plan-time-sheets-submitted"] = opts.ExplicitNotifyWhenAllPlanTimeSheetsSubmitted
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-notify-when-all-plan-time-sheets-approved") {
		attributes["explicit-notify-when-all-plan-time-sheets-approved"] = opts.ExplicitNotifyWhenAllPlanTimeSheetsApproved
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-notify-when-plan-schedule-changes") {
		attributes["explicit-notify-when-plan-schedule-changes"] = opts.ExplicitNotifyWhenPlanScheduleChanges
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-notify-on-excessive-site-wait-time") {
		attributes["explicit-notify-on-excessive-site-wait-time"] = opts.ExplicitNotifyOnExcessiveSiteWaitTime
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-auto-approve-auto-time-cards-with-non-material-quantities") {
		attributes["explicit-auto-approve-auto-time-cards-with-non-material-quantities"] = opts.ExplicitAutoApproveAutoTimeCardsWithNonMaterialQuantities
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-require-admin-approval-time-card-attachments") {
		attributes["explicit-require-admin-approval-time-card-attachments"] = opts.ExplicitRequireAdminApprovalTimeCardAttachments
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-is-time-card-expecting-mtxns-accepted-before-approval") {
		attributes["explicit-is-time-card-expecting-mtxns-accepted-before-approval"] = opts.ExplicitIsTimeCardExpectingMtxnsAcceptedBeforeApproval
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-is-updating-checksum-range-from-segments") {
		attributes["explicit-is-updating-checksum-range-from-segments"] = opts.ExplicitIsUpdatingChecksumRangeFromSegments
		hasChanges = true
	}
	if cmd.Flags().Changed("enable-implicit-time-card-approval") {
		attributes["enable-implicit-time-card-approval"] = opts.EnableImplicitTimeCardApproval
		hasChanges = true
	}
	if cmd.Flags().Changed("is-using-volumetric-measurements") {
		attributes["is-using-volumetric-measurements"] = opts.IsUsingVolumetricMeasurements
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-is-auditing-time-card-approvals") {
		attributes["explicit-is-auditing-time-card-approvals"] = opts.ExplicitIsAuditingTimeCardApprovals
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-submission-requires-different-job-site") {
		attributes["explicit-submission-requires-different-job-site"] = opts.ExplicitSubmissionRequiresDifferentJobSite
		hasChanges = true
	}
	// Tri-state booleans (true/false/null)
	if cmd.Flags().Changed("is-prevailing-wage-explicit") {
		if opts.IsPrevailingWageExplicit == "null" {
			attributes["is-prevailing-wage-explicit"] = nil
		} else {
			attributes["is-prevailing-wage-explicit"] = opts.IsPrevailingWageExplicit == "true"
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("is-certification-required-explicit") {
		if opts.IsCertificationRequiredExplicit == "null" {
			attributes["is-certification-required-explicit"] = nil
		} else {
			attributes["is-certification-required-explicit"] = opts.IsCertificationRequiredExplicit == "true"
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("is-managing-material-site-start-timing-explicit") {
		if opts.IsManagingMaterialSiteStartTimingExplicit == "null" {
			attributes["is-managing-material-site-start-timing-explicit"] = nil
		} else {
			attributes["is-managing-material-site-start-timing-explicit"] = opts.IsManagingMaterialSiteStartTimingExplicit == "true"
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("is-time-card-payroll-certification-required-explicit") {
		if opts.IsTimeCardPayrollCertificationRequiredExplicit == "null" {
			attributes["is-time-card-payroll-certification-required-explicit"] = nil
		} else {
			attributes["is-time-card-payroll-certification-required-explicit"] = opts.IsTimeCardPayrollCertificationRequiredExplicit == "true"
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("is-one-way-job-explicit") {
		if opts.IsOneWayJobExplicit == "null" {
			attributes["is-one-way-job-explicit"] = nil
		} else {
			attributes["is-one-way-job-explicit"] = opts.IsOneWayJobExplicit == "true"
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("is-expecting-safety-meeting") {
		attributes["is-expecting-safety-meeting"] = opts.IsExpectingSafetyMeeting
		hasChanges = true
	}
	if cmd.Flags().Changed("allows-unmanaged-shift") {
		attributes["allows-unmanaged-shift"] = opts.AllowsUnmanagedShift
		hasChanges = true
	}
	if cmd.Flags().Changed("is-job-site-material-site-material-transaction-source") {
		attributes["is-job-site-material-site-material-transaction-source"] = opts.IsJobSiteMaterialSiteMaterialTransactionSource
		hasChanges = true
	}
	if cmd.Flags().Changed("show-loadout-position-to-drivers") {
		attributes["show-loadout-position-to-drivers"] = opts.ShowLoadoutPositionToDrivers
		hasChanges = true
	}
	if cmd.Flags().Changed("is-expecting-driver-field-approval-time-card") {
		attributes["is-expecting-driver-field-approval-time-card"] = opts.IsExpectingDriverFieldApprovalTimeCard
		hasChanges = true
	}
	if cmd.Flags().Changed("are-shifts-expecting-time-cards") {
		attributes["are-shifts-expecting-time-cards"] = opts.AreShiftsExpectingTimeCards
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-requires-inspector") {
		attributes["explicit-requires-inspector"] = opts.ExplicitRequiresInspector
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-requires-certified-weigher") {
		attributes["explicit-requires-certified-weigher"] = opts.ExplicitRequiresCertifiedWeigher
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-requires-project") {
		attributes["explicit-requires-project"] = opts.ExplicitRequiresProject
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-is-material-type-default-cost-code-required") {
		attributes["explicit-is-material-type-default-cost-code-required"] = opts.ExplicitIsMaterialTypeDefaultCostCodeRequired
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-is-validating-project-cost-codes") {
		attributes["explicit-is-validating-project-cost-codes"] = opts.ExplicitIsValidatingProjectCostCodes
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-automatically-create-project-phase-cost-items") {
		attributes["explicit-automatically-create-project-phase-cost-items"] = opts.ExplicitAutomaticallyCreateProjectPhaseCostItems
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-plan-disallows-mtxns-implicit-mix-design-match") {
		attributes["explicit-plan-disallows-mtxns-implicit-mix-design-match"] = opts.ExplicitPlanDisallowsMtxnsImplicitMixDesignMatch
		hasChanges = true
	}
	if cmd.Flags().Changed("are-goals-synced-from-segments") {
		attributes["are-goals-synced-from-segments"] = opts.AreGoalsSyncedFromSegments
		hasChanges = true
	}
	if cmd.Flags().Changed("is-validating-project-material-types") {
		attributes["is-validating-project-material-types"] = opts.IsValidatingProjectMaterialTypes
		hasChanges = true
	}
	if cmd.Flags().Changed("is-managing-job-site-times-explicit") {
		attributes["is-managing-job-site-times-explicit"] = opts.IsManagingJobSiteTimesExplicit
		hasChanges = true
	}
	if cmd.Flags().Changed("is-job-site-times-creation-automated-explicit") {
		attributes["is-job-site-times-creation-automated-explicit"] = opts.IsJobSiteTimesCreationAutomatedExplicit
		hasChanges = true
	}
	if cmd.Flags().Changed("is-trucker-incident-creation-automated-explicit") {
		attributes["is-trucker-incident-creation-automated-explicit"] = opts.IsTruckerIncidentCreationAutomatedExplicit
		hasChanges = true
	}

	relationships := map[string]any{}

	if cmd.Flags().Changed("customer") {
		if opts.Customer == "" {
			relationships["customer"] = map[string]any{"data": nil}
		} else {
			relationships["customer"] = map[string]any{
				"data": map[string]any{
					"type": "customers",
					"id":   opts.Customer,
				},
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("job-site") {
		if opts.JobSite == "" {
			relationships["job-site"] = map[string]any{"data": nil}
		} else {
			relationships["job-site"] = map[string]any{
				"data": map[string]any{
					"type": "job-sites",
					"id":   opts.JobSite,
				},
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("business-unit") {
		if opts.BusinessUnit == "" {
			relationships["business-unit"] = map[string]any{"data": nil}
		} else {
			relationships["business-unit"] = map[string]any{
				"data": map[string]any{
					"type": "business-units",
					"id":   opts.BusinessUnit,
				},
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("planner") {
		if opts.Planner == "" {
			relationships["planner"] = map[string]any{"data": nil}
		} else {
			relationships["planner"] = map[string]any{
				"data": map[string]any{
					"type": "users",
					"id":   opts.Planner,
				},
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("project-manager") {
		if opts.ProjectManager == "" {
			relationships["project-manager"] = map[string]any{"data": nil}
		} else {
			relationships["project-manager"] = map[string]any{
				"data": map[string]any{
					"type": "users",
					"id":   opts.ProjectManager,
				},
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("project") {
		if opts.Project == "" {
			relationships["project"] = map[string]any{"data": nil}
		} else {
			relationships["project"] = map[string]any{
				"data": map[string]any{
					"type": "projects",
					"id":   opts.Project,
				},
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("default-trucker") {
		if opts.DefaultTrucker == "" {
			relationships["default-trucker"] = map[string]any{"data": nil}
		} else {
			relationships["default-trucker"] = map[string]any{
				"data": map[string]any{
					"type": "truckers",
					"id":   opts.DefaultTrucker,
				},
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("contractor") {
		if opts.Contractor == "" {
			relationships["contractor"] = map[string]any{"data": nil}
		} else {
			relationships["contractor"] = map[string]any{
				"data": map[string]any{
					"type": "contractors",
					"id":   opts.Contractor,
				},
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("developer") {
		if opts.Developer == "" {
			relationships["developer"] = map[string]any{"data": nil}
		} else {
			relationships["developer"] = map[string]any{
				"data": map[string]any{
					"type": "developers",
					"id":   opts.Developer,
				},
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("unit-of-measure") {
		if opts.UnitOfMeasure == "" {
			relationships["unit-of-measure"] = map[string]any{"data": nil}
		} else {
			relationships["unit-of-measure"] = map[string]any{
				"data": map[string]any{
					"type": "unit-of-measures",
					"id":   opts.UnitOfMeasure,
				},
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("template") {
		if opts.Template == "" {
			relationships["template"] = map[string]any{"data": nil}
		} else {
			relationships["template"] = map[string]any{
				"data": map[string]any{
					"type": "job-production-plans",
					"id":   opts.Template,
				},
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-dispatch-user") {
		if opts.ExplicitDispatchUser == "" {
			relationships["explicit-dispatch-user"] = map[string]any{"data": nil}
		} else {
			relationships["explicit-dispatch-user"] = map[string]any{
				"data": map[string]any{
					"type": "users",
					"id":   opts.ExplicitDispatchUser,
				},
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("certified-weigher") {
		if opts.CertifiedWeigher == "" {
			relationships["certified-weigher"] = map[string]any{"data": nil}
		} else {
			relationships["certified-weigher"] = map[string]any{
				"data": map[string]any{
					"type": "developer-certified-weighers",
					"id":   opts.CertifiedWeigher,
				},
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("safety-meeting") {
		if opts.SafetyMeeting == "" {
			relationships["safety-meeting"] = map[string]any{"data": nil}
		} else {
			relationships["safety-meeting"] = map[string]any{
				"data": map[string]any{
					"type": "meetings",
					"id":   opts.SafetyMeeting,
				},
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("equipment-movement-trip") {
		if opts.EquipmentMovementTrip == "" {
			relationships["equipment-movement-trip"] = map[string]any{"data": nil}
		} else {
			relationships["equipment-movement-trip"] = map[string]any{
				"data": map[string]any{
					"type": "equipment-movement-trips",
					"id":   opts.EquipmentMovementTrip,
				},
			}
		}
		hasChanges = true
	}
	// Has-many relationships
	if cmd.Flags().Changed("trailer-classifications") {
		if opts.TrailerClassifications == "" {
			relationships["trailer-classifications"] = map[string]any{"data": []any{}}
		} else {
			ids := strings.Split(opts.TrailerClassifications, ",")
			data := make([]map[string]any, len(ids))
			for i, id := range ids {
				data[i] = map[string]any{
					"type": "trailer-classifications",
					"id":   strings.TrimSpace(id),
				}
			}
			relationships["trailer-classifications"] = map[string]any{"data": data}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("cost-codes") {
		if opts.CostCodes == "" {
			relationships["cost-codes"] = map[string]any{"data": []any{}}
		} else {
			ids := strings.Split(opts.CostCodes, ",")
			data := make([]map[string]any, len(ids))
			for i, id := range ids {
				data[i] = map[string]any{
					"type": "cost-codes",
					"id":   strings.TrimSpace(id),
				}
			}
			relationships["cost-codes"] = map[string]any{"data": data}
		}
		hasChanges = true
	}

	if !hasChanges {
		err := fmt.Errorf("at least one attribute or relationship must be specified")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type":       "job-production-plans",
		"id":         opts.ID,
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

	body, _, err := client.Patch(cmd.Context(), "/v1/job-production-plans/"+opts.ID, jsonBody)
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

	row := jobProductionPlanRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated job production plan %s (%s)\n", row.ID, row.JobName)
	return nil
}

func parseDoJobProductionPlansUpdateOptions(cmd *cobra.Command, args []string) (doJobProductionPlansUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	// String attributes
	jobNumber, _ := cmd.Flags().GetString("job-number")
	jobName, _ := cmd.Flags().GetString("job-name")
	phaseName, _ := cmd.Flags().GetString("phase-name")
	rawJobNumber, _ := cmd.Flags().GetString("raw-job-number")
	startOn, _ := cmd.Flags().GetString("start-on")
	startTime, _ := cmd.Flags().GetString("start-time")
	endTime, _ := cmd.Flags().GetString("end-time")
	materialSiteStartOn, _ := cmd.Flags().GetString("material-site-start-on")
	materialSiteStartTime, _ := cmd.Flags().GetString("material-site-start-time")
	explicitJobSiteStartOn, _ := cmd.Flags().GetString("explicit-job-site-start-on")
	explicitJobSiteStartTime, _ := cmd.Flags().GetString("explicit-job-site-start-time")
	notes, _ := cmd.Flags().GetString("notes")
	goalHours, _ := cmd.Flags().GetString("goal-hours")
	goalQuantity, _ := cmd.Flags().GetString("goal-quantity")
	remainingQuantity, _ := cmd.Flags().GetString("remaining-quantity")
	dispatchInstructions, _ := cmd.Flags().GetString("dispatch-instructions")
	templateName, _ := cmd.Flags().GetString("template-name")
	templateStartOnMin, _ := cmd.Flags().GetString("template-start-on-min")
	templateStartOnMax, _ := cmd.Flags().GetString("template-start-on-max")
	explicitLoadedMiles, _ := cmd.Flags().GetString("explicit-loaded-miles")
	explicitPlanValidMaterialTransactionUnitOfMeasures, _ := cmd.Flags().GetString("explicit-plan-valid-material-transaction-unit-of-measures")
	defaultMaterialTransactionTonsMax, _ := cmd.Flags().GetString("default-material-transaction-tons-max")
	costPerTruckHour, _ := cmd.Flags().GetString("cost-per-truck-hour")
	costPerCrewHour, _ := cmd.Flags().GetString("cost-per-crew-hour")
	defaultTimeCardApprovalProcess, _ := cmd.Flags().GetString("default-time-card-approval-process")
	parallelProductionCount, _ := cmd.Flags().GetString("parallel-production-count")
	plannedNonProductionTruckCount, _ := cmd.Flags().GetString("planned-non-production-truck-count")
	explicitTimeZoneID, _ := cmd.Flags().GetString("explicit-time-zone-id")
	explicitColorHex, _ := cmd.Flags().GetString("explicit-color-hex")
	explicitJobSiteProximityMeters, _ := cmd.Flags().GetString("explicit-job-site-proximity-meters")
	explicitMaterialSiteProximityMeters, _ := cmd.Flags().GetString("explicit-material-site-proximity-meters")
	defaultCrewRequirementStartAtOffsetMinutes, _ := cmd.Flags().GetString("default-crew-requirement-start-at-offset-minutes")
	observedPossibleCycleMinutes, _ := cmd.Flags().GetString("observed-possible-cycle-minutes")
	explicitDriverDayMobilizationBeforeMinutes, _ := cmd.Flags().GetString("explicit-driver-day-mobilization-before-minutes")
	explicitExcessiveJobSiteWaitTimeThresholdMinutes, _ := cmd.Flags().GetString("explicit-excessive-job-site-wait-time-threshold-minutes")
	explicitExcessiveMaterialSiteWaitTimeThresholdMinutes, _ := cmd.Flags().GetString("explicit-excessive-material-site-wait-time-threshold-minutes")
	referenceData, _ := cmd.Flags().GetString("reference-data")
	explicitCrewRequirementInboundLatitude, _ := cmd.Flags().GetString("explicit-crew-requirement-inbound-latitude")
	explicitCrewRequirementInboundLongitude, _ := cmd.Flags().GetString("explicit-crew-requirement-inbound-longitude")
	explicitCrewRequirementOutboundLatitude, _ := cmd.Flags().GetString("explicit-crew-requirement-outbound-latitude")
	explicitCrewRequirementOutboundLongitude, _ := cmd.Flags().GetString("explicit-crew-requirement-outbound-longitude")
	benchmarkTonsPerTruckHour, _ := cmd.Flags().GetString("benchmark-tons-per-truck-hour")
	defaultTransactionsStartBeforeSeconds, _ := cmd.Flags().GetString("default-transactions-start-before-seconds")
	defaultTransactionsEndAfterSeconds, _ := cmd.Flags().GetString("default-transactions-end-after-seconds")
	explicitJobSitePhoneNumber, _ := cmd.Flags().GetString("explicit-job-site-phone-number")
	onHoldComment, _ := cmd.Flags().GetString("on-hold-comment")
	notOnHoldComment, _ := cmd.Flags().GetString("not-on-hold-comment")
	// Boolean attributes
	isTemplate, _ := cmd.Flags().GetBool("is-template")
	isOnHold, _ := cmd.Flags().GetBool("is-on-hold")
	isScheduleLocked, _ := cmd.Flags().GetBool("is-schedule-locked")
	isRawJobNumberLocked, _ := cmd.Flags().GetBool("is-raw-job-number-locked")
	isCostCodeRequiredPerSegment, _ := cmd.Flags().GetBool("is-cost-code-required-per-segment")
	isCostCodeAllocationRequiredPerTimeCard, _ := cmd.Flags().GetBool("is-cost-code-allocation-required-per-time-card")
	isCostCodeAllocationRequiredPerTimeSheet, _ := cmd.Flags().GetBool("is-cost-code-allocation-required-per-time-sheet")
	enableRecapNotifications, _ := cmd.Flags().GetBool("enable-recap-notifications")
	createDetectedProductionIncidents, _ := cmd.Flags().GetBool("create-detected-production-incidents")
	approvalRequiresJobNumber, _ := cmd.Flags().GetBool("approval-requires-job-number")
	isMaintainingServiceTypeUnitOfMeasureCohort, _ := cmd.Flags().GetBool("is-maintaining-service-type-unit-of-measure-cohort")
	isManagingCrewRequirements, _ := cmd.Flags().GetBool("is-managing-crew-requirements")
	isMaterialTransactionInspectionEnabled, _ := cmd.Flags().GetBool("is-material-transaction-inspection-enabled")
	isNotifyingCrew, _ := cmd.Flags().GetBool("is-notifying-crew")
	requiresTrucking, _ := cmd.Flags().GetBool("requires-trucking")
	requiresMaterials, _ := cmd.Flags().GetBool("requires-materials")
	lockObservedPossibleCycleMinutes, _ := cmd.Flags().GetBool("lock-observed-possible-cycle-minutes")
	autoCheckInDriverOnArrivalAtStartSite, _ := cmd.Flags().GetBool("auto-check-in-driver-on-arrival-at-start-site")
	requiresDrivingMinutes, _ := cmd.Flags().GetBool("requires-driving-minutes")
	requiresMaterialSiteMinutes, _ := cmd.Flags().GetBool("requires-material-site-minutes")
	explicitNotifyJobProductionPlanTimeCardApprover, _ := cmd.Flags().GetBool("explicit-notify-job-production-plan-time-card-approver")
	explicitEstimatesCostCodesVia, _ := cmd.Flags().GetBool("explicit-estimates-cost-codes-via")
	explicitIsDriverExpectingMaterialTransactionInspection, _ := cmd.Flags().GetBool("explicit-is-driver-expecting-material-transaction-inspection")
	explicitRequiresBusinessUnit, _ := cmd.Flags().GetBool("explicit-requires-business-unit")
	explicitNotifyDriverWhenGPSNotAvailable, _ := cmd.Flags().GetBool("explicit-notify-driver-when-gps-not-available")
	explicitNotifyWhenAllPlanTimeSheetsSubmitted, _ := cmd.Flags().GetBool("explicit-notify-when-all-plan-time-sheets-submitted")
	explicitNotifyWhenAllPlanTimeSheetsApproved, _ := cmd.Flags().GetBool("explicit-notify-when-all-plan-time-sheets-approved")
	explicitNotifyWhenPlanScheduleChanges, _ := cmd.Flags().GetBool("explicit-notify-when-plan-schedule-changes")
	explicitNotifyOnExcessiveSiteWaitTime, _ := cmd.Flags().GetBool("explicit-notify-on-excessive-site-wait-time")
	explicitAutoApproveAutoTimeCardsWithNonMaterialQuantities, _ := cmd.Flags().GetBool("explicit-auto-approve-auto-time-cards-with-non-material-quantities")
	explicitRequireAdminApprovalTimeCardAttachments, _ := cmd.Flags().GetBool("explicit-require-admin-approval-time-card-attachments")
	explicitIsTimeCardExpectingMtxnsAcceptedBeforeApproval, _ := cmd.Flags().GetBool("explicit-is-time-card-expecting-mtxns-accepted-before-approval")
	explicitIsUpdatingChecksumRangeFromSegments, _ := cmd.Flags().GetBool("explicit-is-updating-checksum-range-from-segments")
	enableImplicitTimeCardApproval, _ := cmd.Flags().GetBool("enable-implicit-time-card-approval")
	isUsingVolumetricMeasurements, _ := cmd.Flags().GetBool("is-using-volumetric-measurements")
	explicitIsAuditingTimeCardApprovals, _ := cmd.Flags().GetBool("explicit-is-auditing-time-card-approvals")
	explicitSubmissionRequiresDifferentJobSite, _ := cmd.Flags().GetBool("explicit-submission-requires-different-job-site")
	isPrevailingWageExplicit, _ := cmd.Flags().GetString("is-prevailing-wage-explicit")
	isCertificationRequiredExplicit, _ := cmd.Flags().GetString("is-certification-required-explicit")
	isManagingMaterialSiteStartTimingExplicit, _ := cmd.Flags().GetString("is-managing-material-site-start-timing-explicit")
	isTimeCardPayrollCertificationRequiredExplicit, _ := cmd.Flags().GetString("is-time-card-payroll-certification-required-explicit")
	isOneWayJobExplicit, _ := cmd.Flags().GetString("is-one-way-job-explicit")
	isExpectingSafetyMeeting, _ := cmd.Flags().GetBool("is-expecting-safety-meeting")
	allowsUnmanagedShift, _ := cmd.Flags().GetBool("allows-unmanaged-shift")
	isJobSiteMaterialSiteMaterialTransactionSource, _ := cmd.Flags().GetBool("is-job-site-material-site-material-transaction-source")
	showLoadoutPositionToDrivers, _ := cmd.Flags().GetBool("show-loadout-position-to-drivers")
	isExpectingDriverFieldApprovalTimeCard, _ := cmd.Flags().GetBool("is-expecting-driver-field-approval-time-card")
	areShiftsExpectingTimeCards, _ := cmd.Flags().GetBool("are-shifts-expecting-time-cards")
	explicitRequiresInspector, _ := cmd.Flags().GetBool("explicit-requires-inspector")
	explicitRequiresCertifiedWeigher, _ := cmd.Flags().GetBool("explicit-requires-certified-weigher")
	explicitRequiresProject, _ := cmd.Flags().GetBool("explicit-requires-project")
	explicitIsMaterialTypeDefaultCostCodeRequired, _ := cmd.Flags().GetBool("explicit-is-material-type-default-cost-code-required")
	explicitIsValidatingProjectCostCodes, _ := cmd.Flags().GetBool("explicit-is-validating-project-cost-codes")
	explicitAutomaticallyCreateProjectPhaseCostItems, _ := cmd.Flags().GetBool("explicit-automatically-create-project-phase-cost-items")
	explicitPlanDisallowsMtxnsImplicitMixDesignMatch, _ := cmd.Flags().GetBool("explicit-plan-disallows-mtxns-implicit-mix-design-match")
	areGoalsSyncedFromSegments, _ := cmd.Flags().GetBool("are-goals-synced-from-segments")
	isValidatingProjectMaterialTypes, _ := cmd.Flags().GetBool("is-validating-project-material-types")
	isManagingJobSiteTimesExplicit, _ := cmd.Flags().GetBool("is-managing-job-site-times-explicit")
	isJobSiteTimesCreationAutomatedExplicit, _ := cmd.Flags().GetBool("is-job-site-times-creation-automated-explicit")
	isTruckerIncidentCreationAutomatedExplicit, _ := cmd.Flags().GetBool("is-trucker-incident-creation-automated-explicit")
	// Relationships
	customer, _ := cmd.Flags().GetString("customer")
	jobSite, _ := cmd.Flags().GetString("job-site")
	businessUnit, _ := cmd.Flags().GetString("business-unit")
	planner, _ := cmd.Flags().GetString("planner")
	projectManager, _ := cmd.Flags().GetString("project-manager")
	project, _ := cmd.Flags().GetString("project")
	defaultTrucker, _ := cmd.Flags().GetString("default-trucker")
	contractor, _ := cmd.Flags().GetString("contractor")
	developer, _ := cmd.Flags().GetString("developer")
	unitOfMeasure, _ := cmd.Flags().GetString("unit-of-measure")
	template, _ := cmd.Flags().GetString("template")
	explicitDispatchUser, _ := cmd.Flags().GetString("explicit-dispatch-user")
	certifiedWeigher, _ := cmd.Flags().GetString("certified-weigher")
	safetyMeeting, _ := cmd.Flags().GetString("safety-meeting")
	equipmentMovementTrip, _ := cmd.Flags().GetString("equipment-movement-trip")
	trailerClassifications, _ := cmd.Flags().GetString("trailer-classifications")
	costCodes, _ := cmd.Flags().GetString("cost-codes")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlansUpdateOptions{
		BaseURL:                  baseURL,
		Token:                    token,
		JSON:                     jsonOut,
		ID:                       args[0],
		JobNumber:                jobNumber,
		JobName:                  jobName,
		PhaseName:                phaseName,
		RawJobNumber:             rawJobNumber,
		StartOn:                  startOn,
		StartTime:                startTime,
		EndTime:                  endTime,
		MaterialSiteStartOn:      materialSiteStartOn,
		MaterialSiteStartTime:    materialSiteStartTime,
		ExplicitJobSiteStartOn:   explicitJobSiteStartOn,
		ExplicitJobSiteStartTime: explicitJobSiteStartTime,
		Notes:                    notes,
		GoalHours:                goalHours,
		GoalQuantity:             goalQuantity,
		RemainingQuantity:        remainingQuantity,
		DispatchInstructions:     dispatchInstructions,
		TemplateName:             templateName,
		TemplateStartOnMin:       templateStartOnMin,
		TemplateStartOnMax:       templateStartOnMax,
		ExplicitLoadedMiles:      explicitLoadedMiles,
		ExplicitPlanValidMaterialTransactionUnitOfMeasures:        explicitPlanValidMaterialTransactionUnitOfMeasures,
		DefaultMaterialTransactionTonsMax:                         defaultMaterialTransactionTonsMax,
		CostPerTruckHour:                                          costPerTruckHour,
		CostPerCrewHour:                                           costPerCrewHour,
		DefaultTimeCardApprovalProcess:                            defaultTimeCardApprovalProcess,
		ParallelProductionCount:                                   parallelProductionCount,
		PlannedNonProductionTruckCount:                            plannedNonProductionTruckCount,
		ExplicitTimeZoneID:                                        explicitTimeZoneID,
		ExplicitColorHex:                                          explicitColorHex,
		ExplicitJobSiteProximityMeters:                            explicitJobSiteProximityMeters,
		ExplicitMaterialSiteProximityMeters:                       explicitMaterialSiteProximityMeters,
		DefaultCrewRequirementStartAtOffsetMinutes:                defaultCrewRequirementStartAtOffsetMinutes,
		ObservedPossibleCycleMinutes:                              observedPossibleCycleMinutes,
		ExplicitDriverDayMobilizationBeforeMinutes:                explicitDriverDayMobilizationBeforeMinutes,
		ExplicitExcessiveJobSiteWaitTimeThresholdMinutes:          explicitExcessiveJobSiteWaitTimeThresholdMinutes,
		ExplicitExcessiveMaterialSiteWaitTimeThresholdMinutes:     explicitExcessiveMaterialSiteWaitTimeThresholdMinutes,
		ReferenceData:                                             referenceData,
		ExplicitCrewRequirementInboundLatitude:                    explicitCrewRequirementInboundLatitude,
		ExplicitCrewRequirementInboundLongitude:                   explicitCrewRequirementInboundLongitude,
		ExplicitCrewRequirementOutboundLatitude:                   explicitCrewRequirementOutboundLatitude,
		ExplicitCrewRequirementOutboundLongitude:                  explicitCrewRequirementOutboundLongitude,
		BenchmarkTonsPerTruckHour:                                 benchmarkTonsPerTruckHour,
		DefaultTransactionsStartBeforeSeconds:                     defaultTransactionsStartBeforeSeconds,
		DefaultTransactionsEndAfterSeconds:                        defaultTransactionsEndAfterSeconds,
		ExplicitJobSitePhoneNumber:                                explicitJobSitePhoneNumber,
		OnHoldComment:                                             onHoldComment,
		NotOnHoldComment:                                          notOnHoldComment,
		IsTemplate:                                                isTemplate,
		IsOnHold:                                                  isOnHold,
		IsScheduleLocked:                                          isScheduleLocked,
		IsRawJobNumberLocked:                                      isRawJobNumberLocked,
		IsCostCodeRequiredPerSegment:                              isCostCodeRequiredPerSegment,
		IsCostCodeAllocationRequiredPerTimeCard:                   isCostCodeAllocationRequiredPerTimeCard,
		IsCostCodeAllocationRequiredPerTimeSheet:                  isCostCodeAllocationRequiredPerTimeSheet,
		EnableRecapNotifications:                                  enableRecapNotifications,
		CreateDetectedProductionIncidents:                         createDetectedProductionIncidents,
		ApprovalRequiresJobNumber:                                 approvalRequiresJobNumber,
		IsMaintainingServiceTypeUnitOfMeasureCohort:               isMaintainingServiceTypeUnitOfMeasureCohort,
		IsManagingCrewRequirements:                                isManagingCrewRequirements,
		IsMaterialTransactionInspectionEnabled:                    isMaterialTransactionInspectionEnabled,
		IsNotifyingCrew:                                           isNotifyingCrew,
		RequiresTrucking:                                          requiresTrucking,
		RequiresMaterials:                                         requiresMaterials,
		LockObservedPossibleCycleMinutes:                          lockObservedPossibleCycleMinutes,
		AutoCheckInDriverOnArrivalAtStartSite:                     autoCheckInDriverOnArrivalAtStartSite,
		RequiresDrivingMinutes:                                    requiresDrivingMinutes,
		RequiresMaterialSiteMinutes:                               requiresMaterialSiteMinutes,
		ExplicitNotifyJobProductionPlanTimeCardApprover:           explicitNotifyJobProductionPlanTimeCardApprover,
		ExplicitEstimatesCostCodesVia:                             explicitEstimatesCostCodesVia,
		ExplicitIsDriverExpectingMaterialTransactionInspection:    explicitIsDriverExpectingMaterialTransactionInspection,
		ExplicitRequiresBusinessUnit:                              explicitRequiresBusinessUnit,
		ExplicitNotifyDriverWhenGPSNotAvailable:                   explicitNotifyDriverWhenGPSNotAvailable,
		ExplicitNotifyWhenAllPlanTimeSheetsSubmitted:              explicitNotifyWhenAllPlanTimeSheetsSubmitted,
		ExplicitNotifyWhenAllPlanTimeSheetsApproved:               explicitNotifyWhenAllPlanTimeSheetsApproved,
		ExplicitNotifyWhenPlanScheduleChanges:                     explicitNotifyWhenPlanScheduleChanges,
		ExplicitNotifyOnExcessiveSiteWaitTime:                     explicitNotifyOnExcessiveSiteWaitTime,
		ExplicitAutoApproveAutoTimeCardsWithNonMaterialQuantities: explicitAutoApproveAutoTimeCardsWithNonMaterialQuantities,
		ExplicitRequireAdminApprovalTimeCardAttachments:           explicitRequireAdminApprovalTimeCardAttachments,
		ExplicitIsTimeCardExpectingMtxnsAcceptedBeforeApproval:    explicitIsTimeCardExpectingMtxnsAcceptedBeforeApproval,
		ExplicitIsUpdatingChecksumRangeFromSegments:               explicitIsUpdatingChecksumRangeFromSegments,
		EnableImplicitTimeCardApproval:                            enableImplicitTimeCardApproval,
		IsUsingVolumetricMeasurements:                             isUsingVolumetricMeasurements,
		ExplicitIsAuditingTimeCardApprovals:                       explicitIsAuditingTimeCardApprovals,
		ExplicitSubmissionRequiresDifferentJobSite:                explicitSubmissionRequiresDifferentJobSite,
		IsPrevailingWageExplicit:                                  isPrevailingWageExplicit,
		IsCertificationRequiredExplicit:                           isCertificationRequiredExplicit,
		IsManagingMaterialSiteStartTimingExplicit:                 isManagingMaterialSiteStartTimingExplicit,
		IsTimeCardPayrollCertificationRequiredExplicit:            isTimeCardPayrollCertificationRequiredExplicit,
		IsOneWayJobExplicit:                                       isOneWayJobExplicit,
		IsExpectingSafetyMeeting:                                  isExpectingSafetyMeeting,
		AllowsUnmanagedShift:                                      allowsUnmanagedShift,
		IsJobSiteMaterialSiteMaterialTransactionSource:            isJobSiteMaterialSiteMaterialTransactionSource,
		ShowLoadoutPositionToDrivers:                              showLoadoutPositionToDrivers,
		IsExpectingDriverFieldApprovalTimeCard:                    isExpectingDriverFieldApprovalTimeCard,
		AreShiftsExpectingTimeCards:                               areShiftsExpectingTimeCards,
		ExplicitRequiresInspector:                                 explicitRequiresInspector,
		ExplicitRequiresCertifiedWeigher:                          explicitRequiresCertifiedWeigher,
		ExplicitRequiresProject:                                   explicitRequiresProject,
		ExplicitIsMaterialTypeDefaultCostCodeRequired:             explicitIsMaterialTypeDefaultCostCodeRequired,
		ExplicitIsValidatingProjectCostCodes:                      explicitIsValidatingProjectCostCodes,
		ExplicitAutomaticallyCreateProjectPhaseCostItems:          explicitAutomaticallyCreateProjectPhaseCostItems,
		ExplicitPlanDisallowsMtxnsImplicitMixDesignMatch:          explicitPlanDisallowsMtxnsImplicitMixDesignMatch,
		AreGoalsSyncedFromSegments:                                areGoalsSyncedFromSegments,
		IsValidatingProjectMaterialTypes:                          isValidatingProjectMaterialTypes,
		IsManagingJobSiteTimesExplicit:                            isManagingJobSiteTimesExplicit,
		IsJobSiteTimesCreationAutomatedExplicit:                   isJobSiteTimesCreationAutomatedExplicit,
		IsTruckerIncidentCreationAutomatedExplicit:                isTruckerIncidentCreationAutomatedExplicit,
		Customer:               customer,
		JobSite:                jobSite,
		BusinessUnit:           businessUnit,
		Planner:                planner,
		ProjectManager:         projectManager,
		Project:                project,
		DefaultTrucker:         defaultTrucker,
		Contractor:             contractor,
		Developer:              developer,
		UnitOfMeasure:          unitOfMeasure,
		Template:               template,
		ExplicitDispatchUser:   explicitDispatchUser,
		CertifiedWeigher:       certifiedWeigher,
		SafetyMeeting:          safetyMeeting,
		EquipmentMovementTrip:  equipmentMovementTrip,
		TrailerClassifications: trailerClassifications,
		CostCodes:              costCodes,
	}, nil
}
