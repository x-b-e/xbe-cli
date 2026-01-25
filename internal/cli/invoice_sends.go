package cli

import "github.com/spf13/cobra"

var invoiceSendsCmd = &cobra.Command{
	Use:     "invoice-sends",
	Aliases: []string{"invoice-send"},
	Short:   "View invoice sends",
	Long: `View invoice sends.

Sends record a status change from editing to sent and
may include a comment.

Commands:
  list    List invoice sends
  show    Show invoice send details`,
	Example: `  # List sends
  xbe view invoice-sends list

  # Show a send
  xbe view invoice-sends show 123`,
}

func init() {
	viewCmd.AddCommand(invoiceSendsCmd)
}
