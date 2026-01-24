package cli

import "github.com/spf13/cobra"

var doRetainerPaymentsCmd = &cobra.Command{
	Use:   "retainer-payments",
	Short: "Manage retainer payments",
	Long: `Create, update, and delete retainer payments.

Retainer payments represent scheduled or recorded payments tied to retainer periods.

Commands:
  create    Create a retainer payment
  update    Update a retainer payment
  delete    Delete a retainer payment`,
}

func init() {
	doCmd.AddCommand(doRetainerPaymentsCmd)
}
