package cli

import "github.com/spf13/cobra"

var doBrokerCommitmentsCmd = &cobra.Command{
	Use:     "broker-commitments",
	Aliases: []string{"broker-commitment"},
	Short:   "Manage broker commitments",
	Long: `Manage broker commitments on the XBE platform.

Commands:
  create    Create a broker commitment
  update    Update a broker commitment
  delete    Delete a broker commitment`,
	Example: `  # Create a broker commitment
  xbe do broker-commitments create --status active --broker 123 --trucker 456

  # Update a broker commitment
  xbe do broker-commitments update 789 --status inactive

  # Delete a broker commitment (requires --confirm)
  xbe do broker-commitments delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doBrokerCommitmentsCmd)
}
