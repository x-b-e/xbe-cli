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

type productionIncidentsListOptions struct {
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
	Subject                       string
	SubjectType                   string
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
	StartOn                       string
	StartOnMin                    string
	StartOnMax                    string
	StartAtMin                    string
	StartAtMax                    string
	EndAtMin                      string
	EndAtMax                      string
	NetImpactMinutesMin           string
	NetImpactMinutesMax           string
	NetImpactDollarsMin           string
	NetImpactDollarsMax           string
	Q                             string
}

type productionIncidentRow struct {
	ID               string `json:"id"`
	Status           string `json:"status,omitempty"`
	Kind             string `json:"kind,omitempty"`
	Severity         string `json:"severity,omitempty"`
	TimeValueType    string `json:"time_value_type,omitempty"`
	StartAt          string `json:"start_at,omitempty"`
	NetImpactMinutes string `json:"net_impact_minutes,omitempty"`
	NetImpactDollars string `json:"net_impact_dollars,omitempty"`
	IsDownTime       bool   `json:"is_down_time,omitempty"`
	SubjectType      string `json:"subject_type,omitempty"`
	SubjectID        string `json:"subject_id,omitempty"`
}

func newProductionIncidentsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List production incidents",
		Long: `List production incidents with filters and pagination.

Output Columns:
  ID          Incident identifier
  STATUS      Current status
  KIND        Incident kind
  SEVERITY    Severity level
  TIME TYPE   Time value type
  NET MIN     Net impact minutes
  NET $       Net impact dollars
  DOWN TIME   Down time flag
  START AT    Start timestamp
  SUBJECT     Subject type and ID

Filters:
  --status                        Filter by status
  --kind                          Filter by kind
  --severity                      Filter by severity
  --did-stop-work                 Filter by did stop work (true/false)
  --natures                       Filter by incident natures (comma-separated)
  --subject                       Filter by subject (Type|ID, class name)
  --subject-type                  Filter by subject type (class name)
  --parent                        Filter by parent incident ID
  --has-parent                    Filter by has parent (true/false)
  --broker                        Filter by broker ID
  --customer                      Filter by customer ID
  --developer                     Filter by developer ID
  --trucker                       Filter by trucker ID
  --contractor                    Filter by contractor ID
  --material-supplier             Filter by material supplier ID
  --material-site                 Filter by material site ID
  --job-production-plan           Filter by job production plan ID
  --job-production-plan-project   Filter by job production plan project ID
  --equipment                     Filter by equipment ID
  --assignee                      Filter by assignee user ID
  --created-by                    Filter by created-by user ID
  --tender-job-schedule-shift     Filter by tender job schedule shift ID
  --tender-job-schedule-shift-driver  Filter by tender job schedule shift driver user ID
  --tender-job-schedule-shift-trucker Filter by tender job schedule shift trucker ID
  --job-number                    Filter by job number (partial match)
  --has-equipment                 Filter by has equipment (true/false)
  --has-live-action-items         Filter by has live action items (true/false)
  --incident-tag                  Filter by incident tag ID
  --incident-tag-slug             Filter by incident tag slug
  --zero-incident-tags            Filter by zero incident tags (true/false)
  --root-causes                   Filter by root cause IDs (comma-separated)
  --action-items                  Filter by action item IDs (comma-separated)
  --notifiable-to                 Filter by notifiable user ID
  --user-has-stake                Filter by user ID with stake
  --responsible-person            Filter by responsible person user ID
  --start-on                      Filter by start date (YYYY-MM-DD)
  --start-on-min                  Filter by minimum start date (YYYY-MM-DD)
  --start-on-max                  Filter by maximum start date (YYYY-MM-DD)
  --start-at-min                  Filter by minimum start timestamp (ISO 8601)
  --start-at-max                  Filter by maximum start timestamp (ISO 8601)
  --end-at-min                    Filter by minimum end timestamp (ISO 8601)
  --end-at-max                    Filter by maximum end timestamp (ISO 8601)
  --net-impact-minutes-min        Filter by minimum net impact minutes
  --net-impact-minutes-max        Filter by maximum net impact minutes
  --net-impact-dollars-min        Filter by minimum net impact dollars
  --net-impact-dollars-max        Filter by maximum net impact dollars
  --q                             Search query

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List production incidents
  xbe view production-incidents list

  # Filter by status and kind
  xbe view production-incidents list --status open --kind trucking

  # Filter by net impact
  xbe view production-incidents list --net-impact-minutes-min 15 --net-impact-dollars-min 100

  # Filter by subject
  xbe view production-incidents list --subject \"Customer|123\"

  # Output as JSON
  xbe view production-incidents list --json`,
		RunE: runProductionIncidentsList,
	}
	initProductionIncidentsListFlags(cmd)
	return cmd
}

