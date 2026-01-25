package cli

import "github.com/spf13/cobra"

var doOrganizationInvoicesBatchPdfTemplatesCmd = &cobra.Command{
	Use:     "organization-invoices-batch-pdf-templates",
	Aliases: []string{"organization-invoices-batch-pdf-template"},
	Short:   "Manage organization invoices batch PDF templates",
	Long: `Create and update organization invoices batch PDF templates.

Commands:
  create    Create a PDF template
  update    Update a PDF template`,
	Example: `  # Create a template for a broker organization
  xbe do organization-invoices-batch-pdf-templates create \\
    --organization Broker|123 \\
    --broker 123 \\
    --description "Default batch invoice template" \\
    --template "{{invoice_number}}"

  # Update a template description
  xbe do organization-invoices-batch-pdf-templates update 456 --description "Updated description"`,
}

func init() {
	doCmd.AddCommand(doOrganizationInvoicesBatchPdfTemplatesCmd)
}
