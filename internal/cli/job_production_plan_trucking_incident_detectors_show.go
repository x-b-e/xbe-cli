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

type jobProductionPlanTruckingIncidentDetectorsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type jobProductionPlanTruckingIncidentDetectorDetails struct {
	ID                       string           `json:"id"`
	JobProductionPlanID      string           `json:"job_production_plan_id,omitempty"`
	AsOf                     string           `json:"as_of,omitempty"`
	PersistChanges           bool             `json:"persist_changes"`
	IsPerformed              bool             `json:"is_performed"`
	DetectedIncidents        []map[string]any `json:"detected_incidents,omitempty"`
	PersistedIncidentChanges []map[string]any `json:"persisted_incident_changes,omitempty"`
}

func newJobProductionPlanTruckingIncidentDetectorsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show job production plan trucking incident detector details",
		Long: `Show the full details of a job production plan trucking incident detector.

Output Fields:
  ID                        Detector identifier
  Job Production Plan       Associated job production plan ID
  As Of                     As-of timestamp for detection
  Persist Changes           Whether incident changes are persisted
  Is Performed              Whether detection has been performed
  Detected Incidents         Detected incident payloads
  Persisted Incident Changes Persisted incident change payloads

Arguments:
  <id>    The detector ID (required). You can find IDs using the list command.`,
		Example: `  # Show a trucking incident detector
  xbe view job-production-plan-trucking-incident-detectors show 123

  # Get JSON output
  xbe view job-production-plan-trucking-incident-detectors show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runJobProductionPlanTruckingIncidentDetectorsShow,
	}
	initJobProductionPlanTruckingIncidentDetectorsShowFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanTruckingIncidentDetectorsCmd.AddCommand(newJobProductionPlanTruckingIncidentDetectorsShowCmd())
}

func initJobProductionPlanTruckingIncidentDetectorsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanTruckingIncidentDetectorsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseJobProductionPlanTruckingIncidentDetectorsShowOptions(cmd)
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
		return fmt.Errorf("job production plan trucking incident detector id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-trucking-incident-detectors/"+id, nil)
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

	details := buildJobProductionPlanTruckingIncidentDetectorDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderJobProductionPlanTruckingIncidentDetectorDetails(cmd, details)
}

func parseJobProductionPlanTruckingIncidentDetectorsShowOptions(cmd *cobra.Command) (jobProductionPlanTruckingIncidentDetectorsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanTruckingIncidentDetectorsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildJobProductionPlanTruckingIncidentDetectorDetails(resp jsonAPISingleResponse) jobProductionPlanTruckingIncidentDetectorDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := jobProductionPlanTruckingIncidentDetectorDetails{
		ID:                       resource.ID,
		AsOf:                     formatDateTime(stringAttr(attrs, "as-of")),
		PersistChanges:           boolAttr(attrs, "persist-changes"),
		IsPerformed:              boolAttr(attrs, "is-performed"),
		DetectedIncidents:        mapSliceAttr(attrs, "detected-incidents"),
		PersistedIncidentChanges: mapSliceAttr(attrs, "persisted-incident-changes"),
	}

	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		details.JobProductionPlanID = rel.Data.ID
	}

	return details
}

func renderJobProductionPlanTruckingIncidentDetectorDetails(cmd *cobra.Command, details jobProductionPlanTruckingIncidentDetectorDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.JobProductionPlanID != "" {
		fmt.Fprintf(out, "Job Production Plan: %s\n", details.JobProductionPlanID)
	}
	if details.AsOf != "" {
		fmt.Fprintf(out, "As Of: %s\n", details.AsOf)
	}
	fmt.Fprintf(out, "Persist Changes: %t\n", details.PersistChanges)
	fmt.Fprintf(out, "Is Performed: %t\n", details.IsPerformed)

	if len(details.DetectedIncidents) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Detected Incidents:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, formatJSON(details.DetectedIncidents))
	}

	if len(details.PersistedIncidentChanges) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Persisted Incident Changes:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, formatJSON(details.PersistedIncidentChanges))
	}

	return nil
}

func mapSliceAttr(attrs map[string]any, key string) []map[string]any {
	if attrs == nil {
		return nil
	}
	value, ok := attrs[key]
	if !ok || value == nil {
		return nil
	}
	switch typed := value.(type) {
	case []map[string]any:
		return typed
	case []any:
		items := make([]map[string]any, 0, len(typed))
		for _, item := range typed {
			if item == nil {
				continue
			}
			if mapped, ok := item.(map[string]any); ok {
				items = append(items, mapped)
				continue
			}
			items = append(items, map[string]any{"value": item})
		}
		return items
	default:
		return []map[string]any{{"value": typed}}
	}
}
