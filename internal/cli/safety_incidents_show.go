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

type safetyIncidentsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type safetyIncidentDetails struct {
	ID                               string   `json:"id"`
	Status                           string   `json:"status,omitempty"`
	Kind                             string   `json:"kind,omitempty"`
	Severity                         string   `json:"severity,omitempty"`
	Headline                         string   `json:"headline,omitempty"`
	Description                      string   `json:"description,omitempty"`
	StartAt                          string   `json:"start_at,omitempty"`
	EndAt                            string   `json:"end_at,omitempty"`
	Natures                          []string `json:"natures,omitempty"`
	DidStopWork                      *bool    `json:"did_stop_work,omitempty"`
	NetImpactTons                    string   `json:"net_impact_tons,omitempty"`
	CurrentUserCanUpdate             *bool    `json:"current_user_can_update,omitempty"`
	TimeZoneID                       string   `json:"time_zone_id,omitempty"`
	CreatedByBot                     *bool    `json:"created_by_bot,omitempty"`
	SubjectType                      string   `json:"subject_type,omitempty"`
	SubjectID                        string   `json:"subject_id,omitempty"`
	ParentID                         string   `json:"parent_id,omitempty"`
	ChildIDs                         []string `json:"child_ids,omitempty"`
	EquipmentID                      string   `json:"equipment_id,omitempty"`
	JobProductionPlanID              string   `json:"job_production_plan_id,omitempty"`
	AssigneeID                       string   `json:"assignee_id,omitempty"`
	TenderJobScheduleShiftID         string   `json:"tender_job_schedule_shift_id,omitempty"`
	CreatedByID                      string   `json:"created_by_id,omitempty"`
	IncidentRequestID                string   `json:"incident_request_id,omitempty"`
	BrokerID                         string   `json:"broker_id,omitempty"`
	DeveloperID                      string   `json:"developer_id,omitempty"`
	IncidentUnitOfMeasureQuantityIDs []string `json:"incident_unit_of_measure_quantity_ids,omitempty"`
	RootCauseIDs                     []string `json:"root_cause_ids,omitempty"`
	ActionItemIDs                    []string `json:"action_item_ids,omitempty"`
	IncidentParticipantIDs           []string `json:"incident_participant_ids,omitempty"`
	IncidentHeadlineSuggestionIDs    []string `json:"incident_headline_suggestion_ids,omitempty"`
	IncidentTagIncidentIDs           []string `json:"incident_tag_incident_ids,omitempty"`
	IncidentTagIDs                   []string `json:"incident_tag_ids,omitempty"`
}

func newSafetyIncidentsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show safety incident details",
		Long: `Show the full details of a safety incident.

Output Fields:
  ID                         Safety incident ID
  Status                     Incident status
  Kind                       Incident kind
  Severity                   Severity level
  Headline                   Headline text
  Description                Description text
  Start At                   Start timestamp
  End At                     End timestamp
  Natures                    Incident natures
  Did Stop Work              Whether work stopped
  Net Impact Tons            Net impact tons (overloading only)
  Current User Can Update    Whether current user can update
  Time Zone ID               Incident time zone
  Created By Bot             Whether created by bot
  Subject                    Subject type and ID
  Parent                     Parent incident ID
  Children                   Child incident IDs
  Equipment                  Equipment ID
  Job Production Plan        Job production plan ID
  Assignee                   Assignee user ID
  Tender Job Schedule Shift  Tender job schedule shift ID
  Created By                 Creator user ID
  Incident Request           Incident request ID
  Broker                     Broker ID
  Developer                  Developer ID
  Incident UOM Quantities    Incident unit of measure quantity IDs
  Root Causes                Root cause IDs
  Action Items               Action item IDs
  Incident Participants      Incident participant IDs
  Headline Suggestions       Incident headline suggestion IDs
  Incident Tag Incidents     Incident tag incident IDs
  Incident Tags              Incident tag IDs

Arguments:
  <id>  The safety incident ID (required).`,
		Example: `  # Show a safety incident
  xbe view safety-incidents show 123

  # Output as JSON
  xbe view safety-incidents show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runSafetyIncidentsShow,
	}
	initSafetyIncidentsShowFlags(cmd)
	return cmd
}

func init() {
	safetyIncidentsCmd.AddCommand(newSafetyIncidentsShowCmd())
}

func initSafetyIncidentsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runSafetyIncidentsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseSafetyIncidentsShowOptions(cmd)
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
		return fmt.Errorf("safety incident id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[safety-incidents]", strings.Join([]string{
		"start-at",
		"end-at",
		"status",
		"kind",
		"description",
		"natures",
		"severity",
		"did-stop-work",
		"headline",
		"current-user-can-update",
		"time-zone-id",
		"created-by-bot",
		"net-impact-tons",
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

	body, _, err := client.Get(cmd.Context(), "/v1/safety-incidents/"+id, query)
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

	details := buildSafetyIncidentDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderSafetyIncidentDetails(cmd, details)
}

func parseSafetyIncidentsShowOptions(cmd *cobra.Command) (safetyIncidentsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return safetyIncidentsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildSafetyIncidentDetails(resp jsonAPISingleResponse) safetyIncidentDetails {
	attrs := resp.Data.Attributes
	details := safetyIncidentDetails{
		ID:                   resp.Data.ID,
		Status:               stringAttr(attrs, "status"),
		Kind:                 stringAttr(attrs, "kind"),
		Severity:             stringAttr(attrs, "severity"),
		Headline:             stringAttr(attrs, "headline"),
		Description:          stringAttr(attrs, "description"),
		StartAt:              stringAttr(attrs, "start-at"),
		EndAt:                stringAttr(attrs, "end-at"),
		Natures:              stringSliceAttr(attrs, "natures"),
		NetImpactTons:        numberAttrAsString(attrs, "net-impact-tons"),
		TimeZoneID:           stringAttr(attrs, "time-zone-id"),
		DidStopWork:          boolAttrPtr(attrs, "did-stop-work"),
		CurrentUserCanUpdate: boolAttrPtr(attrs, "current-user-can-update"),
		CreatedByBot:         boolAttrPtr(attrs, "created-by-bot"),
	}

	if rel, ok := resp.Data.Relationships["subject"]; ok && rel.Data != nil {
		details.SubjectType = rel.Data.Type
		details.SubjectID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["parent"]; ok && rel.Data != nil {
		details.ParentID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["children"]; ok {
		details.ChildIDs = relationshipIDsToStrings(rel)
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
	if rel, ok := resp.Data.Relationships["incident-unit-of-measure-quantities"]; ok {
		details.IncidentUnitOfMeasureQuantityIDs = relationshipIDsToStrings(rel)
	}
	if rel, ok := resp.Data.Relationships["root-causes"]; ok {
		details.RootCauseIDs = relationshipIDsToStrings(rel)
	}
	if rel, ok := resp.Data.Relationships["action-items"]; ok {
		details.ActionItemIDs = relationshipIDsToStrings(rel)
	}
	if rel, ok := resp.Data.Relationships["incident-participants"]; ok {
		details.IncidentParticipantIDs = relationshipIDsToStrings(rel)
	}
	if rel, ok := resp.Data.Relationships["incident-headline-suggestions"]; ok {
		details.IncidentHeadlineSuggestionIDs = relationshipIDsToStrings(rel)
	}
	if rel, ok := resp.Data.Relationships["incident-tag-incidents"]; ok {
		details.IncidentTagIncidentIDs = relationshipIDsToStrings(rel)
	}
	if rel, ok := resp.Data.Relationships["incident-tags"]; ok {
		details.IncidentTagIDs = relationshipIDsToStrings(rel)
	}

	return details
}

func renderSafetyIncidentDetails(cmd *cobra.Command, details safetyIncidentDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
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
	if details.DidStopWork != nil {
		fmt.Fprintf(out, "Did Stop Work: %t\n", *details.DidStopWork)
	}
	if details.NetImpactTons != "" {
		fmt.Fprintf(out, "Net Impact Tons: %s\n", details.NetImpactTons)
	}
	if details.CurrentUserCanUpdate != nil {
		fmt.Fprintf(out, "Current User Can Update: %t\n", *details.CurrentUserCanUpdate)
	}
	if details.TimeZoneID != "" {
		fmt.Fprintf(out, "Time Zone ID: %s\n", details.TimeZoneID)
	}
	if details.CreatedByBot != nil {
		fmt.Fprintf(out, "Created By Bot: %t\n", *details.CreatedByBot)
	}
	if details.SubjectID != "" || details.SubjectType != "" {
		fmt.Fprintf(out, "Subject: %s\n", formatPolymorphic(details.SubjectType, details.SubjectID))
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
	if details.DeveloperID != "" {
		fmt.Fprintf(out, "Developer: %s\n", details.DeveloperID)
	}
	if len(details.IncidentUnitOfMeasureQuantityIDs) > 0 {
		fmt.Fprintf(out, "Incident UOM Quantities: %s\n", strings.Join(details.IncidentUnitOfMeasureQuantityIDs, ", "))
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
	if len(details.IncidentHeadlineSuggestionIDs) > 0 {
		fmt.Fprintf(out, "Headline Suggestions: %s\n", strings.Join(details.IncidentHeadlineSuggestionIDs, ", "))
	}
	if len(details.IncidentTagIncidentIDs) > 0 {
		fmt.Fprintf(out, "Incident Tag Incidents: %s\n", strings.Join(details.IncidentTagIncidentIDs, ", "))
	}
	if len(details.IncidentTagIDs) > 0 {
		fmt.Fprintf(out, "Incident Tags: %s\n", strings.Join(details.IncidentTagIDs, ", "))
	}

	return nil
}

func boolAttrPtr(attrs map[string]any, key string) *bool {
	if attrs == nil {
		return nil
	}
	value, ok := attrs[key]
	if !ok || value == nil {
		return nil
	}
	switch typed := value.(type) {
	case bool:
		return &typed
	case string:
		val := strings.EqualFold(strings.TrimSpace(typed), "true")
		return &val
	default:
		val := fmt.Sprintf("%v", typed) == "true"
		return &val
	}
}