func init() {
	productionIncidentsCmd.AddCommand(newProductionIncidentsListCmd())
}

func initProductionIncidentsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("status", "", "Filter by status")
	cmd.Flags().String("kind", "", "Filter by kind")
	cmd.Flags().String("severity", "", "Filter by severity")
	cmd.Flags().String("did-stop-work", "", "Filter by did stop work (true/false)")
	cmd.Flags().String("natures", "", "Filter by incident natures (comma-separated)")
	cmd.Flags().String("subject", "", "Filter by subject (Type|ID, class name)")
	cmd.Flags().String("subject-type", "", "Filter by subject type (class name)")
	cmd.Flags().String("parent", "", "Filter by parent incident ID")
	cmd.Flags().String("has-parent", "", "Filter by has parent (true/false)")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("customer", "", "Filter by customer ID")
	cmd.Flags().String("developer", "", "Filter by developer ID")
	cmd.Flags().String("trucker", "", "Filter by trucker ID")
	cmd.Flags().String("contractor", "", "Filter by contractor ID")
	cmd.Flags().String("material-supplier", "", "Filter by material supplier ID")
	cmd.Flags().String("material-site", "", "Filter by material site ID")
	cmd.Flags().String("job-production-plan", "", "Filter by job production plan ID")
	cmd.Flags().String("job-production-plan-project", "", "Filter by job production plan project ID")
	cmd.Flags().String("equipment", "", "Filter by equipment ID")
	cmd.Flags().String("assignee", "", "Filter by assignee user ID")
	cmd.Flags().String("created-by", "", "Filter by created-by user ID")
	cmd.Flags().String("tender-job-schedule-shift", "", "Filter by tender job schedule shift ID")
	cmd.Flags().String("tender-job-schedule-shift-driver", "", "Filter by tender job schedule shift driver user ID")
	cmd.Flags().String("tender-job-schedule-shift-trucker", "", "Filter by tender job schedule shift trucker ID")
	cmd.Flags().String("job-number", "", "Filter by job number (partial match)")
	cmd.Flags().String("has-equipment", "", "Filter by has equipment (true/false)")
	cmd.Flags().String("has-live-action-items", "", "Filter by has live action items (true/false)")
	cmd.Flags().String("incident-tag", "", "Filter by incident tag ID")
	cmd.Flags().String("incident-tag-slug", "", "Filter by incident tag slug")
	cmd.Flags().String("zero-incident-tags", "", "Filter by zero incident tags (true/false)")
	cmd.Flags().String("root-causes", "", "Filter by root cause IDs (comma-separated)")
	cmd.Flags().String("action-items", "", "Filter by action item IDs (comma-separated)")
	cmd.Flags().String("notifiable-to", "", "Filter by notifiable user ID")
	cmd.Flags().String("user-has-stake", "", "Filter by user ID with stake")
	cmd.Flags().String("responsible-person", "", "Filter by responsible person user ID")
	cmd.Flags().String("start-on", "", "Filter by start date (YYYY-MM-DD)")
	cmd.Flags().String("start-on-min", "", "Filter by minimum start date (YYYY-MM-DD)")
	cmd.Flags().String("start-on-max", "", "Filter by maximum start date (YYYY-MM-DD)")
	cmd.Flags().String("start-at-min", "", "Filter by minimum start timestamp (ISO 8601)")
	cmd.Flags().String("start-at-max", "", "Filter by maximum start timestamp (ISO 8601)")
	cmd.Flags().String("end-at-min", "", "Filter by minimum end timestamp (ISO 8601)")
	cmd.Flags().String("end-at-max", "", "Filter by maximum end timestamp (ISO 8601)")
	cmd.Flags().String("net-impact-minutes-min", "", "Filter by minimum net impact minutes")
	cmd.Flags().String("net-impact-minutes-max", "", "Filter by maximum net impact minutes")
	cmd.Flags().String("net-impact-dollars-min", "", "Filter by minimum net impact dollars")
	cmd.Flags().String("net-impact-dollars-max", "", "Filter by maximum net impact dollars")
	cmd.Flags().String("q", "", "Search query")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProductionIncidentsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProductionIncidentsListOptions(cmd)
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
	query.Set("fields[production-incidents]", "status,kind,severity,time-value-type,start-at,net-impact-minutes,net-impact-dollars,is-down-time,subject")

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
	setFilterIfPresent(query, "filter[did_stop_work]", opts.DidStopWork)
	setFilterIfPresent(query, "filter[natures]", opts.Natures)
	setFilterIfPresent(query, "filter[subject]", opts.Subject)
	setFilterIfPresent(query, "filter[subject_type]", opts.SubjectType)
	setFilterIfPresent(query, "filter[parent]", opts.Parent)
	setFilterIfPresent(query, "filter[has_parent]", opts.HasParent)
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
	setFilterIfPresent(query, "filter[tender_job_schedule_shift]", opts.TenderJobScheduleShift)
	setFilterIfPresent(query, "filter[tender_job_schedule_shift_driver]", opts.TenderJobScheduleShiftDriver)
	setFilterIfPresent(query, "filter[tender_job_schedule_shift_trucker]", opts.TenderJobScheduleShiftTrucker)
	setFilterIfPresent(query, "filter[job_number]", opts.JobNumber)
	setFilterIfPresent(query, "filter[has_equipment]", opts.HasEquipment)
	setFilterIfPresent(query, "filter[has_live_action_items]", opts.HasLiveActionItems)
	setFilterIfPresent(query, "filter[incident_tag]", opts.IncidentTag)
	setFilterIfPresent(query, "filter[incident_tag_slug]", opts.IncidentTagSlug)
	setFilterIfPresent(query, "filter[zero_incident_tags]", opts.ZeroIncidentTags)
	setFilterIfPresent(query, "filter[root_causes]", opts.RootCauses)
	setFilterIfPresent(query, "filter[action_items]", opts.ActionItems)
	setFilterIfPresent(query, "filter[notifiable_to]", opts.NotifiableTo)
	setFilterIfPresent(query, "filter[user_has_stake]", opts.UserHasStake)
	setFilterIfPresent(query, "filter[responsible_person]", opts.ResponsiblePerson)
	setFilterIfPresent(query, "filter[start_on]", opts.StartOn)
	setFilterIfPresent(query, "filter[start_on_min]", opts.StartOnMin)
	setFilterIfPresent(query, "filter[start_on_max]", opts.StartOnMax)
	setFilterIfPresent(query, "filter[start_at_min]", opts.StartAtMin)
	setFilterIfPresent(query, "filter[start_at_max]", opts.StartAtMax)
	setFilterIfPresent(query, "filter[end_at_min]", opts.EndAtMin)
	setFilterIfPresent(query, "filter[end_at_max]", opts.EndAtMax)
	setFilterIfPresent(query, "filter[net_impact_minutes_min]", opts.NetImpactMinutesMin)
	setFilterIfPresent(query, "filter[net_impact_minutes_max]", opts.NetImpactMinutesMax)
	setFilterIfPresent(query, "filter[net_impact_dollars_min]", opts.NetImpactDollarsMin)
	setFilterIfPresent(query, "filter[net_impact_dollars_max]", opts.NetImpactDollarsMax)
	setFilterIfPresent(query, "filter[q]", opts.Q)

	body, _, err := client.Get(cmd.Context(), "/v1/production-incidents", query)
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

	handled, err := renderSparseListIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	rows := buildProductionIncidentRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProductionIncidentsTable(cmd, rows)
}

