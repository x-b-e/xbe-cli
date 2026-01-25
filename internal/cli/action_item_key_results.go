package cli

import "github.com/spf13/cobra"

var actionItemKeyResultsCmd = &cobra.Command{
	Use:     "action-item-key-results",
	Aliases: []string{"action-item-key-result"},
	Short:   "View action item key result links",
	Long: `View action item key result links.

Action item key results link action items to key results (OKRs).

Commands:
  list    List action item key result links
  show    Show action item key result details`,
	Example: `  # List action item key results
  xbe view action-item-key-results list

  # Show a specific action item key result link
  xbe view action-item-key-results show 123

  # Output JSON
  xbe view action-item-key-results list --json`,
}

func init() {
	viewCmd.AddCommand(actionItemKeyResultsCmd)
}
