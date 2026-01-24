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

type doProjectTransportPlanEventTimesUpdateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	Kind    string
	StartAt string
	EndAt   string
	At      string
}

func newDoProjectTransportPlanEventTimesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a project transport plan event time",
		Long: `Update a project transport plan event time.

Provide the event time ID as an argument, then use flags to specify
which fields to update. Only specified fields will be modified.

Updatable fields:
  --kind       Time kind (planned, expected, actual, modeled)
  --start-at   Start timestamp (ISO 8601)
  --end-at     End timestamp (ISO 8601)
  --at         Legacy timestamp; sets start and end to the same value`,
		Example: `  # Update start and end timestamps
  xbe do project-transport-plan-event-times update 123 \\
    --start-at 2025-01-01T12:00:00Z \\
    --end-at 2025-01-01T12:15:00Z

  # Update kind
  xbe do project-transport-plan-event-times update 123 --kind expected`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProjectTransportPlanEventTimesUpdate,
	}
	initDoProjectTransportPlanEventTimesUpdateFlags(cmd)
	return cmd
}

func init() {
	doProjectTransportPlanEventTimesCmd.AddCommand(newDoProjectTransportPlanEventTimesUpdateCmd())
}

func initDoProjectTransportPlanEventTimesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("kind", "", "Time kind (planned, expected, actual, modeled)")
	cmd.Flags().String("start-at", "", "Start timestamp (ISO 8601)")
	cmd.Flags().String("end-at", "", "End timestamp (ISO 8601)")
	cmd.Flags().String("at", "", "Legacy timestamp; sets start and end to the same value")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectTransportPlanEventTimesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectTransportPlanEventTimesUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("kind") {
		attributes["kind"] = opts.Kind
	}
	if cmd.Flags().Changed("start-at") {
		attributes["start-at"] = opts.StartAt
	}
	if cmd.Flags().Changed("end-at") {
		attributes["end-at"] = opts.EndAt
	}
	if cmd.Flags().Changed("at") {
		attributes["at"] = opts.At
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update; specify at least one of --kind, --start-at, --end-at, --at")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "project-transport-plan-event-times",
			"id":         opts.ID,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/project-transport-plan-event-times/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated project transport plan event time %s\n", row.ID)
	return nil
}

func parseDoProjectTransportPlanEventTimesUpdateOptions(cmd *cobra.Command, args []string) (doProjectTransportPlanEventTimesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	kind, _ := cmd.Flags().GetString("kind")
	startAt, _ := cmd.Flags().GetString("start-at")
	endAt, _ := cmd.Flags().GetString("end-at")
	at, _ := cmd.Flags().GetString("at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectTransportPlanEventTimesUpdateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		Kind:    kind,
		StartAt: startAt,
		EndAt:   endAt,
		At:      at,
	}, nil
}
