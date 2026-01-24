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

type administrativeIncidentsListOptions struct {
	BaseURL                       string
	Token                         string
	JSON                          bool
	NoAuth                        bool
	Limit                         int
	Offset                        int
	Sort                          string
	Status                        string
	Kind                          string
	Severity                      string
	Broker                        string
	Customer                      string
	Developer                     string
	Trucker                       string
	Contractor                    string
	MaterialSupplier              string
	MaterialSite                  string
	JobProductionPlan             string
	JobProductionPlanProject      string
	Equipment                     string
	Assignee                      string
	CreatedBy                     string
	Parent                        string
	StartOn                       string
	StartOnMin                    string
	StartOnMax                    string
	StartAtMin                    string
	StartAtMax                    string
	EndAtMin                      string
	EndAtMax                      string
	Subject                       string
	SubjectType                   string
	SubjectID                     string
	NotSubjectType                string
	HasParent                     string
	HasEquipment                  string
	HasLiveActionItems            string
	IncidentTag                   string
	IncidentTagSlug               string
	ZeroIncidentTags              string
	RootCauses                    string
	ActionItems                   string
	TenderJobScheduleShift        string
	TenderJobScheduleShiftDriver  string
	TenderJobScheduleShiftTrucker string
	JobNumber                     string
	NotifiableTo                  string
	UserHasStake                  string
	ResponsiblePerson             string
	Natures                       string
	DidStopWork                   string
	NetImpactDollars              string
	NetImpactDollarsMin           string
	NetImpactDollarsMax           string
	Q                             string
}

type administrativeIncidentRow struct {
	ID                string `json:"id"`
	Status            string `json:"status,omitempty"`
	Kind              string `json:"kind,omitempty"`
	Severity          string `json:"severity,omitempty"`
	Headline          string `json:"headline,omitempty"`
	StartAt           string `json:"start_at,omitempty"`
	NetImpactDollars  string `json:"net_impact_dollars,omitempty"`
	SubjectType       string `json:"subject_type,omitempty"`
	SubjectID         string `json:"subject_id,omitempty"`
	AssigneeID        string `json:"assignee_id,omitempty"`
	CreatedByID       string `json:"created_by_id,omitempty"`
	BrokerID          string `json:"broker_id,omitempty"`
	JobProductionPlan string `json:"job_production_plan_id,omitempty"`
}

func newAdministrativeIncidentsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List administrative incidents",
		Long: `List administrative incidents with filtering and pagination.

Output Columns:
  ID          Incident identifier
  STATUS      Current status
  KIND        Incident kind
  SEVERITY    Severity level
  HEADLINE    Brief description
  START AT    When the incident started
  NET IMPACT  Net impact dollars
  SUBJECT     Subject type and ID

Filters:
  --status                         Filter by status (comma-separated)
  --kind                           Filter by kind (comma-separated)
  --severity                       Filter by severity (comma-separated)
  --broker                         Filter by broker ID (comma-separated)
  --customer                       Filter by customer ID (comma-separated)
  --developer                      Filter by developer ID (comma-separated)
  --trucker                        Filter by trucker ID (comma-separated)
  --contractor                     Filter by contractor ID (comma-separated)
  --material-supplier              Filter by material supplier ID (comma-separated)
  --material-site                  Filter by material site ID (comma-separated)
  --job-production-plan            Filter by job production plan ID (comma-separated)
  --job-production-plan-project    Filter by job production plan project ID (comma-separated)
  --equipment                      Filter by equipment ID (comma-separated)
  --assignee                       Filter by assignee user ID (comma-separated)
  --created-by                     Filter by created-by user ID (comma-separated)
  --parent                         Filter by parent incident ID (comma-separated)
  --subject                        Filter by subject Type|ID (e.g. Broker|123)
  --subject-type                   Filter by subject type (Broker, Customer, JobProductionPlan, etc.)
  --subject-id                     Filter by subject ID (requires --subject-type)
  --not-subject-type               Filter by excluding subject types (comma-separated, server-dependent)
  --has-parent                     Filter by has parent (true/false)
  --has-equipment                  Filter by has equipment (true/false)
  --has-live-action-items          Filter by has live action items (true/false)
  --incident-tag                   Filter by incident tag ID (comma-separated)
  --incident-tag-slug              Filter by incident tag slug
  --zero-incident-tags             Filter by zero incident tags (true/false)
  --root-causes                    Filter by root cause IDs (comma-separated)
  --action-items                   Filter by action item IDs (comma-separated)
  --tender-job-schedule-shift      Filter by tender job schedule shift ID (comma-separated)
  --tender-job-schedule-shift-driver   Filter by tender job schedule shift driver ID (comma-separated)
  --tender-job-schedule-shift-trucker  Filter by tender job schedule shift trucker ID (comma-separated)
  --job-number                     Filter by job number
  --notifiable-to                  Filter by notifiable-to user ID (comma-separated)
  --user-has-stake                 Filter by user-has-stake user ID (comma-separated)
  --responsible-person             Filter by responsible person user ID (comma-separated)
  --natures                        Filter by nature (comma-separated: personal,property)
  --did-stop-work                  Filter by did stop work (true/false)
  --start-on                       Filter by exact start date (YYYY-MM-DD)
  --start-on-min                   Filter by minimum start date (YYYY-MM-DD)
  --start-on-max                   Filter by maximum start date (YYYY-MM-DD)
  --start-at-min                   Filter by minimum start timestamp (ISO 8601)
  --start-at-max                   Filter by maximum start timestamp (ISO 8601)
  --end-at-min                     Filter by minimum end timestamp (ISO 8601)
  --end-at-max                     Filter by maximum end timestamp (ISO 8601)
  --net-impact-dollars             Filter by exact net impact dollars
  --net-impact-dollars-min         Filter by minimum net impact dollars
  --net-impact-dollars-max         Filter by maximum net impact dollars
  --q                              Search query

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List administrative incidents
  xbe view administrative-incidents list

  # Filter by status and kind
  xbe view administrative-incidents list --status open --kind planning

  # Filter by subject and date range
  xbe view administrative-incidents list --subject Broker|123 --start-on-min 2024-01-01 --start-on-max 2024-12-31

  # Output as JSON
  xbe view administrative-incidents list --json`,
		Args: cobra.NoArgs,
		RunE: runAdministrativeIncidentsList,
	}
	initAdministrativeIncidentsListFlags(cmd)
	return cmd
}

func init() {
	administrativeIncidentsCmd.AddCommand(newAdministrativeIncidentsListCmd())
}

func initAdministrativeIncidentsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("status", "", "Filter by status (comma-separated)")
	cmd.Flags().String("kind", "", "Filter by kind (comma-separated)")
	cmd.Flags().String("severity", "", "Filter by severity (comma-separated)")
	cmd.Flags().String("broker", "", "Filter by broker ID (comma-separated)")
	cmd.Flags().String("customer", "", "Filter by customer ID (comma-separated)")
	cmd.Flags().String("developer", "", "Filter by developer ID (comma-separated)")
	cmd.Flags().String("trucker", "", "Filter by trucker ID (comma-separated)")
	cmd.Flags().String("contractor", "", "Filter by contractor ID (comma-separated)")
	cmd.Flags().String("material-supplier", "", "Filter by material supplier ID (comma-separated)")
	cmd.Flags().String("material-site", "", "Filter by material site ID (comma-separated)")
	cmd.Flags().String("job-production-plan", "", "Filter by job production plan ID (comma-separated)")
	cmd.Flags().String("job-production-plan-project", "", "Filter by job production plan project ID (comma-separated)")
	cmd.Flags().String("equipment", "", "Filter by equipment ID (comma-separated)")
	cmd.Flags().String("assignee", "", "Filter by assignee user ID (comma-separated)")
	cmd.Flags().String("created-by", "", "Filter by created-by user ID (comma-separated)")
	cmd.Flags().String("parent", "", "Filter by parent incident ID (comma-separated)")
	cmd.Flags().String("subject", "", "Filter by subject Type|ID (e.g. Broker|123)")
	cmd.Flags().String("subject-type", "", "Filter by subject type (Broker, Customer, JobProductionPlan, etc.)")
	cmd.Flags().String("subject-id", "", "Filter by subject ID (requires --subject-type)")
	cmd.Flags().String("not-subject-type", "", "Filter by excluding subject types (comma-separated)")
	cmd.Flags().String("has-parent", "", "Filter by has parent (true/false)")
	cmd.Flags().String("has-equipment", "", "Filter by has equipment (true/false)")
	cmd.Flags().String("has-live-action-items", "", "Filter by has live action items (true/false)")
	cmd.Flags().String("incident-tag", "", "Filter by incident tag ID (comma-separated)")
	cmd.Flags().String("incident-tag-slug", "", "Filter by incident tag slug")
	cmd.Flags().String("zero-incident-tags", "", "Filter by zero incident tags (true/false)")
	cmd.Flags().String("root-causes", "", "Filter by root cause IDs (comma-separated)")
	cmd.Flags().String("action-items", "", "Filter by action item IDs (comma-separated)")
	cmd.Flags().String("tender-job-schedule-shift", "", "Filter by tender job schedule shift ID (comma-separated)")
	cmd.Flags().String("tender-job-schedule-shift-driver", "", "Filter by tender job schedule shift driver ID (comma-separated)")
	cmd.Flags().String("tender-job-schedule-shift-trucker", "", "Filter by tender job schedule shift trucker ID (comma-separated)")
	cmd.Flags().String("job-number", "", "Filter by job number")
	cmd.Flags().String("notifiable-to", "", "Filter by notifiable-to user ID (comma-separated)")
	cmd.Flags().String("user-has-stake", "", "Filter by user-has-stake user ID (comma-separated)")
	cmd.Flags().String("responsible-person", "", "Filter by responsible person user ID (comma-separated)")
	cmd.Flags().String("natures", "", "Filter by nature (comma-separated: personal,property)")
	cmd.Flags().String("did-stop-work", "", "Filter by did stop work (true/false)")
	cmd.Flags().String("start-on", "", "Filter by exact start date (YYYY-MM-DD)")
	cmd.Flags().String("start-on-min", "", "Filter by minimum start date (YYYY-MM-DD)")
	cmd.Flags().String("start-on-max", "", "Filter by maximum start date (YYYY-MM-DD)")
	cmd.Flags().String("start-at-min", "", "Filter by minimum start timestamp (ISO 8601)")
	cmd.Flags().String("start-at-max", "", "Filter by maximum start timestamp (ISO 8601)")
	cmd.Flags().String("end-at-min", "", "Filter by minimum end timestamp (ISO 8601)")
	cmd.Flags().String("end-at-max", "", "Filter by maximum end timestamp (ISO 8601)")
	cmd.Flags().String("net-impact-dollars", "", "Filter by exact net impact dollars")
	cmd.Flags().String("net-impact-dollars-min", "", "Filter by minimum net impact dollars")
	cmd.Flags().String("net-impact-dollars-max", "", "Filter by maximum net impact dollars")
	cmd.Flags().String("q", "", "Search query")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runAdministrativeIncidentsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseAdministrativeIncidentsListOptions(cmd)
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
	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[kind]", opts.Kind)
	setFilterIfPresent(query, "filter[severity]", opts.Severity)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[customer]", opts.Customer)
	setFilterIfPresent(query, "filter[developer]", opts.Developer)
	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)
	setFilterIfPresent(query, "filter[contractor]", opts.Contractor)
	setFilterIfPresent(query, "filter[material_supplier]", opts.MaterialSupplier)
	setFilterIfPresent(query, "filter[material_site]", opts.MaterialSite)
	setFilterIfPresent(query, "filter[job_production_plan]", opts.JobProductionPlan)
	setFilterIfPresent(query, "filter[job_production_plan_project]", opts.JobProductionPlanProject)
	setFilterIfPresent(query, "filter[equipment]", opts.Equipment)
	setFilterIfPresent(query, "filter[assignee]", opts.Assignee)
	setFilterIfPresent(query, "filter[created_by]", opts.CreatedBy)
	setFilterIfPresent(query, "filter[parent]", opts.Parent)
	setFilterIfPresent(query, "filter[subject_type]", opts.SubjectType)
	setFilterIfPresent(query, "filter[not_subject_type]", opts.NotSubjectType)
	setFilterIfPresent(query, "filter[has_parent]", opts.HasParent)
	setFilterIfPresent(query, "filter[has_equipment]", opts.HasEquipment)
	setFilterIfPresent(query, "filter[has_live_action_items]", opts.HasLiveActionItems)
	setFilterIfPresent(query, "filter[incident_tag]", opts.IncidentTag)
	setFilterIfPresent(query, "filter[incident_tag_slug]", opts.IncidentTagSlug)
	setFilterIfPresent(query, "filter[zero_incident_tags]", opts.ZeroIncidentTags)
	setFilterIfPresent(query, "filter[root_causes]", opts.RootCauses)
	setFilterIfPresent(query, "filter[action_items]", opts.ActionItems)
	setFilterIfPresent(query, "filter[tender_job_schedule_shift]", opts.TenderJobScheduleShift)
	setFilterIfPresent(query, "filter[tender_job_schedule_shift_driver]", opts.TenderJobScheduleShiftDriver)
	setFilterIfPresent(query, "filter[tender_job_schedule_shift_trucker]", opts.TenderJobScheduleShiftTrucker)
	setFilterIfPresent(query, "filter[job_number]", opts.JobNumber)
	setFilterIfPresent(query, "filter[notifiable_to]", opts.NotifiableTo)
	setFilterIfPresent(query, "filter[user_has_stake]", opts.UserHasStake)
	setFilterIfPresent(query, "filter[responsible_person]", opts.ResponsiblePerson)
	setFilterIfPresent(query, "filter[natures]", opts.Natures)
	setFilterIfPresent(query, "filter[did_stop_work]", opts.DidStopWork)
	setFilterIfPresent(query, "filter[start_on]", opts.StartOn)
	setFilterIfPresent(query, "filter[start_on_min]", opts.StartOnMin)
	setFilterIfPresent(query, "filter[start_on_max]", opts.StartOnMax)
	setFilterIfPresent(query, "filter[start_at_min]", opts.StartAtMin)
	setFilterIfPresent(query, "filter[start_at_max]", opts.StartAtMax)
	setFilterIfPresent(query, "filter[end_at_min]", opts.EndAtMin)
	setFilterIfPresent(query, "filter[end_at_max]", opts.EndAtMax)
	if opts.NetImpactDollars != "" {
		if opts.NetImpactDollarsMin == "" {
			opts.NetImpactDollarsMin = opts.NetImpactDollars
		}
		if opts.NetImpactDollarsMax == "" {
			opts.NetImpactDollarsMax = opts.NetImpactDollars
		}
	}
	setFilterIfPresent(query, "filter[net_impact_dollars_min]", opts.NetImpactDollarsMin)
	setFilterIfPresent(query, "filter[net_impact_dollars_max]", opts.NetImpactDollarsMax)
	setFilterIfPresent(query, "filter[q]", opts.Q)

	if opts.Subject != "" {
		if opts.SubjectID != "" {
			err := fmt.Errorf("--subject and --subject-id cannot be used together")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		subjectFilter, err := buildIncidentSubjectFilter(opts.Subject)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		query.Set("filter[subject]", subjectFilter)
	}
	if opts.SubjectID != "" {
		if strings.TrimSpace(opts.SubjectType) == "" {
			err := fmt.Errorf("--subject-type is required when using --subject-id")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		subjectFilter, err := buildIncidentSubjectFilter(opts.SubjectType + "|" + opts.SubjectID)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		query.Set("filter[subject]", subjectFilter)
	}

	body, _, err := client.Get(cmd.Context(), "/v1/administrative-incidents", query)
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

	rows := buildAdministrativeIncidentRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderAdministrativeIncidentsTable(cmd, rows)
}

func parseAdministrativeIncidentsListOptions(cmd *cobra.Command) (administrativeIncidentsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	status, _ := cmd.Flags().GetString("status")
	kind, _ := cmd.Flags().GetString("kind")
	severity, _ := cmd.Flags().GetString("severity")
	broker, _ := cmd.Flags().GetString("broker")
	customer, _ := cmd.Flags().GetString("customer")
	developer, _ := cmd.Flags().GetString("developer")
	trucker, _ := cmd.Flags().GetString("trucker")
	contractor, _ := cmd.Flags().GetString("contractor")
	materialSupplier, _ := cmd.Flags().GetString("material-supplier")
	materialSite, _ := cmd.Flags().GetString("material-site")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	jobProductionPlanProject, _ := cmd.Flags().GetString("job-production-plan-project")
	equipment, _ := cmd.Flags().GetString("equipment")
	assignee, _ := cmd.Flags().GetString("assignee")
	createdBy, _ := cmd.Flags().GetString("created-by")
	parent, _ := cmd.Flags().GetString("parent")
	startOn, _ := cmd.Flags().GetString("start-on")
	startOnMin, _ := cmd.Flags().GetString("start-on-min")
	startOnMax, _ := cmd.Flags().GetString("start-on-max")
	startAtMin, _ := cmd.Flags().GetString("start-at-min")
	startAtMax, _ := cmd.Flags().GetString("start-at-max")
	endAtMin, _ := cmd.Flags().GetString("end-at-min")
	endAtMax, _ := cmd.Flags().GetString("end-at-max")
	subject, _ := cmd.Flags().GetString("subject")
	subjectType, _ := cmd.Flags().GetString("subject-type")
	subjectID, _ := cmd.Flags().GetString("subject-id")
	notSubjectType, _ := cmd.Flags().GetString("not-subject-type")
	hasParent, _ := cmd.Flags().GetString("has-parent")
	hasEquipment, _ := cmd.Flags().GetString("has-equipment")
	hasLiveActionItems, _ := cmd.Flags().GetString("has-live-action-items")
	incidentTag, _ := cmd.Flags().GetString("incident-tag")
	incidentTagSlug, _ := cmd.Flags().GetString("incident-tag-slug")
	zeroIncidentTags, _ := cmd.Flags().GetString("zero-incident-tags")
	rootCauses, _ := cmd.Flags().GetString("root-causes")
	actionItems, _ := cmd.Flags().GetString("action-items")
	tenderJobScheduleShift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	tenderJobScheduleShiftDriver, _ := cmd.Flags().GetString("tender-job-schedule-shift-driver")
	tenderJobScheduleShiftTrucker, _ := cmd.Flags().GetString("tender-job-schedule-shift-trucker")
	jobNumber, _ := cmd.Flags().GetString("job-number")
	notifiableTo, _ := cmd.Flags().GetString("notifiable-to")
	userHasStake, _ := cmd.Flags().GetString("user-has-stake")
	responsiblePerson, _ := cmd.Flags().GetString("responsible-person")
	natures, _ := cmd.Flags().GetString("natures")
	didStopWork, _ := cmd.Flags().GetString("did-stop-work")
	netImpactDollars, _ := cmd.Flags().GetString("net-impact-dollars")
	netImpactDollarsMin, _ := cmd.Flags().GetString("net-impact-dollars-min")
	netImpactDollarsMax, _ := cmd.Flags().GetString("net-impact-dollars-max")
	q, _ := cmd.Flags().GetString("q")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return administrativeIncidentsListOptions{
		BaseURL:                       baseURL,
		Token:                         token,
		JSON:                          jsonOut,
		NoAuth:                        noAuth,
		Limit:                         limit,
		Offset:                        offset,
		Sort:                          sort,
		Status:                        status,
		Kind:                          kind,
		Severity:                      severity,
		Broker:                        broker,
		Customer:                      customer,
		Developer:                     developer,
		Trucker:                       trucker,
		Contractor:                    contractor,
		MaterialSupplier:              materialSupplier,
		MaterialSite:                  materialSite,
		JobProductionPlan:             jobProductionPlan,
		JobProductionPlanProject:      jobProductionPlanProject,
		Equipment:                     equipment,
		Assignee:                      assignee,
		CreatedBy:                     createdBy,
		Parent:                        parent,
		StartOn:                       startOn,
		StartOnMin:                    startOnMin,
		StartOnMax:                    startOnMax,
		StartAtMin:                    startAtMin,
		StartAtMax:                    startAtMax,
		EndAtMin:                      endAtMin,
		EndAtMax:                      endAtMax,
		Subject:                       subject,
		SubjectType:                   subjectType,
		SubjectID:                     subjectID,
		NotSubjectType:                notSubjectType,
		HasParent:                     hasParent,
		HasEquipment:                  hasEquipment,
		HasLiveActionItems:            hasLiveActionItems,
		IncidentTag:                   incidentTag,
		IncidentTagSlug:               incidentTagSlug,
		ZeroIncidentTags:              zeroIncidentTags,
		RootCauses:                    rootCauses,
		ActionItems:                   actionItems,
		TenderJobScheduleShift:        tenderJobScheduleShift,
		TenderJobScheduleShiftDriver:  tenderJobScheduleShiftDriver,
		TenderJobScheduleShiftTrucker: tenderJobScheduleShiftTrucker,
		JobNumber:                     jobNumber,
		NotifiableTo:                  notifiableTo,
		UserHasStake:                  userHasStake,
		ResponsiblePerson:             responsiblePerson,
		Natures:                       natures,
		DidStopWork:                   didStopWork,
		NetImpactDollars:              netImpactDollars,
		NetImpactDollarsMin:           netImpactDollarsMin,
		NetImpactDollarsMax:           netImpactDollarsMax,
		Q:                             q,
	}, nil
}

func buildAdministrativeIncidentRows(resp jsonAPIResponse) []administrativeIncidentRow {
	rows := make([]administrativeIncidentRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := administrativeIncidentRow{
			ID:               resource.ID,
			Status:           stringAttr(resource.Attributes, "status"),
			Kind:             stringAttr(resource.Attributes, "kind"),
			Severity:         stringAttr(resource.Attributes, "severity"),
			Headline:         stringAttr(resource.Attributes, "headline"),
			StartAt:          stringAttr(resource.Attributes, "start-at"),
			NetImpactDollars: stringAttr(resource.Attributes, "net-impact-dollars"),
		}

		if rel, ok := resource.Relationships["subject"]; ok && rel.Data != nil {
			row.SubjectType = rel.Data.Type
			row.SubjectID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["assignee"]; ok && rel.Data != nil {
			row.AssigneeID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
			row.CreatedByID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			row.BrokerID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
			row.JobProductionPlan = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderAdministrativeIncidentsTable(cmd *cobra.Command, rows []administrativeIncidentRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No administrative incidents found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tKIND\tSEVERITY\tHEADLINE\tSTART AT\tNET IMPACT\tSUBJECT")
	for _, row := range rows {
		subject := ""
		if row.SubjectType != "" && row.SubjectID != "" {
			subject = row.SubjectType + "/" + row.SubjectID
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Status,
			row.Kind,
			row.Severity,
			truncateString(row.Headline, 30),
			row.StartAt,
			row.NetImpactDollars,
			subject,
		)
	}
	return writer.Flush()
}
