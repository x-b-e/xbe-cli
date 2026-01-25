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

type incidentsListOptions struct {
	BaseURL            string
	Token              string
	JSON               bool
	NoAuth             bool
	Limit              int
	Offset             int
	Status             string
	Kind               string
	Severity           string
	Broker             string
	Customer           string
	Developer          string
	Trucker            string
	JobProductionPlan  string
	Equipment          string
	Assignee           string
	CreatedBy          string
	StartOnMin         string
	StartOnMax         string
	StartAtMin         string
	StartAtMax         string
	EndAtMin           string
	EndAtMax           string
	SubjectType        string
	HasParent          string
	HasEquipment       string
	HasLiveActionItems string
	IncidentTag        string
	IncidentTagSlug    string
	Q                  string
}

type incidentRow struct {
	ID                  string `json:"id"`
	Type                string `json:"type,omitempty"`
	Status              string `json:"status,omitempty"`
	Kind                string `json:"kind,omitempty"`
	Severity            string `json:"severity,omitempty"`
	Headline            string `json:"headline,omitempty"`
	Description         string `json:"description,omitempty"`
	StartAt             string `json:"start_at,omitempty"`
	EndAt               string `json:"end_at,omitempty"`
	SubjectType         string `json:"subject_type,omitempty"`
	SubjectID           string `json:"subject_id,omitempty"`
	BrokerID            string `json:"broker_id,omitempty"`
	AssigneeID          string `json:"assignee_id,omitempty"`
	CreatedByID         string `json:"created_by_id,omitempty"`
	JobProductionPlanID string `json:"job_production_plan_id,omitempty"`
}

func newIncidentsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List incidents",
		Long: `List incidents across the organization.

Output Columns:
  ID          Incident identifier
  TYPE        Incident type (safety, production, etc.)
  STATUS      Current status
  SEVERITY    Severity level
  HEADLINE    Brief description
  START AT    When the incident started
  SUBJECT     Subject type and ID

Common Filters:
  --status              Filter by status
  --kind                Filter by kind
  --severity            Filter by severity
  --broker              Filter by broker ID
  --customer            Filter by customer ID
  --developer           Filter by developer ID
  --trucker             Filter by trucker ID
  --job-production-plan Filter by job production plan ID
  --equipment           Filter by equipment ID
  --assignee            Filter by assignee user ID
  --created-by          Filter by created-by user ID
  --start-on-min        Filter by minimum start date
  --start-on-max        Filter by maximum start date
  --start-at-min        Filter by minimum start timestamp
  --start-at-max        Filter by maximum start timestamp
  --subject-type        Filter by subject type
  --has-parent          Filter by has parent (true/false)
  --has-equipment       Filter by has equipment (true/false)
  --has-live-action-items  Filter by has live action items (true/false)
  --incident-tag        Filter by incident tag ID
  --incident-tag-slug   Filter by incident tag slug
  --q                   Search query`,
		Example: `  # List all incidents
  xbe view incidents list

  # Filter by status
  xbe view incidents list --status open

  # Filter by broker
  xbe view incidents list --broker 123

  # Filter by severity
  xbe view incidents list --severity critical

  # Filter by date range
  xbe view incidents list --start-on-min 2024-01-01 --start-on-max 2024-12-31

  # Search incidents
  xbe view incidents list --q "delay"

  # Output as JSON
  xbe view incidents list --json`,
		RunE: runIncidentsList,
	}
	initIncidentsListFlags(cmd)
	return cmd
}

func init() {
	incidentsCmd.AddCommand(newIncidentsListCmd())
}

func initIncidentsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("status", "", "Filter by status")
	cmd.Flags().String("kind", "", "Filter by kind")
	cmd.Flags().String("severity", "", "Filter by severity")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("customer", "", "Filter by customer ID")
	cmd.Flags().String("developer", "", "Filter by developer ID")
	cmd.Flags().String("trucker", "", "Filter by trucker ID")
	cmd.Flags().String("job-production-plan", "", "Filter by job production plan ID")
	cmd.Flags().String("equipment", "", "Filter by equipment ID")
	cmd.Flags().String("assignee", "", "Filter by assignee user ID")
	cmd.Flags().String("created-by", "", "Filter by created-by user ID")
	cmd.Flags().String("start-on-min", "", "Filter by minimum start date")
	cmd.Flags().String("start-on-max", "", "Filter by maximum start date")
	cmd.Flags().String("start-at-min", "", "Filter by minimum start timestamp")
	cmd.Flags().String("start-at-max", "", "Filter by maximum start timestamp")
	cmd.Flags().String("end-at-min", "", "Filter by minimum end timestamp")
	cmd.Flags().String("end-at-max", "", "Filter by maximum end timestamp")
	cmd.Flags().String("subject-type", "", "Filter by subject type")
	cmd.Flags().String("has-parent", "", "Filter by has parent (true/false)")
	cmd.Flags().String("has-equipment", "", "Filter by has equipment (true/false)")
	cmd.Flags().String("has-live-action-items", "", "Filter by has live action items (true/false)")
	cmd.Flags().String("incident-tag", "", "Filter by incident tag ID")
	cmd.Flags().String("incident-tag-slug", "", "Filter by incident tag slug")
	cmd.Flags().String("q", "", "Search query")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runIncidentsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseIncidentsListOptions(cmd)
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

	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[kind]", opts.Kind)
	setFilterIfPresent(query, "filter[severity]", opts.Severity)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[customer]", opts.Customer)
	setFilterIfPresent(query, "filter[developer]", opts.Developer)
	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)
	setFilterIfPresent(query, "filter[job_production_plan]", opts.JobProductionPlan)
	setFilterIfPresent(query, "filter[equipment]", opts.Equipment)
	setFilterIfPresent(query, "filter[assignee]", opts.Assignee)
	setFilterIfPresent(query, "filter[created_by]", opts.CreatedBy)
	setFilterIfPresent(query, "filter[start_on_min]", opts.StartOnMin)
	setFilterIfPresent(query, "filter[start_on_max]", opts.StartOnMax)
	setFilterIfPresent(query, "filter[start_at_min]", opts.StartAtMin)
	setFilterIfPresent(query, "filter[start_at_max]", opts.StartAtMax)
	setFilterIfPresent(query, "filter[end_at_min]", opts.EndAtMin)
	setFilterIfPresent(query, "filter[end_at_max]", opts.EndAtMax)
	setFilterIfPresent(query, "filter[subject_type]", opts.SubjectType)
	setFilterIfPresent(query, "filter[has_parent]", opts.HasParent)
	setFilterIfPresent(query, "filter[has_equipment]", opts.HasEquipment)
	setFilterIfPresent(query, "filter[has_live_action_items]", opts.HasLiveActionItems)
	setFilterIfPresent(query, "filter[incident_tag]", opts.IncidentTag)
	setFilterIfPresent(query, "filter[incident_tag_slug]", opts.IncidentTagSlug)
	setFilterIfPresent(query, "filter[q]", opts.Q)

	body, _, err := client.Get(cmd.Context(), "/v1/incidents", query)
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

	rows := buildIncidentRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderIncidentsTable(cmd, rows)
}

