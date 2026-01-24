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

type doTenderJobScheduleShiftTimeCardReviewsCreateOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	TenderJobScheduleShift string
}

func newDoTenderJobScheduleShiftTimeCardReviewsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a tender job schedule shift time card review",
		Long: `Create a tender job schedule shift time card review.

Required flags:
  --tender-job-schedule-shift   Tender job schedule shift ID (required)

Notes:
  Creating a review triggers an automated analysis that runs asynchronously.`,
		Example: `  # Create a time card review
  xbe do tender-job-schedule-shift-time-card-reviews create \
    --tender-job-schedule-shift 123

  # JSON output
  xbe do tender-job-schedule-shift-time-card-reviews create \
    --tender-job-schedule-shift 123 \
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoTenderJobScheduleShiftTimeCardReviewsCreate,
	}
	initDoTenderJobScheduleShiftTimeCardReviewsCreateFlags(cmd)
	return cmd
}

func init() {
	doTenderJobScheduleShiftTimeCardReviewsCmd.AddCommand(newDoTenderJobScheduleShiftTimeCardReviewsCreateCmd())
}

func initDoTenderJobScheduleShiftTimeCardReviewsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("tender-job-schedule-shift", "", "Tender job schedule shift ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTenderJobScheduleShiftTimeCardReviewsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTenderJobScheduleShiftTimeCardReviewsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.TenderJobScheduleShift) == "" {
		err := fmt.Errorf("--tender-job-schedule-shift is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
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
			"type":          "tender-job-schedule-shift-time-card-reviews",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/tender-job-schedule-shift-time-card-reviews", jsonBody)
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

	row := buildTenderJobScheduleShiftTimeCardReviewRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created tender job schedule shift time card review %s\n", row.ID)
	return nil
}

func parseDoTenderJobScheduleShiftTimeCardReviewsCreateOptions(cmd *cobra.Command) (doTenderJobScheduleShiftTimeCardReviewsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	tenderJobScheduleShift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTenderJobScheduleShiftTimeCardReviewsCreateOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		TenderJobScheduleShift: tenderJobScheduleShift,
	}, nil
}

func buildTenderJobScheduleShiftTimeCardReviewRowFromSingle(resp jsonAPISingleResponse) tenderJobScheduleShiftTimeCardReviewRow {
	return buildTenderJobScheduleShiftTimeCardReviewRow(resp.Data)
}
