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

type doTenderJobScheduleShiftCancellationsCreateOptions struct {
	BaseURL                              string
	Token                                string
	JSON                                 bool
	TenderJobScheduleShift               string
	StatusChangeComment                  string
	StatusChangedBy                      string
	IsReturned                           string
	JobProductionPlanCancellationComment string
	SkipTruckerNotifications             string
}

func newDoTenderJobScheduleShiftCancellationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Cancel a tender job schedule shift",
		Long: `Cancel a tender job schedule shift.

Required flags:
  --tender-job-schedule-shift  Tender job schedule shift ID (required)

Optional flags:
  --status-change-comment                    Status change comment
  --status-changed-by                        User ID for status change
  --is-returned                              Return the tender (true/false)
  --job-production-plan-cancellation-comment Job production plan cancellation comment
  --skip-trucker-notifications               Skip trucker notifications (true/false)`,
		Example: `  # Cancel a tender job schedule shift
  xbe do tender-job-schedule-shift-cancellations create --tender-job-schedule-shift 123

  # Cancel with comments and notification settings
  xbe do tender-job-schedule-shift-cancellations create \
    --tender-job-schedule-shift 123 \
    --status-change-comment "Cancelled due to weather" \
    --job-production-plan-cancellation-comment "Weather delay" \
    --skip-trucker-notifications true`,
		Args: cobra.NoArgs,
		RunE: runDoTenderJobScheduleShiftCancellationsCreate,
	}
	initDoTenderJobScheduleShiftCancellationsCreateFlags(cmd)
	return cmd
}

func init() {
	doTenderJobScheduleShiftCancellationsCmd.AddCommand(newDoTenderJobScheduleShiftCancellationsCreateCmd())
}

func initDoTenderJobScheduleShiftCancellationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("tender-job-schedule-shift", "", "Tender job schedule shift ID (required)")
	cmd.Flags().String("status-change-comment", "", "Status change comment")
	cmd.Flags().String("status-changed-by", "", "User ID for status change")
	cmd.Flags().String("is-returned", "", "Return the tender (true/false)")
	cmd.Flags().String("job-production-plan-cancellation-comment", "", "Job production plan cancellation comment")
	cmd.Flags().String("skip-trucker-notifications", "", "Skip trucker notifications (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTenderJobScheduleShiftCancellationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTenderJobScheduleShiftCancellationsCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run xbe auth login first.")
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

	attributes := map[string]any{}
	if strings.TrimSpace(opts.StatusChangeComment) != "" {
		attributes["status-change-comment"] = opts.StatusChangeComment
	}
	if strings.TrimSpace(opts.JobProductionPlanCancellationComment) != "" {
		attributes["job-production-plan-cancellation-comment"] = opts.JobProductionPlanCancellationComment
	}
	if strings.TrimSpace(opts.IsReturned) != "" {
		value, err := parseTenderJobScheduleShiftCancellationBool(opts.IsReturned, "is-returned")
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["is-returned"] = value
	}
	if strings.TrimSpace(opts.SkipTruckerNotifications) != "" {
		value, err := parseTenderJobScheduleShiftCancellationBool(opts.SkipTruckerNotifications, "skip-trucker-notifications")
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["skip-trucker-notifications"] = value
	}

	relationships := map[string]any{
		"tender-job-schedule-shift": map[string]any{
			"data": map[string]any{
				"type": "tender-job-schedule-shifts",
				"id":   opts.TenderJobScheduleShift,
			},
		},
	}
	if strings.TrimSpace(opts.StatusChangedBy) != "" {
		relationships["status-changed-by"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.StatusChangedBy,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "tender-job-schedule-shift-cancellations",
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

	body, _, err := client.Post(cmd.Context(), "/v1/tender-job-schedule-shift-cancellations", jsonBody)
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

	row := tenderJobScheduleShiftCancellationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created tender job schedule shift cancellation %s\n", row.ID)
	return nil
}

func tenderJobScheduleShiftCancellationRowFromSingle(resp jsonAPISingleResponse) tenderJobScheduleShiftCancellationRow {
	attrs := resp.Data.Attributes
	row := tenderJobScheduleShiftCancellationRow{
		ID:                                   resp.Data.ID,
		StatusChangeComment:                  stringAttr(attrs, "status-change-comment"),
		IsReturned:                           boolAttr(attrs, "is-returned"),
		JobProductionPlanCancellationComment: stringAttr(attrs, "job-production-plan-cancellation-comment"),
		SkipTruckerNotifications:             boolAttr(attrs, "skip-trucker-notifications"),
	}

	if rel, ok := resp.Data.Relationships["tender-job-schedule-shift"]; ok && rel.Data != nil {
		row.TenderJobScheduleShiftID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["status-changed-by"]; ok && rel.Data != nil {
		row.StatusChangedByID = rel.Data.ID
	}

	return row
}

func parseDoTenderJobScheduleShiftCancellationsCreateOptions(cmd *cobra.Command) (doTenderJobScheduleShiftCancellationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	shiftID, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	statusChangeComment, _ := cmd.Flags().GetString("status-change-comment")
	statusChangedBy, _ := cmd.Flags().GetString("status-changed-by")
	isReturned, _ := cmd.Flags().GetString("is-returned")
	jobProductionPlanCancellationComment, _ := cmd.Flags().GetString("job-production-plan-cancellation-comment")
	skipTruckerNotifications, _ := cmd.Flags().GetString("skip-trucker-notifications")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTenderJobScheduleShiftCancellationsCreateOptions{
		BaseURL:                              baseURL,
		Token:                                token,
		JSON:                                 jsonOut,
		TenderJobScheduleShift:               shiftID,
		StatusChangeComment:                  statusChangeComment,
		StatusChangedBy:                      statusChangedBy,
		IsReturned:                           isReturned,
		JobProductionPlanCancellationComment: jobProductionPlanCancellationComment,
		SkipTruckerNotifications:             skipTruckerNotifications,
	}, nil
}

func parseTenderJobScheduleShiftCancellationBool(value string, flagName string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, fmt.Errorf("--%s must be true or false", flagName)
	}
}
