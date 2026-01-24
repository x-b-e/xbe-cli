package cli

import "github.com/spf13/cobra"

var keyResultsCmd = &cobra.Command{
	Use:     "key-results",
	Aliases: []string{"key-result"},
	Short:   "View key results",
	Long: `View key results on the XBE platform.

Key results track measurable outcomes tied to objectives.

Commands:
  list    List key results
  show    Show key result details`,
	Example: `  # List key results
  xbe view key-results list

  # Filter by objective
  xbe view key-results list --objective 123

  # Show a key result
  xbe view key-results show 456`,
}

func init() {
	viewCmd.AddCommand(keyResultsCmd)
}
