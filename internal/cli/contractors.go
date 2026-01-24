package cli

import "github.com/spf13/cobra"

var contractorsCmd = &cobra.Command{
	Use:   "contractors",
	Short: "View contractors",
	Long: `View contractors on the XBE platform.

Contractors are broker-associated organizations used in job production plans
and incident tracking.

Commands:
  list    List contractors
  show    Show contractor details`,
	Example: `  # List contractors
  xbe view contractors list

  # Filter by broker
  xbe view contractors list --broker 123

  # Search by name
  xbe view contractors list --name "Acme"

  # Show a contractor
  xbe view contractors show 456`,
}

func init() {
	viewCmd.AddCommand(contractorsCmd)
}
