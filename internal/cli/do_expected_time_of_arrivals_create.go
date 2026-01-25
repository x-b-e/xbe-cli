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

type doExpectedTimeOfArrivalsCreateOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	TenderJobScheduleShift string
	ExpectedAt             string
	Note                   string
	Unsure                 string
}

func newDoExpectedTimeOfArrivalsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new expected time of arrival",
		Long: `Create a new expected time of arrival update.

Required flags:
  --tender-job-schedule-shift  Tender job schedule shift ID (required)

One of:
  --expected-at                Expected arrival time (ISO 8601)
  --unsure                     Mark arrival time as unsure (true/false)

Optional flags:
  --note                       Additional notes`,
		Example: `  # Create an expected time of arrival
  xbe do expected-time-of-arrivals create \
    --tender-job-schedule-shift 123 \
    --expected-at 2025-01-15T12:00:00Z

  # Create an unsure expected time of arrival
  xbe do expected-time-of-arrivals create \
    --tender-job-schedule-shift 123 \
    --unsure true \
    --note "Awaiting confirmation"`,
		Args: cobra.NoArgs,
		RunE: runDoExpectedTimeOfArrivalsCreate,
	}
	initDoExpectedTimeOfArrivalsCreateFlags(cmd)
	return cmd
}

func init() {
	doExpectedTimeOfArrivalsCmd.AddCommand(newDoExpectedTimeOfArrivalsCreateCmd())
}

func initDoExpectedTimeOfArrivalsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("tender-job-schedule-shift", "", "Tender job schedule shift ID (required)")
	cmd.Flags().String("expected-at", "", "Expected arrival time (ISO 8601)")
	cmd.Flags().String("note", "", "Notes")
	cmd.Flags().String("unsure", "", "Mark arrival time as unsure (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoExpectedTimeOfArrivalsCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoExpectedTimeOfArrivalsCreateOptions(cmd)
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

	if opts.TenderJobScheduleShift == "" {
		err := fmt.Errorf("--tender-job-schedule-shift is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.ExpectedAt == "" && opts.Unsure != "true" {
		err := fmt.Errorf("either --expected-at or --unsure true is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.ExpectedAt != "" && opts.Unsure == "true" {
		err := fmt.Errorf("--expected-at cannot be used with --unsure true")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.ExpectedAt != "" {
		attributes["expected-at"] = opts.ExpectedAt
	}
	if opts.Note != "" {
		attributes["note"] = opts.Note
	}
	if opts.Unsure != "" {
		attributes["unsure"] = opts.Unsure == "true"
	}

	relationships := map[string]any{
		"tender-job-schedule-shift": map[string]any{
			"data": map[string]any{
				"type": "tender-job-schedule-shifts",
				"id":   opts.TenderJobScheduleShift,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "expected-time-of-arrivals",
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

	body, _, err := client.Post(cmd.Context(), "/v1/expected-time-of-arrivals", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created expected time of arrival %s\n", row.ID)
	return nil
}

func parseDoExpectedTimeOfArrivalsCreateOptions(cmd *cobra.Command) (doExpectedTimeOfArrivalsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	tenderJobScheduleShift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	expectedAt, _ := cmd.Flags().GetString("expected-at")
	note, _ := cmd.Flags().GetString("note")
	unsure, _ := cmd.Flags().GetString("unsure")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doExpectedTimeOfArrivalsCreateOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		TenderJobScheduleShift: tenderJobScheduleShift,
		ExpectedAt:             expectedAt,
		Note:                   note,
		Unsure:                 unsure,
	}, nil
}

func buildExpectedTimeOfArrivalRowFromSingle(resp jsonAPISingleResponse) expectedTimeOfArrivalRow {
	attrs := resp.Data.Attributes

	row := expectedTimeOfArrivalRow{
		ID:         resp.Data.ID,
		ExpectedAt: formatDateTime(stringAttr(attrs, "expected-at")),
		Note:       stringAttr(attrs, "note"),
		Unsure:     boolAttr(attrs, "unsure"),
	}

	if rel, ok := resp.Data.Relationships["tender-job-schedule-shift"]; ok && rel.Data != nil {
		row.TenderJobScheduleShift = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["job-schedule-shift"]; ok && rel.Data != nil {
		row.JobScheduleShift = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		row.CreatedBy = rel.Data.ID
	}

	return row
}
