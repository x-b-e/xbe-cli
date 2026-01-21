package cli

import "github.com/spf13/cobra"

var doShiftFeedbackReasonsCmd = &cobra.Command{
	Use:   "shift-feedback-reasons",
	Short: "Manage shift feedback reasons",
	Long: `Create, update, and delete shift feedback reasons.

Shift feedback reasons define the types of feedback that can be given for shifts.

Note: Only admin users can create, update, or delete shift feedback reasons.

Commands:
  create  Create a new shift feedback reason
  update  Update an existing shift feedback reason
  delete  Delete a shift feedback reason`,
	Example: `  # Create a shift feedback reason
  xbe do shift-feedback-reasons create --name "Late Arrival" --kind negative --slug "late-arrival"

  # Update a shift feedback reason
  xbe do shift-feedback-reasons update 123 --name "Updated Name"

  # Delete a shift feedback reason
  xbe do shift-feedback-reasons delete 123 --confirm`,
}

func init() {
	doCmd.AddCommand(doShiftFeedbackReasonsCmd)
}
