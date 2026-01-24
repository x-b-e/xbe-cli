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

type doBrokerRetainerPaymentForecastsCreateOptions struct {
	BaseURL  string
	Token    string
	JSON     bool
	BrokerID string
	Date     string
}

func newDoBrokerRetainerPaymentForecastsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a broker retainer payment forecast",
		Long: `Create a broker retainer payment forecast.

Forecasts estimate upcoming retainer payments for a broker based on active
retainers and retainer periods.

Required flags:
  --broker   Broker ID

Optional flags:
  --date     Forecast start date (YYYY-MM-DD)`,
		Example: `  # Forecast broker retainer payments starting today
  xbe do broker-retainer-payment-forecasts create --broker 123

  # Forecast broker retainer payments starting on a specific date
  xbe do broker-retainer-payment-forecasts create --broker 123 --date 2025-01-01

  # Output as JSON
  xbe do broker-retainer-payment-forecasts create --broker 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoBrokerRetainerPaymentForecastsCreate,
	}
	initDoBrokerRetainerPaymentForecastsCreateFlags(cmd)
	return cmd
}

func init() {
	doBrokerRetainerPaymentForecastsCmd.AddCommand(newDoBrokerRetainerPaymentForecastsCreateCmd())
}

func initDoBrokerRetainerPaymentForecastsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("date", "", "Forecast start date (YYYY-MM-DD)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoBrokerRetainerPaymentForecastsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoBrokerRetainerPaymentForecastsCreateOptions(cmd)
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

	if opts.BrokerID == "" {
		err := fmt.Errorf("--broker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.Date != "" {
		attributes["date"] = opts.Date
	}

	relationships := map[string]any{
		"broker": map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.BrokerID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "broker-retainer-payment-forecasts",
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

	body, _, err := client.Post(cmd.Context(), "/v1/broker-retainer-payment-forecasts", jsonBody)
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

	row := buildBrokerRetainerPaymentForecastRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	return renderBrokerRetainerPaymentForecastDetails(cmd, row)
}

func parseDoBrokerRetainerPaymentForecastsCreateOptions(cmd *cobra.Command) (doBrokerRetainerPaymentForecastsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	brokerID, _ := cmd.Flags().GetString("broker")
	date, _ := cmd.Flags().GetString("date")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doBrokerRetainerPaymentForecastsCreateOptions{
		BaseURL:  baseURL,
		Token:    token,
		JSON:     jsonOut,
		BrokerID: brokerID,
		Date:     date,
	}, nil
}
