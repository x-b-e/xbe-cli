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

type doDriverDayTripsAdjustmentsCreateOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	Description            string
	Status                 string
	OldTripsAttributes     string
	TenderJobScheduleShift string
}

func newDoDriverDayTripsAdjustmentsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a driver day trips adjustment",
		Long: `Create a driver day trips adjustment.

Required flags:
  --tender-job-schedule-shift  Tender job schedule shift ID (required)
  --old-trips-attributes       JSON array of original trip attributes (required)

Optional flags:
  --description                Description of the adjustment
  --status                     Adjustment status (e.g., editing)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create an adjustment
  xbe do driver-day-trips-adjustments create \
    --tender-job-schedule-shift 123 \
    --old-trips-attributes '[{\"note\":\"Original trip\"}]' \
    --description \"Adjust trip order\" \
    --status editing`,
		Args: cobra.NoArgs,
		RunE: runDoDriverDayTripsAdjustmentsCreate,
	}
	initDoDriverDayTripsAdjustmentsCreateFlags(cmd)
	return cmd
}

func init() {
	doDriverDayTripsAdjustmentsCmd.AddCommand(newDoDriverDayTripsAdjustmentsCreateCmd())
}

func initDoDriverDayTripsAdjustmentsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("tender-job-schedule-shift", "", "Tender job schedule shift ID (required)")
	cmd.Flags().String("old-trips-attributes", "", "JSON array of original trip attributes (required)")
	cmd.Flags().String("description", "", "Description of the adjustment")
	cmd.Flags().String("status", "", "Adjustment status (e.g., editing)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoDriverDayTripsAdjustmentsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoDriverDayTripsAdjustmentsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.OldTripsAttributes) == "" {
		err := fmt.Errorf("--old-trips-attributes is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var oldTrips any
	if err := json.Unmarshal([]byte(opts.OldTripsAttributes), &oldTrips); err != nil {
		return fmt.Errorf("invalid old-trips-attributes JSON: %w", err)
	}
	if _, ok := oldTrips.([]any); !ok {
		return fmt.Errorf("--old-trips-attributes must be a JSON array")
	}

	attributes := map[string]any{
		"old-trips-attributes": oldTrips,
	}
	if opts.Description != "" {
		attributes["description"] = opts.Description
	}
	if opts.Status != "" {
		attributes["status"] = opts.Status
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
			"type":          "driver-day-trips-adjustments",
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

	body, _, err := client.Post(cmd.Context(), "/v1/driver-day-trips-adjustments", jsonBody)
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

	row := driverDayTripsAdjustmentRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created driver day trips adjustment %s\n", row.ID)
	return nil
}

func parseDoDriverDayTripsAdjustmentsCreateOptions(cmd *cobra.Command) (doDriverDayTripsAdjustmentsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	tenderJobScheduleShift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	oldTripsAttributes, _ := cmd.Flags().GetString("old-trips-attributes")
	description, _ := cmd.Flags().GetString("description")
	status, _ := cmd.Flags().GetString("status")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doDriverDayTripsAdjustmentsCreateOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		Description:            description,
		Status:                 status,
		OldTripsAttributes:     oldTripsAttributes,
		TenderJobScheduleShift: tenderJobScheduleShift,
	}, nil
}
