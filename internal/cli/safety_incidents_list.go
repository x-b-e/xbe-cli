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

type safetyIncidentsListOptions struct {
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
	StartOn                       string
	StartOnMin                    string
	StartOnMax                    string
	StartAt                       string
	StartAtMin                    string
	StartAtMax                    string
	EndAt                         string
	EndAtMin                      string
	EndAtMax                      string
	Subject                       string
	SubjectType                   string
	SubjectID                     string
	Parent                        string
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
	DidStopWork                   string
	Natures                       string
	Q                             string
	NetImpactTons                 string
	NetImpactTonsMin              string
	NetImpactTonsMax              string
}

type safetyIncidentRow struct {
	ID            string `json:"id"`
	Status        string `json:"status,omitempty"`
	Kind          string `json:"kind,omitempty"`
	Severity      string `json:"severity,omitempty"`
	Headline      string `json:"headline,omitempty"`
	StartAt       string `json:"start_at,omitempty"`
	EndAt         string `json:"end_at,omitempty"`
	NetImpactTons string `json:"net_impact_tons,omitempty"`
	SubjectType   string `json:"subject_type,omitempty"`
	SubjectID     string `json:"subject_id,omitempty"`
}

func newSafetyIncidentsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List safety incidents",
		Long: `List safety incidents with filtering and pagination.

Output Columns:
  ID             Safety incident identifier
  STATUS         Current status
  KIND           Incident kind
  SEVERITY       Severity level
  NET IMPACT     Net impact tons (overloading only)
  HEADLINE       Brief description
  START AT       Start time
  SUBJECT        Subject type and ID

Filters:
  --status                      Filter by status
  --kind                        Filter by kind
  --severity                    Filter by severity
  --broker                      Filter by broker ID
  --customer                    Filter by customer ID
  --developer                   Filter by developer ID
  --trucker                     Filter by trucker ID
  --contractor                  Filter by contractor ID
  --material-supplier           Filter by material supplier ID
  --material-site               Filter by material site ID
  --job-production-plan         Filter by job production plan ID
  --job-production-plan-project Filter by job production plan project ID
  --equipment                   Filter by equipment ID
  --assignee                    Filter by assignee user ID
  --created-by                  Filter by created-by user ID
  --start-on                    Filter by start date (YYYY-MM-DD)
  --start-on-min                Filter by minimum start date
  --start-on-max                Filter by maximum start date
  --start-at                    Filter by start timestamp
  --start-at-min                Filter by minimum start timestamp
  --start-at-max                Filter by maximum start timestamp
  --end-at                      Filter by end timestamp
  --end-at-min                  Filter by minimum end timestamp
  --end-at-max                  Filter by maximum end timestamp
  --subject                     Filter by subject (Type|ID, e.g., Broker|123)
  --subject-type                Filter by subject type (e.g., Broker, JobProductionPlan)
  --subject-id                  Filter by subject (Type|ID, e.g., Broker|123)
  --parent                      Filter by parent incident ID
  --has-parent                  Filter by presence of parent (true/false)
  --has-equipment               Filter by presence of equipment (true/false)
  --has-live-action-items       Filter by presence of live action items (true/false)
  --incident-tag                Filter by incident tag ID
  --incident-tag-slug           Filter by incident tag slug
  --zero-incident-tags          Filter by zero incident tags (true/false)
  --root-causes                 Filter by root cause IDs (comma-separated)
  --action-items                Filter by action item IDs (comma-separated)
  --tender-job-schedule-shift   Filter by tender job schedule shift ID
  --tender-job-schedule-shift-driver  Filter by tender job schedule shift driver ID
  --tender-job-schedule-shift-trucker Filter by tender job schedule shift trucker ID
  --job-number                  Filter by job number
  --notifiable-to               Filter by notifiable user ID
  --user-has-stake              Filter by user stake (user ID)
  --responsible-person          Filter by responsible person (user ID)
  --did-stop-work               Filter by did stop work (true/false)
  --natures                     Filter by incident natures (comma-separated: personal,property)
  --q                           Search query
  --net-impact-tons             Filter by net impact tons
  --net-impact-tons-min         Filter by minimum net impact tons
  --net-impact-tons-max         Filter by maximum net impact tons

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List safety incidents
  xbe view safety-incidents list

  # Filter by status
  xbe view safety-incidents list --status open

  # Filter by kind
  xbe view safety-incidents list --kind near_miss

  # Filter by subject
  xbe view safety-incidents list --subject "Broker|123"

  # Filter by net impact tons
  xbe view safety-incidents list --net-impact-tons-min 10

  # Output as JSON
  xbe view safety-incidents list --json`,
		RunE: runSafetyIncidentsList,
	}
	initSafetyIncidentsListFlags(cmd)
	return cmd
}

