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

type doSafetyIncidentsUpdateOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	ID                     string
	SubjectType            string
	SubjectID              string
	StartAt                string
	EndAt                  string
	Status                 string
	Kind                   string
	Severity               string
	Description            string
	Headline               string
	Natures                []string
	DidStopWork            bool
	NetImpactTons          string
	NewType                string
	Equipment              string
	JobProductionPlan      string
	Assignee               string
	TenderJobScheduleShift string
	Parent                 string
}

func newDoSafetyIncidentsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a safety incident",
		Long: `Update an existing safety incident.

Optional fields:
  --subject-type            Subject type (JSON:API type, must be paired with --subject-id)
  --subject-id              Subject ID (must be paired with --subject-type)
  --start-at                Start timestamp (ISO 8601)
  --end-at                  End timestamp (ISO 8601)
  --status                  Status (open, closed, abandoned, processing, parked)
  --kind                    Kind (near_miss, good_catch, damage, overloading, work_zone_intrusion)
  --severity                Severity (low, medium, high, catastrophic)
  --headline                Headline text
  --description             Description text
  --natures                 Incident natures (comma-separated: personal,property)
  --did-stop-work           Whether work stopped
  --net-impact-tons         Net impact tons (overloading only)
  --new-type                Change incident type (Incident subclass name, e.g., SafetyIncident)
  --parent                  Parent incident ID
  --equipment               Equipment ID
  --job-production-plan     Job production plan ID
  --assignee                Assignee user ID
  --tender-job-schedule-shift  Tender job schedule shift ID`,
		Example: `  # Update status and end time
  xbe do safety-incidents update 123 --status closed --end-at 2025-01-15T12:00:00Z

  # Update kind and net impact tons
  xbe do safety-incidents update 123 --kind overloading --net-impact-tons 8.5`,
		Args: cobra.ExactArgs(1),
		RunE: runDoSafetyIncidentsUpdate,
	}
	initDoSafetyIncidentsUpdateFlags(cmd)
	return cmd
}

func init() {
	doSafetyIncidentsCmd.AddCommand(newDoSafetyIncidentsUpdateCmd())
}

func initDoSafetyIncidentsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("subject-type", "", "Subject type (JSON:API type, must be paired with --subject-id)")
	cmd.Flags().String("subject-id", "", "Subject ID (must be paired with --subject-type)")
	cmd.Flags().String("start-at", "", "Start timestamp (ISO 8601)")
	cmd.Flags().String("end-at", "", "End timestamp (ISO 8601)")
	cmd.Flags().String("status", "", "Status (open, closed, abandoned, processing, parked)")
	cmd.Flags().String("kind", "", "Kind (near_miss, good_catch, damage, overloading, work_zone_intrusion)")
	cmd.Flags().String("severity", "", "Severity (low, medium, high, catastrophic)")
	cmd.Flags().String("headline", "", "Headline text")
	cmd.Flags().String("description", "", "Description text")
	cmd.Flags().StringSlice("natures", nil, "Incident natures (comma-separated: personal,property)")
	cmd.Flags().Bool("did-stop-work", false, "Whether work stopped")
	cmd.Flags().String("net-impact-tons", "", "Net impact tons (overloading only)")
	cmd.Flags().String("new-type", "", "Change incident type (Incident subclass name, e.g., SafetyIncident)")
	cmd.Flags().String("parent", "", "Parent incident ID")
	cmd.Flags().String("equipment", "", "Equipment ID")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID")
	cmd.Flags().String("assignee", "", "Assignee user ID")
	cmd.Flags().String("tender-job-schedule-shift", "", "Tender job schedule shift ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoSafetyIncidentsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoSafetyIncidentsUpdateOptions(cmd, args)
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
	relationships := map[string]any{}
	hasUpdate := false

	if opts.StartAt != "" {
		attributes["start-at"] = opts.StartAt
		hasUpdate = true
	}
	if opts.EndAt != "" {
		attributes["end-at"] = opts.EndAt
		hasUpdate = true
	}
	if opts.Status != "" {
		attributes["status"] = opts.Status
		hasUpdate = true
	}
	if opts.Kind != "" {
		attributes["kind"] = opts.Kind
		hasUpdate = true
	}
	if opts.Severity != "" {
		attributes["severity"] = opts.Severity
		hasUpdate = true
	}
	if opts.Description != "" {
		attributes["description"] = opts.Description
		hasUpdate = true
	}
	if opts.Headline != "" {
		attributes["headline"] = opts.Headline
		hasUpdate = true
	}
	if cmd.Flags().Changed("natures") {
		attributes["natures"] = opts.Natures
		hasUpdate = true
	}
	if cmd.Flags().Changed("did-stop-work") {
		attributes["did-stop-work"] = opts.DidStopWork
		hasUpdate = true
	}
	if opts.NetImpactTons != "" {
		attributes["net-impact-tons"] = opts.NetImpactTons
		hasUpdate = true
	}
	if opts.NewType != "" {
		attributes["new-type"] = opts.NewType
		hasUpdate = true
	}

	subjectChanged := cmd.Flags().Changed("subject-type") || cmd.Flags().Changed("subject-id")
	if subjectChanged {
		if opts.SubjectType == "" || opts.SubjectID == "" {
			err := fmt.Errorf("--subject-type and --subject-id must be used together")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["subject"] = map[string]any{
			"data": map[string]any{
				"type": opts.SubjectType,
				"id":   opts.SubjectID,
			},
		}
		hasUpdate = true
	}

	if opts.Parent != "" {
		relationships["parent"] = map[string]any{
			"data": map[string]any{
				"type": "incidents",
				"id":   opts.Parent,
			},
		}
		hasUpdate = true
	}
	if opts.Equipment != "" {
		relationships["equipment"] = map[string]any{
			"data": map[string]any{
				"type": "equipment",
				"id":   opts.Equipment,
			},
		}
		hasUpdate = true
	}
	if opts.JobProductionPlan != "" {
		relationships["job-production-plan"] = map[string]any{
			"data": map[string]any{
				"type": "job-production-plans",
				"id":   opts.JobProductionPlan,
			},
		}
		hasUpdate = true
	}
	if opts.Assignee != "" {
		relationships["assignee"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.Assignee,
			},
		}
		hasUpdate = true
	}
	if opts.TenderJobScheduleShift != "" {
		relationships["tender-job-schedule-shift"] = map[string]any{
			"data": map[string]any{
				"type": "tender-job-schedule-shifts",
				"id":   opts.TenderJobScheduleShift,
			},
		}
		hasUpdate = true
	}

	if !hasUpdate {
		err := fmt.Errorf("no fields to update; specify at least one flag")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"id":   opts.ID,
		"type": "safety-incidents",
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{"data": data}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/safety-incidents/"+opts.ID, jsonBody)
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

	details := buildSafetyIncidentDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated safety incident %s\n", details.ID)
	return renderSafetyIncidentDetails(cmd, details)
}

func parseDoSafetyIncidentsUpdateOptions(cmd *cobra.Command, args []string) (doSafetyIncidentsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	subjectType, _ := cmd.Flags().GetString("subject-type")
	subjectID, _ := cmd.Flags().GetString("subject-id")
	startAt, _ := cmd.Flags().GetString("start-at")
	endAt, _ := cmd.Flags().GetString("end-at")
	status, _ := cmd.Flags().GetString("status")
	kind, _ := cmd.Flags().GetString("kind")
	severity, _ := cmd.Flags().GetString("severity")
	headline, _ := cmd.Flags().GetString("headline")
	description, _ := cmd.Flags().GetString("description")
	natures, _ := cmd.Flags().GetStringSlice("natures")
	didStopWork, _ := cmd.Flags().GetBool("did-stop-work")
	netImpactTons, _ := cmd.Flags().GetString("net-impact-tons")
	newType, _ := cmd.Flags().GetString("new-type")
	parent, _ := cmd.Flags().GetString("parent")
	equipment, _ := cmd.Flags().GetString("equipment")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	assignee, _ := cmd.Flags().GetString("assignee")
	tenderJobScheduleShift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doSafetyIncidentsUpdateOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		ID:                     args[0],
		SubjectType:            subjectType,
		SubjectID:              subjectID,
		StartAt:                startAt,
		EndAt:                  endAt,
		Status:                 status,
		Kind:                   kind,
		Severity:               severity,
		Headline:               headline,
		Description:            description,
		Natures:                natures,
		DidStopWork:            didStopWork,
		NetImpactTons:          netImpactTons,
		NewType:                newType,
		Parent:                 parent,
		Equipment:              equipment,
		JobProductionPlan:      jobProductionPlan,
		Assignee:               assignee,
		TenderJobScheduleShift: tenderJobScheduleShift,
	}, nil
}
