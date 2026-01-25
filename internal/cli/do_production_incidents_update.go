package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doProductionIncidentsUpdateOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	ID                     string
	StartAt                string
	EndAt                  string
	Status                 string
	Kind                   string
	Description            string
	Natures                []string
	Severity               string
	TimeValueType          string
	Headline               string
	NetImpactMinutes       string
	NetImpactDollars       string
	IsDownTime             bool
	DidStopWork            bool
	NewType                string
	SubjectType            string
	SubjectID              string
	JobProductionPlan      string
	Equipment              string
	TenderJobScheduleShift string
	Assignee               string
	ParentID               string
	ParentType             string
}

func newDoProductionIncidentsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a production incident",
		Long: `Update a production incident.

Provide the incident ID as an argument and the fields to update.

Updatable fields:
  --start-at                 Start timestamp (ISO 8601)
  --end-at                   End timestamp (ISO 8601)
  --status                   Status
  --kind                     Incident kind
  --description              Description
  --natures                  Incident natures (comma-separated, not valid for production incidents)
  --severity                 Severity (low, medium, high, catastrophic)
  --time-value-type          Time value type (deducted_time, credited_time)
  --headline                 Headline
  --net-impact-minutes       Net impact minutes
  --net-impact-dollars       Net impact dollars
  --is-down-time             Down time flag (true/false)
  --did-stop-work            Did stop work (true/false, not valid for production incidents)
  --new-type                 Change incident type (e.g., ProductionIncident)

Relationship updates:
  --subject-type             Subject type (e.g., customers)
  --subject-id               Subject ID
  --job-production-plan      Job production plan ID
  --equipment                Equipment ID
  --tender-job-schedule-shift  Tender job schedule shift ID
  --assignee                 Assignee user ID
  --parent-id                Parent incident ID
  --parent-type              Parent incident type (e.g., production-incidents)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update status
  xbe do production-incidents update 123 --status closed

  # Update net impact values
  xbe do production-incidents update 123 --net-impact-minutes 90 --net-impact-dollars 1500

  # Update time value type
  xbe do production-incidents update 123 --time-value-type credited_time

  # Change incident type
  xbe do production-incidents update 123 --new-type ProductionIncident`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProductionIncidentsUpdate,
	}
	initDoProductionIncidentsUpdateFlags(cmd)
	return cmd
}

func init() {
	doProductionIncidentsCmd.AddCommand(newDoProductionIncidentsUpdateCmd())
}

func initDoProductionIncidentsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("start-at", "", "Start timestamp (ISO 8601)")
	cmd.Flags().String("end-at", "", "End timestamp (ISO 8601)")
	cmd.Flags().String("status", "", "Status")
	cmd.Flags().String("kind", "", "Incident kind")
	cmd.Flags().String("description", "", "Description")
	cmd.Flags().StringSlice("natures", nil, "Incident natures (comma-separated, not valid for production incidents)")
	cmd.Flags().String("severity", "", "Severity (low, medium, high, catastrophic)")
	cmd.Flags().String("time-value-type", "", "Time value type (deducted_time, credited_time)")
	cmd.Flags().String("headline", "", "Headline")
	cmd.Flags().String("net-impact-minutes", "", "Net impact minutes")
	cmd.Flags().String("net-impact-dollars", "", "Net impact dollars")
	cmd.Flags().Bool("is-down-time", false, "Down time flag")
	cmd.Flags().Bool("did-stop-work", false, "Did stop work (not valid for production incidents)")
	cmd.Flags().String("new-type", "", "Change incident type (e.g., ProductionIncident)")
	cmd.Flags().String("subject-type", "", "Subject type (use with --subject-id)")
	cmd.Flags().String("subject-id", "", "Subject ID (use with --subject-type)")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID")
	cmd.Flags().String("equipment", "", "Equipment ID")
	cmd.Flags().String("tender-job-schedule-shift", "", "Tender job schedule shift ID")
	cmd.Flags().String("assignee", "", "Assignee user ID")
	cmd.Flags().String("parent-id", "", "Parent incident ID")
	cmd.Flags().String("parent-type", "", "Parent incident type (e.g., production-incidents)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProductionIncidentsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProductionIncidentsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("start-at") {
		attributes["start-at"] = opts.StartAt
	}
	if cmd.Flags().Changed("end-at") {
		attributes["end-at"] = opts.EndAt
	}
	if cmd.Flags().Changed("status") {
		attributes["status"] = opts.Status
	}
	if cmd.Flags().Changed("kind") {
		attributes["kind"] = opts.Kind
	}
	if cmd.Flags().Changed("description") {
		attributes["description"] = opts.Description
	}
	if cmd.Flags().Changed("natures") {
		attributes["natures"] = opts.Natures
	}
	if cmd.Flags().Changed("severity") {
		attributes["severity"] = opts.Severity
	}
	if cmd.Flags().Changed("time-value-type") {
		attributes["time-value-type"] = opts.TimeValueType
	}
	if cmd.Flags().Changed("headline") {
		attributes["headline"] = opts.Headline
	}
	if cmd.Flags().Changed("net-impact-minutes") {
		if strings.TrimSpace(opts.NetImpactMinutes) != "" {
			minutes, err := strconv.ParseFloat(opts.NetImpactMinutes, 64)
			if err != nil {
				err := fmt.Errorf("--net-impact-minutes must be a number")
				fmt.Fprintln(cmd.ErrOrStderr(), err)
				return err
			}
			attributes["net-impact-minutes"] = minutes
		}
	}
	if cmd.Flags().Changed("net-impact-dollars") {
		if strings.TrimSpace(opts.NetImpactDollars) != "" {
			dollars, err := strconv.ParseFloat(opts.NetImpactDollars, 64)
			if err != nil {
				err := fmt.Errorf("--net-impact-dollars must be a number")
				fmt.Fprintln(cmd.ErrOrStderr(), err)
				return err
			}
			attributes["net-impact-dollars"] = dollars
		}
	}
	if cmd.Flags().Changed("is-down-time") {
		attributes["is-down-time"] = opts.IsDownTime
	}
	if cmd.Flags().Changed("did-stop-work") {
		attributes["did-stop-work"] = opts.DidStopWork
	}
	if cmd.Flags().Changed("new-type") {
		attributes["new-type"] = opts.NewType
	}

	relationships := map[string]any{}
	if cmd.Flags().Changed("subject-type") || cmd.Flags().Changed("subject-id") {
		if strings.TrimSpace(opts.SubjectType) == "" || strings.TrimSpace(opts.SubjectID) == "" {
			err := fmt.Errorf("--subject-type and --subject-id must be set together")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["subject"] = map[string]any{
			"data": map[string]any{
				"type": normalizeIncidentSubjectRelationshipType(opts.SubjectType),
				"id":   opts.SubjectID,
			},
		}
	}
	if cmd.Flags().Changed("job-production-plan") && opts.JobProductionPlan != "" {
		relationships["job-production-plan"] = map[string]any{
			"data": map[string]any{
				"type": "job-production-plans",
				"id":   opts.JobProductionPlan,
			},
		}
	}
	if cmd.Flags().Changed("equipment") && opts.Equipment != "" {
		relationships["equipment"] = map[string]any{
			"data": map[string]any{
				"type": "equipment",
				"id":   opts.Equipment,
			},
		}
	}
	if cmd.Flags().Changed("tender-job-schedule-shift") && opts.TenderJobScheduleShift != "" {
		relationships["tender-job-schedule-shift"] = map[string]any{
			"data": map[string]any{
				"type": "tender-job-schedule-shifts",
				"id":   opts.TenderJobScheduleShift,
			},
		}
	}
	if cmd.Flags().Changed("assignee") && opts.Assignee != "" {
		relationships["assignee"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.Assignee,
			},
		}
	}
	if cmd.Flags().Changed("parent-id") || cmd.Flags().Changed("parent-type") {
		if strings.TrimSpace(opts.ParentID) == "" {
			err := fmt.Errorf("--parent-id is required when setting parent relationship")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		parentType := opts.ParentType
		if strings.TrimSpace(parentType) == "" {
			parentType = "production-incidents"
		}
		relationships["parent"] = map[string]any{
			"data": map[string]any{
				"type": normalizeIncidentRelationshipType(parentType),
				"id":   opts.ParentID,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no fields to update; specify at least one field or relationship")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type": "production-incidents",
			"id":   opts.ID,
		},
	}
	if len(attributes) > 0 {
		requestBody["data"].(map[string]any)["attributes"] = attributes
	}
	if len(relationships) > 0 {
		requestBody["data"].(map[string]any)["relationships"] = relationships
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/production-incidents/"+opts.ID, jsonBody)
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

	row := buildProductionIncidentRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated production incident %s\n", row.ID)
	return nil
}

func parseDoProductionIncidentsUpdateOptions(cmd *cobra.Command, args []string) (doProductionIncidentsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	startAt, _ := cmd.Flags().GetString("start-at")
	endAt, _ := cmd.Flags().GetString("end-at")
	status, _ := cmd.Flags().GetString("status")
	kind, _ := cmd.Flags().GetString("kind")
	description, _ := cmd.Flags().GetString("description")
	natures, _ := cmd.Flags().GetStringSlice("natures")
	severity, _ := cmd.Flags().GetString("severity")
	timeValueType, _ := cmd.Flags().GetString("time-value-type")
	headline, _ := cmd.Flags().GetString("headline")
	netImpactMinutes, _ := cmd.Flags().GetString("net-impact-minutes")
	netImpactDollars, _ := cmd.Flags().GetString("net-impact-dollars")
	isDownTime, _ := cmd.Flags().GetBool("is-down-time")
	didStopWork, _ := cmd.Flags().GetBool("did-stop-work")
	newType, _ := cmd.Flags().GetString("new-type")
	subjectType, _ := cmd.Flags().GetString("subject-type")
	subjectID, _ := cmd.Flags().GetString("subject-id")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	equipment, _ := cmd.Flags().GetString("equipment")
	tenderJobScheduleShift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	assignee, _ := cmd.Flags().GetString("assignee")
	parentID, _ := cmd.Flags().GetString("parent-id")
	parentType, _ := cmd.Flags().GetString("parent-type")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProductionIncidentsUpdateOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		ID:                     args[0],
		StartAt:                startAt,
		EndAt:                  endAt,
		Status:                 status,
		Kind:                   kind,
		Description:            description,
		Natures:                natures,
		Severity:               severity,
		TimeValueType:          timeValueType,
		Headline:               headline,
		NetImpactMinutes:       netImpactMinutes,
		NetImpactDollars:       netImpactDollars,
		IsDownTime:             isDownTime,
		DidStopWork:            didStopWork,
		NewType:                newType,
		SubjectType:            subjectType,
		SubjectID:              subjectID,
		JobProductionPlan:      jobProductionPlan,
		Equipment:              equipment,
		TenderJobScheduleShift: tenderJobScheduleShift,
		Assignee:               assignee,
		ParentID:               parentID,
		ParentType:             parentType,
	}, nil
}