func init() {
	safetyIncidentsCmd.AddCommand(newSafetyIncidentsListCmd())
}

func initSafetyIncidentsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort order (prefix with - for descending)")
	cmd.Flags().String("status", "", "Filter by status")
	cmd.Flags().String("kind", "", "Filter by kind")
	cmd.Flags().String("severity", "", "Filter by severity")
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
	cmd.Flags().String("start-on", "", "Filter by start date (YYYY-MM-DD)")
	cmd.Flags().String("start-on-min", "", "Filter by minimum start date (YYYY-MM-DD)")
	cmd.Flags().String("start-on-max", "", "Filter by maximum start date (YYYY-MM-DD)")
	cmd.Flags().String("start-at", "", "Filter by start timestamp")
	cmd.Flags().String("start-at-min", "", "Filter by minimum start timestamp")
	cmd.Flags().String("start-at-max", "", "Filter by maximum start timestamp")
	cmd.Flags().String("end-at", "", "Filter by end timestamp")
	cmd.Flags().String("end-at-min", "", "Filter by minimum end timestamp")
	cmd.Flags().String("end-at-max", "", "Filter by maximum end timestamp")
	cmd.Flags().String("subject", "", "Filter by subject (Type|ID, e.g., Broker|123)")
	cmd.Flags().String("subject-type", "", "Filter by subject type (e.g., Broker, JobProductionPlan)")
	cmd.Flags().String("subject-id", "", "Filter by subject (Type|ID, e.g., Broker|123)")
	cmd.Flags().String("parent", "", "Filter by parent incident ID")
	cmd.Flags().String("has-parent", "", "Filter by presence of parent (true/false)")
	cmd.Flags().String("has-equipment", "", "Filter by presence of equipment (true/false)")
	cmd.Flags().String("has-live-action-items", "", "Filter by presence of live action items (true/false)")
	cmd.Flags().String("incident-tag", "", "Filter by incident tag ID")
	cmd.Flags().String("incident-tag-slug", "", "Filter by incident tag slug")
	cmd.Flags().String("zero-incident-tags", "", "Filter by zero incident tags (true/false)")
	cmd.Flags().String("root-causes", "", "Filter by root cause IDs (comma-separated)")
	cmd.Flags().String("action-items", "", "Filter by action item IDs (comma-separated)")
	cmd.Flags().String("tender-job-schedule-shift", "", "Filter by tender job schedule shift ID")
	cmd.Flags().String("tender-job-schedule-shift-driver", "", "Filter by tender job schedule shift driver ID")
	cmd.Flags().String("tender-job-schedule-shift-trucker", "", "Filter by tender job schedule shift trucker ID")
	cmd.Flags().String("job-number", "", "Filter by job number")
	cmd.Flags().String("notifiable-to", "", "Filter by notifiable user ID")
	cmd.Flags().String("user-has-stake", "", "Filter by user stake (user ID)")
	cmd.Flags().String("responsible-person", "", "Filter by responsible person (user ID)")
	cmd.Flags().String("did-stop-work", "", "Filter by did stop work (true/false)")
	cmd.Flags().String("natures", "", "Filter by natures (comma-separated: personal,property)")
	cmd.Flags().String("q", "", "Search query")
	cmd.Flags().String("net-impact-tons", "", "Filter by net impact tons")
	cmd.Flags().String("net-impact-tons-min", "", "Filter by minimum net impact tons")
	cmd.Flags().String("net-impact-tons-max", "", "Filter by maximum net impact tons")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runSafetyIncidentsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseSafetyIncidentsListOptions(cmd)
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
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}
	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
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
	setFilterIfPresent(query, "filter[start_on]", opts.StartOn)
	setFilterIfPresent(query, "filter[start_on_min]", opts.StartOnMin)
	setFilterIfPresent(query, "filter[start_on_max]", opts.StartOnMax)
	setFilterIfPresent(query, "filter[start_at_min]", opts.StartAtMin)
	setFilterIfPresent(query, "filter[start_at_max]", opts.StartAtMax)
	if opts.StartAt != "" {
		if opts.StartAtMin == "" {
			query.Set("filter[start_at_min]", opts.StartAt)
		}
		if opts.StartAtMax == "" {
			query.Set("filter[start_at_max]", opts.StartAt)
		}
	}
	setFilterIfPresent(query, "filter[end_at_min]", opts.EndAtMin)
	setFilterIfPresent(query, "filter[end_at_max]", opts.EndAtMax)
	if opts.EndAt != "" {
		if opts.EndAtMin == "" {
			query.Set("filter[end_at_min]", opts.EndAt)
		}
		if opts.EndAtMax == "" {
			query.Set("filter[end_at_max]", opts.EndAt)
		}
	}
	setFilterIfPresent(query, "filter[subject]", opts.Subject)
	setFilterIfPresent(query, "filter[subject_type]", opts.SubjectType)
	setFilterIfPresent(query, "filter[subject_id]", opts.SubjectID)
	setFilterIfPresent(query, "filter[parent]", opts.Parent)
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
	setFilterIfPresent(query, "filter[did_stop_work]", opts.DidStopWork)
	setFilterIfPresent(query, "filter[natures]", opts.Natures)
	setFilterIfPresent(query, "filter[q]", opts.Q)
	setFilterIfPresent(query, "filter[net_impact_tons_min]", opts.NetImpactTonsMin)
	setFilterIfPresent(query, "filter[net_impact_tons_max]", opts.NetImpactTonsMax)
	if opts.NetImpactTons != "" {
		if opts.NetImpactTonsMin == "" {
			query.Set("filter[net_impact_tons_min]", opts.NetImpactTons)
		}
		if opts.NetImpactTonsMax == "" {
			query.Set("filter[net_impact_tons_max]", opts.NetImpactTons)
		}
	}

	body, _, err := client.Get(cmd.Context(), "/v1/safety-incidents", query)
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

	rows := buildSafetyIncidentRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderSafetyIncidentsTable(cmd, rows)
}

