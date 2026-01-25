package cli

import "github.com/spf13/cobra"

var lineupScenarioGeneratorsCmd = &cobra.Command{
	Use:     "lineup-scenario-generators",
	Aliases: []string{"lineup-scenario-generator"},
	Short:   "Browse lineup scenario generators",
	Long: `Browse lineup scenario generators.

Lineup scenario generators create lineup scenarios for a broker/date/window,
including optional constraints and assignment limits.`,
	Example: `  # List lineup scenario generators
  xbe view lineup-scenario-generators list

  # Show a generator
  xbe view lineup-scenario-generators show 123`,
}

func init() {
	viewCmd.AddCommand(lineupScenarioGeneratorsCmd)
}
