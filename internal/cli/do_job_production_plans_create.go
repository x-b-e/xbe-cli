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

type doJobProductionPlansCreateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
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
	// Boolean attributes
	IsTemplate                                                bool
	IsOnHold                                                  bool
	OnHoldComment                                             string
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
	IsLord                                                    bool
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

func newDoJobProductionPlansCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new job production plan",
		Long: `Create a new job production plan.

Key attributes:
  --job-number              Job number identifier
  --job-name                Job name/description
  --start-on                Start date (YYYY-MM-DD)
  --start-time              Start time (HH:MM)
  --end-time                End time (HH:MM)
  --material-site-start-on  Material site start date
  --material-site-start-time Material site start time
  --notes                   Notes about the plan
  --goal-hours              Goal hours
  --goal-quantity           Goal quantity (e.g., tons)
  --remaining-quantity      Remaining quantity
  --dispatch-instructions   Instructions for dispatch
  --phase-name              Phase name
  --is-template             Mark as template
  --template-name           Template name (for templates)
  --is-on-hold              Put plan on hold
  --on-hold-comment         Comment for hold status

Relationships:
  --customer                Customer ID
  --job-site                Job site ID
  --business-unit           Business unit ID
  --planner                 Planner user ID
  --project-manager         Project manager user ID
  --project                 Project ID
  --default-trucker         Default trucker ID
  --contractor              Contractor ID
  --developer               Developer ID
  --unit-of-measure         Unit of measure ID
  --template                Template job production plan ID
  --trailer-classifications Trailer classification IDs (comma-separated)
  --cost-codes              Cost code IDs (comma-separated)`,
		Example: `  # Create a basic job production plan
  xbe do job-production-plans create --job-name "Main Street Paving" --start-on 2025-01-20 --start-time 06:00 --customer 123

  # Create a plan with goals
  xbe do job-production-plans create --job-name "Highway Project" --start-on 2025-01-20 --goal-quantity 500 --customer 123

  # Create a template
  xbe do job-production-plans create --job-name "Standard Job" --is-template --template-name "Standard Template"

  # Create with cost tracking
  xbe do job-production-plans create --job-name "Project X" --start-on 2025-01-20 --customer 123 --cost-per-truck-hour 150 --cost-per-crew-hour 75`,
		RunE: runDoJobProductionPlansCreate,
	}
	initDoJobProductionPlansCreateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlansCmd.AddCommand(newDoJobProductionPlansCreateCmd())
}

