package cli

import "github.com/spf13/cobra"

var doProcessNonProcessedTimeCardTimeChangesCmd = &cobra.Command{
	Use:     "process-non-processed-time-card-time-changes",
	Aliases: []string{"process-non-processed-time-card-time-change"},
	Short:   "Process non-processed time card time changes",
	Long: `Process non-processed time card time changes.

This operation processes the specified time card time changes, updates related
invoices, and optionally deletes unprocessed changes after processing.

Commands:
  create    Process time card time changes`,
	Example: `  # Process time card time changes
  xbe do process-non-processed-time-card-time-changes create --time-card-time-change-ids 123,456

  # Keep unprocessed changes
  xbe do process-non-processed-time-card-time-changes create --time-card-time-change-ids 123 --delete-unprocessed false`,
}

func init() {
	doCmd.AddCommand(doProcessNonProcessedTimeCardTimeChangesCmd)
}
