package cli

import "github.com/spf13/cobra"

var projectPhaseRevenueItemActualExportsCmd = &cobra.Command{
	Use:     "project-phase-revenue-item-actual-exports",
	Aliases: []string{"project-phase-revenue-item-actual-export"},
	Short:   "Browse project phase revenue item actual exports",
	Long: `Browse project phase revenue item actual exports.

Project phase revenue item actual exports generate formatted files for selected
project phase revenue items using an organization formatter.

Commands:
  list    List project phase revenue item actual exports
  show    Show project phase revenue item actual export details`,
	Example: `  # List exports
  xbe view project-phase-revenue-item-actual-exports list

  # Show export details
  xbe view project-phase-revenue-item-actual-exports show 123`,
}

func init() {
	viewCmd.AddCommand(projectPhaseRevenueItemActualExportsCmd)
}
