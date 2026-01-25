package cli

import "github.com/spf13/cobra"

var postChildrenCmd = &cobra.Command{
	Use:     "post-children",
	Aliases: []string{"post-child"},
	Short:   "View post child links",
	Long:    "Commands for viewing post child links.",
}

func init() {
	viewCmd.AddCommand(postChildrenCmd)
}
