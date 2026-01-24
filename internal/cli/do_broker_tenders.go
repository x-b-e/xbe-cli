package cli

import "github.com/spf13/cobra"

var doBrokerTendersCmd = &cobra.Command{
	Use:     "broker-tenders",
	Aliases: []string{"broker-tender"},
	Short:   "Manage broker tenders",
	Long: `Manage broker tenders on the XBE platform.

Commands:
  create    Create a broker tender
  update    Update a broker tender
  delete    Delete a broker tender`,
	Example: `  # Create a broker tender
  xbe do broker-tenders create --job 123 --broker 456 --trucker 789

  # Update a broker tender
  xbe do broker-tenders update 123 --note "Updated note"

  # Delete a broker tender (requires --confirm)
  xbe do broker-tenders delete 123 --confirm`,
}

func init() {
	doCmd.AddCommand(doBrokerTendersCmd)
}
