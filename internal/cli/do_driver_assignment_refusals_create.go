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

type doDriverAssignmentRefusalsCreateOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	TenderJobScheduleShift string
	Driver                 string
	Comment                string
}

func newDoDriverAssignmentRefusalsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a driver assignment refusal",
		Long: `Create a driver assignment refusal.

Required flags:
  --tender-job-schedule-shift   Tender job schedule shift ID (required)
  --driver                      Driver user ID (required)

Optional flags:
  --comment                     Refusal comment`,
		Example: `  # Refuse an assigned shift
  xbe do driver-assignment-refusals create \
    --tender-job-schedule-shift 123 \
    --driver 456 \
    --comment "Unable to cover the shift"`,
		Args: cobra.NoArgs,
		RunE: runDoDriverAssignmentRefusalsCreate,
	}
	initDoDriverAssignmentRefusalsCreateFlags(cmd)
	return cmd
}

func init() {
	doDriverAssignmentRefusalsCmd.AddCommand(newDoDriverAssignmentRefusalsCreateCmd())
}

func initDoDriverAssignmentRefusalsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("tender-job-schedule-shift", "", "Tender job schedule shift ID (required)")
	cmd.Flags().String("driver", "", "Driver user ID (required)")
	cmd.Flags().String("comment", "", "Refusal comment")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoDriverAssignmentRefusalsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoDriverAssignmentRefusalsCreateOptions(cmd)
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

	if opts.Driver == "" {
		err := fmt.Errorf("--driver is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.Comment != "" {
		attributes["comment"] = opts.Comment
	}

	relationships := map[string]any{
		"tender-job-schedule-shift": map[string]any{
			"data": map[string]any{
				"type": "tender-job-schedule-shifts",
				"id":   opts.TenderJobScheduleShift,
			},
		},
		"driver": map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.Driver,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "driver-assignment-refusals",
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

	body, _, err := client.Post(cmd.Context(), "/v1/driver-assignment-refusals", jsonBody)
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

	row := buildDriverAssignmentRefusalRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created driver assignment refusal %s\n", row.ID)
	return nil
}

func parseDoDriverAssignmentRefusalsCreateOptions(cmd *cobra.Command) (doDriverAssignmentRefusalsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	tenderJobScheduleShift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	driver, _ := cmd.Flags().GetString("driver")
	comment, _ := cmd.Flags().GetString("comment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doDriverAssignmentRefusalsCreateOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		TenderJobScheduleShift: tenderJobScheduleShift,
		Driver:                 driver,
		Comment:                comment,
	}, nil
}
