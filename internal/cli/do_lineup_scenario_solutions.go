package cli

import "github.com/spf13/cobra"

var doLineupScenarioSolutionsCmd = &cobra.Command{
	Use:     "lineup-scenario-solutions",
	Aliases: []string{"lineup-scenario-solution"},
	Short:   "Solve lineup scenarios",
	Long:    "Commands for generating lineup scenario solutions.",
}

func init() {
	doCmd.AddCommand(doLineupScenarioSolutionsCmd)
}
