package cli

import "github.com/spf13/cobra"

var doProjectPhaseRevenueItemActualExportsCmd = &cobra.Command{
	Use:   "project-phase-revenue-item-actual-exports",
	Short: "Manage project phase revenue item actual exports",
	Long: `Manage project phase revenue item actual exports.

Exports generate formatted files for selected project phase revenue items using
an organization formatter.

Commands:
  create    Create a project phase revenue item actual export`,
	Example: `  # Create an export
  xbe do project-phase-revenue-item-actual-exports create \
    --organization-formatter 123 \
    --project-phase-revenue-item-ids 456`,
}

func init() {
	doCmd.AddCommand(doProjectPhaseRevenueItemActualExportsCmd)
}
