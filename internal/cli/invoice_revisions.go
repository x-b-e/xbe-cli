package cli

import "github.com/spf13/cobra"

var invoiceRevisionsCmd = &cobra.Command{
	Use:     "invoice-revisions",
	Aliases: []string{"invoice-revision"},
	Short:   "Browse invoice revisions",
	Long: `Browse invoice revisions.

Invoice revisions capture the line items and metadata for a specific invoice revision.

Commands:
  list    List invoice revisions with filtering
  show    Show invoice revision details`,
	Example: `  # List invoice revisions
  xbe view invoice-revisions list

  # Show an invoice revision
  xbe view invoice-revisions show 123`,
}

func init() {
	viewCmd.AddCommand(invoiceRevisionsCmd)
}
