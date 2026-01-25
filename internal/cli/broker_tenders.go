package cli

import "github.com/spf13/cobra"

var brokerTendersCmd = &cobra.Command{
	Use:     "broker-tenders",
	Aliases: []string{"broker-tender"},
	Short:   "Browse broker tenders",
	Long: `Browse broker tenders on the XBE platform.

Broker tenders are offers from brokers to truckers for job work.

Commands:
  list    List broker tenders with filtering
  show    Show broker tender details`,
	Example: `  # List broker tenders
  xbe view broker-tenders list

  # Show a broker tender
  xbe view broker-tenders show 123`,
}

func init() {
	viewCmd.AddCommand(brokerTendersCmd)
}
