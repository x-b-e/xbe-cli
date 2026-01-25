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

type doProjectTransportPlanEventTimesCreateOptions struct {
	BaseURL                   string
	Token                     string
	JSON                      bool
	ProjectTransportPlanEvent string
	Kind                      string
	StartAt                   string
	EndAt                     string
	At                        string
}

func newDoProjectTransportPlanEventTimesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project transport plan event time",
		Long: `Create a project transport plan event time.

Required flags:
  --project-transport-plan-event   Project transport plan event ID (required)
  --kind                           Time kind (planned, expected, actual, modeled) (required)
  --start-at or --at               Timestamp for the event time (ISO 8601) (required)

Optional flags:
  --end-at                         End timestamp (ISO 8601)
  --at                             Legacy timestamp; sets start and end to the same value`,
		Example: `  # Create an expected time for a plan event
  xbe do project-transport-plan-event-times create \\
    --project-transport-plan-event 123 \\
    --kind expected \\
    --start-at 2025-01-01T12:00:00Z \\
    --end-at 2025-01-01T12:15:00Z

  # Create a planned time using legacy --at
  xbe do project-transport-plan-event-times create \\
    --project-transport-plan-event 123 \\
    --kind planned \\
    --at 2025-01-01T12:00:00Z

  # JSON output
  xbe do project-transport-plan-event-times create \\
    --project-transport-plan-event 123 \\
    --kind expected \\
    --start-at 2025-01-01T12:00:00Z \\
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoProjectTransportPlanEventTimesCreate,
	}
	initDoProjectTransportPlanEventTimesCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectTransportPlanEventTimesCmd.AddCommand(newDoProjectTransportPlanEventTimesCreateCmd())
}

func initDoProjectTransportPlanEventTimesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project-transport-plan-event", "", "Project transport plan event ID (required)")
	cmd.Flags().String("kind", "", "Time kind (planned, expected, actual, modeled) (required)")
	cmd.Flags().String("start-at", "", "Start timestamp (ISO 8601)")
	cmd.Flags().String("end-at", "", "End timestamp (ISO 8601)")
	cmd.Flags().String("at", "", "Legacy timestamp; sets start and end to the same value")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectTransportPlanEventTimesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectTransportPlanEventTimesCreateOptions(cmd)
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

	if strings.TrimSpace(opts.ProjectTransportPlanEvent) == "" {
		err := fmt.Errorf("--project-transport-plan-event is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Kind) == "" {
		err := fmt.Errorf("--kind is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.StartAt) == "" && strings.TrimSpace(opts.At) == "" {
		err := fmt.Errorf("--start-at or --at is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if cmd.Flags().Changed("start-at") {
		attributes["start-at"] = opts.StartAt
	}
	if cmd.Flags().Changed("end-at") {
		attributes["end-at"] = opts.EndAt
	}
	if cmd.Flags().Changed("at") {
		attributes["at"] = opts.At
	}
	if cmd.Flags().Changed("kind") {
		attributes["kind"] = opts.Kind
	}

	relationships := map[string]any{
		"project-transport-plan-event": map[string]any{
			"data": map[string]any{
				"type": "project-transport-plan-events",
				"id":   opts.ProjectTransportPlanEvent,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "project-transport-plan-event-times",
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

	body, _, err := client.Post(cmd.Context(), "/v1/project-transport-plan-event-times", jsonBody)
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

	row := buildProjectTransportPlanEventTimeRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project transport plan event time %s\n", row.ID)
	return nil
}

func parseDoProjectTransportPlanEventTimesCreateOptions(cmd *cobra.Command) (doProjectTransportPlanEventTimesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	projectTransportPlanEvent, _ := cmd.Flags().GetString("project-transport-plan-event")
	kind, _ := cmd.Flags().GetString("kind")
	startAt, _ := cmd.Flags().GetString("start-at")
	endAt, _ := cmd.Flags().GetString("end-at")
	at, _ := cmd.Flags().GetString("at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectTransportPlanEventTimesCreateOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		ProjectTransportPlanEvent: projectTransportPlanEvent,
		Kind:                      kind,
		StartAt:                   startAt,
		EndAt:                     endAt,
		At:                        at,
	}, nil
}

func buildProjectTransportPlanEventTimeRowFromSingle(resp jsonAPISingleResponse) projectTransportPlanEventTimeRow {
	attrs := resp.Data.Attributes

	row := projectTransportPlanEventTimeRow{
		ID:      resp.Data.ID,
		Kind:    stringAttr(attrs, "kind"),
		StartAt: formatDateTime(stringAttr(attrs, "start-at")),
		EndAt:   formatDateTime(stringAttr(attrs, "end-at")),
	}

	if rel, ok := resp.Data.Relationships["project-transport-plan-event"]; ok && rel.Data != nil {
		row.ProjectTransportPlanEventID = rel.Data.ID
	}

	return row
}