func initDoJobProductionPlansCreateFlags(cmd *cobra.Command) {
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
	cmd.Flags().Bool("is-lord", false, "Is lord (create only)")
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

func runDoJobProductionPlansCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoJobProductionPlansCreateOptions(cmd)
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

	// String attributes
	if opts.JobNumber != "" {
		attributes["job-number"] = opts.JobNumber
	}
	if opts.JobName != "" {
		attributes["job-name"] = opts.JobName
	}
	if opts.PhaseName != "" {
		attributes["phase-name"] = opts.PhaseName
	}
	if opts.RawJobNumber != "" {
		attributes["raw-job-number"] = opts.RawJobNumber
	}
	if opts.StartOn != "" {
		attributes["start-on"] = opts.StartOn
	}
	if opts.StartTime != "" {
		attributes["start-time"] = opts.StartTime
	}
	if opts.EndTime != "" {
		attributes["end-time"] = opts.EndTime
	}
	if opts.MaterialSiteStartOn != "" {
		attributes["material-site-start-on"] = opts.MaterialSiteStartOn
	}
	if opts.MaterialSiteStartTime != "" {
		attributes["material-site-start-time"] = opts.MaterialSiteStartTime
	}
	if opts.ExplicitJobSiteStartOn != "" {
		attributes["explicit-job-site-start-on"] = opts.ExplicitJobSiteStartOn
	}
	if opts.ExplicitJobSiteStartTime != "" {
		attributes["explicit-job-site-start-time"] = opts.ExplicitJobSiteStartTime
	}
	if opts.Notes != "" {
		attributes["notes"] = opts.Notes
	}
	if opts.GoalHours != "" {
		attributes["goal-hours"] = opts.GoalHours
	}
	if opts.GoalQuantity != "" {
		attributes["goal-quantity"] = opts.GoalQuantity
	}
	if opts.RemainingQuantity != "" {
		attributes["remaining-quantity"] = opts.RemainingQuantity
	}
	if opts.DispatchInstructions != "" {
		attributes["dispatch-instructions"] = opts.DispatchInstructions
	}
	if opts.TemplateName != "" {
		attributes["template-name"] = opts.TemplateName
	}
	if opts.TemplateStartOnMin != "" {
		attributes["template-start-on-min"] = opts.TemplateStartOnMin
	}
	if opts.TemplateStartOnMax != "" {
		attributes["template-start-on-max"] = opts.TemplateStartOnMax
	}
	if opts.ExplicitLoadedMiles != "" {
		attributes["explicit-loaded-miles"] = opts.ExplicitLoadedMiles
	}
	if opts.ExplicitPlanValidMaterialTransactionUnitOfMeasures != "" {
		attributes["explicit-plan-valid-material-transaction-unit-of-measures"] = opts.ExplicitPlanValidMaterialTransactionUnitOfMeasures
	}
	if opts.DefaultMaterialTransactionTonsMax != "" {
		attributes["default-material-transaction-tons-max"] = opts.DefaultMaterialTransactionTonsMax
	}
	if opts.CostPerTruckHour != "" {
		attributes["cost-per-truck-hour"] = opts.CostPerTruckHour
	}
	if opts.CostPerCrewHour != "" {
		attributes["cost-per-crew-hour"] = opts.CostPerCrewHour
	}
	if opts.DefaultTimeCardApprovalProcess != "" {
		attributes["default-time-card-approval-process"] = opts.DefaultTimeCardApprovalProcess
	}
	if opts.ParallelProductionCount != "" {
		attributes["parallel-production-count"] = opts.ParallelProductionCount
	}
	if opts.PlannedNonProductionTruckCount != "" {
		attributes["planned-non-production-truck-count"] = opts.PlannedNonProductionTruckCount
	}
	if opts.ExplicitTimeZoneID != "" {
		attributes["explicit-time-zone-id"] = opts.ExplicitTimeZoneID
	}
	if opts.ExplicitColorHex != "" {
		attributes["explicit-color-hex"] = opts.ExplicitColorHex
	}
	if opts.ExplicitJobSiteProximityMeters != "" {
		attributes["explicit-job-site-proximity-meters"] = opts.ExplicitJobSiteProximityMeters
	}
	if opts.ExplicitMaterialSiteProximityMeters != "" {
		attributes["explicit-material-site-proximity-meters"] = opts.ExplicitMaterialSiteProximityMeters
	}
	if opts.DefaultCrewRequirementStartAtOffsetMinutes != "" {
		attributes["default-crew-requirement-start-at-offset-minutes"] = opts.DefaultCrewRequirementStartAtOffsetMinutes
	}
	if opts.ObservedPossibleCycleMinutes != "" {
		attributes["observed-possible-cycle-minutes"] = opts.ObservedPossibleCycleMinutes
	}
	if opts.ExplicitDriverDayMobilizationBeforeMinutes != "" {
		attributes["explicit-driver-day-mobilization-before-minutes"] = opts.ExplicitDriverDayMobilizationBeforeMinutes
	}
	if opts.ExplicitExcessiveJobSiteWaitTimeThresholdMinutes != "" {
		attributes["explicit-excessive-job-site-wait-time-threshold-minutes"] = opts.ExplicitExcessiveJobSiteWaitTimeThresholdMinutes
	}
	if opts.ExplicitExcessiveMaterialSiteWaitTimeThresholdMinutes != "" {
		attributes["explicit-excessive-material-site-wait-time-threshold-minutes"] = opts.ExplicitExcessiveMaterialSiteWaitTimeThresholdMinutes
	}
	if opts.ReferenceData != "" {
		// Parse as JSON object
		var refData map[string]any
		if err := json.Unmarshal([]byte(opts.ReferenceData), &refData); err != nil {
			err = fmt.Errorf("invalid JSON for reference-data: %w", err)
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["reference-data"] = refData
	}
	if opts.ExplicitCrewRequirementInboundLatitude != "" {
		attributes["explicit-crew-requirement-inbound-latitude"] = opts.ExplicitCrewRequirementInboundLatitude
	}
	if opts.ExplicitCrewRequirementInboundLongitude != "" {
		attributes["explicit-crew-requirement-inbound-longitude"] = opts.ExplicitCrewRequirementInboundLongitude
	}
	if opts.ExplicitCrewRequirementOutboundLatitude != "" {
		attributes["explicit-crew-requirement-outbound-latitude"] = opts.ExplicitCrewRequirementOutboundLatitude
	}
	if opts.ExplicitCrewRequirementOutboundLongitude != "" {
		attributes["explicit-crew-requirement-outbound-longitude"] = opts.ExplicitCrewRequirementOutboundLongitude
	}
	if opts.BenchmarkTonsPerTruckHour != "" {
		attributes["benchmark-tons-per-truck-hour"] = opts.BenchmarkTonsPerTruckHour
	}
	if opts.DefaultTransactionsStartBeforeSeconds != "" {
		attributes["default-transactions-start-before-seconds"] = opts.DefaultTransactionsStartBeforeSeconds
	}
	if opts.DefaultTransactionsEndAfterSeconds != "" {
		attributes["default-transactions-end-after-seconds"] = opts.DefaultTransactionsEndAfterSeconds
	}
	if opts.ExplicitJobSitePhoneNumber != "" {
		attributes["explicit-job-site-phone-number"] = opts.ExplicitJobSitePhoneNumber
	}
	if opts.OnHoldComment != "" {
		attributes["on-hold-comment"] = opts.OnHoldComment
	}

	// Boolean attributes
	if cmd.Flags().Changed("is-template") {
		attributes["is-template"] = opts.IsTemplate
	}
	if cmd.Flags().Changed("is-on-hold") {
		attributes["is-on-hold"] = opts.IsOnHold
	}
	if cmd.Flags().Changed("is-schedule-locked") {
		attributes["is-schedule-locked"] = opts.IsScheduleLocked
	}
	if cmd.Flags().Changed("is-raw-job-number-locked") {
		attributes["is-raw-job-number-locked"] = opts.IsRawJobNumberLocked
	}
	if cmd.Flags().Changed("is-cost-code-required-per-segment") {
		attributes["is-cost-code-required-per-segment"] = opts.IsCostCodeRequiredPerSegment
	}
	if cmd.Flags().Changed("is-cost-code-allocation-required-per-time-card") {
		attributes["is-cost-code-allocation-required-per-time-card"] = opts.IsCostCodeAllocationRequiredPerTimeCard
	}
	if cmd.Flags().Changed("is-cost-code-allocation-required-per-time-sheet") {
		attributes["is-cost-code-allocation-required-per-time-sheet"] = opts.IsCostCodeAllocationRequiredPerTimeSheet
	}
	if cmd.Flags().Changed("enable-recap-notifications") {
		attributes["enable-recap-notifications"] = opts.EnableRecapNotifications
	}
	if cmd.Flags().Changed("create-detected-production-incidents") {
		attributes["create-detected-production-incidents"] = opts.CreateDetectedProductionIncidents
	}
	if cmd.Flags().Changed("approval-requires-job-number") {
		attributes["approval-requires-job-number"] = opts.ApprovalRequiresJobNumber
	}
	if cmd.Flags().Changed("is-maintaining-service-type-unit-of-measure-cohort") {
		attributes["is-maintaining-service-type-unit-of-measure-cohort"] = opts.IsMaintainingServiceTypeUnitOfMeasureCohort
	}
	if cmd.Flags().Changed("is-managing-crew-requirements") {
		attributes["is-managing-crew-requirements"] = opts.IsManagingCrewRequirements
	}
	if cmd.Flags().Changed("is-material-transaction-inspection-enabled") {
		attributes["is-material-transaction-inspection-enabled"] = opts.IsMaterialTransactionInspectionEnabled
	}
	if cmd.Flags().Changed("is-notifying-crew") {
		attributes["is-notifying-crew"] = opts.IsNotifyingCrew
	}
	if cmd.Flags().Changed("requires-trucking") {
		attributes["requires-trucking"] = opts.RequiresTrucking
	}
	if cmd.Flags().Changed("requires-materials") {
		attributes["requires-materials"] = opts.RequiresMaterials
	}
	if cmd.Flags().Changed("lock-observed-possible-cycle-minutes") {
		attributes["lock-observed-possible-cycle-minutes"] = opts.LockObservedPossibleCycleMinutes
	}
	if cmd.Flags().Changed("auto-check-in-driver-on-arrival-at-start-site") {
		attributes["auto-check-in-driver-on-arrival-at-start-site"] = opts.AutoCheckInDriverOnArrivalAtStartSite
	}
	if cmd.Flags().Changed("requires-driving-minutes") {
		attributes["requires-driving-minutes"] = opts.RequiresDrivingMinutes
	}
	if cmd.Flags().Changed("requires-material-site-minutes") {
		attributes["requires-material-site-minutes"] = opts.RequiresMaterialSiteMinutes
	}
	if cmd.Flags().Changed("explicit-notify-job-production-plan-time-card-approver") {
		attributes["explicit-notify-job-production-plan-time-card-approver"] = opts.ExplicitNotifyJobProductionPlanTimeCardApprover
	}
	if cmd.Flags().Changed("explicit-estimates-cost-codes-via") {
		attributes["explicit-estimates-cost-codes-via"] = opts.ExplicitEstimatesCostCodesVia
	}
	if cmd.Flags().Changed("explicit-is-driver-expecting-material-transaction-inspection") {
		attributes["explicit-is-driver-expecting-material-transaction-inspection"] = opts.ExplicitIsDriverExpectingMaterialTransactionInspection
	}
	if cmd.Flags().Changed("explicit-requires-business-unit") {
		attributes["explicit-requires-business-unit"] = opts.ExplicitRequiresBusinessUnit
	}
	if cmd.Flags().Changed("explicit-notify-driver-when-gps-not-available") {
		attributes["explicit-notify-driver-when-gps-not-available"] = opts.ExplicitNotifyDriverWhenGPSNotAvailable
	}
	if cmd.Flags().Changed("explicit-notify-when-all-plan-time-sheets-submitted") {
		attributes["explicit-notify-when-all-plan-time-sheets-submitted"] = opts.ExplicitNotifyWhenAllPlanTimeSheetsSubmitted
	}
	if cmd.Flags().Changed("explicit-notify-when-all-plan-time-sheets-approved") {
		attributes["explicit-notify-when-all-plan-time-sheets-approved"] = opts.ExplicitNotifyWhenAllPlanTimeSheetsApproved
	}
	if cmd.Flags().Changed("explicit-notify-when-plan-schedule-changes") {
		attributes["explicit-notify-when-plan-schedule-changes"] = opts.ExplicitNotifyWhenPlanScheduleChanges
	}
	if cmd.Flags().Changed("explicit-notify-on-excessive-site-wait-time") {
		attributes["explicit-notify-on-excessive-site-wait-time"] = opts.ExplicitNotifyOnExcessiveSiteWaitTime
	}
	if cmd.Flags().Changed("explicit-auto-approve-auto-time-cards-with-non-material-quantities") {
		attributes["explicit-auto-approve-auto-time-cards-with-non-material-quantities"] = opts.ExplicitAutoApproveAutoTimeCardsWithNonMaterialQuantities
	}
	if cmd.Flags().Changed("explicit-require-admin-approval-time-card-attachments") {
		attributes["explicit-require-admin-approval-time-card-attachments"] = opts.ExplicitRequireAdminApprovalTimeCardAttachments
	}
	if cmd.Flags().Changed("explicit-is-time-card-expecting-mtxns-accepted-before-approval") {
		attributes["explicit-is-time-card-expecting-mtxns-accepted-before-approval"] = opts.ExplicitIsTimeCardExpectingMtxnsAcceptedBeforeApproval
	}
	if cmd.Flags().Changed("explicit-is-updating-checksum-range-from-segments") {
		attributes["explicit-is-updating-checksum-range-from-segments"] = opts.ExplicitIsUpdatingChecksumRangeFromSegments
	}
	if cmd.Flags().Changed("enable-implicit-time-card-approval") {
		attributes["enable-implicit-time-card-approval"] = opts.EnableImplicitTimeCardApproval
	}
	if cmd.Flags().Changed("is-using-volumetric-measurements") {
		attributes["is-using-volumetric-measurements"] = opts.IsUsingVolumetricMeasurements
	}
	if cmd.Flags().Changed("explicit-is-auditing-time-card-approvals") {
		attributes["explicit-is-auditing-time-card-approvals"] = opts.ExplicitIsAuditingTimeCardApprovals
	}
	if cmd.Flags().Changed("explicit-submission-requires-different-job-site") {
		attributes["explicit-submission-requires-different-job-site"] = opts.ExplicitSubmissionRequiresDifferentJobSite
	}
	// Tri-state booleans (true/false/null)
	if cmd.Flags().Changed("is-prevailing-wage-explicit") {
		if opts.IsPrevailingWageExplicit == "null" {
			attributes["is-prevailing-wage-explicit"] = nil
		} else {
			attributes["is-prevailing-wage-explicit"] = opts.IsPrevailingWageExplicit == "true"
		}
	}
	if cmd.Flags().Changed("is-certification-required-explicit") {
		if opts.IsCertificationRequiredExplicit == "null" {
			attributes["is-certification-required-explicit"] = nil
		} else {
			attributes["is-certification-required-explicit"] = opts.IsCertificationRequiredExplicit == "true"
		}
	}
	if cmd.Flags().Changed("is-managing-material-site-start-timing-explicit") {
		if opts.IsManagingMaterialSiteStartTimingExplicit == "null" {
			attributes["is-managing-material-site-start-timing-explicit"] = nil
		} else {
			attributes["is-managing-material-site-start-timing-explicit"] = opts.IsManagingMaterialSiteStartTimingExplicit == "true"
		}
	}
	if cmd.Flags().Changed("is-time-card-payroll-certification-required-explicit") {
		if opts.IsTimeCardPayrollCertificationRequiredExplicit == "null" {
			attributes["is-time-card-payroll-certification-required-explicit"] = nil
		} else {
			attributes["is-time-card-payroll-certification-required-explicit"] = opts.IsTimeCardPayrollCertificationRequiredExplicit == "true"
		}
	}
	if cmd.Flags().Changed("is-one-way-job-explicit") {
		if opts.IsOneWayJobExplicit == "null" {
			attributes["is-one-way-job-explicit"] = nil
		} else {
			attributes["is-one-way-job-explicit"] = opts.IsOneWayJobExplicit == "true"
		}
	}
	if cmd.Flags().Changed("is-expecting-safety-meeting") {
		attributes["is-expecting-safety-meeting"] = opts.IsExpectingSafetyMeeting
	}
	if cmd.Flags().Changed("allows-unmanaged-shift") {
		attributes["allows-unmanaged-shift"] = opts.AllowsUnmanagedShift
	}
	if cmd.Flags().Changed("is-job-site-material-site-material-transaction-source") {
		attributes["is-job-site-material-site-material-transaction-source"] = opts.IsJobSiteMaterialSiteMaterialTransactionSource
	}
	if cmd.Flags().Changed("show-loadout-position-to-drivers") {
		attributes["show-loadout-position-to-drivers"] = opts.ShowLoadoutPositionToDrivers
	}
	if cmd.Flags().Changed("is-expecting-driver-field-approval-time-card") {
		attributes["is-expecting-driver-field-approval-time-card"] = opts.IsExpectingDriverFieldApprovalTimeCard
	}
	if cmd.Flags().Changed("are-shifts-expecting-time-cards") {
		attributes["are-shifts-expecting-time-cards"] = opts.AreShiftsExpectingTimeCards
	}
	if cmd.Flags().Changed("explicit-requires-inspector") {
		attributes["explicit-requires-inspector"] = opts.ExplicitRequiresInspector
	}
	if cmd.Flags().Changed("explicit-requires-certified-weigher") {
		attributes["explicit-requires-certified-weigher"] = opts.ExplicitRequiresCertifiedWeigher
	}
	if cmd.Flags().Changed("explicit-requires-project") {
		attributes["explicit-requires-project"] = opts.ExplicitRequiresProject
	}
	if cmd.Flags().Changed("explicit-is-material-type-default-cost-code-required") {
		attributes["explicit-is-material-type-default-cost-code-required"] = opts.ExplicitIsMaterialTypeDefaultCostCodeRequired
	}
	if cmd.Flags().Changed("explicit-is-validating-project-cost-codes") {
		attributes["explicit-is-validating-project-cost-codes"] = opts.ExplicitIsValidatingProjectCostCodes
	}
	if cmd.Flags().Changed("explicit-automatically-create-project-phase-cost-items") {
		attributes["explicit-automatically-create-project-phase-cost-items"] = opts.ExplicitAutomaticallyCreateProjectPhaseCostItems
	}
	if cmd.Flags().Changed("explicit-plan-disallows-mtxns-implicit-mix-design-match") {
		attributes["explicit-plan-disallows-mtxns-implicit-mix-design-match"] = opts.ExplicitPlanDisallowsMtxnsImplicitMixDesignMatch
	}
	if cmd.Flags().Changed("are-goals-synced-from-segments") {
		attributes["are-goals-synced-from-segments"] = opts.AreGoalsSyncedFromSegments
	}
	if cmd.Flags().Changed("is-validating-project-material-types") {
		attributes["is-validating-project-material-types"] = opts.IsValidatingProjectMaterialTypes
	}
	if cmd.Flags().Changed("is-managing-job-site-times-explicit") {
		attributes["is-managing-job-site-times-explicit"] = opts.IsManagingJobSiteTimesExplicit
	}
	if cmd.Flags().Changed("is-job-site-times-creation-automated-explicit") {
		attributes["is-job-site-times-creation-automated-explicit"] = opts.IsJobSiteTimesCreationAutomatedExplicit
	}
	if cmd.Flags().Changed("is-trucker-incident-creation-automated-explicit") {
		attributes["is-trucker-incident-creation-automated-explicit"] = opts.IsTruckerIncidentCreationAutomatedExplicit
	}
	if cmd.Flags().Changed("is-lord") {
		attributes["is-lord"] = opts.IsLord
	}

	relationships := map[string]any{}

	if opts.Customer != "" {
		relationships["customer"] = map[string]any{
			"data": map[string]any{
				"type": "customers",
				"id":   opts.Customer,
			},
		}
	}
	if opts.JobSite != "" {
		relationships["job-site"] = map[string]any{
			"data": map[string]any{
				"type": "job-sites",
				"id":   opts.JobSite,
			},
		}
	}
	if opts.BusinessUnit != "" {
		relationships["business-unit"] = map[string]any{
			"data": map[string]any{
				"type": "business-units",
				"id":   opts.BusinessUnit,
			},
		}
	}
	if opts.Planner != "" {
		relationships["planner"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.Planner,
			},
		}
	}
	if opts.ProjectManager != "" {
		relationships["project-manager"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.ProjectManager,
			},
		}
	}
	if opts.Project != "" {
		relationships["project"] = map[string]any{
			"data": map[string]any{
				"type": "projects",
				"id":   opts.Project,
			},
		}
	}
	if opts.DefaultTrucker != "" {
		relationships["default-trucker"] = map[string]any{
			"data": map[string]any{
				"type": "truckers",
				"id":   opts.DefaultTrucker,
			},
		}
	}
	if opts.Contractor != "" {
		relationships["contractor"] = map[string]any{
			"data": map[string]any{
				"type": "contractors",
				"id":   opts.Contractor,
			},
		}
	}
	if opts.Developer != "" {
		relationships["developer"] = map[string]any{
			"data": map[string]any{
				"type": "developers",
				"id":   opts.Developer,
			},
		}
	}
	if opts.UnitOfMeasure != "" {
		relationships["unit-of-measure"] = map[string]any{
			"data": map[string]any{
				"type": "unit-of-measures",
				"id":   opts.UnitOfMeasure,
			},
		}
	}
	if opts.Template != "" {
		relationships["template"] = map[string]any{
			"data": map[string]any{
				"type": "job-production-plans",
				"id":   opts.Template,
			},
		}
	}
	if opts.ExplicitDispatchUser != "" {
		relationships["explicit-dispatch-user"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.ExplicitDispatchUser,
			},
		}
	}
	if opts.CertifiedWeigher != "" {
		relationships["certified-weigher"] = map[string]any{
			"data": map[string]any{
				"type": "developer-certified-weighers",
				"id":   opts.CertifiedWeigher,
			},
		}
	}
	if opts.SafetyMeeting != "" {
		relationships["safety-meeting"] = map[string]any{
			"data": map[string]any{
				"type": "meetings",
				"id":   opts.SafetyMeeting,
			},
		}
	}
	if opts.EquipmentMovementTrip != "" {
		relationships["equipment-movement-trip"] = map[string]any{
			"data": map[string]any{
				"type": "equipment-movement-trips",
				"id":   opts.EquipmentMovementTrip,
			},
		}
	}
	// Has-many relationships
	if opts.TrailerClassifications != "" {
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
	if opts.CostCodes != "" {
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

	data := map[string]any{
		"type":       "job-production-plans",
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

	body, _, err := client.Post(cmd.Context(), "/v1/job-production-plans", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created job production plan %s (%s)\n", row.ID, row.JobName)
	return nil
}

func parseDoJobProductionPlansCreateOptions(cmd *cobra.Command) (doJobProductionPlansCreateOptions, error) {
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
	isLord, _ := cmd.Flags().GetBool("is-lord")
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

	return doJobProductionPlansCreateOptions{
		BaseURL:                  baseURL,
		Token:                    token,
		JSON:                     jsonOut,
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
		IsLord:                 isLord,
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

func jobProductionPlanRowFromSingle(resp jsonAPISingleResponse) jobProductionPlanRow {
	return jobProductionPlanRow{
		ID:        resp.Data.ID,
		Status:    stringAttr(resp.Data.Attributes, "status"),
		JobNumber: stringAttr(resp.Data.Attributes, "job-number"),
		JobName:   stringAttr(resp.Data.Attributes, "job-name"),
		StartOn:   formatDate(stringAttr(resp.Data.Attributes, "start-on")),
		StartTime: formatTime(stringAttr(resp.Data.Attributes, "start-time")),
		GoalTons:  floatAttr(resp.Data.Attributes, "goal-quantity"),
	}
}
