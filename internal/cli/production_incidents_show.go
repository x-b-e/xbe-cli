package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type productionIncidentsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type productionIncidentDetails struct {
	ID                                           string   `json:"id"`
	Status                                       string   `json:"status,omitempty"`
	Kind                                         string   `json:"kind,omitempty"`
	Severity                                     string   `json:"severity,omitempty"`
	TimeValueType                                string   `json:"time_value_type,omitempty"`
	Headline                                     string   `json:"headline,omitempty"`
	Description                                  string   `json:"description,omitempty"`
	StartAt                                      string   `json:"start_at,omitempty"`
	EndAt                                        string   `json:"end_at,omitempty"`
	Natures                                      []string `json:"natures,omitempty"`
	DidStopWork                                  bool     `json:"did_stop_work,omitempty"`
	NetImpactMinutes                             string   `json:"net_impact_minutes,omitempty"`
	NetImpactDollars                             string   `json:"net_impact_dollars,omitempty"`
	IsDownTime                                   bool     `json:"is_down_time,omitempty"`
	CurrentUserCanUpdate                         bool     `json:"current_user_can_update,omitempty"`
	TimeZoneID                                   string   `json:"time_zone_id,omitempty"`
	CreatedByBot                                 bool     `json:"created_by_bot,omitempty"`
	IsCreatedByJobProductionPlanTruckingDetector bool     `json:"is_created_by_job_production_plan_trucking_incident_detector,omitempty"`
	SubjectType                                  string   `json:"subject_type,omitempty"`
	SubjectID                                    string   `json:"subject_id,omitempty"`
	ParentType                                   string   `json:"parent_type,omitempty"`
	ParentID                                     string   `json:"parent_id,omitempty"`
	Children                                     []string `json:"children,omitempty"`
	EquipmentID                                  string   `json:"equipment_id,omitempty"`
	JobProductionPlanID                          string   `json:"job_production_plan_id,omitempty"`
	AssigneeID                                   string   `json:"assignee_id,omitempty"`
	TenderJobScheduleShiftID                     string   `json:"tender_job_schedule_shift_id,omitempty"`
	CreatedByID                                  string   `json:"created_by_id,omitempty"`
	IncidentRequestID                            string   `json:"incident_request_id,omitempty"`
	BrokerID                                     string   `json:"broker_id,omitempty"`
	DeveloperID                                  string   `json:"developer_id,omitempty"`
	IncidentUnitOfMeasureQuantities              []string `json:"incident_unit_of_measure_quantities,omitempty"`
	RootCauses                                   []string `json:"root_causes,omitempty"`
	ActionItems                                  []string `json:"action_items,omitempty"`
	IncidentParticipants                         []string `json:"incident_participants,omitempty"`
	IncidentHeadlineSuggestions                  []string `json:"incident_headline_suggestions,omitempty"`
	IncidentTagIncidents                         []string `json:"incident_tag_incidents,omitempty"`
	IncidentTags                                 []string `json:"incident_tags,omitempty"`
}

func newProductionIncidentsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show production incident details",
		Long: `Show full details of a production incident, including key relationships.

Arguments:
  <id>    Production incident ID (required)

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a production incident
  xbe view production-incidents show 123

  # JSON output
  xbe view production-incidents show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProductionIncidentsShow,
	}
	initProductionIncidentsShowFlags(cmd)
	return cmd
}

func init() {
	productionIncidentsCmd.AddCommand(newProductionIncidentsShowCmd())
}

func initProductionIncidentsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProductionIncidentsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseProductionIncidentsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("production incident id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[production-incidents]", "start-at,end-at,status,kind,description,natures,severity,time-value-type,did-stop-work,headline,net-impact-minutes,net-impact-dollars,is-down-time,current-user-can-update,time-zone-id,created-by-bot,is-created-by-job-production-plan-trucking-incident-detector,subject,parent,children,equipment,job-production-plan,assignee,tender-job-schedule-shift,created-by,incident-request,broker,developer,incident-unit-of-measure-quantities,root-causes,action-items,incident-participants,incident-headline-suggestions,incident-tag-incidents,incident-tags")
	query.Set("include", strings.Join([]string{
		"subject",
		"parent",
		"children",
		"equipment",
		"job-production-plan",
		"assignee",
		"tender-job-schedule-shift",
		"created-by",
		"incident-request",
		"broker",
		"developer",
		"incident-unit-of-measure-quantities",
		"root-causes",
		"action-items",
		"incident-participants",
		"incident-headline-suggestions",
		"incident-tag-incidents",
		"incident-tags",
	}, ","))

	body, _, err := client.Get(cmd.Context(), "/v1/production-incidents/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildProductionIncidentDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProductionIncidentDetails(cmd, details)
}

func parseProductionIncidentsShowOptions(cmd *cobra.Command) (productionIncidentsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return productionIncidentsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProductionIncidentDetails(resp jsonAPISingleResponse) productionIncidentDetails {
	attrs := resp.Data.Attributes
	details := productionIncidentDetails{
		ID:                   resp.Data.ID,
		Status:               stringAttr(attrs, "status"),
		Kind:                 stringAttr(attrs, "kind"),
		Severity:             stringAttr(attrs, "severity"),
		TimeValueType:        stringAttr(attrs, "time-value-type"),
		Headline:             stringAttr(attrs, "headline"),
		Description:          stringAttr(attrs, "description"),
		StartAt:              stringAttr(attrs, "start-at"),
		EndAt:                stringAttr(attrs, "end-at"),
		Natures:              stringSliceAttr(attrs, "natures"),
		DidStopWork:          boolAttr(attrs, "did-stop-work"),
		NetImpactMinutes:     stringAttr(attrs, "net-impact-minutes"),
		NetImpactDollars:     stringAttr(attrs, "net-impact-dollars"),
		IsDownTime:           boolAttr(attrs, "is-down-time"),
		CurrentUserCanUpdate: boolAttr(attrs, "current-user-can-update"),
		TimeZoneID:           stringAttr(attrs, "time-zone-id"),
		CreatedByBot:         boolAttr(attrs, "created-by-bot"),
		IsCreatedByJobProductionPlanTruckingDetector: boolAttr(attrs, "is-created-by-job-production-plan-trucking-incident-detector"),
	}

	if rel, ok := resp.Data.Relationships["subject"]; ok && rel.Data != nil {
		details.SubjectType = rel.Data.Type
		details.SubjectID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["parent"]; ok && rel.Data != nil {
		details.ParentType = rel.Data.Type
		details.ParentID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["equipment"]; ok && rel.Data != nil {
		details.EquipmentID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["job-production-plan"]; ok && rel.Data != nil {
		details.JobProductionPlanID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["assignee"]; ok && rel.Data != nil {
		details.AssigneeID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["tender-job-schedule-shift"]; ok && rel.Data != nil {
		details.TenderJobScheduleShiftID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["incident-request"]; ok && rel.Data != nil {
		details.IncidentRequestID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["developer"]; ok && rel.Data != nil {
		details.DeveloperID = rel.Data.ID
	}

	details.Children = relationshipReferencesProduction(resp.Data.Relationships["children"])
	details.IncidentUnitOfMeasureQuantities = relationshipReferencesProduction(resp.Data.Relationships["incident-unit-of-measure-quantities"])
	details.RootCauses = relationshipReferencesProduction(resp.Data.Relationships["root-causes"])
	details.ActionItems = relationshipReferencesProduction(resp.Data.Relationships["action-items"])
	details.IncidentParticipants = relationshipReferencesProduction(resp.Data.Relationships["incident-participants"])
	details.IncidentHeadlineSuggestions = relationshipReferencesProduction(resp.Data.Relationships["incident-headline-suggestions"])
	details.IncidentTagIncidents = relationshipReferencesProduction(resp.Data.Relationships["incident-tag-incidents"])
	details.IncidentTags = relationshipReferencesProduction(resp.Data.Relationships["incident-tags"])

	return details
}

func renderProductionIncidentDetails(cmd *cobra.Command, details productionIncidentDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	fmt.Fprintf(out, "Status: %s\n", formatOptional(details.Status))
	fmt.Fprintf(out, "Kind: %s\n", formatOptional(details.Kind))
	fmt.Fprintf(out, "Severity: %s\n", formatOptional(details.Severity))
	fmt.Fprintf(out, "Time Value Type: %s\n", formatOptional(details.TimeValueType))
	fmt.Fprintf(out, "Headline: %s\n", formatOptional(details.Headline))
	fmt.Fprintf(out, "Description: %s\n", formatOptional(details.Description))
	fmt.Fprintf(out, "Start At: %s\n", formatOptional(details.StartAt))
	fmt.Fprintf(out, "End At: %s\n", formatOptional(details.EndAt))
	fmt.Fprintf(out, "Natures: %s\n", formatOptional(strings.Join(details.Natures, ", ")))
	fmt.Fprintf(out, "Did Stop Work: %s\n", yesNo(details.DidStopWork))
	fmt.Fprintf(out, "Net Impact Minutes: %s\n", formatOptional(details.NetImpactMinutes))
	fmt.Fprintf(out, "Net Impact Dollars: %s\n", formatOptional(details.NetImpactDollars))
	fmt.Fprintf(out, "Down Time: %s\n", yesNo(details.IsDownTime))
	fmt.Fprintf(out, "Current User Can Update: %s\n", yesNo(details.CurrentUserCanUpdate))
	fmt.Fprintf(out, "Time Zone ID: %s\n", formatOptional(details.TimeZoneID))
	fmt.Fprintf(out, "Created By Bot: %s\n", yesNo(details.CreatedByBot))
	fmt.Fprintf(out, "Created By JPP Trucking Incident Detector: %s\n", yesNo(details.IsCreatedByJobProductionPlanTruckingDetector))
	fmt.Fprintf(out, "Subject: %s\n", formatOptional(formatIncidentReference(details.SubjectType, details.SubjectID)))
	fmt.Fprintf(out, "Parent: %s\n", formatOptional(formatIncidentReference(details.ParentType, details.ParentID)))
	fmt.Fprintf(out, "Children: %s\n", formatOptional(strings.Join(details.Children, ", ")))
	fmt.Fprintf(out, "Equipment: %s\n", formatOptional(details.EquipmentID))
	fmt.Fprintf(out, "Job Production Plan: %s\n", formatOptional(details.JobProductionPlanID))
	fmt.Fprintf(out, "Assignee: %s\n", formatOptional(details.AssigneeID))
	fmt.Fprintf(out, "Tender Job Schedule Shift: %s\n", formatOptional(details.TenderJobScheduleShiftID))
	fmt.Fprintf(out, "Created By: %s\n", formatOptional(details.CreatedByID))
	fmt.Fprintf(out, "Incident Request: %s\n", formatOptional(details.IncidentRequestID))
	fmt.Fprintf(out, "Broker: %s\n", formatOptional(details.BrokerID))
	fmt.Fprintf(out, "Developer: %s\n", formatOptional(details.DeveloperID))
	fmt.Fprintf(out, "Incident Unit Of Measure Quantities: %s\n", formatOptional(strings.Join(details.IncidentUnitOfMeasureQuantities, ", ")))
	fmt.Fprintf(out, "Root Causes: %s\n", formatOptional(strings.Join(details.RootCauses, ", ")))
	fmt.Fprintf(out, "Action Items: %s\n", formatOptional(strings.Join(details.ActionItems, ", ")))
	fmt.Fprintf(out, "Incident Participants: %s\n", formatOptional(strings.Join(details.IncidentParticipants, ", ")))
	fmt.Fprintf(out, "Incident Headline Suggestions: %s\n", formatOptional(strings.Join(details.IncidentHeadlineSuggestions, ", ")))
	fmt.Fprintf(out, "Incident Tag Incidents: %s\n", formatOptional(strings.Join(details.IncidentTagIncidents, ", ")))
	fmt.Fprintf(out, "Incident Tags: %s\n", formatOptional(strings.Join(details.IncidentTags, ", ")))

	return nil
}

func relationshipReferencesProduction(rel jsonAPIRelationship) []string {
	if len(rel.raw) == 0 {
		return nil
	}

	var ids []jsonAPIResourceIdentifier
	if err := json.Unmarshal(rel.raw, &ids); err != nil {
		return nil
	}

	values := make([]string, 0, len(ids))
	for _, item := range ids {
		values = append(values, formatIncidentReference(item.Type, item.ID))
	}
	return values
}
