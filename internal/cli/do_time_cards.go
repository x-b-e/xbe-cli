package cli

import "github.com/spf13/cobra"

var doTimeCardsCmd = &cobra.Command{
	Use:   "time-cards",
	Short: "Manage time cards",
	Long: `Manage time cards on the XBE platform.

Commands:
  create    Create a new time card
  update    Update an existing time card
  delete    Delete a time card`,
	Example: `  # Create a time card
  xbe do time-cards create --broker-tender 123 --tender-job-schedule-shift 456

  # Update a time card
  xbe do time-cards update 789 --ticket-number "TC-001"

  # Delete a time card
  xbe do time-cards delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doTimeCardsCmd)
}
