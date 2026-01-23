package cli

import "github.com/spf13/cobra"

var brokerCommitmentsCmd = &cobra.Command{
	Use:     "broker-commitments",
	Aliases: []string{"broker-commitment"},
	Short:   "View broker commitments",
	Long: `View broker commitments on the XBE platform.

Broker commitments represent agreements between brokers and truckers for
capacity or service commitments.

Commands:
  list    List broker commitments
  show    Show broker commitment details`,
	Example: `  # List broker commitments
  xbe view broker-commitments list

  # Show a broker commitment
  xbe view broker-commitments show 123`,
}

func init() {
	viewCmd.AddCommand(brokerCommitmentsCmd)
}
