package cli

import "github.com/spf13/cobra"

var organizationInvoicesBatchesCmd = &cobra.Command{
	Use:   "organization-invoices-batches",
	Short: "Browse organization invoices batches",
	Long: `Browse organization invoices batches on the XBE platform.

Organization invoices batches group invoices for organization-level processing.

Commands:
  list    List organization invoices batches with filtering
  show    Show organization invoices batch details`,
	Example: `  # List organization invoices batches
  xbe view organization-invoices-batches list

  # Show an organization invoices batch
  xbe view organization-invoices-batches show 123`,
}

func init() {
	viewCmd.AddCommand(organizationInvoicesBatchesCmd)
}
