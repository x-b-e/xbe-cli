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

type doJobScheduleShiftSplitsCreateOptions struct {
	BaseURL                          string
	Token                            string
	JSON                             bool
	JobScheduleShift                 string
	ExpectedMaterialTransactionCount int
	ExpectedMaterialTransactionTons  float64
	NewStartAt                       string
}

func newDoJobScheduleShiftSplitsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Split a job schedule shift",
		Long: `Split a job schedule shift into a new shift.

Splits create a new job schedule shift from an existing flexible shift. You can
optionally set expected loads or adjust the new shift start time.

Required flags:
  --job-schedule-shift   Job schedule shift ID

Optional flags:
  --expected-material-transaction-count  Expected material transaction count
  --expected-material-transaction-tons   Expected material transaction tons
  --new-start-at                         New shift start time (RFC3339)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Split a shift into a new shift with expected count
  xbe do job-schedule-shift-splits create \\
    --job-schedule-shift 123 \\
    --expected-material-transaction-count 1

  # Split a shift and set tons and a new start time
  xbe do job-schedule-shift-splits create \\
    --job-schedule-shift 123 \\
    --expected-material-transaction-tons 12.5 \\
    --new-start-at 2026-01-23T08:30:00Z`,
		Args: cobra.NoArgs,
		RunE: runDoJobScheduleShiftSplitsCreate,
	}
	initDoJobScheduleShiftSplitsCreateFlags(cmd)
	return cmd
}

func init() {
	doJobScheduleShiftSplitsCmd.AddCommand(newDoJobScheduleShiftSplitsCreateCmd())
}

func initDoJobScheduleShiftSplitsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-schedule-shift", "", "Job schedule shift ID")
	cmd.Flags().Int("expected-material-transaction-count", 0, "Expected material transaction count")
	cmd.Flags().Float64("expected-material-transaction-tons", 0, "Expected material transaction tons")
	cmd.Flags().String("new-start-at", "", "New shift start time (RFC3339)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobScheduleShiftSplitsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoJobScheduleShiftSplitsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.JobScheduleShift) == "" {
		err := fmt.Errorf("--job-schedule-shift is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if cmd.Flags().Changed("expected-material-transaction-count") {
		attributes["expected-material-transaction-count"] = opts.ExpectedMaterialTransactionCount
	}
	if cmd.Flags().Changed("expected-material-transaction-tons") {
		attributes["expected-material-transaction-tons"] = opts.ExpectedMaterialTransactionTons
	}
	if cmd.Flags().Changed("new-start-at") && strings.TrimSpace(opts.NewStartAt) != "" {
		attributes["new-start-at"] = opts.NewStartAt
	}

	relationships := map[string]any{
		"job-schedule-shift": map[string]any{
			"data": map[string]any{
				"type": "job-schedule-shifts",
				"id":   opts.JobScheduleShift,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "job-schedule-shift-splits",
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

	body, _, err := client.Post(cmd.Context(), "/v1/job-schedule-shift-splits", jsonBody)
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

	row := buildJobScheduleShiftSplitRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created job schedule shift split %s\n", row.ID)
	return nil
}

func parseDoJobScheduleShiftSplitsCreateOptions(cmd *cobra.Command) (doJobScheduleShiftSplitsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobScheduleShift, _ := cmd.Flags().GetString("job-schedule-shift")
	expectedCount, _ := cmd.Flags().GetInt("expected-material-transaction-count")
	expectedTons, _ := cmd.Flags().GetFloat64("expected-material-transaction-tons")
	newStartAt, _ := cmd.Flags().GetString("new-start-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobScheduleShiftSplitsCreateOptions{
		BaseURL:                          baseURL,
		Token:                            token,
		JSON:                             jsonOut,
		JobScheduleShift:                 jobScheduleShift,
		ExpectedMaterialTransactionCount: expectedCount,
		ExpectedMaterialTransactionTons:  expectedTons,
		NewStartAt:                       newStartAt,
	}, nil
}
