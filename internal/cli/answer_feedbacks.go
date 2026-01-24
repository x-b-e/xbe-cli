package cli

import "github.com/spf13/cobra"

var answerFeedbacksCmd = &cobra.Command{
	Use:     "answer-feedbacks",
	Aliases: []string{"answer-feedback"},
	Short:   "View answer feedbacks",
	Long:    "Commands for viewing answer feedbacks tied to answers.",
}

func init() {
	viewCmd.AddCommand(answerFeedbacksCmd)
}
