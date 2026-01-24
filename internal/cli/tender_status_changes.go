package cli

import "github.com/spf13/cobra"

var tenderStatusChangesCmd = &cobra.Command{
	Use:     "tender-status-changes",
	Aliases: []string{"tender-status-change"},
	Short:   "View tender status changes",
	Long: `View tender status changes on the XBE platform.

Tender status changes track workflow transitions for tenders.

Commands:
  list    List tender status changes
  show    Show tender status change details`,
	Example: `  # List tender status changes
  xbe view tender-status-changes list

  # Filter by tender
  xbe view tender-status-changes list --tender 123

  # Show a status change
  xbe view tender-status-changes show 456`,
}

func init() {
	viewCmd.AddCommand(tenderStatusChangesCmd)
}
