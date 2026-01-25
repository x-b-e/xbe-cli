package cli

import "github.com/spf13/cobra"

var doShiftFeedbacksCmd = &cobra.Command{
	Use:     "shift-feedbacks",
	Aliases: []string{"shift-feedback"},
	Short:   "Manage shift feedbacks",
	Long:    "Commands for creating, updating, and deleting shift feedbacks.",
}

func init() {
	doCmd.AddCommand(doShiftFeedbacksCmd)
}
