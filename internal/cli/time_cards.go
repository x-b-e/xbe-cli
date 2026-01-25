package cli

import "github.com/spf13/cobra"

var timeCardsCmd = &cobra.Command{
	Use:   "time-cards",
	Short: "Browse time cards",
	Long: `Browse time cards on the XBE platform.

Time cards track shift work details, hours, approvals, and invoicing status.

Commands:
  list    List time cards with filtering
  show    Show time card details`,
	Example: `  # List time cards
  xbe view time-cards list

  # Filter by status
  xbe view time-cards list --status approved

  # Show a time card
  xbe view time-cards show 123`,
}

func init() {
	viewCmd.AddCommand(timeCardsCmd)
}