func parseSafetyIncidentsListOptions(cmd *cobra.Command) (safetyIncidentsListOptions, error) {
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
	startOn, _ := cmd.Flags().GetString("start-on")
	startOnMin, _ := cmd.Flags().GetString("start-on-min")
	startOnMax, _ := cmd.Flags().GetString("start-on-max")
	startAt, _ := cmd.Flags().GetString("start-at")
	startAtMin, _ := cmd.Flags().GetString("start-at-min")
	startAtMax, _ := cmd.Flags().GetString("start-at-max")
	endAt, _ := cmd.Flags().GetString("end-at")
	endAtMin, _ := cmd.Flags().GetString("end-at-min")
	endAtMax, _ := cmd.Flags().GetString("end-at-max")
	subject, _ := cmd.Flags().GetString("subject")
	subjectType, _ := cmd.Flags().GetString("subject-type")
	subjectID, _ := cmd.Flags().GetString("subject-id")
	parent, _ := cmd.Flags().GetString("parent")
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
	didStopWork, _ := cmd.Flags().GetString("did-stop-work")
	natures, _ := cmd.Flags().GetString("natures")
	q, _ := cmd.Flags().GetString("q")
	netImpactTons, _ := cmd.Flags().GetString("net-impact-tons")
	netImpactTonsMin, _ := cmd.Flags().GetString("net-impact-tons-min")
	netImpactTonsMax, _ := cmd.Flags().GetString("net-impact-tons-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return safetyIncidentsListOptions{
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
		StartOn:                       startOn,
		StartOnMin:                    startOnMin,
		StartOnMax:                    startOnMax,
		StartAt:                       startAt,
		StartAtMin:                    startAtMin,
		StartAtMax:                    startAtMax,
		EndAt:                         endAt,
		EndAtMin:                      endAtMin,
		EndAtMax:                      endAtMax,
		Subject:                       subject,
		SubjectType:                   subjectType,
		SubjectID:                     subjectID,
		Parent:                        parent,
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
		DidStopWork:                   didStopWork,
		Natures:                       natures,
		Q:                             q,
		NetImpactTons:                 netImpactTons,
		NetImpactTonsMin:              netImpactTonsMin,
		NetImpactTonsMax:              netImpactTonsMax,
	}, nil
}

func buildSafetyIncidentRows(resp jsonAPIResponse) []safetyIncidentRow {
	rows := make([]safetyIncidentRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := safetyIncidentRow{
			ID:            resource.ID,
			Status:        stringAttr(resource.Attributes, "status"),
			Kind:          stringAttr(resource.Attributes, "kind"),
			Severity:      stringAttr(resource.Attributes, "severity"),
			Headline:      stringAttr(resource.Attributes, "headline"),
			StartAt:       stringAttr(resource.Attributes, "start-at"),
			EndAt:         stringAttr(resource.Attributes, "end-at"),
			NetImpactTons: numberAttrAsString(resource.Attributes, "net-impact-tons"),
		}

		if rel, ok := resource.Relationships["subject"]; ok && rel.Data != nil {
			row.SubjectType = rel.Data.Type
			row.SubjectID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderSafetyIncidentsTable(cmd *cobra.Command, rows []safetyIncidentRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No safety incidents found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tKIND\tSEVERITY\tNET IMPACT\tHEADLINE\tSTART AT\tSUBJECT")
	for _, row := range rows {
		subject := formatPolymorphic(row.SubjectType, row.SubjectID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Status,
			row.Kind,
			row.Severity,
			row.NetImpactTons,
			truncateString(row.Headline, 30),
			row.StartAt,
			subject,
		)
	}
	return writer.Flush()
}
