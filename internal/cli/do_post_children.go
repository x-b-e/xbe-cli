package cli

import "github.com/spf13/cobra"

var doPostChildrenCmd = &cobra.Command{
	Use:   "post-children",
	Short: "Manage post child links",
	Long:  "Commands for creating and deleting post child links.",
}

func init() {
	doCmd.AddCommand(doPostChildrenCmd)
}
