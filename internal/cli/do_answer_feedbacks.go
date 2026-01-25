package cli

import "github.com/spf13/cobra"

var doAnswerFeedbacksCmd = &cobra.Command{
	Use:     "answer-feedbacks",
	Aliases: []string{"answer-feedback"},
	Short:   "Manage answer feedbacks",
	Long:    "Commands for creating, updating, and deleting answer feedbacks.",
}

func init() {
	doCmd.AddCommand(doAnswerFeedbacksCmd)
}
