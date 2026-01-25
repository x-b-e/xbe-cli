package cli

import "github.com/spf13/cobra"

var organizationInvoicesBatchPdfTemplatesCmd = &cobra.Command{
	Use:     "organization-invoices-batch-pdf-templates",
	Aliases: []string{"organization-invoices-batch-pdf-template"},
	Short:   "View organization invoices batch PDF templates",
	Long: `View organization invoices batch PDF templates on the XBE platform.

Organization invoices batch PDF templates define the template content used to
render invoice batch PDFs for organizations or global defaults.

Commands:
  list    List organization invoices batch PDF templates
  show    Show organization invoices batch PDF template details`,
	Example: `  # List templates
  xbe view organization-invoices-batch-pdf-templates list

  # Filter by organization
  xbe view organization-invoices-batch-pdf-templates list --organization "Broker|123"

  # Show a template
  xbe view organization-invoices-batch-pdf-templates show 456`,
}

func init() {
	viewCmd.AddCommand(organizationInvoicesBatchPdfTemplatesCmd)
}
