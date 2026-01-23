package cli

import "github.com/spf13/cobra"

var shiftFeedbacksCmd = &cobra.Command{
	Use:     "shift-feedbacks",
	Aliases: []string{"shift-feedback"},
	Short:   "View shift feedbacks",
	Long:    "Commands for viewing shift feedbacks (trucker/driver performance feedback).",
}

func init() {
	viewCmd.AddCommand(shiftFeedbacksCmd)
}
