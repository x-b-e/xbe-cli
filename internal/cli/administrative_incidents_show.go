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

type administrativeIncidentsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type administrativeIncidentDetails struct {
	ID                   string   `json:"id"`
	Type                 string   `json:"type,omitempty"`
	Status               string   `json:"status,omitempty"`
	Kind                 string   `json:"kind,omitempty"`
	Severity             string   `json:"severity,omitempty"`
	Headline             string   `json:"headline,omitempty"`
	Description          string   `json:"description,omitempty"`
	StartAt              string   `json:"start_at,omitempty"`
	EndAt                string   `json:"end_at,omitempty"`
	Natures              []string `json:"natures,omitempty"`
	DidStopWork          bool     `json:"did_stop_work"`
	TimeZoneID           string   `json:"time_zone_id,omitempty"`
	CurrentUserCanUpdate bool     `json:"current_user_can_update"`
	CreatedByBot         bool     `json:"created_by_bot"`
	NetImpactDollars     string   `json:"net_impact_dollars,omitempty"`

	SubjectType              string   `json:"subject_type,omitempty"`
	SubjectID                string   `json:"subject_id,omitempty"`
	ParentID                 string   `json:"parent_id,omitempty"`
	ChildIDs                 []string `json:"child_ids,omitempty"`
	EquipmentID              string   `json:"equipment_id,omitempty"`
	JobProductionPlanID      string   `json:"job_production_plan_id,omitempty"`
	AssigneeID               string   `json:"assignee_id,omitempty"`
	TenderJobScheduleShiftID string   `json:"tender_job_schedule_shift_id,omitempty"`
	CreatedByID              string   `json:"created_by_id,omitempty"`
	IncidentRequestID        string   `json:"incident_request_id,omitempty"`
	BrokerID                 string   `json:"broker_id,omitempty"`
	CustomerID               string   `json:"customer_id,omitempty"`
	TruckerID                string   `json:"trucker_id,omitempty"`
	DeveloperID              string   `json:"developer_id,omitempty"`
	ContractorID             string   `json:"contractor_id,omitempty"`
	MaterialSupplierID       string   `json:"material_supplier_id,omitempty"`
	MaterialSiteID           string   `json:"material_site_id,omitempty"`
	IncidentUnitOfMeasureIDs []string `json:"incident_unit_of_measure_quantity_ids,omitempty"`
	RootCauseIDs             []string `json:"root_cause_ids,omitempty"`
	ActionItemIDs            []string `json:"action_item_ids,omitempty"`
	IncidentParticipantIDs   []string `json:"incident_participant_ids,omitempty"`
	IncidentHeadlineIDs      []string `json:"incident_headline_suggestion_ids,omitempty"`
	IncidentTagIncidentIDs   []string `json:"incident_tag_incident_ids,omitempty"`
	IncidentTagIDs           []string `json:"incident_tag_ids,omitempty"`
	CommentIDs               []string `json:"comment_ids,omitempty"`
	AttachmentIDs            []string `json:"attachment_ids,omitempty"`
}

func newAdministrativeIncidentsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show administrative incident details",
		Long: `Show the full details of an administrative incident.

Output Fields:
  ID, type, status, kind, severity
  Headline, description, start/end timestamps
  Natures, did stop work, time zone
  Current user can update, created by bot
  Net impact dollars
  Subject, parent, children
  Related resource IDs (assignee, broker, customer, etc.)
  Related collections (root causes, action items, tags, participants)

Arguments:
  <id>    The administrative incident ID (required). Use the list command to find IDs.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show an administrative incident
  xbe view administrative-incidents show 123

  # JSON output
  xbe view administrative-incidents show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runAdministrativeIncidentsShow,
	}
	initAdministrativeIncidentsShowFlags(cmd)
	return cmd
}

func init() {
	administrativeIncidentsCmd.AddCommand(newAdministrativeIncidentsShowCmd())
}

func initAdministrativeIncidentsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runAdministrativeIncidentsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseAdministrativeIncidentsShowOptions(cmd)
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
		return fmt.Errorf("administrative incident id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("include", "subject,parent,children,equipment,job-production-plan,assignee,tender-job-schedule-shift,created-by,incident-request,broker,developer,incident-unit-of-measure-quantities,root-causes,action-items,incident-participants,incident-headline-suggestions,incident-tag-incidents,incident-tags,comments,file-attachments")

	body, _, err := client.Get(cmd.Context(), "/v1/administrative-incidents/"+id, query)
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

	details := buildAdministrativeIncidentDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderAdministrativeIncidentDetails(cmd, details)
}

func parseAdministrativeIncidentsShowOptions(cmd *cobra.Command) (administrativeIncidentsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return administrativeIncidentsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildAdministrativeIncidentDetails(resp jsonAPISingleResponse) administrativeIncidentDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := administrativeIncidentDetails{
		ID:                   resource.ID,
		Type:                 resource.Type,
		Status:               stringAttr(attrs, "status"),
		Kind:                 stringAttr(attrs, "kind"),
		Severity:             stringAttr(attrs, "severity"),
		Headline:             stringAttr(attrs, "headline"),
		Description:          stringAttr(attrs, "description"),
		StartAt:              stringAttr(attrs, "start-at"),
		EndAt:                stringAttr(attrs, "end-at"),
		Natures:              stringSliceAttr(attrs, "natures"),
		DidStopWork:          boolAttr(attrs, "did-stop-work"),
		TimeZoneID:           stringAttr(attrs, "time-zone-id"),
		CurrentUserCanUpdate: boolAttr(attrs, "current-user-can-update"),
		CreatedByBot:         boolAttr(attrs, "created-by-bot"),
		NetImpactDollars:     stringAttr(attrs, "net-impact-dollars"),
	}

	if rel, ok := resource.Relationships["subject"]; ok && rel.Data != nil {
		details.SubjectType = rel.Data.Type
		details.SubjectID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["parent"]; ok && rel.Data != nil {
		details.ParentID = rel.Data.ID
	}
	details.ChildIDs = relationshipIDsFromMap(resource.Relationships, "children")
	if rel, ok := resource.Relationships["equipment"]; ok && rel.Data != nil {
		details.EquipmentID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		details.JobProductionPlanID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["assignee"]; ok && rel.Data != nil {
		details.AssigneeID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["tender-job-schedule-shift"]; ok && rel.Data != nil {
		details.TenderJobScheduleShiftID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["incident-request"]; ok && rel.Data != nil {
		details.IncidentRequestID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["customer"]; ok && rel.Data != nil {
		details.CustomerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["trucker"]; ok && rel.Data != nil {
		details.TruckerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["developer"]; ok && rel.Data != nil {
		details.DeveloperID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["contractor"]; ok && rel.Data != nil {
		details.ContractorID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["material-supplier"]; ok && rel.Data != nil {
		details.MaterialSupplierID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["material-site"]; ok && rel.Data != nil {
		details.MaterialSiteID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["incident-unit-of-measure-quantities"]; ok && rel.raw != nil {
		details.IncidentUnitOfMeasureIDs = relationshipIDsFromMap(resource.Relationships, "incident-unit-of-measure-quantities")
	}
	details.RootCauseIDs = relationshipIDsFromMap(resource.Relationships, "root-causes")
	details.ActionItemIDs = relationshipIDsFromMap(resource.Relationships, "action-items")
	details.IncidentParticipantIDs = relationshipIDsFromMap(resource.Relationships, "incident-participants")
	details.IncidentHeadlineIDs = relationshipIDsFromMap(resource.Relationships, "incident-headline-suggestions")
	details.IncidentTagIncidentIDs = relationshipIDsFromMap(resource.Relationships, "incident-tag-incidents")
	details.IncidentTagIDs = relationshipIDsFromMap(resource.Relationships, "incident-tags")
	details.CommentIDs = relationshipIDsFromMap(resource.Relationships, "comments")
	details.AttachmentIDs = relationshipIDsFromMap(resource.Relationships, "file-attachments")

	return details
}

func renderAdministrativeIncidentDetails(cmd *cobra.Command, details administrativeIncidentDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Type != "" {
		fmt.Fprintf(out, "Type: %s\n", details.Type)
	}
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.Kind != "" {
		fmt.Fprintf(out, "Kind: %s\n", details.Kind)
	}
	if details.Severity != "" {
		fmt.Fprintf(out, "Severity: %s\n", details.Severity)
	}
	if details.Headline != "" {
		fmt.Fprintf(out, "Headline: %s\n", details.Headline)
	}
	if details.Description != "" {
		fmt.Fprintf(out, "Description: %s\n", details.Description)
	}
	if details.StartAt != "" {
		fmt.Fprintf(out, "Start At: %s\n", details.StartAt)
	}
	if details.EndAt != "" {
		fmt.Fprintf(out, "End At: %s\n", details.EndAt)
	}
	if len(details.Natures) > 0 {
		fmt.Fprintf(out, "Natures: %s\n", strings.Join(details.Natures, ", "))
	}
	fmt.Fprintf(out, "Did Stop Work: %t\n", details.DidStopWork)
	if details.TimeZoneID != "" {
		fmt.Fprintf(out, "Time Zone: %s\n", details.TimeZoneID)
	}
	fmt.Fprintf(out, "Current User Can Update: %t\n", details.CurrentUserCanUpdate)
	fmt.Fprintf(out, "Created By Bot: %t\n", details.CreatedByBot)
	if details.NetImpactDollars != "" {
		fmt.Fprintf(out, "Net Impact Dollars: %s\n", details.NetImpactDollars)
	}
	if details.SubjectType != "" && details.SubjectID != "" {
		fmt.Fprintf(out, "Subject: %s/%s\n", details.SubjectType, details.SubjectID)
	}
	if details.ParentID != "" {
		fmt.Fprintf(out, "Parent: %s\n", details.ParentID)
	}
	if len(details.ChildIDs) > 0 {
		fmt.Fprintf(out, "Children: %s\n", strings.Join(details.ChildIDs, ", "))
	}
	if details.EquipmentID != "" {
		fmt.Fprintf(out, "Equipment: %s\n", details.EquipmentID)
	}
	if details.JobProductionPlanID != "" {
		fmt.Fprintf(out, "Job Production Plan: %s\n", details.JobProductionPlanID)
	}
	if details.AssigneeID != "" {
		fmt.Fprintf(out, "Assignee: %s\n", details.AssigneeID)
	}
	if details.TenderJobScheduleShiftID != "" {
		fmt.Fprintf(out, "Tender Job Schedule Shift: %s\n", details.TenderJobScheduleShiftID)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By: %s\n", details.CreatedByID)
	}
	if details.IncidentRequestID != "" {
		fmt.Fprintf(out, "Incident Request: %s\n", details.IncidentRequestID)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker: %s\n", details.BrokerID)
	}
	if details.CustomerID != "" {
		fmt.Fprintf(out, "Customer: %s\n", details.CustomerID)
	}
	if details.TruckerID != "" {
		fmt.Fprintf(out, "Trucker: %s\n", details.TruckerID)
	}
	if details.DeveloperID != "" {
		fmt.Fprintf(out, "Developer: %s\n", details.DeveloperID)
	}
	if details.ContractorID != "" {
		fmt.Fprintf(out, "Contractor: %s\n", details.ContractorID)
	}
	if details.MaterialSupplierID != "" {
		fmt.Fprintf(out, "Material Supplier: %s\n", details.MaterialSupplierID)
	}
	if details.MaterialSiteID != "" {
		fmt.Fprintf(out, "Material Site: %s\n", details.MaterialSiteID)
	}
	if len(details.IncidentUnitOfMeasureIDs) > 0 {
		fmt.Fprintf(out, "Incident Unit Of Measure Quantities: %s\n", strings.Join(details.IncidentUnitOfMeasureIDs, ", "))
	}
	if len(details.RootCauseIDs) > 0 {
		fmt.Fprintf(out, "Root Causes: %s\n", strings.Join(details.RootCauseIDs, ", "))
	}
	if len(details.ActionItemIDs) > 0 {
		fmt.Fprintf(out, "Action Items: %s\n", strings.Join(details.ActionItemIDs, ", "))
	}
	if len(details.IncidentParticipantIDs) > 0 {
		fmt.Fprintf(out, "Incident Participants: %s\n", strings.Join(details.IncidentParticipantIDs, ", "))
	}
	if len(details.IncidentHeadlineIDs) > 0 {
		fmt.Fprintf(out, "Incident Headline Suggestions: %s\n", strings.Join(details.IncidentHeadlineIDs, ", "))
	}
	if len(details.IncidentTagIncidentIDs) > 0 {
		fmt.Fprintf(out, "Incident Tag Incidents: %s\n", strings.Join(details.IncidentTagIncidentIDs, ", "))
	}
	if len(details.IncidentTagIDs) > 0 {
		fmt.Fprintf(out, "Incident Tags: %s\n", strings.Join(details.IncidentTagIDs, ", "))
	}
	if len(details.CommentIDs) > 0 {
		fmt.Fprintf(out, "Comments: %s\n", strings.Join(details.CommentIDs, ", "))
	}
	if len(details.AttachmentIDs) > 0 {
		fmt.Fprintf(out, "Attachments: %s\n", strings.Join(details.AttachmentIDs, ", "))
	}

	return nil
}
