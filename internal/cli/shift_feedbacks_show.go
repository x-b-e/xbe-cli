package cli

import "github.com/spf13/cobra"

func newShiftFeedbacksShowCmd() *cobra.Command {
	return newGenericShowCmd("shift-feedbacks")
}

func init() {
	shiftFeedbacksCmd.AddCommand(newShiftFeedbacksShowCmd())
}
