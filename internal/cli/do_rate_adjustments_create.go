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

type doRateAdjustmentsCreateOptions struct {
	BaseURL                            string
	Token                              string
	JSON                               bool
	Rate                               string
	CostIndex                          string
	ZeroInterceptValue                 string
	ZeroInterceptRatio                 string
	AdjustmentMin                      string
	AdjustmentMax                      string
	PreventRatingWhenIndexValueMissing bool
}

func newDoRateAdjustmentsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a rate adjustment",
		Long: `Create a rate adjustment.

Required:
  --rate                   Rate ID
  --cost-index             Cost index ID
  --zero-intercept-value   Zero intercept value
  --zero-intercept-ratio   Zero intercept ratio (0 < ratio < 1)

Optional:
  --adjustment-min         Minimum adjustment value
  --adjustment-max         Maximum adjustment value
  --prevent-rating-when-index-value-missing  Prevent rating when index value is missing

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a rate adjustment
  xbe do rate-adjustments create --rate 123 --cost-index 456 \
    --zero-intercept-value 100 --zero-intercept-ratio 0.25

  # Create with bounds
  xbe do rate-adjustments create --rate 123 --cost-index 456 \
    --zero-intercept-value 100 --zero-intercept-ratio 0.25 \
    --adjustment-min 1.00 --adjustment-max 5.00

  # Create and prevent rating when index value is missing
  xbe do rate-adjustments create --rate 123 --cost-index 456 \
    --zero-intercept-value 100 --zero-intercept-ratio 0.25 \
    --prevent-rating-when-index-value-missing`,
		RunE: runDoRateAdjustmentsCreate,
	}
	initDoRateAdjustmentsCreateFlags(cmd)
	return cmd
}

func init() {
	doRateAdjustmentsCmd.AddCommand(newDoRateAdjustmentsCreateCmd())
}

func initDoRateAdjustmentsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("rate", "", "Rate ID")
	cmd.Flags().String("cost-index", "", "Cost index ID")
	cmd.Flags().String("zero-intercept-value", "", "Zero intercept value")
	cmd.Flags().String("zero-intercept-ratio", "", "Zero intercept ratio (0 < ratio < 1)")
	cmd.Flags().String("adjustment-min", "", "Minimum adjustment value")
	cmd.Flags().String("adjustment-max", "", "Maximum adjustment value")
	cmd.Flags().Bool("prevent-rating-when-index-value-missing", false, "Prevent rating when index value is missing")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("rate")
	_ = cmd.MarkFlagRequired("cost-index")
	_ = cmd.MarkFlagRequired("zero-intercept-value")
	_ = cmd.MarkFlagRequired("zero-intercept-ratio")
}

func runDoRateAdjustmentsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoRateAdjustmentsCreateOptions(cmd)
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

	attributes := map[string]any{
		"zero-intercept-value": opts.ZeroInterceptValue,
		"zero-intercept-ratio": opts.ZeroInterceptRatio,
	}
	if opts.AdjustmentMin != "" {
		attributes["adjustment-min"] = opts.AdjustmentMin
	}
	if opts.AdjustmentMax != "" {
		attributes["adjustment-max"] = opts.AdjustmentMax
	}
	if cmd.Flags().Changed("prevent-rating-when-index-value-missing") {
		attributes["prevent-rating-when-index-value-missing"] = opts.PreventRatingWhenIndexValueMissing
	}

	relationships := map[string]any{
		"rate": map[string]any{
			"data": map[string]any{
				"type": "rates",
				"id":   opts.Rate,
			},
		},
		"cost-index": map[string]any{
			"data": map[string]any{
				"type": "cost-indexes",
				"id":   opts.CostIndex,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "rate-adjustments",
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

	body, _, err := client.Post(cmd.Context(), "/v1/rate-adjustments", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created rate adjustment %s\n", row.ID)
	return nil
}

func parseDoRateAdjustmentsCreateOptions(cmd *cobra.Command) (doRateAdjustmentsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	rate, _ := cmd.Flags().GetString("rate")
	costIndex, _ := cmd.Flags().GetString("cost-index")
	zeroInterceptValue, _ := cmd.Flags().GetString("zero-intercept-value")
	zeroInterceptRatio, _ := cmd.Flags().GetString("zero-intercept-ratio")
	adjustmentMin, _ := cmd.Flags().GetString("adjustment-min")
	adjustmentMax, _ := cmd.Flags().GetString("adjustment-max")
	preventRatingWhenIndexValueMissing, _ := cmd.Flags().GetBool("prevent-rating-when-index-value-missing")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doRateAdjustmentsCreateOptions{
		BaseURL:                            baseURL,
		Token:                              token,
		JSON:                               jsonOut,
		Rate:                               rate,
		CostIndex:                          costIndex,
		ZeroInterceptValue:                 zeroInterceptValue,
		ZeroInterceptRatio:                 zeroInterceptRatio,
		AdjustmentMin:                      adjustmentMin,
		AdjustmentMax:                      adjustmentMax,
		PreventRatingWhenIndexValueMissing: preventRatingWhenIndexValueMissing,
	}, nil
}
