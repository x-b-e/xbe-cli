package cli

import "github.com/spf13/cobra"

var shiftFeedbackReasonsCmd = &cobra.Command{
	Use:   "shift-feedback-reasons",
	Short: "View shift feedback reasons",
	Long: `View shift feedback reasons on the XBE platform.

Shift feedback reasons define the types of feedback that can be given for shifts,
including ratings and corrective actions.

Commands:
  list    List shift feedback reasons`,
	Example: `  # List shift feedback reasons
  xbe view shift-feedback-reasons list

  # Filter by kind
  xbe view shift-feedback-reasons list --kind positive

  # Output as JSON
  xbe view shift-feedback-reasons list --json`,
}

func init() {
	viewCmd.AddCommand(shiftFeedbackReasonsCmd)
}
