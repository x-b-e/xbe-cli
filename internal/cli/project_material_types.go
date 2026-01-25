package cli

import "github.com/spf13/cobra"

var projectMaterialTypesCmd = &cobra.Command{
	Use:   "project-material-types",
	Short: "View project material types",
	Long: `View project material types.

Project material types define material requirements for a project and can be
scoped to material sites, job sites, or transport-only pickup/delivery locations.

Commands:
  list    List project material types with filtering
  show    Show project material type details`,
	Example: `  # List project material types
  xbe view project-material-types list

  # Filter by project
  xbe view project-material-types list --project 123

  # Show a specific record
  xbe view project-material-types show 456`,
}

func init() {
	viewCmd.AddCommand(projectMaterialTypesCmd)
}
