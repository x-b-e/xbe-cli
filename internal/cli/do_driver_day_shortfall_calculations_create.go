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

type doDriverDayShortfallCalculationsCreateOptions struct {
	BaseURL                        string
	Token                          string
	JSON                           bool
	TimeCardIDs                    []string
	UnallocatableTimeCardIDs       []string
	DriverDayTimeCardConstraintIDs []string
}

func newDoDriverDayShortfallCalculationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a driver day shortfall calculation",
		Long: `Create a driver day shortfall calculation.

Required flags:
  --time-card-ids                        Time card IDs to include (comma-separated or repeated)
  --driver-day-time-card-constraint-ids  Driver day time card constraint IDs (comma-separated or repeated)

Optional flags:
  --unallocatable-time-card-ids          Time card IDs excluded from allocation (comma-separated or repeated)`,
		Example: `  # Calculate shortfall for time cards and constraints
  xbe do driver-day-shortfall-calculations create \\
    --time-card-ids 101,102 \\
    --driver-day-time-card-constraint-ids 55,56

  # Exclude time cards from allocation
  xbe do driver-day-shortfall-calculations create \\
    --time-card-ids 101,102,103 \\
    --unallocatable-time-card-ids 103 \\
    --driver-day-time-card-constraint-ids 55,56`,
		Args: cobra.NoArgs,
		RunE: runDoDriverDayShortfallCalculationsCreate,
	}
	initDoDriverDayShortfallCalculationsCreateFlags(cmd)
	return cmd
}

func init() {
	doDriverDayShortfallCalculationsCmd.AddCommand(newDoDriverDayShortfallCalculationsCreateCmd())
}

func initDoDriverDayShortfallCalculationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().StringSlice("time-card-ids", nil, "Time card IDs (comma-separated or repeated)")
	cmd.Flags().StringSlice("unallocatable-time-card-ids", nil, "Unallocatable time card IDs (comma-separated or repeated)")
	cmd.Flags().StringSlice("driver-day-time-card-constraint-ids", nil, "Driver day time card constraint IDs (comma-separated or repeated)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoDriverDayShortfallCalculationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoDriverDayShortfallCalculationsCreateOptions(cmd)
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

	if len(opts.TimeCardIDs) == 0 {
		err := fmt.Errorf("--time-card-ids is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if len(opts.DriverDayTimeCardConstraintIDs) == 0 {
		err := fmt.Errorf("--driver-day-time-card-constraint-ids is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"time-card-ids":                       opts.TimeCardIDs,
		"driver-day-time-card-constraint-ids": opts.DriverDayTimeCardConstraintIDs,
	}
	if len(opts.UnallocatableTimeCardIDs) > 0 {
		attributes["unallocatable-time-card-ids"] = opts.UnallocatableTimeCardIDs
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "driver-day-shortfall-calculations",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/driver-day-shortfall-calculations", jsonBody)
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

	details := buildDriverDayShortfallCalculationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderDriverDayShortfallCalculationDetails(cmd, details)
}

func parseDoDriverDayShortfallCalculationsCreateOptions(cmd *cobra.Command) (doDriverDayShortfallCalculationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	timeCardIDs, _ := cmd.Flags().GetStringSlice("time-card-ids")
	unallocatableTimeCardIDs, _ := cmd.Flags().GetStringSlice("unallocatable-time-card-ids")
	driverDayTimeCardConstraintIDs, _ := cmd.Flags().GetStringSlice("driver-day-time-card-constraint-ids")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doDriverDayShortfallCalculationsCreateOptions{
		BaseURL:                        baseURL,
		Token:                          token,
		JSON:                           jsonOut,
		TimeCardIDs:                    timeCardIDs,
		UnallocatableTimeCardIDs:       unallocatableTimeCardIDs,
		DriverDayTimeCardConstraintIDs: driverDayTimeCardConstraintIDs,
	}, nil
}
