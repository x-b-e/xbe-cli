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

type doDriverAssignmentAcknowledgementsCreateOptions struct {
	BaseURL                  string
	Token                    string
	JSON                     bool
	TenderJobScheduleShiftID string
	DriverID                 string
}

func newDoDriverAssignmentAcknowledgementsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a driver assignment acknowledgement",
		Long: `Create a driver assignment acknowledgement.

Required flags:
  --tender-job-schedule-shift  Tender job schedule shift ID (required)
  --driver                     Driver user ID (required)`,
		Example: `  # Acknowledge a driver assignment
  xbe do driver-assignment-acknowledgements create \
    --tender-job-schedule-shift 123 \
    --driver 456

  # Get JSON output
  xbe do driver-assignment-acknowledgements create \
    --tender-job-schedule-shift 123 \
    --driver 456 \
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoDriverAssignmentAcknowledgementsCreate,
	}
	initDoDriverAssignmentAcknowledgementsCreateFlags(cmd)
	return cmd
}

func init() {
	doDriverAssignmentAcknowledgementsCmd.AddCommand(newDoDriverAssignmentAcknowledgementsCreateCmd())
}

func initDoDriverAssignmentAcknowledgementsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("tender-job-schedule-shift", "", "Tender job schedule shift ID (required)")
	cmd.Flags().String("driver", "", "Driver user ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoDriverAssignmentAcknowledgementsCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoDriverAssignmentAcknowledgementsCreateOptions(cmd)
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

	if opts.TenderJobScheduleShiftID == "" {
		err := fmt.Errorf("--tender-job-schedule-shift is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.DriverID == "" {
		err := fmt.Errorf("--driver is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"tender-job-schedule-shift": map[string]any{
			"data": map[string]any{
				"type": "tender-job-schedule-shifts",
				"id":   opts.TenderJobScheduleShiftID,
			},
		},
		"driver": map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.DriverID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "driver-assignment-acknowledgements",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/driver-assignment-acknowledgements", jsonBody)
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

	row := buildDriverAssignmentAcknowledgementRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created driver assignment acknowledgement %s\n", row.ID)
	return nil
}

func parseDoDriverAssignmentAcknowledgementsCreateOptions(cmd *cobra.Command) (doDriverAssignmentAcknowledgementsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	tenderJobScheduleShiftID, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	driverID, _ := cmd.Flags().GetString("driver")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doDriverAssignmentAcknowledgementsCreateOptions{
		BaseURL:                  baseURL,
		Token:                    token,
		JSON:                     jsonOut,
		TenderJobScheduleShiftID: tenderJobScheduleShiftID,
		DriverID:                 driverID,
	}, nil
}
