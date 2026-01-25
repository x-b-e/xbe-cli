package cli

import "github.com/spf13/cobra"

var doBrokerTruckerRatingsCmd = &cobra.Command{
	Use:   "broker-trucker-ratings",
	Short: "Manage broker trucker ratings",
	Long: `Manage broker trucker ratings on the XBE platform.

Broker trucker ratings capture a broker's rating for a trucker.

Commands:
  create    Create a broker trucker rating
  update    Update a broker trucker rating
  delete    Delete a broker trucker rating`,
	Example: `  # Create a broker trucker rating
  xbe do broker-trucker-ratings create --broker 123 --trucker 456 --rating 5

  # Update a broker trucker rating
  xbe do broker-trucker-ratings update 789 --rating 4

  # Delete a broker trucker rating (requires --confirm)
  xbe do broker-trucker-ratings delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doBrokerTruckerRatingsCmd)
}
