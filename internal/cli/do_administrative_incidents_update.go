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

type doAdministrativeIncidentsUpdateOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	ID                     string
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
	NewType                string
	Assignee               string
	JobProductionPlan      string
	Equipment              string
	Parent                 string
	TenderJobScheduleShift string
}

func newDoAdministrativeIncidentsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an administrative incident",
		Long: `Update an existing administrative incident.

Only the fields you specify will be updated. Fields not provided remain unchanged.

Arguments:
  <id>    The administrative incident ID (required)

Flags:
  --subject                   Update subject (Type|ID, e.g. Broker|123)
  --start-at                  Update start timestamp (ISO 8601)
  --end-at                    Update end timestamp (ISO 8601)
  --status                    Update status (open, closed, abandoned, processing, parked)
  --kind                      Update kind (capacity, good_catch, near_miss, planning, quality, trucking)
  --severity                  Update severity (low, medium, high, catastrophic)
  --description               Update description
  --headline                  Update headline
  --natures                   Update natures (comma-separated: personal,property)
  --did-stop-work             Update did stop work (true/false)
  --net-impact-dollars        Update net impact dollars
  --new-type                  Update incident type (Incident subclass name)
  --assignee                  Update assignee user ID
  --job-production-plan        Update job production plan ID
  --equipment                 Update equipment ID
  --parent                    Update parent incident ID
  --tender-job-schedule-shift Update tender job schedule shift ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update status and end time
  xbe do administrative-incidents update 123 --status closed --end-at 2025-01-01T10:00:00Z

  # Update net impact dollars
  xbe do administrative-incidents update 123 --net-impact-dollars 1500`,
		Args: cobra.ExactArgs(1),
		RunE: runDoAdministrativeIncidentsUpdate,
	}
	initDoAdministrativeIncidentsUpdateFlags(cmd)
	return cmd
}

func init() {
	doAdministrativeIncidentsCmd.AddCommand(newDoAdministrativeIncidentsUpdateCmd())
}

func initDoAdministrativeIncidentsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("subject", "", "Subject in Type|ID format")
	cmd.Flags().String("start-at", "", "Start timestamp (ISO 8601)")
	cmd.Flags().String("end-at", "", "End timestamp (ISO 8601)")
	cmd.Flags().String("status", "", "Status (open, closed, abandoned, processing, parked)")
	cmd.Flags().String("kind", "", "Kind (capacity, good_catch, near_miss, planning, quality, trucking)")
	cmd.Flags().String("severity", "", "Severity (low, medium, high, catastrophic)")
	cmd.Flags().String("description", "", "Description text")
	cmd.Flags().String("headline", "", "Headline text")
	cmd.Flags().String("natures", "", "Natures (comma-separated: personal,property)")
	cmd.Flags().String("did-stop-work", "", "Did stop work (true/false)")
	cmd.Flags().String("net-impact-dollars", "", "Net impact dollars")
	cmd.Flags().String("new-type", "", "Incident subclass type name")
	cmd.Flags().String("assignee", "", "Assignee user ID")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID")
	cmd.Flags().String("equipment", "", "Equipment ID")
	cmd.Flags().String("parent", "", "Parent incident ID")
	cmd.Flags().String("tender-job-schedule-shift", "", "Tender job schedule shift ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoAdministrativeIncidentsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoAdministrativeIncidentsUpdateOptions(cmd, args)
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

	id := strings.TrimSpace(opts.ID)
	if id == "" {
		return fmt.Errorf("administrative incident id is required")
	}

	attributes := map[string]any{}
	setStringAttrIfPresent(attributes, "start-at", opts.StartAt)
	setStringAttrIfPresent(attributes, "end-at", opts.EndAt)
	setStringAttrIfPresent(attributes, "status", opts.Status)
	setStringAttrIfPresent(attributes, "kind", opts.Kind)
	setStringAttrIfPresent(attributes, "severity", opts.Severity)
	setStringAttrIfPresent(attributes, "description", opts.Description)
	setStringAttrIfPresent(attributes, "headline", opts.Headline)
	if opts.Natures != "" {
		attributes["natures"] = splitCommaList(opts.Natures)
	}
	setBoolAttrIfPresent(attributes, "did-stop-work", opts.DidStopWork)
	setStringAttrIfPresent(attributes, "net-impact-dollars", opts.NetImpactDollars)
	setStringAttrIfPresent(attributes, "new-type", opts.NewType)

	relationships := map[string]any{}
	if cmd.Flags().Changed("subject") {
		if strings.TrimSpace(opts.Subject) == "" {
			err := fmt.Errorf("--subject cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		_, subjectType, subjectID, err := parseIncidentSubjectRef(opts.Subject)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["subject"] = map[string]any{
			"data": map[string]string{
				"type": subjectType,
				"id":   subjectID,
			},
		}
	}
	if cmd.Flags().Changed("parent") {
		if strings.TrimSpace(opts.Parent) == "" {
			err := fmt.Errorf("--parent cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["parent"] = map[string]any{
			"data": map[string]string{
				"type": "incidents",
				"id":   opts.Parent,
			},
		}
	}
	if cmd.Flags().Changed("equipment") {
		if strings.TrimSpace(opts.Equipment) == "" {
			err := fmt.Errorf("--equipment cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["equipment"] = map[string]any{
			"data": map[string]string{
				"type": "equipment",
				"id":   opts.Equipment,
			},
		}
	}
	if cmd.Flags().Changed("job-production-plan") {
		if strings.TrimSpace(opts.JobProductionPlan) == "" {
			err := fmt.Errorf("--job-production-plan cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["job-production-plan"] = map[string]any{
			"data": map[string]string{
				"type": "job-production-plans",
				"id":   opts.JobProductionPlan,
			},
		}
	}
	if cmd.Flags().Changed("assignee") {
		if strings.TrimSpace(opts.Assignee) == "" {
			err := fmt.Errorf("--assignee cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["assignee"] = map[string]any{
			"data": map[string]string{
				"type": "users",
				"id":   opts.Assignee,
			},
		}
	}
	if cmd.Flags().Changed("tender-job-schedule-shift") {
		if strings.TrimSpace(opts.TenderJobScheduleShift) == "" {
			err := fmt.Errorf("--tender-job-schedule-shift cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["tender-job-schedule-shift"] = map[string]any{
			"data": map[string]string{
				"type": "tender-job-schedule-shifts",
				"id":   opts.TenderJobScheduleShift,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("at least one field or relationship to update is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"id":   id,
		"type": "administrative-incidents",
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
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

	body, _, err := client.Patch(cmd.Context(), "/v1/administrative-incidents/"+id, jsonBody)
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

	details := buildAdministrativeIncidentDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated administrative incident %s\n", details.ID)
	return renderAdministrativeIncidentDetails(cmd, details)
}

func parseDoAdministrativeIncidentsUpdateOptions(cmd *cobra.Command, args []string) (doAdministrativeIncidentsUpdateOptions, error) {
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
	newType, _ := cmd.Flags().GetString("new-type")
	assignee, _ := cmd.Flags().GetString("assignee")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	equipment, _ := cmd.Flags().GetString("equipment")
	parent, _ := cmd.Flags().GetString("parent")
	tenderJobScheduleShift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doAdministrativeIncidentsUpdateOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		ID:                     args[0],
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
		NewType:                newType,
		Assignee:               assignee,
		JobProductionPlan:      jobProductionPlan,
		Equipment:              equipment,
		Parent:                 parent,
		TenderJobScheduleShift: tenderJobScheduleShift,
	}, nil
}
