package cli

import "github.com/spf13/cobra"

var doCostIndexEntriesCmd = &cobra.Command{
	Use:   "cost-index-entries",
	Short: "Manage cost index entries",
	Long: `Create, update, and delete cost index entries.

Cost index entries are time-series values for cost indexes, used in rate adjustments.

Commands:
  create  Create a new cost index entry
  update  Update an existing cost index entry
  delete  Delete a cost index entry`,
	Example: `  # Create a cost index entry
  xbe do cost-index-entries create --cost-index 123 --start-on "2024-01-01" --value 1.05

  # Update a cost index entry
  xbe do cost-index-entries update 456 --value 1.10

  # Delete a cost index entry
  xbe do cost-index-entries delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doCostIndexEntriesCmd)
}
