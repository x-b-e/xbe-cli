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
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	StartOn      string
	StartOnMin   string
	StartOnMax   string
	Status       string
	Customer     string
	Planner      string
	ProjectMgr   string
	JobSite      string
	MaterialSite string
	BusinessUnit string
	Q            string
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
	cmd.Flags().String("customer", "", "Filter by customer ID")
	cmd.Flags().String("planner", "", "Filter by planner user ID")
	cmd.Flags().String("project-manager", "", "Filter by project manager user ID")
	cmd.Flags().String("job-site", "", "Filter by job site ID")
	cmd.Flags().String("material-site", "", "Filter by material site ID")
	cmd.Flags().String("business-unit", "", "Filter by business unit ID")
	cmd.Flags().String("q", "", "Search by job name or number")
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
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return jobProductionPlansListOptions{}, err
	}

	return jobProductionPlansListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		StartOn:      startOn,
		StartOnMin:   startOnMin,
		StartOnMax:   startOnMax,
		Status:       status,
		Customer:     customer,
		Planner:      planner,
		ProjectMgr:   projectMgr,
		JobSite:      jobSite,
		MaterialSite: materialSite,
		BusinessUnit: businessUnit,
		Q:            q,
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
