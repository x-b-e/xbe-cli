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

type doRateAdjustmentsUpdateOptions struct {
	BaseURL                            string
	Token                              string
	JSON                               bool
	ID                                 string
	CostIndex                          string
	ZeroInterceptValue                 string
	ZeroInterceptRatio                 string
	AdjustmentMin                      string
	AdjustmentMax                      string
	PreventRatingWhenIndexValueMissing bool
}

func newDoRateAdjustmentsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a rate adjustment",
		Long: `Update a rate adjustment.

Optional:
  --cost-index             Cost index ID
  --zero-intercept-value   Zero intercept value
  --zero-intercept-ratio   Zero intercept ratio (0 < ratio < 1)
  --adjustment-min         Minimum adjustment value
  --adjustment-max         Maximum adjustment value
  --prevent-rating-when-index-value-missing  Prevent rating when index value is missing

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update intercept values
  xbe do rate-adjustments update 123 --zero-intercept-value 110 --zero-intercept-ratio 0.3

  # Update bounds
  xbe do rate-adjustments update 123 --adjustment-min 2.0 --adjustment-max 6.0

  # Update cost index
  xbe do rate-adjustments update 123 --cost-index 789

  # Update prevent rating flag
  xbe do rate-adjustments update 123 --prevent-rating-when-index-value-missing=false`,
		Args: cobra.ExactArgs(1),
		RunE: runDoRateAdjustmentsUpdate,
	}
	initDoRateAdjustmentsUpdateFlags(cmd)
	return cmd
}

func init() {
	doRateAdjustmentsCmd.AddCommand(newDoRateAdjustmentsUpdateCmd())
}

func initDoRateAdjustmentsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("cost-index", "", "Cost index ID")
	cmd.Flags().String("zero-intercept-value", "", "Zero intercept value")
	cmd.Flags().String("zero-intercept-ratio", "", "Zero intercept ratio (0 < ratio < 1)")
	cmd.Flags().String("adjustment-min", "", "Minimum adjustment value")
	cmd.Flags().String("adjustment-max", "", "Maximum adjustment value")
	cmd.Flags().Bool("prevent-rating-when-index-value-missing", false, "Prevent rating when index value is missing")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoRateAdjustmentsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoRateAdjustmentsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("zero-intercept-value") {
		attributes["zero-intercept-value"] = opts.ZeroInterceptValue
	}
	if cmd.Flags().Changed("zero-intercept-ratio") {
		attributes["zero-intercept-ratio"] = opts.ZeroInterceptRatio
	}
	if cmd.Flags().Changed("adjustment-min") {
		attributes["adjustment-min"] = opts.AdjustmentMin
	}
	if cmd.Flags().Changed("adjustment-max") {
		attributes["adjustment-max"] = opts.AdjustmentMax
	}
	if cmd.Flags().Changed("prevent-rating-when-index-value-missing") {
		attributes["prevent-rating-when-index-value-missing"] = opts.PreventRatingWhenIndexValueMissing
	}

	relationships := map[string]any{}
	if cmd.Flags().Changed("cost-index") {
		relationships["cost-index"] = map[string]any{
			"data": map[string]any{
				"type": "cost-indexes",
				"id":   opts.CostIndex,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no fields to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type":       "rate-adjustments",
		"id":         opts.ID,
		"attributes": attributes,
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/rate-adjustments/"+opts.ID, jsonBody)
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

	row := buildRateAdjustmentRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated rate adjustment %s\n", row.ID)
	return nil
}

func parseDoRateAdjustmentsUpdateOptions(cmd *cobra.Command, args []string) (doRateAdjustmentsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	costIndex, _ := cmd.Flags().GetString("cost-index")
	zeroInterceptValue, _ := cmd.Flags().GetString("zero-intercept-value")
	zeroInterceptRatio, _ := cmd.Flags().GetString("zero-intercept-ratio")
	adjustmentMin, _ := cmd.Flags().GetString("adjustment-min")
	adjustmentMax, _ := cmd.Flags().GetString("adjustment-max")
	preventRatingWhenIndexValueMissing, _ := cmd.Flags().GetBool("prevent-rating-when-index-value-missing")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doRateAdjustmentsUpdateOptions{
		BaseURL:                            baseURL,
		Token:                              token,
		JSON:                               jsonOut,
		ID:                                 args[0],
		CostIndex:                          costIndex,
		ZeroInterceptValue:                 zeroInterceptValue,
		ZeroInterceptRatio:                 zeroInterceptRatio,
		AdjustmentMin:                      adjustmentMin,
		AdjustmentMax:                      adjustmentMax,
		PreventRatingWhenIndexValueMissing: preventRatingWhenIndexValueMissing,
	}, nil
}
