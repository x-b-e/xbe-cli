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

type doProductionIncidentsCreateOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	SubjectType            string
	SubjectID              string
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
	JobProductionPlan      string
	Equipment              string
	TenderJobScheduleShift string
	Assignee               string
	ParentID               string
	ParentType             string
}

func newDoProductionIncidentsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a production incident",
		Long: `Create a production incident.

Required flags:
  --subject-type     Subject type (e.g., customers, job-production-plans) (required)
  --subject-id       Subject ID (required)
  --start-at         Start timestamp (ISO 8601) (required)
  --status           Status (required)

Optional flags:
  --end-at                 End timestamp (ISO 8601)
  --kind                   Incident kind
  --description            Description
  --natures                Incident natures (comma-separated, not valid for production incidents)
  --severity               Severity (low, medium, high, catastrophic)
  --time-value-type        Time value type (deducted_time, credited_time)
  --headline               Headline
  --net-impact-minutes     Net impact minutes
  --net-impact-dollars     Net impact dollars
  --is-down-time           Down time flag (true/false)
  --did-stop-work          Did stop work (true/false, not valid for production incidents)
  --job-production-plan    Job production plan ID
  --equipment              Equipment ID
  --tender-job-schedule-shift  Tender job schedule shift ID
  --assignee               Assignee user ID
  --parent-id              Parent incident ID
  --parent-type            Parent incident type (e.g., production-incidents, safety-incidents)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a production incident for a customer
  xbe do production-incidents create \\
    --subject-type customers \\
    --subject-id 123 \\
    --start-at 2026-01-01T08:00:00Z \\
    --status open \\
    --kind trucking \\
    --net-impact-minutes 60

  # Create with net impact dollars and down time
  xbe do production-incidents create \\
    --subject-type customers \\
    --subject-id 123 \\
    --start-at 2026-01-01T08:00:00Z \\
    --status open \\
    --time-value-type deducted_time \\
    --net-impact-dollars 500 \\
    --is-down-time`,
		Args: cobra.NoArgs,
		RunE: runDoProductionIncidentsCreate,
	}
	initDoProductionIncidentsCreateFlags(cmd)
	return cmd
}

func init() {
	doProductionIncidentsCmd.AddCommand(newDoProductionIncidentsCreateCmd())
}

func initDoProductionIncidentsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("subject-type", "", "Subject type (required)")
	cmd.Flags().String("subject-id", "", "Subject ID (required)")
	cmd.Flags().String("start-at", "", "Start timestamp (ISO 8601) (required)")
	cmd.Flags().String("end-at", "", "End timestamp (ISO 8601)")
	cmd.Flags().String("status", "", "Status (required)")
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
	cmd.Flags().String("job-production-plan", "", "Job production plan ID")
	cmd.Flags().String("equipment", "", "Equipment ID")
	cmd.Flags().String("tender-job-schedule-shift", "", "Tender job schedule shift ID")
	cmd.Flags().String("assignee", "", "Assignee user ID")
	cmd.Flags().String("parent-id", "", "Parent incident ID")
	cmd.Flags().String("parent-type", "", "Parent incident type (e.g., production-incidents)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProductionIncidentsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProductionIncidentsCreateOptions(cmd)
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

	if opts.SubjectType == "" {
		err := fmt.Errorf("--subject-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.SubjectID == "" {
		err := fmt.Errorf("--subject-id is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.StartAt == "" {
		err := fmt.Errorf("--start-at is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.Status == "" {
		err := fmt.Errorf("--status is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"start-at": opts.StartAt,
		"status":   opts.Status,
	}
	if opts.EndAt != "" {
		attributes["end-at"] = opts.EndAt
	}
	if opts.Kind != "" {
		attributes["kind"] = opts.Kind
	}
	if opts.Description != "" {
		attributes["description"] = opts.Description
	}
	if cmd.Flags().Changed("natures") {
		attributes["natures"] = opts.Natures
	}
	if opts.Severity != "" {
		attributes["severity"] = opts.Severity
	}
	if opts.TimeValueType != "" {
		attributes["time-value-type"] = opts.TimeValueType
	}
	if opts.Headline != "" {
		attributes["headline"] = opts.Headline
	}
	if opts.NetImpactMinutes != "" {
		minutes, err := strconv.ParseFloat(opts.NetImpactMinutes, 64)
		if err != nil {
			err := fmt.Errorf("--net-impact-minutes must be a number")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["net-impact-minutes"] = minutes
	}
	if opts.NetImpactDollars != "" {
		dollars, err := strconv.ParseFloat(opts.NetImpactDollars, 64)
		if err != nil {
			err := fmt.Errorf("--net-impact-dollars must be a number")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["net-impact-dollars"] = dollars
	}
	if cmd.Flags().Changed("is-down-time") {
		attributes["is-down-time"] = opts.IsDownTime
	}
	if cmd.Flags().Changed("did-stop-work") {
		attributes["did-stop-work"] = opts.DidStopWork
	}

	relationships := map[string]any{
		"subject": map[string]any{
			"data": map[string]any{
				"type": normalizeIncidentSubjectRelationshipType(opts.SubjectType),
				"id":   opts.SubjectID,
			},
		},
	}

	if opts.JobProductionPlan != "" {
		relationships["job-production-plan"] = map[string]any{
			"data": map[string]any{
				"type": "job-production-plans",
				"id":   opts.JobProductionPlan,
			},
		}
	}
	if opts.Equipment != "" {
		relationships["equipment"] = map[string]any{
			"data": map[string]any{
				"type": "equipment",
				"id":   opts.Equipment,
			},
		}
	}
	if opts.TenderJobScheduleShift != "" {
		relationships["tender-job-schedule-shift"] = map[string]any{
			"data": map[string]any{
				"type": "tender-job-schedule-shifts",
				"id":   opts.TenderJobScheduleShift,
			},
		}
	}
	if opts.Assignee != "" {
		relationships["assignee"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.Assignee,
			},
		}
	}
	if opts.ParentID != "" {
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

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "production-incidents",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/production-incidents", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created production incident %s\n", row.ID)
	return nil
}

func parseDoProductionIncidentsCreateOptions(cmd *cobra.Command) (doProductionIncidentsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	subjectType, _ := cmd.Flags().GetString("subject-type")
	subjectID, _ := cmd.Flags().GetString("subject-id")
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
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	equipment, _ := cmd.Flags().GetString("equipment")
	tenderJobScheduleShift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	assignee, _ := cmd.Flags().GetString("assignee")
	parentID, _ := cmd.Flags().GetString("parent-id")
	parentType, _ := cmd.Flags().GetString("parent-type")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProductionIncidentsCreateOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		SubjectType:            subjectType,
		SubjectID:              subjectID,
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
		JobProductionPlan:      jobProductionPlan,
		Equipment:              equipment,
		TenderJobScheduleShift: tenderJobScheduleShift,
		Assignee:               assignee,
		ParentID:               parentID,
		ParentType:             parentType,
	}, nil
}
