package cli

import "github.com/spf13/cobra"

var doBrokerRetainerPaymentForecastsCmd = &cobra.Command{
	Use:     "broker-retainer-payment-forecasts",
	Aliases: []string{"broker-retainer-payment-forecast"},
	Short:   "Generate broker retainer payment forecasts",
	Long:    "Commands for generating broker retainer payment forecasts.",
}

func init() {
	doCmd.AddCommand(doBrokerRetainerPaymentForecastsCmd)
}
