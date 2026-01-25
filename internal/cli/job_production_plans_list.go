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

type jobProductionPlansListOptions struct {
	BaseURL                                    string
	Token                                      string
	JSON                                       bool
	NoAuth                                     bool
	Limit                                      int
	Offset                                     int
	StartOn                                    string
	StartOnMin                                 string
	StartOnMax                                 string
	Status                                     string
	Customer                                   string
	Planner                                    string
	ProjectMgr                                 string
	JobSite                                    string
	MaterialSite                               string
	BusinessUnit                               string
	Q                                          string
	Broker                                     string
	BrokerID                                   string
	Project                                    string
	Trucker                                    string
	JobNumber                                  string
	JobName                                    string
	MaterialType                               string
	MaterialSupplier                           string
	Contractor                                 string
	IsTemplate                                 string
	CreatedBy                                  string
	CostCode                                   string
	StartTimeMin                               string
	StartTimeMax                               string
	RemainingQuantityMin                       string
	RemainingQuantityMax                       string
	DefaultTrucker                             string
	NotCustomer                                string
	TrailerClassificationOrEquiv               string
	IsOnlyForEquipmentMovement                 string
	IsAuditingTimeCardApprovals                string
	PlannedTonsPerProductiveSegmentMin         string
	PlannedTonsPerProductiveSegmentMax         string
	DefaultTimeCardApprovalProcess             string
	IsUsingVolumetricMeasurements              string
	HasSupplyDemandBalanceCannotComputeReasons string
	StartAtMin                                 string
	StartAtMax                                 string
	ActiveOn                                   string
	PracticallyStartOn                         string
	PracticallyStartOnMin                      string
	PracticallyStartOnMax                      string
	ChecksumDifference                         string
	ChecksumDifferenceMin                      string
	ChecksumDifferenceMax                      string
	HasChecksumDifference                      string
	HasManagerAssignment                       string
	UserHasStake                               string
	TemplateName                               string
	TemplateStartOnMin                         string
	TemplateStartOnMax                         string
	Template                                   string
	DuplicationToken                           string
	JobSiteActiveAround                        string
	HasProjectPhaseRevenueItems                string
	HasCrewRequirements                        string
	HasLaborRequirements                       string
	HasEquipmentRequirements                   string
	CouldHaveLaborRequirements                 string
	UltimateMaterialTypes                      string
	MaterialTypeUltimateParentCountMin         string
	MaterialTypeUltimateParentCountMax         string
	HasMaterialTypesWithQCRequirements         string
	WithNonDeletableLineupJPPs                 string
	QSegments                                  string
	ExternalIdentificationValue                string
	ReferenceData                              string
	PracticallyStartOnBetween                  string
}

type jobProductionPlanRow struct {
	ID             string  `json:"id"`
	Status         string  `json:"status"`
	JobNumber      string  `json:"job_number"`
	JobName        string  `json:"job_name"`
	StartOn        string  `json:"start_on"`
	StartTime      string  `json:"start_time"`
	Customer       string  `json:"customer"`
	Planner        string  `json:"planner"`
	ProjectManager string  `json:"project_manager,omitempty"`
	JobSite        string  `json:"job_site"`
	MaterialSite   string  `json:"material_site"`
	GoalTons       float64 `json:"goal_tons"`
	// Additional fields
	BusinessUnit       string  `json:"business_unit,omitempty"`
	ProjectName        string  `json:"project_name,omitempty"`
	MixTypes           string  `json:"mix_types,omitempty"`
	TonsPct            float64 `json:"pct_goal,omitempty"`
	ApprovedSurplusPct float64 `json:"approved_surplus_pct,omitempty"`
	ActualSurplusPct   float64 `json:"actual_surplus_pct,omitempty"`
}

func newJobProductionPlansListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List job production plans",
		Long: `List job production plans with filtering and pagination.

Returns a list of job production plans matching the specified criteria.
Plans are sorted by start date/time.

Date Filtering:
  --start-on        Exact date match (can be comma-separated for multiple dates)
  --start-on-min    Start of date range (inclusive)
  --start-on-max    End of date range (inclusive)

  Either --start-on or --start-on-min is required. Dates use YYYY-MM-DD format.

Status Values:
  editing, submitted, rejected, approved, cancelled, complete, abandoned, scrapped

Output Columns (table format):
  ID          Plan identifier
  STATUS      Current status (abbreviated)
  DATE        Start date
  TIME        Start time
  CUSTOMER    Customer name
  JOB NAME    Job name
  PLANNER     Planner name
  PM          Project manager
  SITES       Material sites
  MATERIALS   Material types
  TONS        Goal tonnage
  % GOAL      Percentage of goal achieved
  SURP APPR   Approved surplus percentage
  SURP ACT    Actual surplus percentage`,
		Example: `  # List plans for a specific date
  xbe view job-production-plans list --start-on 2025-01-18

  # List plans for multiple specific dates
  xbe view job-production-plans list --start-on 2025-01-18,2025-01-19,2025-01-20

  # List plans for a date range
  xbe view job-production-plans list --start-on-min 2025-01-01 --start-on-max 2025-01-31

  # Filter by status
  xbe view job-production-plans list --start-on 2025-01-18 --status approved

  # Search by job name or number
  xbe view job-production-plans list --start-on 2025-01-18 --q "Main Street"

  # Filter by customer ID
  xbe view job-production-plans list --start-on 2025-01-18 --customer 123

  # Filter by planner user ID
  xbe view job-production-plans list --start-on 2025-01-18 --planner 456

  # Combine multiple filters
  xbe view job-production-plans list --start-on 2025-01-18 --status approved --customer 123

  # Get JSON output
  xbe view job-production-plans list --start-on 2025-01-18 --json`,
		RunE: runJobProductionPlansList,
	}
	initJobProductionPlansListFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlansCmd.AddCommand(newJobProductionPlansListCmd())
}

func initJobProductionPlansListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 0, "Page size (defaults to server default)")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("start-on", "", "Exact start date(s), comma-separated (YYYY-MM-DD)")
	cmd.Flags().String("start-on-min", "", "Start of date range (YYYY-MM-DD)")
	cmd.Flags().String("start-on-max", "", "End of date range (YYYY-MM-DD)")
	cmd.Flags().String("status", "", "Filter by status (editing/submitted/rejected/approved/cancelled/complete/abandoned/scrapped)")
	cmd.Flags().String("customer", "", "Filter by customer ID (comma-separated for multiple)")
	cmd.Flags().String("planner", "", "Filter by planner user ID (comma-separated for multiple)")
	cmd.Flags().String("project-manager", "", "Filter by project manager user ID (comma-separated for multiple)")
	cmd.Flags().String("job-site", "", "Filter by job site ID (comma-separated for multiple)")
	cmd.Flags().String("material-site", "", "Filter by material site ID (comma-separated for multiple)")
	cmd.Flags().String("business-unit", "", "Filter by business unit ID (comma-separated for multiple)")
	cmd.Flags().String("q", "", "Search by job name or number")
	cmd.Flags().String("broker", "", "Filter by broker ID (comma-separated for multiple)")
	cmd.Flags().String("project", "", "Filter by project ID (comma-separated for multiple)")
	cmd.Flags().String("trucker", "", "Filter by trucker ID (comma-separated for multiple)")
	cmd.Flags().String("job-number", "", "Filter by job number (partial match)")
	cmd.Flags().String("job-name", "", "Filter by job name (partial match)")
	cmd.Flags().String("material-type", "", "Filter by material type ID (comma-separated for multiple)")
	cmd.Flags().String("material-supplier", "", "Filter by material supplier ID (comma-separated for multiple)")
	cmd.Flags().String("contractor", "", "Filter by contractor ID (comma-separated for multiple)")
	cmd.Flags().String("is-template", "", "Filter by template status (true/false)")
	cmd.Flags().String("created-by", "", "Filter by creator user ID (comma-separated for multiple)")
	cmd.Flags().String("cost-code", "", "Filter by cost code ID (comma-separated for multiple)")
	cmd.Flags().String("start-time-min", "", "Filter by minimum start time (HH:MM)")
	cmd.Flags().String("start-time-max", "", "Filter by maximum start time (HH:MM)")
	cmd.Flags().String("remaining-quantity-min", "", "Filter by minimum remaining quantity")
	cmd.Flags().String("remaining-quantity-max", "", "Filter by maximum remaining quantity")
	cmd.Flags().String("default-trucker", "", "Filter by default trucker ID (comma-separated for multiple)")
	cmd.Flags().String("not-customer", "", "Exclude plans for customer ID (comma-separated for multiple)")
	cmd.Flags().String("trailer-classification-or-equivalent", "", "Filter by trailer classification")
	cmd.Flags().String("is-only-for-equipment-movement", "", "Filter by equipment movement only status (true/false)")
	cmd.Flags().String("broker-id", "", "Filter by broker ID (comma-separated for multiple)")
	cmd.Flags().String("is-auditing-time-card-approvals", "", "Filter by time card audit status (true/false)")
	cmd.Flags().String("planned-tons-per-productive-segment-min", "", "Filter by minimum planned tons per productive segment")
	cmd.Flags().String("planned-tons-per-productive-segment-max", "", "Filter by maximum planned tons per productive segment")
	cmd.Flags().String("default-time-card-approval-process", "", "Filter by approval process (admin/field)")
	cmd.Flags().String("is-using-volumetric-measurements", "", "Filter by volumetric measurements (true/false)")
	cmd.Flags().String("has-supply-demand-balance-cannot-compute-reasons", "", "Filter by supply/demand balance compute issues (true/false)")
	cmd.Flags().String("start-at-min", "", "Filter by minimum start datetime (ISO 8601)")
	cmd.Flags().String("start-at-max", "", "Filter by maximum start datetime (ISO 8601)")
	cmd.Flags().String("active-on", "", "Filter by active on date (YYYY-MM-DD)")
	cmd.Flags().String("practically-start-on", "", "Filter by practical start date (YYYY-MM-DD)")
	cmd.Flags().String("practically-start-on-min", "", "Filter by minimum practical start date (YYYY-MM-DD)")
	cmd.Flags().String("practically-start-on-max", "", "Filter by maximum practical start date (YYYY-MM-DD)")
	cmd.Flags().String("checksum-difference", "", "Filter by exact checksum difference")
	cmd.Flags().String("checksum-difference-min", "", "Filter by minimum checksum difference")
	cmd.Flags().String("checksum-difference-max", "", "Filter by maximum checksum difference")
	cmd.Flags().String("has-checksum-difference", "", "Filter by having checksum difference (true/false)")
	cmd.Flags().String("has-manager-assignment", "", "Filter by manager assignment (true/false)")
	cmd.Flags().String("user-has-stake", "", "Filter by user ID with stake (comma-separated for multiple)")
	cmd.Flags().String("template-name", "", "Filter by template name")
	cmd.Flags().String("template-start-on-min", "", "Filter by minimum template start date (YYYY-MM-DD)")
	cmd.Flags().String("template-start-on-max", "", "Filter by maximum template start date (YYYY-MM-DD)")
	cmd.Flags().String("template", "", "Filter by template ID (comma-separated for multiple)")
	cmd.Flags().String("duplication-token", "", "Filter by duplication token")
	cmd.Flags().String("job-site-active-around", "", "Filter by job site active around datetime (ISO 8601)")
	cmd.Flags().String("has-project-phase-revenue-items", "", "Filter by project phase revenue items (true/false)")
	cmd.Flags().String("has-crew-requirements", "", "Filter by crew requirements (true/false)")
	cmd.Flags().String("has-labor-requirements", "", "Filter by labor requirements (true/false)")
	cmd.Flags().String("has-equipment-requirements", "", "Filter by equipment requirements (true/false)")
	cmd.Flags().String("could-have-labor-requirements", "", "Filter by potential labor requirements (true/false)")
	cmd.Flags().String("ultimate-material-types", "", "Filter by ultimate material type names (comma-separated)")
	cmd.Flags().String("material-type-ultimate-parent-count-min", "", "Filter by minimum material type ultimate parent count")
	cmd.Flags().String("material-type-ultimate-parent-count-max", "", "Filter by maximum material type ultimate parent count")
	cmd.Flags().String("has-material-types-with-qc-requirements", "", "Filter by QC requirements (true/false)")
	cmd.Flags().String("with-non-deletable-lineup-jpps", "", "Filter by non-deletable lineup plans (true/false)")
	cmd.Flags().String("q-segments", "", "Search segments by description")
	cmd.Flags().String("external-identification-value", "", "Filter by external identification value")
	cmd.Flags().String("reference-data", "", "Filter by reference data (format: key|value)")
	cmd.Flags().String("practically-start-on-between", "", "Filter by practical start date range (format: date1|date2)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlansList(cmd *cobra.Command, _ []string) error {
	opts, err := parseJobProductionPlansListOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Require at least one date filter
	if strings.TrimSpace(opts.StartOn) == "" && strings.TrimSpace(opts.StartOnMin) == "" {
		err := fmt.Errorf("either --start-on or --start-on-min is required (use YYYY-MM-DD format)")
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
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("sort", "start-at")
	query.Set("fields[job-production-plans]", "job-number,job-name,status,start-on,start-time,goal-quantity,tons,tons-matched,material-type-ultimate-parent-qualified-names,customer,planner,project-manager,job-site,business-unit,project,planned-supply-demand-balance,actual-supply-demand-balance")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[users]", "name")
	query.Set("fields[job-sites]", "name")
	query.Set("fields[material-sites]", "name")
	query.Set("fields[business-units]", "company-name")
	query.Set("fields[projects]", "name")
	query.Set("fields[job-production-plan-supply-demand-balances]", "planned-practical-surplus-pct,actual-practical-surplus-pct")
	query.Set("include", "customer,planner,project-manager,job-site,business-unit,project,job-production-plan-material-sites.material-site,planned-supply-demand-balance,actual-supply-demand-balance")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	// Date filtering
	if opts.StartOn != "" {
		// Exact date(s)
		query.Set("filter[start-on]", opts.StartOn)
	}
	if opts.StartOnMin != "" {
		query.Set("filter[start-on-min]", opts.StartOnMin)
	}
	if opts.StartOnMax != "" {
		query.Set("filter[start-on-max]", opts.StartOnMax)
	}

	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[customer]", opts.Customer)
	setFilterIfPresent(query, "filter[planner]", opts.Planner)
	setFilterIfPresent(query, "filter[project-manager]", opts.ProjectMgr)
	setFilterIfPresent(query, "filter[job-site]", opts.JobSite)
	setFilterIfPresent(query, "filter[material-site]", opts.MaterialSite)
	setFilterIfPresent(query, "filter[business-unit]", opts.BusinessUnit)
	setFilterIfPresent(query, "filter[q]", opts.Q)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[project]", opts.Project)
	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)
	setFilterIfPresent(query, "filter[job-number]", opts.JobNumber)
	setFilterIfPresent(query, "filter[job-name]", opts.JobName)
	setFilterIfPresent(query, "filter[material-type]", opts.MaterialType)
	setFilterIfPresent(query, "filter[material-supplier]", opts.MaterialSupplier)
	setFilterIfPresent(query, "filter[contractor]", opts.Contractor)
	setFilterIfPresent(query, "filter[is-template]", opts.IsTemplate)
	setFilterIfPresent(query, "filter[created-by]", opts.CreatedBy)
	setFilterIfPresent(query, "filter[cost-code]", opts.CostCode)
	setFilterIfPresent(query, "filter[start-time-min]", opts.StartTimeMin)
	setFilterIfPresent(query, "filter[start-time-max]", opts.StartTimeMax)
	setFilterIfPresent(query, "filter[remaining-quantity-min]", opts.RemainingQuantityMin)
	setFilterIfPresent(query, "filter[remaining-quantity-max]", opts.RemainingQuantityMax)
	setFilterIfPresent(query, "filter[default-trucker]", opts.DefaultTrucker)
	setFilterIfPresent(query, "filter[not-customer]", opts.NotCustomer)
	setFilterIfPresent(query, "filter[trailer-classification-or-equivalent]", opts.TrailerClassificationOrEquiv)
	setFilterIfPresent(query, "filter[is-only-for-equipment-movement]", opts.IsOnlyForEquipmentMovement)
	setFilterIfPresent(query, "filter[broker-id]", opts.BrokerID)
	setFilterIfPresent(query, "filter[is-auditing-time-card-approvals]", opts.IsAuditingTimeCardApprovals)
	setFilterIfPresent(query, "filter[with-planned-tons-per-productive-segment-min]", opts.PlannedTonsPerProductiveSegmentMin)
	setFilterIfPresent(query, "filter[with-planned-tons-per-productive-segment-max]", opts.PlannedTonsPerProductiveSegmentMax)
	setFilterIfPresent(query, "filter[default-time-card-approval-process]", opts.DefaultTimeCardApprovalProcess)
	setFilterIfPresent(query, "filter[is-using-volumetric-measurements]", opts.IsUsingVolumetricMeasurements)
	setFilterIfPresent(query, "filter[has-supply-demand-balance-cannot-compute-reasons]", opts.HasSupplyDemandBalanceCannotComputeReasons)
	setFilterIfPresent(query, "filter[start-at-min]", opts.StartAtMin)
	setFilterIfPresent(query, "filter[start-at-max]", opts.StartAtMax)
	setFilterIfPresent(query, "filter[active-on]", opts.ActiveOn)
	setFilterIfPresent(query, "filter[practically-start-on]", opts.PracticallyStartOn)
	setFilterIfPresent(query, "filter[practically-start-on-min]", opts.PracticallyStartOnMin)
	setFilterIfPresent(query, "filter[practically-start-on-max]", opts.PracticallyStartOnMax)
	setFilterIfPresent(query, "filter[checksum-difference]", opts.ChecksumDifference)
	setFilterIfPresent(query, "filter[checksum-difference-min]", opts.ChecksumDifferenceMin)
	setFilterIfPresent(query, "filter[checksum-difference-max]", opts.ChecksumDifferenceMax)
	setFilterIfPresent(query, "filter[has-checksum-difference]", opts.HasChecksumDifference)
	setFilterIfPresent(query, "filter[has-manager-assignment]", opts.HasManagerAssignment)
	setFilterIfPresent(query, "filter[user-has-stake]", opts.UserHasStake)
	setFilterIfPresent(query, "filter[template-name]", opts.TemplateName)
	setFilterIfPresent(query, "filter[template-start-on-min]", opts.TemplateStartOnMin)
	setFilterIfPresent(query, "filter[template-start-on-max]", opts.TemplateStartOnMax)
	setFilterIfPresent(query, "filter[template]", opts.Template)
	setFilterIfPresent(query, "filter[duplication-token]", opts.DuplicationToken)
	setFilterIfPresent(query, "filter[job-site-active-around]", opts.JobSiteActiveAround)
	setFilterIfPresent(query, "filter[has-project-phase-revenue-items]", opts.HasProjectPhaseRevenueItems)
	setFilterIfPresent(query, "filter[has-crew-requirements]", opts.HasCrewRequirements)
	setFilterIfPresent(query, "filter[has-labor-requirements]", opts.HasLaborRequirements)
	setFilterIfPresent(query, "filter[has-equipment-requirements]", opts.HasEquipmentRequirements)
	setFilterIfPresent(query, "filter[could-have-labor-requirements]", opts.CouldHaveLaborRequirements)
	setFilterIfPresent(query, "filter[ultimate-material-types]", opts.UltimateMaterialTypes)
	setFilterIfPresent(query, "filter[material-type-ultimate-parent-count-min]", opts.MaterialTypeUltimateParentCountMin)
	setFilterIfPresent(query, "filter[material-type-ultimate-parent-count-max]", opts.MaterialTypeUltimateParentCountMax)
	setFilterIfPresent(query, "filter[has-material-types-with-quality-control-requirements]", opts.HasMaterialTypesWithQCRequirements)
	setFilterIfPresent(query, "filter[with-non-deletable-lineup-job-production-plans]", opts.WithNonDeletableLineupJPPs)
	setFilterIfPresent(query, "filter[q-segments]", opts.QSegments)
	setFilterIfPresent(query, "filter[external-identification-value]", opts.ExternalIdentificationValue)

	// reference-data filter (format: key|value) - server expects the raw string
	setFilterIfPresent(query, "filter[reference-data]", opts.ReferenceData)

	// practically-start-on-between filter (format: date1|date2)
	if opts.PracticallyStartOnBetween != "" {
		parts := strings.SplitN(opts.PracticallyStartOnBetween, "|", 2)
		if len(parts) == 2 {
			query.Set("filter[practically-start-on-between]", parts[0]+","+parts[1])
		}
	}

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plans", query)
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

	rows := buildJobProductionPlanRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderJobProductionPlansTable(cmd, rows)
}

