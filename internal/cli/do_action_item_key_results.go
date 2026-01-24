package cli

import "github.com/spf13/cobra"

var doActionItemKeyResultsCmd = &cobra.Command{
	Use:     "action-item-key-results",
	Aliases: []string{"action-item-key-result"},
	Short:   "Manage action item key result links",
	Long: `Create and delete action item key result links.

Commands:
  create    Create an action item key result link
  delete    Delete an action item key result link`,
}

func init() {
	doCmd.AddCommand(doActionItemKeyResultsCmd)
}
