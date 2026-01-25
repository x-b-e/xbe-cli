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

type doSafetyIncidentsCreateOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
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
	Equipment              string
	JobProductionPlan      string
	Assignee               string
	TenderJobScheduleShift string
	Parent                 string
}

func newDoSafetyIncidentsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a safety incident",
		Long: `Create a new safety incident.

Required flags:
  --subject-type   Subject type (JSON:API type, e.g., brokers, job-production-plans)
  --subject-id     Subject ID
  --start-at       Start timestamp (ISO 8601)
  --status         Status (open, closed, abandoned, processing, parked)

Optional flags:
  --kind                    Kind (near_miss, good_catch, damage, overloading, work_zone_intrusion)
  --severity                Severity (low, medium, high, catastrophic)
  --headline                Headline text
  --description             Description text
  --natures                 Incident natures (comma-separated: personal,property)
  --did-stop-work           Whether work stopped
  --end-at                  End timestamp (ISO 8601)
  --net-impact-tons         Net impact tons (overloading only)
  --parent                  Parent incident ID
  --equipment               Equipment ID
  --job-production-plan     Job production plan ID
  --assignee                Assignee user ID
  --tender-job-schedule-shift  Tender job schedule shift ID`,
		Example: `  # Create a safety incident on a broker
  xbe do safety-incidents create \\
    --subject-type brokers \\
    --subject-id 123 \\
    --start-at 2025-01-15T10:00:00Z \\
    --status open \\
    --kind near_miss \\
    --headline "Near miss at plant"

  # Create an overloading incident with net impact tons
  xbe do safety-incidents create \\
    --subject-type brokers \\
    --subject-id 123 \\
    --start-at 2025-01-15T10:00:00Z \\
    --status open \\
    --kind overloading \\
    --net-impact-tons 12.5`,
		Args: cobra.NoArgs,
		RunE: runDoSafetyIncidentsCreate,
	}
	initDoSafetyIncidentsCreateFlags(cmd)
	return cmd
}

func init() {
	doSafetyIncidentsCmd.AddCommand(newDoSafetyIncidentsCreateCmd())
}

func initDoSafetyIncidentsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("subject-type", "", "Subject type (JSON:API type, e.g., brokers)")
	cmd.Flags().String("subject-id", "", "Subject ID")
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
	cmd.Flags().String("parent", "", "Parent incident ID")
	cmd.Flags().String("equipment", "", "Equipment ID")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID")
	cmd.Flags().String("assignee", "", "Assignee user ID")
	cmd.Flags().String("tender-job-schedule-shift", "", "Tender job schedule shift ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoSafetyIncidentsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoSafetyIncidentsCreateOptions(cmd)
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
	if opts.Severity != "" {
		attributes["severity"] = opts.Severity
	}
	if opts.Description != "" {
		attributes["description"] = opts.Description
	}
	if opts.Headline != "" {
		attributes["headline"] = opts.Headline
	}
	if len(opts.Natures) > 0 {
		attributes["natures"] = opts.Natures
	}
	if cmd.Flags().Changed("did-stop-work") {
		attributes["did-stop-work"] = opts.DidStopWork
	}
	if opts.NetImpactTons != "" {
		attributes["net-impact-tons"] = opts.NetImpactTons
	}

	relationships := map[string]any{
		"subject": map[string]any{
			"data": map[string]any{
				"type": opts.SubjectType,
				"id":   opts.SubjectID,
			},
		},
	}
	if opts.Parent != "" {
		relationships["parent"] = map[string]any{
			"data": map[string]any{
				"type": "incidents",
				"id":   opts.Parent,
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
	if opts.JobProductionPlan != "" {
		relationships["job-production-plan"] = map[string]any{
			"data": map[string]any{
				"type": "job-production-plans",
				"id":   opts.JobProductionPlan,
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
	if opts.TenderJobScheduleShift != "" {
		relationships["tender-job-schedule-shift"] = map[string]any{
			"data": map[string]any{
				"type": "tender-job-schedule-shifts",
				"id":   opts.TenderJobScheduleShift,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "safety-incidents",
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

	body, _, err := client.Post(cmd.Context(), "/v1/safety-incidents", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created safety incident %s\n", details.ID)
	return renderSafetyIncidentDetails(cmd, details)
}

func parseDoSafetyIncidentsCreateOptions(cmd *cobra.Command) (doSafetyIncidentsCreateOptions, error) {
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
	parent, _ := cmd.Flags().GetString("parent")
	equipment, _ := cmd.Flags().GetString("equipment")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	assignee, _ := cmd.Flags().GetString("assignee")
	tenderJobScheduleShift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doSafetyIncidentsCreateOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
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
		Parent:                 parent,
		Equipment:              equipment,
		JobProductionPlan:      jobProductionPlan,
		Assignee:               assignee,
		TenderJobScheduleShift: tenderJobScheduleShift,
	}, nil
}
