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

type doShiftTimeCardRequisitionsCreateOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	InvalidWhenHourly      bool
	TenderJobScheduleShift string
}

func newDoShiftTimeCardRequisitionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a shift time card requisition",
		Long: `Create a shift time card requisition.

Required flags:
  --tender-job-schedule-shift  Tender job schedule shift ID (required)

Optional flags:
  --invalid-when-hourly        Reject hourly-quantified shifts (true/false, default: true)`,
		Example: `  # Create a requisition
  xbe do shift-time-card-requisitions create \
    --tender-job-schedule-shift 123

  # Allow hourly shifts
  xbe do shift-time-card-requisitions create \
    --tender-job-schedule-shift 123 \
    --invalid-when-hourly=false`,
		Args: cobra.NoArgs,
		RunE: runDoShiftTimeCardRequisitionsCreate,
	}
	initDoShiftTimeCardRequisitionsCreateFlags(cmd)
	return cmd
}

func init() {
	doShiftTimeCardRequisitionsCmd.AddCommand(newDoShiftTimeCardRequisitionsCreateCmd())
}

func initDoShiftTimeCardRequisitionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("invalid-when-hourly", true, "Reject hourly-quantified shifts (true/false)")
	cmd.Flags().String("tender-job-schedule-shift", "", "Tender job schedule shift ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoShiftTimeCardRequisitionsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoShiftTimeCardRequisitionsCreateOptions(cmd)
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

	attributes := map[string]any{}
	if cmd.Flags().Changed("invalid-when-hourly") {
		attributes["invalid-when-hourly"] = opts.InvalidWhenHourly
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
			"type":          "shift-time-card-requisitions",
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

	body, _, err := client.Post(cmd.Context(), "/v1/shift-time-card-requisitions", jsonBody)
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

	row := buildShiftTimeCardRequisitionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created shift time card requisition %s\n", row.ID)
	return nil
}

func parseDoShiftTimeCardRequisitionsCreateOptions(cmd *cobra.Command) (doShiftTimeCardRequisitionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	invalidWhenHourly, _ := cmd.Flags().GetBool("invalid-when-hourly")
	tenderJobScheduleShift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doShiftTimeCardRequisitionsCreateOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		InvalidWhenHourly:      invalidWhenHourly,
		TenderJobScheduleShift: tenderJobScheduleShift,
	}, nil
}

func buildShiftTimeCardRequisitionRowFromSingle(resp jsonAPISingleResponse) shiftTimeCardRequisitionRow {
	attrs := resp.Data.Attributes

	row := shiftTimeCardRequisitionRow{
		ID:          resp.Data.ID,
		Status:      stringAttr(attrs, "status"),
		IsSubmitted: boolAttr(attrs, "is-submitted"),
	}

	if rel, ok := resp.Data.Relationships["tender-job-schedule-shift"]; ok && rel.Data != nil {
		row.TenderJobScheduleShiftID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["time-card"]; ok && rel.Data != nil {
		row.TimeCardID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		row.CreatedByID = rel.Data.ID
	}

	return row
}
