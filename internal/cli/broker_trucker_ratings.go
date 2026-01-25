package cli

import "github.com/spf13/cobra"

var brokerTruckerRatingsCmd = &cobra.Command{
	Use:   "broker-trucker-ratings",
	Short: "View broker trucker ratings",
	Long: `View broker trucker ratings on the XBE platform.

Broker trucker ratings capture a broker's rating for a trucker.

Commands:
  list    List broker trucker ratings
  show    Show broker trucker rating details`,
	Example: `  # List broker trucker ratings
  xbe view broker-trucker-ratings list

  # Show a broker trucker rating
  xbe view broker-trucker-ratings show 123`,
}

func init() {
	viewCmd.AddCommand(brokerTruckerRatingsCmd)
}