func parseJobProductionPlansListOptions(cmd *cobra.Command) (jobProductionPlansListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	startOn, err := cmd.Flags().GetString("start-on")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	startOnMin, err := cmd.Flags().GetString("start-on-min")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	startOnMax, err := cmd.Flags().GetString("start-on-max")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	status, err := cmd.Flags().GetString("status")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	customer, err := cmd.Flags().GetString("customer")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	planner, err := cmd.Flags().GetString("planner")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	projectMgr, err := cmd.Flags().GetString("project-manager")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	jobSite, err := cmd.Flags().GetString("job-site")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	materialSite, err := cmd.Flags().GetString("material-site")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	businessUnit, err := cmd.Flags().GetString("business-unit")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	q, err := cmd.Flags().GetString("q")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	broker, err := cmd.Flags().GetString("broker")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	project, err := cmd.Flags().GetString("project")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	trucker, err := cmd.Flags().GetString("trucker")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	jobNumber, err := cmd.Flags().GetString("job-number")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	jobName, err := cmd.Flags().GetString("job-name")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	materialType, err := cmd.Flags().GetString("material-type")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	materialSupplier, err := cmd.Flags().GetString("material-supplier")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	contractor, err := cmd.Flags().GetString("contractor")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	isTemplate, err := cmd.Flags().GetString("is-template")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	createdBy, err := cmd.Flags().GetString("created-by")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	costCode, err := cmd.Flags().GetString("cost-code")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	startTimeMin, err := cmd.Flags().GetString("start-time-min")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	startTimeMax, err := cmd.Flags().GetString("start-time-max")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	remainingQuantityMin, err := cmd.Flags().GetString("remaining-quantity-min")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	remainingQuantityMax, err := cmd.Flags().GetString("remaining-quantity-max")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	defaultTrucker, err := cmd.Flags().GetString("default-trucker")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	notCustomer, err := cmd.Flags().GetString("not-customer")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	trailerClassificationOrEquiv, err := cmd.Flags().GetString("trailer-classification-or-equivalent")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	isOnlyForEquipmentMovement, err := cmd.Flags().GetString("is-only-for-equipment-movement")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	brokerID, err := cmd.Flags().GetString("broker-id")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	isAuditingTimeCardApprovals, err := cmd.Flags().GetString("is-auditing-time-card-approvals")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	plannedTonsPerProductiveSegmentMin, err := cmd.Flags().GetString("planned-tons-per-productive-segment-min")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	plannedTonsPerProductiveSegmentMax, err := cmd.Flags().GetString("planned-tons-per-productive-segment-max")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	defaultTimeCardApprovalProcess, err := cmd.Flags().GetString("default-time-card-approval-process")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	isUsingVolumetricMeasurements, err := cmd.Flags().GetString("is-using-volumetric-measurements")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	hasSupplyDemandBalanceCannotComputeReasons, err := cmd.Flags().GetString("has-supply-demand-balance-cannot-compute-reasons")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	startAtMin, err := cmd.Flags().GetString("start-at-min")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	startAtMax, err := cmd.Flags().GetString("start-at-max")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	activeOn, err := cmd.Flags().GetString("active-on")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	practicallyStartOn, err := cmd.Flags().GetString("practically-start-on")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	practicallyStartOnMin, err := cmd.Flags().GetString("practically-start-on-min")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	practicallyStartOnMax, err := cmd.Flags().GetString("practically-start-on-max")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	checksumDifference, err := cmd.Flags().GetString("checksum-difference")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	checksumDifferenceMin, err := cmd.Flags().GetString("checksum-difference-min")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	checksumDifferenceMax, err := cmd.Flags().GetString("checksum-difference-max")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	hasChecksumDifference, err := cmd.Flags().GetString("has-checksum-difference")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	hasManagerAssignment, err := cmd.Flags().GetString("has-manager-assignment")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	userHasStake, err := cmd.Flags().GetString("user-has-stake")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	templateName, err := cmd.Flags().GetString("template-name")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	templateStartOnMin, err := cmd.Flags().GetString("template-start-on-min")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	templateStartOnMax, err := cmd.Flags().GetString("template-start-on-max")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	template, err := cmd.Flags().GetString("template")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	duplicationToken, err := cmd.Flags().GetString("duplication-token")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	jobSiteActiveAround, err := cmd.Flags().GetString("job-site-active-around")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	hasProjectPhaseRevenueItems, err := cmd.Flags().GetString("has-project-phase-revenue-items")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	hasCrewRequirements, err := cmd.Flags().GetString("has-crew-requirements")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	hasLaborRequirements, err := cmd.Flags().GetString("has-labor-requirements")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	hasEquipmentRequirements, err := cmd.Flags().GetString("has-equipment-requirements")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	couldHaveLaborRequirements, err := cmd.Flags().GetString("could-have-labor-requirements")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	ultimateMaterialTypes, err := cmd.Flags().GetString("ultimate-material-types")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	materialTypeUltimateParentCountMin, err := cmd.Flags().GetString("material-type-ultimate-parent-count-min")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	materialTypeUltimateParentCountMax, err := cmd.Flags().GetString("material-type-ultimate-parent-count-max")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	hasMaterialTypesWithQCRequirements, err := cmd.Flags().GetString("has-material-types-with-qc-requirements")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	withNonDeletableLineupJPPs, err := cmd.Flags().GetString("with-non-deletable-lineup-jpps")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	qSegments, err := cmd.Flags().GetString("q-segments")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	externalIdentificationValue, err := cmd.Flags().GetString("external-identification-value")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	referenceData, err := cmd.Flags().GetString("reference-data")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	practicallyStartOnBetween, err := cmd.Flags().GetString("practically-start-on-between")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}

	return jobProductionPlansListOptions{
		BaseURL:                            baseURL,
		Token:                              token,
		JSON:                               jsonOut,
		NoAuth:                             noAuth,
		Limit:                              limit,
		Offset:                             offset,
		StartOn:                            startOn,
		StartOnMin:                         startOnMin,
		StartOnMax:                         startOnMax,
		Status:                             status,
		Customer:                           customer,
		Planner:                            planner,
		ProjectMgr:                         projectMgr,
		JobSite:                            jobSite,
		MaterialSite:                       materialSite,
		BusinessUnit:                       businessUnit,
		Q:                                  q,
		Broker:                             broker,
		BrokerID:                           brokerID,
		Project:                            project,
		Trucker:                            trucker,
		JobNumber:                          jobNumber,
		JobName:                            jobName,
		MaterialType:                       materialType,
		MaterialSupplier:                   materialSupplier,
		Contractor:                         contractor,
		IsTemplate:                         isTemplate,
		CreatedBy:                          createdBy,
		CostCode:                           costCode,
		StartTimeMin:                       startTimeMin,
		StartTimeMax:                       startTimeMax,
		RemainingQuantityMin:               remainingQuantityMin,
		RemainingQuantityMax:               remainingQuantityMax,
		DefaultTrucker:                     defaultTrucker,
		NotCustomer:                        notCustomer,
		TrailerClassificationOrEquiv:       trailerClassificationOrEquiv,
		IsOnlyForEquipmentMovement:         isOnlyForEquipmentMovement,
		IsAuditingTimeCardApprovals:        isAuditingTimeCardApprovals,
		PlannedTonsPerProductiveSegmentMin: plannedTonsPerProductiveSegmentMin,
		PlannedTonsPerProductiveSegmentMax: plannedTonsPerProductiveSegmentMax,
		DefaultTimeCardApprovalProcess:     defaultTimeCardApprovalProcess,
		IsUsingVolumetricMeasurements:      isUsingVolumetricMeasurements,
		HasSupplyDemandBalanceCannotComputeReasons: hasSupplyDemandBalanceCannotComputeReasons,
		StartAtMin:                         startAtMin,
		StartAtMax:                         startAtMax,
		ActiveOn:                           activeOn,
		PracticallyStartOn:                 practicallyStartOn,
		PracticallyStartOnMin:              practicallyStartOnMin,
		PracticallyStartOnMax:              practicallyStartOnMax,
		ChecksumDifference:                 checksumDifference,
		ChecksumDifferenceMin:              checksumDifferenceMin,
		ChecksumDifferenceMax:              checksumDifferenceMax,
		HasChecksumDifference:              hasChecksumDifference,
		HasManagerAssignment:               hasManagerAssignment,
		UserHasStake:                       userHasStake,
		TemplateName:                       templateName,
		TemplateStartOnMin:                 templateStartOnMin,
		TemplateStartOnMax:                 templateStartOnMax,
		Template:                           template,
		DuplicationToken:                   duplicationToken,
		JobSiteActiveAround:                jobSiteActiveAround,
		HasProjectPhaseRevenueItems:        hasProjectPhaseRevenueItems,
		HasCrewRequirements:                hasCrewRequirements,
		HasLaborRequirements:               hasLaborRequirements,
		HasEquipmentRequirements:           hasEquipmentRequirements,
		CouldHaveLaborRequirements:         couldHaveLaborRequirements,
		UltimateMaterialTypes:              ultimateMaterialTypes,
		MaterialTypeUltimateParentCountMin: materialTypeUltimateParentCountMin,
		MaterialTypeUltimateParentCountMax: materialTypeUltimateParentCountMax,
		HasMaterialTypesWithQCRequirements: hasMaterialTypesWithQCRequirements,
		WithNonDeletableLineupJPPs:         withNonDeletableLineupJPPs,
		QSegments:                          qSegments,
		ExternalIdentificationValue:        externalIdentificationValue,
		ReferenceData:                      referenceData,
		PracticallyStartOnBetween:          practicallyStartOnBetween,
	}, nil
}

