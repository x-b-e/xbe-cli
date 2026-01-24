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

type doEfficiencyIncidentsCreateOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	Subject                string
	StartAt                string
	EndAt                  string
	Status                 string
	Kind                   string
	Severity               string
	Description            string
	Headline               string
	Natures                string
	DidStopWork            string
	NetImpactDollars       string
	Assignee               string
	JobProductionPlan      string
	Equipment              string
	Parent                 string
	TenderJobScheduleShift string
}

func newDoEfficiencyIncidentsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an efficiency incident",
		Long: `Create a new efficiency incident.

Required flags:
  --subject    Subject in Type|ID format (e.g. Broker|123)
  --start-at   Incident start timestamp (ISO 8601)
  --status     Status (open, closed, abandoned, processing, parked)

Optional flags:
  --kind                     Incident kind (over_trucking)
  --severity                 Severity (low, medium, high, catastrophic)
  --description              Description text
  --headline                 Headline text
  --natures                  Natures (comma-separated: personal,property)
  --did-stop-work            Did stop work (true/false)
  --end-at                   End timestamp (ISO 8601)
  --net-impact-dollars       Net impact dollars
  --assignee                 Assignee user ID
  --job-production-plan       Job production plan ID
  --equipment                Equipment ID
  --parent                   Parent incident ID
  --tender-job-schedule-shift Tender job schedule shift ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create an efficiency incident
  xbe do efficiency-incidents create --subject Broker|123 --start-at 2025-01-01T08:00:00Z --status open --kind over_trucking

  # Create with net impact dollars
  xbe do efficiency-incidents create --subject Broker|123 --start-at 2025-01-01T08:00:00Z --status open --net-impact-dollars 1500`,
		Args: cobra.NoArgs,
		RunE: runDoEfficiencyIncidentsCreate,
	}
	initDoEfficiencyIncidentsCreateFlags(cmd)
	return cmd
}

func init() {
	doEfficiencyIncidentsCmd.AddCommand(newDoEfficiencyIncidentsCreateCmd())
}

func initDoEfficiencyIncidentsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("subject", "", "Subject in Type|ID format (required)")
	cmd.Flags().String("start-at", "", "Start timestamp (ISO 8601, required)")
	cmd.Flags().String("end-at", "", "End timestamp (ISO 8601)")
	cmd.Flags().String("status", "", "Status (open, closed, abandoned, processing, parked)")
	cmd.Flags().String("kind", "", "Kind (over_trucking)")
	cmd.Flags().String("severity", "", "Severity (low, medium, high, catastrophic)")
	cmd.Flags().String("description", "", "Description text")
	cmd.Flags().String("headline", "", "Headline text")
	cmd.Flags().String("natures", "", "Natures (comma-separated: personal,property)")
	cmd.Flags().String("did-stop-work", "", "Did stop work (true/false)")
	cmd.Flags().String("net-impact-dollars", "", "Net impact dollars")
	cmd.Flags().String("assignee", "", "Assignee user ID")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID")
	cmd.Flags().String("equipment", "", "Equipment ID")
	cmd.Flags().String("parent", "", "Parent incident ID")
	cmd.Flags().String("tender-job-schedule-shift", "", "Tender job schedule shift ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("subject")
	cmd.MarkFlagRequired("start-at")
	cmd.MarkFlagRequired("status")
}

func runDoEfficiencyIncidentsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoEfficiencyIncidentsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.Subject) == "" {
		err := fmt.Errorf("--subject is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.StartAt) == "" {
		err := fmt.Errorf("--start-at is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.Status) == "" {
		err := fmt.Errorf("--status is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	subjectClass, subjectType, subjectID, err := parseIncidentSubjectRef(opts.Subject)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"start-at": opts.StartAt,
		"status":   opts.Status,
	}
	setStringAttrIfPresent(attributes, "end-at", opts.EndAt)
	setStringAttrIfPresent(attributes, "kind", opts.Kind)
	setStringAttrIfPresent(attributes, "severity", opts.Severity)
	setStringAttrIfPresent(attributes, "description", opts.Description)
	setStringAttrIfPresent(attributes, "headline", opts.Headline)
	if opts.Natures != "" {
		attributes["natures"] = splitCommaList(opts.Natures)
	}
	setBoolAttrIfPresent(attributes, "did-stop-work", opts.DidStopWork)
	setStringAttrIfPresent(attributes, "net-impact-dollars", opts.NetImpactDollars)

	relationships := map[string]any{
		"subject": map[string]any{
			"data": map[string]string{
				"type": subjectType,
				"id":   subjectID,
			},
		},
	}

	if opts.Parent != "" {
		relationships["parent"] = map[string]any{
			"data": map[string]string{
				"type": "incidents",
				"id":   opts.Parent,
			},
		}
	}
	if opts.Equipment != "" {
		relationships["equipment"] = map[string]any{
			"data": map[string]string{
				"type": "equipment",
				"id":   opts.Equipment,
			},
		}
	}
	if opts.JobProductionPlan != "" {
		relationships["job-production-plan"] = map[string]any{
			"data": map[string]string{
				"type": "job-production-plans",
				"id":   opts.JobProductionPlan,
			},
		}
	}
	if opts.Assignee != "" {
		relationships["assignee"] = map[string]any{
			"data": map[string]string{
				"type": "users",
				"id":   opts.Assignee,
			},
		}
	}
	if opts.TenderJobScheduleShift != "" {
		relationships["tender-job-schedule-shift"] = map[string]any{
			"data": map[string]string{
				"type": "tender-job-schedule-shifts",
				"id":   opts.TenderJobScheduleShift,
			},
		}
	}

	data := map[string]any{
		"type":          "efficiency-incidents",
		"attributes":    attributes,
		"relationships": relationships,
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/efficiency-incidents", jsonBody)
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

	details := buildEfficiencyIncidentDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created efficiency incident %s (%s)\n", details.ID, subjectClass)
	return renderEfficiencyIncidentDetails(cmd, details)
}

func parseDoEfficiencyIncidentsCreateOptions(cmd *cobra.Command) (doEfficiencyIncidentsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	subject, _ := cmd.Flags().GetString("subject")
	startAt, _ := cmd.Flags().GetString("start-at")
	endAt, _ := cmd.Flags().GetString("end-at")
	status, _ := cmd.Flags().GetString("status")
	kind, _ := cmd.Flags().GetString("kind")
	severity, _ := cmd.Flags().GetString("severity")
	description, _ := cmd.Flags().GetString("description")
	headline, _ := cmd.Flags().GetString("headline")
	natures, _ := cmd.Flags().GetString("natures")
	didStopWork, _ := cmd.Flags().GetString("did-stop-work")
	netImpactDollars, _ := cmd.Flags().GetString("net-impact-dollars")
	assignee, _ := cmd.Flags().GetString("assignee")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	equipment, _ := cmd.Flags().GetString("equipment")
	parent, _ := cmd.Flags().GetString("parent")
	tenderJobScheduleShift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doEfficiencyIncidentsCreateOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		Subject:                subject,
		StartAt:                startAt,
		EndAt:                  endAt,
		Status:                 status,
		Kind:                   kind,
		Severity:               severity,
		Description:            description,
		Headline:               headline,
		Natures:                natures,
		DidStopWork:            didStopWork,
		NetImpactDollars:       netImpactDollars,
		Assignee:               assignee,
		JobProductionPlan:      jobProductionPlan,
		Equipment:              equipment,
		Parent:                 parent,
		TenderJobScheduleShift: tenderJobScheduleShift,
	}, nil
}
