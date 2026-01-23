package cli

import "github.com/spf13/cobra"

var invoiceApprovalsCmd = &cobra.Command{
	Use:     "invoice-approvals",
	Aliases: []string{"invoice-approval"},
	Short:   "View invoice approvals",
	Long: `View invoice approvals.

Approvals record a status change from sent/batched to approved and
may include a comment.

Commands:
  list    List invoice approvals
  show    Show invoice approval details`,
	Example: `  # List approvals
  xbe view invoice-approvals list

  # Show an approval
  xbe view invoice-approvals show 123`,
}

func init() {
	viewCmd.AddCommand(invoiceApprovalsCmd)
}
