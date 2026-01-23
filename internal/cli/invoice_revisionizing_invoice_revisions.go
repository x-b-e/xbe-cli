package cli

import "github.com/spf13/cobra"

var invoiceRevisionizingInvoiceRevisionsCmd = &cobra.Command{
	Use:     "invoice-revisionizing-invoice-revisions",
	Aliases: []string{"invoice-revisionizing-invoice-revision"},
	Short:   "Browse invoice revisionizing invoice revisions",
	Long: `Browse invoice revisionizing invoice revisions on the XBE platform.

Invoice revisionizing invoice revisions link revisionizing work to the
specific invoice revisions and invoices being updated.

Commands:
  list    List invoice revisionizing invoice revisions
  show    Show invoice revisionizing invoice revision details`,
	Example: `  # List invoice revisionizing invoice revisions
  xbe view invoice-revisionizing-invoice-revisions list

  # Show a specific invoice revisionizing invoice revision
  xbe view invoice-revisionizing-invoice-revisions show 123`,
}

func init() {
	viewCmd.AddCommand(invoiceRevisionizingInvoiceRevisionsCmd)
}
