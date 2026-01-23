package cli

import "github.com/spf13/cobra"

var doLineupScenarioLineupsCmd = &cobra.Command{
	Use:   "lineup-scenario-lineups",
	Short: "Manage lineup scenario lineups",
	Long:  "Commands for creating and deleting lineup scenario lineups.",
}

func init() {
	doCmd.AddCommand(doLineupScenarioLineupsCmd)
}
