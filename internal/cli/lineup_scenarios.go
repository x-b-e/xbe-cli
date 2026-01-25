package cli

import "github.com/spf13/cobra"

var lineupScenariosCmd = &cobra.Command{
	Use:     "lineup-scenarios",
	Aliases: []string{"lineup-scenario"},
	Short:   "Browse lineup scenarios",
	Long: `Browse lineup scenarios.

Lineup scenarios capture broker/date/window constraints and generated
options for lineup planning.`,
	Example: `  # List lineup scenarios
  xbe view lineup-scenarios list

  # Filter by broker/date/window
  xbe view lineup-scenarios list --broker 123 --date 2026-01-23 --window day

  # Show a lineup scenario
  xbe view lineup-scenarios show 456`,
}

func init() {
	viewCmd.AddCommand(lineupScenariosCmd)
}
