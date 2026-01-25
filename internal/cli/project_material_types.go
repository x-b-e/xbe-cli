package cli

import "github.com/spf13/cobra"

var projectMaterialTypesCmd = &cobra.Command{
	Use:   "project-material-types",
	Short: "View project material types",
	Long: `View project material types on the XBE platform.

Project material types associate material types with projects, including
optional quantities, units of measure, and pickup/delivery windows.

Commands:
  list    List project material types
  show    Show project material type details`,
	Example: `  # List project material types
  xbe view project-material-types list

  # Show a project material type
  xbe view project-material-types show 123`,
}

func init() {
	viewCmd.AddCommand(projectMaterialTypesCmd)
}
