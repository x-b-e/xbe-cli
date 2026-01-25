package cli

import "github.com/spf13/cobra"

func newTagsShowCmd() *cobra.Command {
	return newGenericShowCmd("tags")
}

func init() {
	tagsCmd.AddCommand(newTagsShowCmd())
}
