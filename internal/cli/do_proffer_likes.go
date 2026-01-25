package cli

import "github.com/spf13/cobra"

var doProfferLikesCmd = &cobra.Command{
	Use:     "proffer-likes",
	Aliases: []string{"proffer-like"},
	Short:   "Manage proffer likes",
	Long:    "Commands for creating and deleting proffer likes.",
}

func init() {
	doCmd.AddCommand(doProfferLikesCmd)
}
