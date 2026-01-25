package cli

import "github.com/spf13/cobra"

var doInvoiceApprovalsCmd = &cobra.Command{
	Use:     "invoice-approvals",
	Aliases: []string{"invoice-approval"},
	Short:   "Approve invoices",
	Long:    "Commands for approving invoices.",
}

func init() {
	doCmd.AddCommand(doInvoiceApprovalsCmd)
}
