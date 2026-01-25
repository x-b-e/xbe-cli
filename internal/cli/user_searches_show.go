package cli

import "github.com/spf13/cobra"

func newUserSearchesShowCmd() *cobra.Command {
	return newGenericShowCmd("user-searches")
}

func init() {
	userSearchesCmd.AddCommand(newUserSearchesShowCmd())
}
