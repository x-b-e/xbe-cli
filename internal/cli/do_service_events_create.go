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

type doServiceEventsCreateOptions struct {
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

func newDoServiceEventsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a service event",
		Long: `Create a service event for a tender job schedule shift.

Required flags:
  --tender-job-schedule-shift  Tender job schedule shift ID (required)
  --occurred-at                Event timestamp (ISO 8601, required)
  --kind                       Event kind (ready_to_work, work_start_at)

Optional flags:
  --note                       Event note
  --occurred-latitude          Event latitude
  --occurred-longitude         Event longitude`,
		Example: `  # Create a ready-to-work event
  xbe do service-events create \
    --tender-job-schedule-shift 123 \
    --occurred-at 2026-01-23T12:00:00Z \
    --kind ready_to_work

  # Create a work-start event with coordinates and note
  xbe do service-events create \
    --tender-job-schedule-shift 123 \
    --occurred-at 2026-01-23T12:30:00Z \
    --kind work_start_at \
    --occurred-latitude 34.05 \
    --occurred-longitude -118.25 \
    --note "Started work"`,
		Args: cobra.NoArgs,
		RunE: runDoServiceEventsCreate,
	}
	initDoServiceEventsCreateFlags(cmd)
	return cmd
}

func init() {
	doServiceEventsCmd.AddCommand(newDoServiceEventsCreateCmd())
}

func initDoServiceEventsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("tender-job-schedule-shift", "", "Tender job schedule shift ID (required)")
	cmd.Flags().String("occurred-at", "", "Event timestamp (ISO 8601, required)")
	cmd.Flags().String("kind", "", "Event kind (ready_to_work, work_start_at)")
	cmd.Flags().String("note", "", "Event note")
	cmd.Flags().String("occurred-latitude", "", "Event latitude")
	cmd.Flags().String("occurred-longitude", "", "Event longitude")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoServiceEventsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoServiceEventsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.TenderJobScheduleShiftID) == "" {
		err := fmt.Errorf("--tender-job-schedule-shift is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.OccurredAt) == "" {
		err := fmt.Errorf("--occurred-at is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.Kind) == "" {
		err := fmt.Errorf("--kind is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if (strings.TrimSpace(opts.OccurredLatitude) == "") != (strings.TrimSpace(opts.OccurredLongitude) == "") {
		err := fmt.Errorf("--occurred-latitude and --occurred-longitude must be provided together")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"occurred-at": opts.OccurredAt,
		"kind":        opts.Kind,
	}
	if opts.Note != "" {
		attributes["note"] = opts.Note
	}
	if opts.OccurredLatitude != "" && opts.OccurredLongitude != "" {
		attributes["occurred-latitude"] = opts.OccurredLatitude
		attributes["occurred-longitude"] = opts.OccurredLongitude
	}

	relationships := map[string]any{
		"tender-job-schedule-shift": map[string]any{
			"data": map[string]any{
				"type": "tender-job-schedule-shifts",
				"id":   opts.TenderJobScheduleShiftID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "service-events",
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

	body, _, err := client.Post(cmd.Context(), "/v1/service-events", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created service event %s\n", row.ID)
	return nil
}

func parseDoServiceEventsCreateOptions(cmd *cobra.Command) (doServiceEventsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	tenderJobScheduleShiftID, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	occurredAt, _ := cmd.Flags().GetString("occurred-at")
	kind, _ := cmd.Flags().GetString("kind")
	note, _ := cmd.Flags().GetString("note")
	occurredLatitude, _ := cmd.Flags().GetString("occurred-latitude")
	occurredLongitude, _ := cmd.Flags().GetString("occurred-longitude")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doServiceEventsCreateOptions{
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