func parseProductionIncidentsListOptions(cmd *cobra.Command) (productionIncidentsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	status, _ := cmd.Flags().GetString("status")
	kind, _ := cmd.Flags().GetString("kind")
	severity, _ := cmd.Flags().GetString("severity")
	didStopWork, _ := cmd.Flags().GetString("did-stop-work")
	natures, _ := cmd.Flags().GetString("natures")
	subject, _ := cmd.Flags().GetString("subject")
	subjectType, _ := cmd.Flags().GetString("subject-type")
	parent, _ := cmd.Flags().GetString("parent")
	hasParent, _ := cmd.Flags().GetString("has-parent")
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
	tenderJobScheduleShift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	tenderJobScheduleShiftDriver, _ := cmd.Flags().GetString("tender-job-schedule-shift-driver")
	tenderJobScheduleShiftTrucker, _ := cmd.Flags().GetString("tender-job-schedule-shift-trucker")
	jobNumber, _ := cmd.Flags().GetString("job-number")
	hasEquipment, _ := cmd.Flags().GetString("has-equipment")
	hasLiveActionItems, _ := cmd.Flags().GetString("has-live-action-items")
	incidentTag, _ := cmd.Flags().GetString("incident-tag")
	incidentTagSlug, _ := cmd.Flags().GetString("incident-tag-slug")
	zeroIncidentTags, _ := cmd.Flags().GetString("zero-incident-tags")
	rootCauses, _ := cmd.Flags().GetString("root-causes")
	actionItems, _ := cmd.Flags().GetString("action-items")
	notifiableTo, _ := cmd.Flags().GetString("notifiable-to")
	userHasStake, _ := cmd.Flags().GetString("user-has-stake")
	responsiblePerson, _ := cmd.Flags().GetString("responsible-person")
	startOn, _ := cmd.Flags().GetString("start-on")
	startOnMin, _ := cmd.Flags().GetString("start-on-min")
	startOnMax, _ := cmd.Flags().GetString("start-on-max")
	startAtMin, _ := cmd.Flags().GetString("start-at-min")
	startAtMax, _ := cmd.Flags().GetString("start-at-max")
	endAtMin, _ := cmd.Flags().GetString("end-at-min")
	endAtMax, _ := cmd.Flags().GetString("end-at-max")
	netImpactMinutesMin, _ := cmd.Flags().GetString("net-impact-minutes-min")
	netImpactMinutesMax, _ := cmd.Flags().GetString("net-impact-minutes-max")
	netImpactDollarsMin, _ := cmd.Flags().GetString("net-impact-dollars-min")
	netImpactDollarsMax, _ := cmd.Flags().GetString("net-impact-dollars-max")
	q, _ := cmd.Flags().GetString("q")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return productionIncidentsListOptions{
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
		DidStopWork:                   didStopWork,
		Natures:                       natures,
		Subject:                       subject,
		SubjectType:                   subjectType,
		Parent:                        parent,
		HasParent:                     hasParent,
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
		TenderJobScheduleShift:        tenderJobScheduleShift,
		TenderJobScheduleShiftDriver:  tenderJobScheduleShiftDriver,
		TenderJobScheduleShiftTrucker: tenderJobScheduleShiftTrucker,
		JobNumber:                     jobNumber,
		HasEquipment:                  hasEquipment,
		HasLiveActionItems:            hasLiveActionItems,
		IncidentTag:                   incidentTag,
		IncidentTagSlug:               incidentTagSlug,
		ZeroIncidentTags:              zeroIncidentTags,
		RootCauses:                    rootCauses,
		ActionItems:                   actionItems,
		NotifiableTo:                  notifiableTo,
		UserHasStake:                  userHasStake,
		ResponsiblePerson:             responsiblePerson,
		StartOn:                       startOn,
		StartOnMin:                    startOnMin,
		StartOnMax:                    startOnMax,
		StartAtMin:                    startAtMin,
		StartAtMax:                    startAtMax,
		EndAtMin:                      endAtMin,
		EndAtMax:                      endAtMax,
		NetImpactMinutesMin:           netImpactMinutesMin,
		NetImpactMinutesMax:           netImpactMinutesMax,
		NetImpactDollarsMin:           netImpactDollarsMin,
		NetImpactDollarsMax:           netImpactDollarsMax,
		Q:                             q,
	}, nil
}

