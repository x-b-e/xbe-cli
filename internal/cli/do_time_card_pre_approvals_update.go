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

type doTimeCardPreApprovalsUpdateOptions struct {
	BaseURL                            string
	Token                              string
	JSON                               bool
	ID                                 string
	MaximumQuantitiesAttributes        string
	ExplicitStartAt                    string
	ExplicitEndAt                      string
	ExplicitDownMinutes                int
	ShouldAutomaticallyCreateAndSubmit bool
	AutomaticSubmissionDelayMinutes    int
	DelayAutomaticSubmissionAfterHours bool
	Note                               string
	SkipQuantityValidation             bool
}

func newDoTimeCardPreApprovalsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a time card pre-approval",
		Long: `Update a time card pre-approval.

Optional flags:
  --maximum-quantities-attributes   Maximum quantities (JSON array)
  --explicit-start-at               Explicit start time (ISO 8601)
  --explicit-end-at                 Explicit end time (ISO 8601)
  --explicit-down-minutes           Explicit down minutes
  --should-automatically-create-and-submit   Enable automatic submission
  --automatic-submission-delay-minutes      Auto-submit delay (minutes)
  --delay-automatic-submission-after-hours  Delay auto-submit after hours
  --note                            Note
  --skip-quantity-validation        Skip quantity validation`,
		Example: `  # Update note
  xbe do time-card-pre-approvals update 123 --note "Updated note"

  # Update explicit timing
  xbe do time-card-pre-approvals update 123 \
    --explicit-start-at 2026-01-23T07:00:00Z \
    --explicit-end-at 2026-01-23T15:00:00Z \
    --explicit-down-minutes 15`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTimeCardPreApprovalsUpdate,
	}
	initDoTimeCardPreApprovalsUpdateFlags(cmd)
	return cmd
}

func init() {
	doTimeCardPreApprovalsCmd.AddCommand(newDoTimeCardPreApprovalsUpdateCmd())
}

func initDoTimeCardPreApprovalsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("maximum-quantities-attributes", "", "Maximum quantities (JSON array)")
	cmd.Flags().String("explicit-start-at", "", "Explicit start time (ISO 8601)")
	cmd.Flags().String("explicit-end-at", "", "Explicit end time (ISO 8601)")
	cmd.Flags().Int("explicit-down-minutes", 0, "Explicit down minutes")
	cmd.Flags().Bool("should-automatically-create-and-submit", false, "Enable automatic submission")
	cmd.Flags().Int("automatic-submission-delay-minutes", 0, "Auto-submit delay (minutes)")
	cmd.Flags().Bool("delay-automatic-submission-after-hours", false, "Delay auto-submit after hours")
	cmd.Flags().String("note", "", "Note")
	cmd.Flags().Bool("skip-quantity-validation", false, "Skip quantity validation")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTimeCardPreApprovalsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTimeCardPreApprovalsUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("maximum-quantities-attributes") {
		if strings.TrimSpace(opts.MaximumQuantitiesAttributes) == "" {
			return fmt.Errorf("maximum-quantities-attributes must be valid JSON (use [] to clear)")
		}
		var maxQuantities any
		if err := json.Unmarshal([]byte(opts.MaximumQuantitiesAttributes), &maxQuantities); err != nil {
			return fmt.Errorf("invalid maximum-quantities-attributes JSON: %w", err)
		}
		attributes["maximum-quantities-attributes"] = maxQuantities
	}
	if cmd.Flags().Changed("explicit-start-at") {
		attributes["explicit-start-at"] = opts.ExplicitStartAt
	}
	if cmd.Flags().Changed("explicit-end-at") {
		attributes["explicit-end-at"] = opts.ExplicitEndAt
	}
	if cmd.Flags().Changed("explicit-down-minutes") {
		attributes["explicit-down-minutes"] = opts.ExplicitDownMinutes
	}
	if cmd.Flags().Changed("should-automatically-create-and-submit") {
		attributes["should-automatically-create-and-submit"] = opts.ShouldAutomaticallyCreateAndSubmit
	}
	if cmd.Flags().Changed("automatic-submission-delay-minutes") {
		attributes["automatic-submission-delay-minutes"] = opts.AutomaticSubmissionDelayMinutes
	}
	if cmd.Flags().Changed("delay-automatic-submission-after-hours") {
		attributes["delay-automatic-submission-after-hours"] = opts.DelayAutomaticSubmissionAfterHours
	}
	if cmd.Flags().Changed("note") {
		attributes["note"] = opts.Note
	}
	if cmd.Flags().Changed("skip-quantity-validation") {
		attributes["skip-quantity-validation"] = opts.SkipQuantityValidation
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"id":         opts.ID,
			"type":       "time-card-pre-approvals",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/time-card-pre-approvals/"+opts.ID, jsonBody)
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

	row := buildTimeCardPreApprovalRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated time card pre-approval %s\n", row.ID)
	return nil
}

func parseDoTimeCardPreApprovalsUpdateOptions(cmd *cobra.Command, args []string) (doTimeCardPreApprovalsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	maximumQuantitiesAttributes, _ := cmd.Flags().GetString("maximum-quantities-attributes")
	explicitStartAt, _ := cmd.Flags().GetString("explicit-start-at")
	explicitEndAt, _ := cmd.Flags().GetString("explicit-end-at")
	explicitDownMinutes, _ := cmd.Flags().GetInt("explicit-down-minutes")
	shouldAutomaticallyCreateAndSubmit, _ := cmd.Flags().GetBool("should-automatically-create-and-submit")
	automaticSubmissionDelayMinutes, _ := cmd.Flags().GetInt("automatic-submission-delay-minutes")
	delayAutomaticSubmissionAfterHours, _ := cmd.Flags().GetBool("delay-automatic-submission-after-hours")
	note, _ := cmd.Flags().GetString("note")
	skipQuantityValidation, _ := cmd.Flags().GetBool("skip-quantity-validation")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTimeCardPreApprovalsUpdateOptions{
		BaseURL:                            baseURL,
		Token:                              token,
		JSON:                               jsonOut,
		ID:                                 args[0],
		MaximumQuantitiesAttributes:        maximumQuantitiesAttributes,
		ExplicitStartAt:                    explicitStartAt,
		ExplicitEndAt:                      explicitEndAt,
		ExplicitDownMinutes:                explicitDownMinutes,
		ShouldAutomaticallyCreateAndSubmit: shouldAutomaticallyCreateAndSubmit,
		AutomaticSubmissionDelayMinutes:    automaticSubmissionDelayMinutes,
		DelayAutomaticSubmissionAfterHours: delayAutomaticSubmissionAfterHours,
		Note:                               note,
		SkipQuantityValidation:             skipQuantityValidation,
	}, nil
}

func buildTimeCardPreApprovalRowFromSingle(resp jsonAPISingleResponse) timeCardPreApprovalRow {
	row := buildTimeCardPreApprovalRow(resp.Data)
	return row
}
