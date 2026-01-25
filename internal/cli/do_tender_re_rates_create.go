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

type doTenderReRatesCreateOptions struct {
	BaseURL                               string
	Token                                 string
	JSON                                  bool
	TenderIDs                             []string
	ReRate                                bool
	ReConstrain                           bool
	UpdateTimeCardQuantities              bool
	SkipUpdateTravelMinutes               bool
	SkipValidateCustomerTenderHourlyRates bool
}

func newDoTenderReRatesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Re-rate tenders",
		Long: `Re-rate tenders and optionally re-constrain rates.

Required flags:
  --tender-ids   Tender IDs to re-rate (comma-separated or repeated)
  --re-rate or --re-constrain  At least one action must be enabled

Optional flags:
  --update-time-card-quantities              Update time card quantities after re-rate
  --skip-update-travel-minutes               Skip updating travel minutes (deprecated)
  --skip-validate-customer-tender-hourly-rates  Skip validation of customer tender hourly rates`,
		Example: `  # Re-rate tenders
  xbe do tender-re-rates create --tender-ids 123,124 --re-rate

  # Re-constrain and skip validation
  xbe do tender-re-rates create --tender-ids 123 \\
    --re-constrain --skip-validate-customer-tender-hourly-rates

  # Re-rate and disable time card quantity updates
  xbe do tender-re-rates create --tender-ids 123 --re-rate \\
    --update-time-card-quantities=false

  # JSON output
  xbe do tender-re-rates create --tender-ids 123 --re-rate --json`,
		Args: cobra.NoArgs,
		RunE: runDoTenderReRatesCreate,
	}
	initDoTenderReRatesCreateFlags(cmd)
	return cmd
}

func init() {
	doTenderReRatesCmd.AddCommand(newDoTenderReRatesCreateCmd())
}

func initDoTenderReRatesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().StringSlice("tender-ids", nil, "Tender IDs (comma-separated or repeated)")
	cmd.Flags().Bool("re-rate", false, "Re-rate tenders")
	cmd.Flags().Bool("re-constrain", false, "Re-constrain tenders")
	cmd.Flags().Bool("update-time-card-quantities", false, "Update time card quantities")
	cmd.Flags().Bool("skip-update-travel-minutes", false, "Skip updating travel minutes (deprecated)")
	cmd.Flags().Bool("skip-validate-customer-tender-hourly-rates", false, "Skip validation of customer tender hourly rates")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTenderReRatesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTenderReRatesCreateOptions(cmd)
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

	if len(opts.TenderIDs) == 0 {
		err := fmt.Errorf("--tender-ids is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if !opts.ReRate && !opts.ReConstrain {
		err := fmt.Errorf("either --re-rate or --re-constrain must be set")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"tender-ids": opts.TenderIDs,
	}
	if cmd.Flags().Changed("re-rate") {
		attributes["re-rate"] = opts.ReRate
	}
	if cmd.Flags().Changed("re-constrain") {
		attributes["re-constrain"] = opts.ReConstrain
	}
	if cmd.Flags().Changed("update-time-card-quantities") {
		attributes["update-time-card-quantities"] = opts.UpdateTimeCardQuantities
	}
	if cmd.Flags().Changed("skip-update-travel-minutes") {
		attributes["skip-update-travel-minutes"] = opts.SkipUpdateTravelMinutes
	}
	if cmd.Flags().Changed("skip-validate-customer-tender-hourly-rates") {
		attributes["skip-validate-customer-tender-hourly-rates"] = opts.SkipValidateCustomerTenderHourlyRates
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "tender-re-rates",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/sombreros/tender-re-rates", jsonBody)
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

	details := buildTenderReRateDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTenderReRateDetails(cmd, details)
}

func parseDoTenderReRatesCreateOptions(cmd *cobra.Command) (doTenderReRatesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	tenderIDs, _ := cmd.Flags().GetStringSlice("tender-ids")
	reRate, _ := cmd.Flags().GetBool("re-rate")
	reConstrain, _ := cmd.Flags().GetBool("re-constrain")
	updateTimeCardQuantities, _ := cmd.Flags().GetBool("update-time-card-quantities")
	skipUpdateTravelMinutes, _ := cmd.Flags().GetBool("skip-update-travel-minutes")
	skipValidateCustomerTenderHourlyRates, _ := cmd.Flags().GetBool("skip-validate-customer-tender-hourly-rates")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTenderReRatesCreateOptions{
		BaseURL:                               baseURL,
		Token:                                 token,
		JSON:                                  jsonOut,
		TenderIDs:                             tenderIDs,
		ReRate:                                reRate,
		ReConstrain:                           reConstrain,
		UpdateTimeCardQuantities:              updateTimeCardQuantities,
		SkipUpdateTravelMinutes:               skipUpdateTravelMinutes,
		SkipValidateCustomerTenderHourlyRates: skipValidateCustomerTenderHourlyRates,
	}, nil
}
