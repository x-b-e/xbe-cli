package cli

import "github.com/spf13/cobra"

var doPostsCmd = &cobra.Command{
	Use:     "posts",
	Aliases: []string{"post"},
	Short:   "Manage posts",
	Long:    `Create, update, and delete posts.`,
}

func init() {
	doCmd.AddCommand(doPostsCmd)
}
