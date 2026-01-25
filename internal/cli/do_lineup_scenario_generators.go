package cli

import "github.com/spf13/cobra"

var doLineupScenarioGeneratorsCmd = &cobra.Command{
	Use:   "lineup-scenario-generators",
	Short: "Generate lineup scenarios",
	Long: `Generate lineup scenarios on the XBE platform.

Commands:
  create    Create a lineup scenario generator
  delete    Delete a lineup scenario generator`,
	Example: `  # Create a generator
  xbe do lineup-scenario-generators create --broker 123 --date 2026-01-23 --window day`,
}

func init() {
	doCmd.AddCommand(doLineupScenarioGeneratorsCmd)
}
