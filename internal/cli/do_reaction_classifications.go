package cli

import "github.com/spf13/cobra"

var doReactionClassificationsCmd = &cobra.Command{
	Use:     "reaction-classifications",
	Aliases: []string{"reaction-classification"},
	Short:   "Manage reaction classifications",
	Long:    `Create reaction classifications.`,
}

func init() {
	doCmd.AddCommand(doReactionClassificationsCmd)
}
