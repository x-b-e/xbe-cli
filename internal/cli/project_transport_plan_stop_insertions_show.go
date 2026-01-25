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

type projectTransportPlanStopInsertionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectTransportPlanStopInsertionDetails struct {
	ID                                string `json:"id"`
	Status                            string `json:"status,omitempty"`
	Mode                              string `json:"mode,omitempty"`
	ReuseHalf                         string `json:"reuse_half,omitempty"`
	BoundaryChoice                    string `json:"boundary_choice,omitempty"`
	PlannedEventTimeStartAt           string `json:"planned_event_time_start_at,omitempty"`
	PlannedEventTimeEndAt             string `json:"planned_event_time_end_at,omitempty"`
	PreserveStopOnDelete              bool   `json:"preserve_stop_on_delete,omitempty"`
	ErrorMessage                      string `json:"error_message,omitempty"`
	Result                            any    `json:"result,omitempty"`
	ProjectTransportPlanID            string `json:"project_transport_plan_id,omitempty"`
	ReferenceProjectTransportPlanStop string `json:"reference_project_transport_plan_stop_id,omitempty"`
	ProjectTransportLocationID        string `json:"project_transport_location_id,omitempty"`
	ProjectTransportPlanSegmentSetID  string `json:"project_transport_plan_segment_set_id,omitempty"`
	PlannedCompletionEventTypeID      string `json:"planned_completion_event_type_id,omitempty"`
	ExistingProjectTransportPlanStop  string `json:"existing_project_transport_plan_stop_id,omitempty"`
	StopToMoveID                      string `json:"stop_to_move_id,omitempty"`
	CreatedByID                       string `json:"created_by_id,omitempty"`
}

func newProjectTransportPlanStopInsertionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project transport plan stop insertion details",
		Long: `Show the full details of a project transport plan stop insertion.

Output Fields:
  ID
  Status
  Mode
  Reuse Half
  Boundary Choice
  Planned Event Time Start At
  Planned Event Time End At
  Preserve Stop On Delete
  Error Message
  Result
  Project Transport Plan ID
  Reference Project Transport Plan Stop ID
  Project Transport Location ID
  Project Transport Plan Segment Set ID
  Planned Completion Event Type ID
  Existing Project Transport Plan Stop ID
  Stop To Move ID
  Created By ID

Arguments:
  <id>    The stop insertion ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a stop insertion
  xbe view project-transport-plan-stop-insertions show 123

  # Output as JSON
  xbe view project-transport-plan-stop-insertions show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectTransportPlanStopInsertionsShow,
	}
	initProjectTransportPlanStopInsertionsShowFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanStopInsertionsCmd.AddCommand(newProjectTransportPlanStopInsertionsShowCmd())
}

func initProjectTransportPlanStopInsertionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanStopInsertionsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseProjectTransportPlanStopInsertionsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("project transport plan stop insertion id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-transport-plan-stop-insertions]", "mode,boundary-choice,reuse-half,status,result,error-message,planned-event-time-start-at,planned-event-time-end-at,preserve-stop-on-delete,project-transport-plan,reference-project-transport-plan-stop,project-transport-location,project-transport-plan-segment-set,planned-completion-event-type,existing-project-transport-plan-stop,stop-to-move,created-by")

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-stop-insertions/"+id, query)
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

	details := buildProjectTransportPlanStopInsertionDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectTransportPlanStopInsertionDetails(cmd, details)
}

func parseProjectTransportPlanStopInsertionsShowOptions(cmd *cobra.Command) (projectTransportPlanStopInsertionsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanStopInsertionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectTransportPlanStopInsertionDetails(resp jsonAPISingleResponse) projectTransportPlanStopInsertionDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := projectTransportPlanStopInsertionDetails{
		ID:                      resource.ID,
		Status:                  stringAttr(attrs, "status"),
		Mode:                    stringAttr(attrs, "mode"),
		ReuseHalf:               stringAttr(attrs, "reuse-half"),
		BoundaryChoice:          stringAttr(attrs, "boundary-choice"),
		PlannedEventTimeStartAt: formatDateTime(stringAttr(attrs, "planned-event-time-start-at")),
		PlannedEventTimeEndAt:   formatDateTime(stringAttr(attrs, "planned-event-time-end-at")),
		PreserveStopOnDelete:    boolAttr(attrs, "preserve-stop-on-delete"),
		ErrorMessage:            stringAttr(attrs, "error-message"),
		Result:                  attrs["result"],
	}

	if rel, ok := resource.Relationships["project-transport-plan"]; ok && rel.Data != nil {
		details.ProjectTransportPlanID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["reference-project-transport-plan-stop"]; ok && rel.Data != nil {
		details.ReferenceProjectTransportPlanStop = rel.Data.ID
	}
	if rel, ok := resource.Relationships["project-transport-location"]; ok && rel.Data != nil {
		details.ProjectTransportLocationID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["project-transport-plan-segment-set"]; ok && rel.Data != nil {
		details.ProjectTransportPlanSegmentSetID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["planned-completion-event-type"]; ok && rel.Data != nil {
		details.PlannedCompletionEventTypeID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["existing-project-transport-plan-stop"]; ok && rel.Data != nil {
		details.ExistingProjectTransportPlanStop = rel.Data.ID
	}
	if rel, ok := resource.Relationships["stop-to-move"]; ok && rel.Data != nil {
		details.StopToMoveID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
	}

	return details
}

func renderProjectTransportPlanStopInsertionDetails(cmd *cobra.Command, details projectTransportPlanStopInsertionDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.Mode != "" {
		fmt.Fprintf(out, "Mode: %s\n", details.Mode)
	}
	if details.ReuseHalf != "" {
		fmt.Fprintf(out, "Reuse Half: %s\n", details.ReuseHalf)
	}
	if details.BoundaryChoice != "" {
		fmt.Fprintf(out, "Boundary Choice: %s\n", details.BoundaryChoice)
	}
	if details.PlannedEventTimeStartAt != "" {
		fmt.Fprintf(out, "Planned Event Time Start At: %s\n", details.PlannedEventTimeStartAt)
	}
	if details.PlannedEventTimeEndAt != "" {
		fmt.Fprintf(out, "Planned Event Time End At: %s\n", details.PlannedEventTimeEndAt)
	}
	if details.PreserveStopOnDelete {
		fmt.Fprintf(out, "Preserve Stop On Delete: %s\n", formatBool(details.PreserveStopOnDelete))
	}
	if details.ErrorMessage != "" {
		fmt.Fprintf(out, "Error Message: %s\n", details.ErrorMessage)
	}
	if details.ProjectTransportPlanID != "" {
		fmt.Fprintf(out, "Project Transport Plan ID: %s\n", details.ProjectTransportPlanID)
	}
	if details.ReferenceProjectTransportPlanStop != "" {
		fmt.Fprintf(out, "Reference Project Transport Plan Stop ID: %s\n", details.ReferenceProjectTransportPlanStop)
	}
	if details.ProjectTransportLocationID != "" {
		fmt.Fprintf(out, "Project Transport Location ID: %s\n", details.ProjectTransportLocationID)
	}
	if details.ProjectTransportPlanSegmentSetID != "" {
		fmt.Fprintf(out, "Project Transport Plan Segment Set ID: %s\n", details.ProjectTransportPlanSegmentSetID)
	}
	if details.PlannedCompletionEventTypeID != "" {
		fmt.Fprintf(out, "Planned Completion Event Type ID: %s\n", details.PlannedCompletionEventTypeID)
	}
	if details.ExistingProjectTransportPlanStop != "" {
		fmt.Fprintf(out, "Existing Project Transport Plan Stop ID: %s\n", details.ExistingProjectTransportPlanStop)
	}
	if details.StopToMoveID != "" {
		fmt.Fprintf(out, "Stop To Move ID: %s\n", details.StopToMoveID)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By ID: %s\n", details.CreatedByID)
	}

	if details.Result != nil {
		pretty := formatJSONValue(details.Result)
		if pretty != "" {
			fmt.Fprintln(out, "")
			fmt.Fprintln(out, "Result:")
			fmt.Fprintln(out, strings.Repeat("-", 40))
			fmt.Fprintln(out, pretty)
		}
	}

	return nil
}
