package cli

import "github.com/spf13/cobra"

var doOrganizationInvoicesBatchesCmd = &cobra.Command{
	Use:   "organization-invoices-batches",
	Short: "Manage organization invoices batches",
	Long: `Manage organization invoices batches.

Organization invoices batches group invoices for organization-level processing.

Commands:
  create    Create an organization invoices batch`,
	Example: `  # Create an organization invoices batch
  xbe do organization-invoices-batches create \
    --organization "Broker|123" \
    --invoices 111,222`,
}

func init() {
	doCmd.AddCommand(doOrganizationInvoicesBatchesCmd)
}
