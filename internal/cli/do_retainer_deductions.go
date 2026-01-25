package cli

import "github.com/spf13/cobra"

var doRetainerDeductionsCmd = &cobra.Command{
	Use:     "retainer-deductions",
	Aliases: []string{"retainer-deduction"},
	Short:   "Manage retainer deductions",
	Long:    "Commands for creating, updating, and deleting retainer deductions.",
}

func init() {
	doCmd.AddCommand(doRetainerDeductionsCmd)
}
