package cli

import "github.com/spf13/cobra"

var keyResultStatusChangesCmd = &cobra.Command{
	Use:     "key-result-status-changes",
	Aliases: []string{"key-result-status-change"},
	Short:   "View key result status changes",
	Long: `View key result status changes on the XBE platform.

Key result status changes track workflow transitions for key results.

Commands:
  list    List key result status changes
  show    Show key result status change details`,
	Example: `  # List key result status changes
  xbe view key-result-status-changes list

  # Filter by key result
  xbe view key-result-status-changes list --key-result 123

  # Show a status change
  xbe view key-result-status-changes show 456`,
}

func init() {
	viewCmd.AddCommand(keyResultStatusChangesCmd)
}
