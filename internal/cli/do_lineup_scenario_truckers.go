package cli

import "github.com/spf13/cobra"

var doLineupScenarioTruckersCmd = &cobra.Command{
	Use:   "lineup-scenario-truckers",
	Short: "Manage lineup scenario truckers",
	Long:  "Commands for creating, updating, and deleting lineup scenario truckers.",
}

func init() {
	doCmd.AddCommand(doLineupScenarioTruckersCmd)
}
