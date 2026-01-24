package cli

import "github.com/spf13/cobra"

var doProjectActualsExportsCmd = &cobra.Command{
	Use:   "project-actuals-exports",
	Short: "Manage project actuals exports",
	Long: `Manage project actuals exports.

Exports generate formatted files for selected job production plans using an
organization formatter.

Commands:
  create    Create a project actuals export`,
	Example: `  # Create an export
  xbe do project-actuals-exports create --organization-formatter 123 --job-production-plan-ids 456`,
}

func init() {
	doCmd.AddCommand(doProjectActualsExportsCmd)
}
