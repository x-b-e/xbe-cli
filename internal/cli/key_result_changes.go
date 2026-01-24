package cli

import "github.com/spf13/cobra"

var keyResultChangesCmd = &cobra.Command{
	Use:     "key-result-changes",
	Aliases: []string{"key-result-change"},
	Short:   "Browse key result changes",
	Long: `Browse key result changes.

Key result changes track updates to key result schedules, including start and end dates.

Commands:
  list    List key result changes with filtering and pagination
  show    Show key result change details`,
	Example: `  # List key result changes
  xbe view key-result-changes list

  # Filter by key result
  xbe view key-result-changes list --key-result 123

  # Filter by objective
  xbe view key-result-changes list --objective 456

  # Show a key result change
  xbe view key-result-changes show 789`,
}

func init() {
	viewCmd.AddCommand(keyResultChangesCmd)
}
