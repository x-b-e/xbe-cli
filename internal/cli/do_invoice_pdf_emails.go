package cli

import "github.com/spf13/cobra"

var doInvoicePdfEmailsCmd = &cobra.Command{
	Use:     "invoice-pdf-emails",
	Aliases: []string{"invoice-pdf-email"},
	Short:   "Email invoice PDFs",
	Long:    "Commands for emailing invoice PDFs.",
}

func init() {
	doCmd.AddCommand(doInvoicePdfEmailsCmd)
}
