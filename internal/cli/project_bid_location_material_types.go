package cli

import "github.com/spf13/cobra"

var projectBidLocationMaterialTypesCmd = &cobra.Command{
	Use:   "project-bid-location-material-types",
	Short: "View project bid location material types",
	Long: `View project bid location material types.

Project bid location material types define quantities and notes for material
requirements at a specific project bid location.

Commands:
  list    List project bid location material types with filtering
  show    Show project bid location material type details`,
	Example: `  # List project bid location material types
  xbe view project-bid-location-material-types list

  # Filter by project bid location
  xbe view project-bid-location-material-types list --project-bid-location 123

  # Show a specific record
  xbe view project-bid-location-material-types show 456`,
}

func init() {
	viewCmd.AddCommand(projectBidLocationMaterialTypesCmd)
}
