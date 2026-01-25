package cli

import "github.com/spf13/cobra"

var doPostViewsCmd = &cobra.Command{
	Use:     "post-views",
	Aliases: []string{"post-view"},
	Short:   "Manage post views",
	Long:    "Commands for recording post views.",
}

func init() {
	doCmd.AddCommand(doPostViewsCmd)
}
