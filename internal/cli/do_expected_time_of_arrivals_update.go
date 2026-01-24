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

type doExpectedTimeOfArrivalsUpdateOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	ID                     string
	TenderJobScheduleShift string
	ExpectedAt             string
	Note                   string
	Unsure                 string
}

func newDoExpectedTimeOfArrivalsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an expected time of arrival",
		Long: `Update an expected time of arrival.

Provide the expected time of arrival ID as an argument, then use flags
to specify which fields to update.

Updatable fields:
  --tender-job-schedule-shift  Tender job schedule shift ID
  --expected-at                Expected arrival time (ISO 8601)
  --note                       Notes
  --unsure                     Mark arrival time as unsure (true/false)`,
		Example: `  # Update expected arrival time
  xbe do expected-time-of-arrivals update 123 --expected-at 2025-01-15T12:30:00Z

  # Update note and unsure status
  xbe do expected-time-of-arrivals update 123 --unsure true --note "ETA pending"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoExpectedTimeOfArrivalsUpdate,
	}
	initDoExpectedTimeOfArrivalsUpdateFlags(cmd)
	return cmd
}

func init() {
	doExpectedTimeOfArrivalsCmd.AddCommand(newDoExpectedTimeOfArrivalsUpdateCmd())
}

func initDoExpectedTimeOfArrivalsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("tender-job-schedule-shift", "", "Tender job schedule shift ID")
	cmd.Flags().String("expected-at", "", "Expected arrival time (ISO 8601)")
	cmd.Flags().String("note", "", "Notes")
	cmd.Flags().String("unsure", "", "Mark arrival time as unsure (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoExpectedTimeOfArrivalsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoExpectedTimeOfArrivalsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("expected-at") {
		attributes["expected-at"] = opts.ExpectedAt
	}
	if cmd.Flags().Changed("note") {
		attributes["note"] = opts.Note
	}
	if cmd.Flags().Changed("unsure") {
		attributes["unsure"] = opts.Unsure == "true"
	}

	relationships := map[string]any{}
	if cmd.Flags().Changed("tender-job-schedule-shift") {
		relationships["tender-job-schedule-shift"] = map[string]any{
			"data": map[string]any{
				"type": "tender-job-schedule-shifts",
				"id":   opts.TenderJobScheduleShift,
			},
		}
	}

	if cmd.Flags().Changed("unsure") && opts.Unsure == "true" && cmd.Flags().Changed("expected-at") && strings.TrimSpace(opts.ExpectedAt) != "" {
		err := fmt.Errorf("--expected-at cannot be used with --unsure true")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no fields to update; specify --expected-at, --note, --unsure, or --tender-job-schedule-shift")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestData := map[string]any{
		"type": "expected-time-of-arrivals",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		requestData["attributes"] = attributes
	}
	if len(relationships) > 0 {
		requestData["relationships"] = relationships
	}

	requestBody := map[string]any{
		"data": requestData,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/expected-time-of-arrivals/"+opts.ID, jsonBody)
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

	row := buildExpectedTimeOfArrivalRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated expected time of arrival %s\n", row.ID)
	return nil
}

func parseDoExpectedTimeOfArrivalsUpdateOptions(cmd *cobra.Command, args []string) (doExpectedTimeOfArrivalsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	tenderJobScheduleShift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	expectedAt, _ := cmd.Flags().GetString("expected-at")
	note, _ := cmd.Flags().GetString("note")
	unsure, _ := cmd.Flags().GetString("unsure")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doExpectedTimeOfArrivalsUpdateOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		ID:                     args[0],
		TenderJobScheduleShift: tenderJobScheduleShift,
		ExpectedAt:             expectedAt,
		Note:                   note,
		Unsure:                 unsure,
	}, nil
}