func buildProductionIncidentRows(resp jsonAPIResponse) []productionIncidentRow {
	rows := make([]productionIncidentRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := buildProductionIncidentRow(resource)
		rows = append(rows, row)
	}
	return rows
}

func buildProductionIncidentRow(resource jsonAPIResource) productionIncidentRow {
	row := productionIncidentRow{
		ID:               resource.ID,
		Status:           stringAttr(resource.Attributes, "status"),
		Kind:             stringAttr(resource.Attributes, "kind"),
		Severity:         stringAttr(resource.Attributes, "severity"),
		TimeValueType:    stringAttr(resource.Attributes, "time-value-type"),
		StartAt:          stringAttr(resource.Attributes, "start-at"),
		NetImpactMinutes: stringAttr(resource.Attributes, "net-impact-minutes"),
		NetImpactDollars: stringAttr(resource.Attributes, "net-impact-dollars"),
		IsDownTime:       boolAttr(resource.Attributes, "is-down-time"),
	}

	if rel, ok := resource.Relationships["subject"]; ok && rel.Data != nil {
		row.SubjectType = rel.Data.Type
		row.SubjectID = rel.Data.ID
	}

	return row
}

func buildProductionIncidentRowFromSingle(resp jsonAPISingleResponse) productionIncidentRow {
	return buildProductionIncidentRow(resp.Data)
}

func renderProductionIncidentsTable(cmd *cobra.Command, rows []productionIncidentRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No production incidents found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tKIND\tSEVERITY\tTIME TYPE\tNET MIN\tNET $\tDOWN TIME\tSTART AT\tSUBJECT")
	for _, row := range rows {
		subject := formatIncidentReference(row.SubjectType, row.SubjectID)
		downTime := "no"
		if row.IsDownTime {
			downTime = "yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Status,
			row.Kind,
			row.Severity,
			row.TimeValueType,
			row.NetImpactMinutes,
			row.NetImpactDollars,
			downTime,
			row.StartAt,
			subject,
		)
	}
	return writer.Flush()
}
