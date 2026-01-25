package cli

import "github.com/spf13/cobra"

var doTimeCardTimeChangesCmd = &cobra.Command{
	Use:   "time-card-time-changes",
	Short: "Manage time card time changes",
	Long: `Create, update, and delete time card time changes.

Time card time changes track requested updates to time card times.

Note: Updates and deletes are only allowed for unprocessed changes.

Commands:
  create  Create a time card time change
  update  Update a time card time change
  delete  Delete a time card time change`,
	Example: `  # Create a time card time change
  xbe do time-card-time-changes create --time-card 123 --time-changes-attributes '{"down_minutes":15}'

  # Update a time card time change
  xbe do time-card-time-changes update 123 --comment "Adjusted down time"

  # Delete a time card time change
  xbe do time-card-time-changes delete 123 --confirm`,
}

func init() {
	doCmd.AddCommand(doTimeCardTimeChangesCmd)
}
