package cli

import "github.com/spf13/cobra"

var doCostIndexesCmd = &cobra.Command{
	Use:   "cost-indexes",
	Short: "Manage cost indexes",
	Long: `Create, update, and delete cost indexes.

Cost indexes define pricing indexes that can be used for rate adjustments.

Commands:
  create  Create a new cost index
  update  Update an existing cost index
  delete  Delete a cost index`,
	Example: `  # Create a cost index
  xbe do cost-indexes create --name "Fuel Index" --broker 123

  # Create a global cost index (no broker)
  xbe do cost-indexes create --name "National CPI"

  # Update a cost index
  xbe do cost-indexes update 123 --name "Updated Name"

  # Delete a cost index
  xbe do cost-indexes delete 123 --confirm`,
}

func init() {
	doCmd.AddCommand(doCostIndexesCmd)
}
