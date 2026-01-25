package cli

import "github.com/spf13/cobra"

func newShiftFeedbackReasonsShowCmd() *cobra.Command {
	return newGenericShowCmd("shift-feedback-reasons")
}

func init() {
	shiftFeedbackReasonsCmd.AddCommand(newShiftFeedbackReasonsShowCmd())
}