func buildJobProductionPlanRows(resp jsonAPIResponse) []jobProductionPlanRow {
	// Build included lookup
	included := make(map[string]map[string]any)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc.Attributes
	}

	rows := make([]jobProductionPlanRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := jobProductionPlanRow{
			ID:        resource.ID,
			Status:    stringAttr(resource.Attributes, "status"),
			JobNumber: stringAttr(resource.Attributes, "job-number"),
			JobName:   stringAttr(resource.Attributes, "job-name"),
			StartOn:   formatDate(stringAttr(resource.Attributes, "start-on")),
			StartTime: formatTime(stringAttr(resource.Attributes, "start-time")),
			GoalTons:  floatAttr(resource.Attributes, "goal-quantity"),
		}

		// Resolve customer
		if rel, ok := resource.Relationships["customer"]; ok && rel.Data != nil {
			key := resourceKey(rel.Data.Type, rel.Data.ID)
			if attrs, ok := included[key]; ok {
				row.Customer = stringAttr(attrs, "company-name")
			}
		}

		// Resolve planner
		if rel, ok := resource.Relationships["planner"]; ok && rel.Data != nil {
			key := resourceKey(rel.Data.Type, rel.Data.ID)
			if attrs, ok := included[key]; ok {
				row.Planner = stringAttr(attrs, "name")
			}
		}

		// Resolve job site
		if rel, ok := resource.Relationships["job-site"]; ok && rel.Data != nil {
			key := resourceKey(rel.Data.Type, rel.Data.ID)
			if attrs, ok := included[key]; ok {
				row.JobSite = stringAttr(attrs, "name")
			}
		}

		// Additional fields
		row.MixTypes = stringAttr(resource.Attributes, "material-type-ultimate-parent-qualified-names")

		// Resolve planned supply demand balance for approved surplus %
		if rel, ok := resource.Relationships["planned-supply-demand-balance"]; ok && rel.Data != nil {
			key := resourceKey(rel.Data.Type, rel.Data.ID)
			if attrs, ok := included[key]; ok {
				row.ApprovedSurplusPct = floatAttr(attrs, "planned-practical-surplus-pct")
			}
		}

		// Resolve actual supply demand balance for actual surplus %
		if rel, ok := resource.Relationships["actual-supply-demand-balance"]; ok && rel.Data != nil {
			key := resourceKey(rel.Data.Type, rel.Data.ID)
			if attrs, ok := included[key]; ok {
				row.ActualSurplusPct = floatAttr(attrs, "actual-practical-surplus-pct")
			}
		}

		// Calculate tons percentage
		tonsMatched := floatAttr(resource.Attributes, "tons-matched")
		if row.GoalTons > 0 {
			row.TonsPct = (tonsMatched / row.GoalTons) * 100
		}

		// Resolve project manager
		if rel, ok := resource.Relationships["project-manager"]; ok && rel.Data != nil {
			key := resourceKey(rel.Data.Type, rel.Data.ID)
			if attrs, ok := included[key]; ok {
				row.ProjectManager = stringAttr(attrs, "name")
			}
		}

		// Resolve business unit
		if rel, ok := resource.Relationships["business-unit"]; ok && rel.Data != nil {
			key := resourceKey(rel.Data.Type, rel.Data.ID)
			if attrs, ok := included[key]; ok {
				row.BusinessUnit = stringAttr(attrs, "company-name")
			}
		}

		// Resolve project
		if rel, ok := resource.Relationships["project"]; ok && rel.Data != nil {
			key := resourceKey(rel.Data.Type, rel.Data.ID)
			if attrs, ok := included[key]; ok {
				row.ProjectName = stringAttr(attrs, "name")
			}
		}

		// Resolve material sites
		row.MaterialSite = resolveMaterialSites(resource, included)

		rows = append(rows, row)
	}

	return rows
}

