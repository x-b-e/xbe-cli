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

type shiftTimeCardRequisitionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type shiftTimeCardRequisitionDetails struct {
	ID                       string `json:"id"`
	Status                   string `json:"status,omitempty"`
	IsSubmitted              bool   `json:"is_submitted,omitempty"`
	InvalidWhenHourly        bool   `json:"invalid_when_hourly,omitempty"`
	TenderJobScheduleShiftID string `json:"tender_job_schedule_shift_id,omitempty"`
	TimeCardID               string `json:"time_card_id,omitempty"`
	CreatedByID              string `json:"created_by_id,omitempty"`
	CreatedByName            string `json:"created_by_name,omitempty"`
	FulfillmentErrorMessages any    `json:"fulfillment_error_messages,omitempty"`
}

func newShiftTimeCardRequisitionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show shift time card requisition details",
		Long: `Show the full details of a shift time card requisition.

Includes fulfillment status and related shift/time card details.

Arguments:
  <id>  The requisition ID (required).`,
		Example: `  # Show a shift time card requisition
  xbe view shift-time-card-requisitions show 123

  # Output as JSON
  xbe view shift-time-card-requisitions show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runShiftTimeCardRequisitionsShow,
	}
	initShiftTimeCardRequisitionsShowFlags(cmd)
	return cmd
}

func init() {
	shiftTimeCardRequisitionsCmd.AddCommand(newShiftTimeCardRequisitionsShowCmd())
}

func initShiftTimeCardRequisitionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runShiftTimeCardRequisitionsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseShiftTimeCardRequisitionsShowOptions(cmd)
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
		return fmt.Errorf("shift time card requisition id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[shift-time-card-requisitions]", "status,is-submitted,invalid-when-hourly,fulfillment-error-messages,tender-job-schedule-shift,time-card,created-by")
	query.Set("include", "created-by")
	query.Set("fields[users]", "name")

	body, _, err := client.Get(cmd.Context(), "/v1/shift-time-card-requisitions/"+id, query)
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

	details := buildShiftTimeCardRequisitionDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderShiftTimeCardRequisitionDetails(cmd, details)
}

func parseShiftTimeCardRequisitionsShowOptions(cmd *cobra.Command) (shiftTimeCardRequisitionsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return shiftTimeCardRequisitionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildShiftTimeCardRequisitionDetails(resp jsonAPISingleResponse) shiftTimeCardRequisitionDetails {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	attrs := resp.Data.Attributes
	details := shiftTimeCardRequisitionDetails{
		ID:                       resp.Data.ID,
		Status:                   stringAttr(attrs, "status"),
		IsSubmitted:              boolAttr(attrs, "is-submitted"),
		InvalidWhenHourly:        boolAttr(attrs, "invalid-when-hourly"),
		FulfillmentErrorMessages: anyAttr(attrs, "fulfillment-error-messages"),
	}

	if rel, ok := resp.Data.Relationships["tender-job-schedule-shift"]; ok && rel.Data != nil {
		details.TenderJobScheduleShiftID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["time-card"]; ok && rel.Data != nil {
		details.TimeCardID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.CreatedByName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
		}
	}

	return details
}

func renderShiftTimeCardRequisitionDetails(cmd *cobra.Command, details shiftTimeCardRequisitionDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	fmt.Fprintf(out, "Submitted: %s\n", formatBool(details.IsSubmitted))
	fmt.Fprintf(out, "Invalid When Hourly: %s\n", formatBool(details.InvalidWhenHourly))
	if details.TenderJobScheduleShiftID != "" {
		fmt.Fprintf(out, "Tender Job Schedule Shift: %s\n", details.TenderJobScheduleShiftID)
	}
	if details.TimeCardID != "" {
		fmt.Fprintf(out, "Time Card: %s\n", details.TimeCardID)
	}
	if details.CreatedByID != "" || details.CreatedByName != "" {
		fmt.Fprintf(out, "Created By: %s\n", formatRelated(details.CreatedByName, details.CreatedByID))
	}

	if details.FulfillmentErrorMessages != nil {
		if formatted := formatAnyJSON(details.FulfillmentErrorMessages); formatted != "" {
			fmt.Fprintln(out, "")
			fmt.Fprintln(out, "Fulfillment Errors:")
			fmt.Fprintln(out, strings.Repeat("-", 40))
			fmt.Fprintln(out, formatted)
		}
	}

	return nil
}
