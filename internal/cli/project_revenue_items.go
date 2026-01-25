package cli

import "github.com/spf13/cobra"

var projectRevenueItemsCmd = &cobra.Command{
	Use:   "project-revenue-items",
	Short: "View project revenue items",
	Long: `View project revenue items on the XBE platform.

Project revenue items define billable line items for a project and link to
revenue classifications, units of measure, and revenue estimates.

Commands:
  list    List project revenue items
  show    Show project revenue item details`,
	Example: `  # List project revenue items
  xbe view project-revenue-items list

  # Filter by project
  xbe view project-revenue-items list --project 123

  # Show a project revenue item
  xbe view project-revenue-items show 456`,
}

func init() {
	viewCmd.AddCommand(projectRevenueItemsCmd)
}
