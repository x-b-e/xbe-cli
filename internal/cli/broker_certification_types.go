package cli

import "github.com/spf13/cobra"

var brokerCertificationTypesCmd = &cobra.Command{
	Use:   "broker-certification-types",
	Short: "View broker certification types",
	Long: `View broker certification types on the XBE platform.

Broker certification types link brokers to certification types they can
track or require.

Commands:
  list    List broker certification types
  show    Show broker certification type details`,
	Example: `  # List broker certification types
  xbe view broker-certification-types list

  # Show a broker certification type
  xbe view broker-certification-types show 123`,
}

func init() {
	viewCmd.AddCommand(brokerCertificationTypesCmd)
}
