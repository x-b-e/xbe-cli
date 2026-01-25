package cli

import "github.com/spf13/cobra"

var digitalFleetTrucksCmd = &cobra.Command{
	Use:     "digital-fleet-trucks",
	Aliases: []string{"digital-fleet-truck"},
	Short:   "Browse digital fleet trucks",
	Long: `Browse Digital Fleet truck integrations.

Digital fleet trucks represent vehicles imported from Digital Fleet
integrations and show assignment status to tractors and trailers.

Commands:
  list    List digital fleet trucks with filtering
  show    Show digital fleet truck details`,
	Example: `  # List digital fleet trucks
  xbe view digital-fleet-trucks list

  # Show digital fleet truck details
  xbe view digital-fleet-trucks show 123`,
}

func init() {
	viewCmd.AddCommand(digitalFleetTrucksCmd)
}
