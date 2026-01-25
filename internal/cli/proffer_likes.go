package cli

import "github.com/spf13/cobra"

var profferLikesCmd = &cobra.Command{
	Use:     "proffer-likes",
	Aliases: []string{"proffer-like"},
	Short:   "View proffer likes",
	Long:    "Commands for viewing likes on proffers.",
}

func init() {
	viewCmd.AddCommand(profferLikesCmd)
}
