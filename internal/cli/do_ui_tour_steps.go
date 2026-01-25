package cli

import "github.com/spf13/cobra"

var doUiTourStepsCmd = &cobra.Command{
	Use:     "ui-tour-steps",
	Aliases: []string{"ui-tour-step"},
	Short:   "Manage UI tour steps",
	Long:    "Commands for creating, updating, and deleting UI tour steps.",
}

func init() {
	doCmd.AddCommand(doUiTourStepsCmd)
}
