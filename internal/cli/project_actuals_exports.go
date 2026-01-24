package cli

import "github.com/spf13/cobra"

var projectActualsExportsCmd = &cobra.Command{
	Use:     "project-actuals-exports",
	Aliases: []string{"project-actuals-export"},
	Short:   "Browse project actuals exports",
	Long: `Browse project actuals exports.

Project actuals exports generate formatted files for selected job production
plans using an organization formatter.

Commands:
  list    List project actuals exports
  show    Show project actuals export details`,
	Example: `  # List exports
  xbe view project-actuals-exports list

  # Show export details
  xbe view project-actuals-exports show 123`,
}

func init() {
	viewCmd.AddCommand(projectActualsExportsCmd)
}
