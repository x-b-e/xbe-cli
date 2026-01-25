package cli

import "github.com/spf13/cobra"

var doProjectRevenueItemsCmd = &cobra.Command{
	Use:   "project-revenue-items",
	Short: "Manage project revenue items",
	Long: `Create, update, and delete project revenue items.

Project revenue items define billable line items for a project and map to
revenue classifications and units of measure.

Commands:
  create  Create a project revenue item
  update  Update a project revenue item
  delete  Delete a project revenue item`,
	Example: `  # Create a project revenue item
  xbe do project-revenue-items create \
    --project 123 \
    --revenue-classification 456 \
    --unit-of-measure 789 \
    --description "Base material"

  # Update description and quantity estimate
  xbe do project-revenue-items update 321 --description "Updated description" --developer-quantity-estimate 1200

  # Delete a project revenue item
  xbe do project-revenue-items delete 321 --confirm`,
}

func init() {
	doCmd.AddCommand(doProjectRevenueItemsCmd)
}
