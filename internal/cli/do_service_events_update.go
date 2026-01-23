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

type doServiceEventsUpdateOptions struct {
	BaseURL                  string
	Token                    string
	JSON                     bool
	TenderJobScheduleShiftID string
	OccurredAt               string
	Kind                     string
	Note                     string
	OccurredLatitude         string
	OccurredLongitude        string
}

func newDoServiceEventsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a service event",
		Long: `Update a service event.

Arguments:
  <id>    The service event ID (required).

Optional flags:
  --tender-job-schedule-shift  Tender job schedule shift ID
  --occurred-at                Event timestamp (ISO 8601)
  --kind                       Event kind (ready_to_work, work_start_at)
  --note                       Event note
  --occurred-latitude          Event latitude
  --occurred-longitude         Event longitude`,
		Example: `  # Update a service event note
  xbe do service-events update 123 --note "Updated note"

  # Update occurred time and kind
  xbe do service-events update 123 \
    --occurred-at 2026-01-23T12:15:00Z \
    --kind work_start_at`,
		Args: cobra.ExactArgs(1),
		RunE: runDoServiceEventsUpdate,
	}
	initDoServiceEventsUpdateFlags(cmd)
	return cmd
}

func init() {
	doServiceEventsCmd.AddCommand(newDoServiceEventsUpdateCmd())
}

func initDoServiceEventsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("tender-job-schedule-shift", "", "Tender job schedule shift ID")
	cmd.Flags().String("occurred-at", "", "Event timestamp (ISO 8601)")
	cmd.Flags().String("kind", "", "Event kind (ready_to_work, work_start_at)")
	cmd.Flags().String("note", "", "Event note")
	cmd.Flags().String("occurred-latitude", "", "Event latitude")
	cmd.Flags().String("occurred-longitude", "", "Event longitude")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoServiceEventsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoServiceEventsUpdateOptions(cmd)
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("service event id is required")
	}

	tenderJobScheduleShiftChanged := cmd.Flags().Changed("tender-job-schedule-shift")
	occurredAtChanged := cmd.Flags().Changed("occurred-at")
	kindChanged := cmd.Flags().Changed("kind")
	noteChanged := cmd.Flags().Changed("note")
	latitudeChanged := cmd.Flags().Changed("occurred-latitude")
	longitudeChanged := cmd.Flags().Changed("occurred-longitude")

	if !tenderJobScheduleShiftChanged && !occurredAtChanged && !kindChanged && !noteChanged && !latitudeChanged && !longitudeChanged {
		err := fmt.Errorf("no fields provided to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if tenderJobScheduleShiftChanged && strings.TrimSpace(opts.TenderJobScheduleShiftID) == "" {
		err := fmt.Errorf("--tender-job-schedule-shift cannot be empty")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if occurredAtChanged && strings.TrimSpace(opts.OccurredAt) == "" {
		err := fmt.Errorf("--occurred-at cannot be empty")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if kindChanged && strings.TrimSpace(opts.Kind) == "" {
		err := fmt.Errorf("--kind cannot be empty")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if latitudeChanged != longitudeChanged {
		err := fmt.Errorf("--occurred-latitude and --occurred-longitude must be provided together")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if occurredAtChanged {
		attributes["occurred-at"] = opts.OccurredAt
	}
	if kindChanged {
		attributes["kind"] = opts.Kind
	}
	if noteChanged {
		attributes["note"] = opts.Note
	}
	if latitudeChanged && longitudeChanged {
		attributes["occurred-latitude"] = opts.OccurredLatitude
		attributes["occurred-longitude"] = opts.OccurredLongitude
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "service-events",
			"id":         id,
			"attributes": attributes,
		},
	}

	if tenderJobScheduleShiftChanged {
		requestBody["data"].(map[string]any)["relationships"] = map[string]any{
			"tender-job-schedule-shift": map[string]any{
				"data": map[string]any{
					"type": "tender-job-schedule-shifts",
					"id":   opts.TenderJobScheduleShiftID,
				},
			},
		}
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/service-events/"+id, jsonBody)
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

	row := buildServiceEventRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated service event %s\n", row.ID)
	return nil
}

func parseDoServiceEventsUpdateOptions(cmd *cobra.Command) (doServiceEventsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	tenderJobScheduleShiftID, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	occurredAt, _ := cmd.Flags().GetString("occurred-at")
	kind, _ := cmd.Flags().GetString("kind")
	note, _ := cmd.Flags().GetString("note")
	occurredLatitude, _ := cmd.Flags().GetString("occurred-latitude")
	occurredLongitude, _ := cmd.Flags().GetString("occurred-longitude")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doServiceEventsUpdateOptions{
		BaseURL:                  baseURL,
		Token:                    token,
		JSON:                     jsonOut,
		TenderJobScheduleShiftID: tenderJobScheduleShiftID,
		OccurredAt:               occurredAt,
		Kind:                     kind,
		Note:                     note,
		OccurredLatitude:         occurredLatitude,
		OccurredLongitude:        occurredLongitude,
	}, nil
}
