package cli

import "github.com/spf13/cobra"

var lineupScenarioSolutionsCmd = &cobra.Command{
	Use:     "lineup-scenario-solutions",
	Aliases: []string{"lineup-scenario-solution"},
	Short:   "Browse lineup scenario solutions",
	Long: `Browse lineup scenario solutions.

Lineup scenario solutions capture the solver output for a lineup scenario,
including assignments, cost, and status.

Commands:
  list    List lineup scenario solutions
  show    Show lineup scenario solution details`,
	Example: `  # List lineup scenario solutions
  xbe view lineup-scenario-solutions list

  # Filter by lineup scenario
  xbe view lineup-scenario-solutions list --lineup-scenario 123

  # Show a lineup scenario solution
  xbe view lineup-scenario-solutions show 456`,
}

func init() {
	viewCmd.AddCommand(lineupScenarioSolutionsCmd)
}
