package cli

import "github.com/spf13/cobra"

var doProjectBidLocationMaterialTypesCmd = &cobra.Command{
	Use:   "project-bid-location-material-types",
	Short: "Manage project bid location material types",
	Long: `Create, update, and delete project bid location material types.

Project bid location material types define planned quantities and notes for a
material type at a specific project bid location.

Commands:
  create    Create a project bid location material type
  update    Update a project bid location material type
  delete    Delete a project bid location material type`,
	Example: `  # Create a project bid location material type
  xbe do project-bid-location-material-types create \
    --project-bid-location 123 \
    --material-type 456 \
    --quantity 10.5

  # Update quantity and notes
  xbe do project-bid-location-material-types update 789 --quantity 12 --notes "Updated"

  # Delete a project bid location material type
  xbe do project-bid-location-material-types delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doProjectBidLocationMaterialTypesCmd)
}
