package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type timeCardPreApprovalsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type timeCardPreApprovalDetails struct {
	ID                                 string `json:"id"`
	TenderJobScheduleShiftID           string `json:"tender_job_schedule_shift_id,omitempty"`
	CreatedByID                        string `json:"created_by_id,omitempty"`
	CreatedBy                          string `json:"created_by,omitempty"`
	CreatedByEmail                     string `json:"created_by_email,omitempty"`
	MaximumQuantitiesAttributes        any    `json:"maximum_quantities_attributes,omitempty"`
	ExplicitStartAt                    string `json:"explicit_start_at,omitempty"`
	ExplicitEndAt                      string `json:"explicit_end_at,omitempty"`
	ExplicitDownMinutes                int    `json:"explicit_down_minutes,omitempty"`
	ShouldAutomaticallyCreateAndSubmit bool   `json:"should_automatically_create_and_submit"`
	AutomaticSubmissionDelayMinutes    int    `json:"automatic_submission_delay_minutes,omitempty"`
	DelayAutomaticSubmissionAfterHours bool   `json:"delay_automatic_submission_after_hours"`
	Note                               string `json:"note,omitempty"`
	SkipQuantityValidation             bool   `json:"skip_quantity_validation"`
	SubmitAt                           string `json:"submit_at,omitempty"`
	CanUpdate                          bool   `json:"can_update"`
}

func newTimeCardPreApprovalsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show time card pre-approval details",
		Long: `Show full details of a time card pre-approval.

Output Fields:
  ID                                 Pre-approval identifier
  Tender Job Schedule Shift          Shift ID the pre-approval applies to
  Created By                         Creator (if available)
  Maximum Quantities Attributes      Maximum quantity overrides (JSON)
  Explicit Start/End/Down Minutes    Explicit time overrides
  Auto Submission                    Auto-submit settings
  Note                               Optional note
  Skip Quantity Validation           Whether quantity validation is skipped
  Submit At                          Calculated submission time
  Can Update                         Whether the pre-approval can be updated

Arguments:
  <id>    Time card pre-approval ID (required)

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a time card pre-approval
  xbe view time-card-pre-approvals show 123

  # JSON output
  xbe view time-card-pre-approvals show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTimeCardPreApprovalsShow,
	}
	initTimeCardPreApprovalsShowFlags(cmd)
	return cmd
}

func init() {
	timeCardPreApprovalsCmd.AddCommand(newTimeCardPreApprovalsShowCmd())
}

func initTimeCardPreApprovalsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTimeCardPreApprovalsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseTimeCardPreApprovalsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("time card pre-approval id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[time-card-pre-approvals]", "tender-job-schedule-shift,created-by,maximum-quantities-attributes,explicit-start-at,explicit-end-at,explicit-down-minutes,should-automatically-create-and-submit,automatic-submission-delay-minutes,delay-automatic-submission-after-hours,note,skip-quantity-validation,submit-at,can-update")
	query.Set("fields[users]", "name,email-address")
	query.Set("include", "created-by")

	body, _, err := client.Get(cmd.Context(), "/v1/time-card-pre-approvals/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildTimeCardPreApprovalDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTimeCardPreApprovalDetails(cmd, details)
}

func parseTimeCardPreApprovalsShowOptions(cmd *cobra.Command) (timeCardPreApprovalsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return timeCardPreApprovalsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTimeCardPreApprovalDetails(resp jsonAPISingleResponse) timeCardPreApprovalDetails {
	attrs := resp.Data.Attributes
	details := timeCardPreApprovalDetails{
		ID:                                 resp.Data.ID,
		MaximumQuantitiesAttributes:        anyAttr(attrs, "maximum-quantities-attributes"),
		ExplicitStartAt:                    formatDateTime(stringAttr(attrs, "explicit-start-at")),
		ExplicitEndAt:                      formatDateTime(stringAttr(attrs, "explicit-end-at")),
		ExplicitDownMinutes:                intAttr(attrs, "explicit-down-minutes"),
		ShouldAutomaticallyCreateAndSubmit: boolAttr(attrs, "should-automatically-create-and-submit"),
		AutomaticSubmissionDelayMinutes:    intAttr(attrs, "automatic-submission-delay-minutes"),
		DelayAutomaticSubmissionAfterHours: boolAttr(attrs, "delay-automatic-submission-after-hours"),
		Note:                               strings.TrimSpace(stringAttr(attrs, "note")),
		SkipQuantityValidation:             boolAttr(attrs, "skip-quantity-validation"),
		SubmitAt:                           formatDateTime(stringAttr(attrs, "submit-at")),
		CanUpdate:                          boolAttr(attrs, "can-update"),
	}

	createdByType := ""
	if rel, ok := resp.Data.Relationships["tender-job-schedule-shift"]; ok && rel.Data != nil {
		details.TenderJobScheduleShiftID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
		createdByType = rel.Data.Type
	}

	if len(resp.Included) == 0 || details.CreatedByID == "" || createdByType == "" {
		return details
	}

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	if user, ok := included[resourceKey(createdByType, details.CreatedByID)]; ok {
		details.CreatedBy = strings.TrimSpace(stringAttr(user.Attributes, "name"))
		details.CreatedByEmail = strings.TrimSpace(stringAttr(user.Attributes, "email-address"))
	}

	return details
}

func renderTimeCardPreApprovalDetails(cmd *cobra.Command, details timeCardPreApprovalDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TenderJobScheduleShiftID != "" {
		fmt.Fprintf(out, "Tender Job Schedule Shift: %s\n", details.TenderJobScheduleShiftID)
	}
	if details.CreatedBy != "" {
		if details.CreatedByID != "" {
			fmt.Fprintf(out, "Created By: %s (%s)\n", details.CreatedBy, details.CreatedByID)
		} else {
			fmt.Fprintf(out, "Created By: %s\n", details.CreatedBy)
		}
		if details.CreatedByEmail != "" {
			fmt.Fprintf(out, "Created By Email: %s\n", details.CreatedByEmail)
		}
	} else if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By ID: %s\n", details.CreatedByID)
	}

	fmt.Fprintf(out, "Explicit Start At: %s\n", formatOptional(details.ExplicitStartAt))
	fmt.Fprintf(out, "Explicit End At: %s\n", formatOptional(details.ExplicitEndAt))
	fmt.Fprintf(out, "Explicit Down Minutes: %d\n", details.ExplicitDownMinutes)
	fmt.Fprintf(out, "Auto Submit: %s\n", formatBool(details.ShouldAutomaticallyCreateAndSubmit))
	fmt.Fprintf(out, "Auto Submit Delay Minutes: %d\n", details.AutomaticSubmissionDelayMinutes)
	fmt.Fprintf(out, "Delay After Hours: %s\n", formatBool(details.DelayAutomaticSubmissionAfterHours))
	fmt.Fprintf(out, "Skip Quantity Validation: %s\n", formatBool(details.SkipQuantityValidation))
	fmt.Fprintf(out, "Submit At: %s\n", formatOptional(details.SubmitAt))
	fmt.Fprintf(out, "Can Update: %s\n", formatBool(details.CanUpdate))

	if details.Note != "" {
		fmt.Fprintf(out, "Note: %s\n", details.Note)
	}

	fmt.Fprintln(out, "Maximum Quantities Attributes:")
	if details.MaximumQuantitiesAttributes == nil {
		fmt.Fprintln(out, "  (none)")
	} else {
		formatted := formatAny(details.MaximumQuantitiesAttributes)
		if formatted == "" {
			fmt.Fprintln(out, "  (none)")
		} else {
			fmt.Fprintln(out, indentLines(formatted, "  "))
		}
	}

	return nil
}

func formatOptional(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "(none)"
	}
	return value
}
