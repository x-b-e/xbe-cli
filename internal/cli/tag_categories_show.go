package cli

import "github.com/spf13/cobra"

func newTagCategoriesShowCmd() *cobra.Command {
	return newGenericShowCmd("tag-categories")
}

func init() {
	tagCategoriesCmd.AddCommand(newTagCategoriesShowCmd())
}
