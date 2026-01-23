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

type doDriverDayAdjustmentsCreateOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	DriverDayID    string
	AmountExplicit string
}

func newDoDriverDayAdjustmentsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a driver day adjustment",
		Long: `Create a driver day adjustment.

Required flags:
  --driver-day      Driver day ID (required)

Optional flags:
  --amount-explicit Explicit adjustment amount (overrides generated amount)`,
		Example: `  # Create an adjustment with explicit amount
  xbe do driver-day-adjustments create --driver-day 123 --amount-explicit "25.00"

  # Create an adjustment that uses the plan-generated amount
  xbe do driver-day-adjustments create --driver-day 123

  # Get JSON output
  xbe do driver-day-adjustments create --driver-day 123 --amount-explicit "25.00" --json`,
		Args: cobra.NoArgs,
		RunE: runDoDriverDayAdjustmentsCreate,
	}
	initDoDriverDayAdjustmentsCreateFlags(cmd)
	return cmd
}

func init() {
	doDriverDayAdjustmentsCmd.AddCommand(newDoDriverDayAdjustmentsCreateCmd())
}

func initDoDriverDayAdjustmentsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("driver-day", "", "Driver day ID (required)")
	cmd.Flags().String("amount-explicit", "", "Explicit adjustment amount")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoDriverDayAdjustmentsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoDriverDayAdjustmentsCreateOptions(cmd)
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

	if opts.DriverDayID == "" {
		err := fmt.Errorf("--driver-day is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.AmountExplicit != "" {
		attributes["amount-explicit"] = opts.AmountExplicit
	}

	relationships := map[string]any{
		"driver-day": map[string]any{
			"data": map[string]any{
				"type": "driver-days",
				"id":   opts.DriverDayID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "driver-day-adjustments",
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

	body, _, err := client.Post(cmd.Context(), "/v1/driver-day-adjustments", jsonBody)
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

	row := buildDriverDayAdjustmentRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created driver day adjustment %s\n", row.ID)
	return nil
}

func parseDoDriverDayAdjustmentsCreateOptions(cmd *cobra.Command) (doDriverDayAdjustmentsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	driverDayID, _ := cmd.Flags().GetString("driver-day")
	amountExplicit, _ := cmd.Flags().GetString("amount-explicit")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doDriverDayAdjustmentsCreateOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		DriverDayID:    driverDayID,
		AmountExplicit: amountExplicit,
	}, nil
}