func parseIncidentsListOptions(cmd *cobra.Command) (incidentsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	status, _ := cmd.Flags().GetString("status")
	kind, _ := cmd.Flags().GetString("kind")
	severity, _ := cmd.Flags().GetString("severity")
	broker, _ := cmd.Flags().GetString("broker")
	customer, _ := cmd.Flags().GetString("customer")
	developer, _ := cmd.Flags().GetString("developer")
	trucker, _ := cmd.Flags().GetString("trucker")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	equipment, _ := cmd.Flags().GetString("equipment")
	assignee, _ := cmd.Flags().GetString("assignee")
	createdBy, _ := cmd.Flags().GetString("created-by")
	startOnMin, _ := cmd.Flags().GetString("start-on-min")
	startOnMax, _ := cmd.Flags().GetString("start-on-max")
	startAtMin, _ := cmd.Flags().GetString("start-at-min")
	startAtMax, _ := cmd.Flags().GetString("start-at-max")
	endAtMin, _ := cmd.Flags().GetString("end-at-min")
	endAtMax, _ := cmd.Flags().GetString("end-at-max")
	subjectType, _ := cmd.Flags().GetString("subject-type")
	hasParent, _ := cmd.Flags().GetString("has-parent")
	hasEquipment, _ := cmd.Flags().GetString("has-equipment")
	hasLiveActionItems, _ := cmd.Flags().GetString("has-live-action-items")
	incidentTag, _ := cmd.Flags().GetString("incident-tag")
	incidentTagSlug, _ := cmd.Flags().GetString("incident-tag-slug")
	q, _ := cmd.Flags().GetString("q")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return incidentsListOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		NoAuth:             noAuth,
		Limit:              limit,
		Offset:             offset,
		Status:             status,
		Kind:               kind,
		Severity:           severity,
		Broker:             broker,
		Customer:           customer,
		Developer:          developer,
		Trucker:            trucker,
		JobProductionPlan:  jobProductionPlan,
		Equipment:          equipment,
		Assignee:           assignee,
		CreatedBy:          createdBy,
		StartOnMin:         startOnMin,
		StartOnMax:         startOnMax,
		StartAtMin:         startAtMin,
		StartAtMax:         startAtMax,
		EndAtMin:           endAtMin,
		EndAtMax:           endAtMax,
		SubjectType:        subjectType,
		HasParent:          hasParent,
		HasEquipment:       hasEquipment,
		HasLiveActionItems: hasLiveActionItems,
		IncidentTag:        incidentTag,
		IncidentTagSlug:    incidentTagSlug,
		Q:                  q,
	}, nil
}

func buildIncidentRows(resp jsonAPIResponse) []incidentRow {
	rows := make([]incidentRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := incidentRow{
			ID:          resource.ID,
			Type:        resource.Type,
			Status:      stringAttr(resource.Attributes, "status"),
			Kind:        stringAttr(resource.Attributes, "kind"),
			Severity:    stringAttr(resource.Attributes, "severity"),
			Headline:    stringAttr(resource.Attributes, "headline"),
			Description: stringAttr(resource.Attributes, "description"),
			StartAt:     stringAttr(resource.Attributes, "start-at"),
			EndAt:       stringAttr(resource.Attributes, "end-at"),
		}

		if rel, ok := resource.Relationships["subject"]; ok && rel.Data != nil {
			row.SubjectType = rel.Data.Type
			row.SubjectID = rel.Data.ID
		}

		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			row.BrokerID = rel.Data.ID
		}

		if rel, ok := resource.Relationships["assignee"]; ok && rel.Data != nil {
			row.AssigneeID = rel.Data.ID
		}

		if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
			row.CreatedByID = rel.Data.ID
		}

		if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
			row.JobProductionPlanID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderIncidentsTable(cmd *cobra.Command, rows []incidentRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No incidents found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTYPE\tSTATUS\tSEVERITY\tHEADLINE\tSTART AT\tSUBJECT")
	for _, row := range rows {
		subject := ""
		if row.SubjectType != "" && row.SubjectID != "" {
			subject = row.SubjectType + "/" + row.SubjectID
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Type,
			row.Status,
			row.Severity,
			truncateString(row.Headline, 30),
			row.StartAt,
			subject,
		)
	}
	return writer.Flush()
}
