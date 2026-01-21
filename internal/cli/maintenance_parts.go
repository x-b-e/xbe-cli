package cli

import "github.com/spf13/cobra"

var maintenancePartsCmd = &cobra.Command{
	Use:   "parts",
	Short: "View maintenance requirement parts catalog",
	Long: `View maintenance requirement parts catalog.

Parts are components that can be used in maintenance requirements,
including their specifications, costs, and manufacturers.

Commands:
  list    List parts in the catalog
  show    View detailed part information`,
	Example: `  # List all parts
  xbe view maintenance parts list

  # View part details
  xbe view maintenance parts show 123`,
}

func init() {
	maintenanceCmd.AddCommand(maintenancePartsCmd)
}
