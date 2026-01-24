package cli

import "github.com/spf13/cobra"

var doOrganizationInvoicesBatchFilesCmd = &cobra.Command{
	Use:   "organization-invoices-batch-files",
	Short: "Manage organization invoices batch files",
	Long: `Create organization invoices batch files.

Organization invoices batch files represent formatted outputs for invoice
batches created with organization formatters.

Commands:
  create    Create a new organization invoices batch file`,
}

func init() {
	doCmd.AddCommand(doOrganizationInvoicesBatchFilesCmd)
}
