package cli

import "github.com/spf13/cobra"

var doProjectMaterialTypesCmd = &cobra.Command{
	Use:   "project-material-types",
	Short: "Manage project material types",
	Long: `Create, update, and delete project material types.

Project material types define material requirements for a project and can be
scoped to material sites, job sites, or transport-only pickup/delivery locations.

Commands:
  create    Create a project material type
  update    Update a project material type
  delete    Delete a project material type`,
	Example: `  # Create a project material type
  xbe do project-material-types create \
    --project 123 \
    --material-type 456 \
    --quantity 10.5

  # Update quantity and display name
  xbe do project-material-types update 789 --quantity 12 --explicit-display-name "Washed Rock"

  # Delete a project material type
  xbe do project-material-types delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doProjectMaterialTypesCmd)
}