func resolveMaterialSites(resource jsonAPIResource, included map[string]map[string]any) string {
	var sites []string
	seen := make(map[string]bool)

	// Look for material-sites in included
	for key, attrs := range included {
		if strings.HasPrefix(key, "material-sites|") {
			siteName := stringAttr(attrs, "name")
			if siteName != "" && !seen[siteName] {
				seen[siteName] = true
				sites = append(sites, siteName)
			}
		}
	}

	if len(sites) == 0 {
		return ""
	}
	if len(sites) == 1 {
		return sites[0]
	}
	return fmt.Sprintf("%s (+%d)", sites[0], len(sites)-1)
}

func formatTime(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	// Time is typically in HH:MM:SS or similar format
	// Return first 5 chars (HH:MM) if longer
	if len(value) >= 5 {
		return value[:5]
	}
	return value
}

func floatAttr(attrs map[string]any, key string) float64 {
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
	case float32:
		return float64(typed)
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	case string:
		if f, err := strconv.ParseFloat(typed, 64); err == nil {
			return f
		}
	}
	return 0
}

func renderJobProductionPlansTable(cmd *cobra.Command, rows []jobProductionPlanRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No job production plans found.")
		return nil
	}

	const (
		maxCustomer = 18
		maxJobName  = 25
		maxPlanner  = 10
		maxPM       = 10
		maxMaterial = 15
		maxMixTypes = 15
	)

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "\t\t-- SCHEDULE --\t\t-- JOB --\t\t-- ASSIGNED --\t\t-- MATERIALS --\t\t-- PRODUCTION --\t\t-- SURPLUS --")
	fmt.Fprintln(writer, "ID\tSTATUS\tDATE\tTIME\tCUSTOMER\tJOB NAME\tPLANNER\tPM\tSITES\tMATERIALS\tTONS\t% GOAL\tAPPR\tACTUAL")
	for _, row := range rows {
		tonsPctStr := ""
		if row.TonsPct > 0 {
			tonsPctStr = fmt.Sprintf("%.0f%%", row.TonsPct)
		}
		apprSurpStr := ""
		apprPct := row.ApprovedSurplusPct * 100
		if apprPct > 0.5 || apprPct < -0.5 {
			apprSurpStr = fmt.Sprintf("%.0f%%", apprPct)
		}
		actSurpStr := ""
		actPct := row.ActualSurplusPct * 100
		if actPct > 0.5 || actPct < -0.5 {
			actSurpStr = fmt.Sprintf("%.0f%%", actPct)
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%.0f\t%s\t%s\t%s\n",
			row.ID,
			abbreviateStatus(row.Status),
			row.StartOn,
			row.StartTime,
			truncateString(row.Customer, maxCustomer),
			truncateString(row.JobName, maxJobName),
			truncateString(row.Planner, maxPlanner),
			truncateString(row.ProjectManager, maxPM),
			truncateString(row.MaterialSite, maxMaterial),
			truncateString(row.MixTypes, maxMixTypes),
			row.GoalTons,
			tonsPctStr,
			apprSurpStr,
			actSurpStr,
		)
	}
	return writer.Flush()
}

func abbreviateStatus(status string) string {
	switch strings.ToLower(status) {
	case "editing":
		return "EDIT"
	case "submitted":
		return "SUBM"
	case "rejected":
		return "REJ"
	case "approved":
		return "APPR"
	case "cancelled":
		return "CANC"
	case "complete":
		return "COMP"
	case "abandoned":
		return "ABAN"
	case "scrapped":
		return "SCRP"
	default:
		return strings.ToUpper(status[:4])
	}
}
