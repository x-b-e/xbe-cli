package cli

import "github.com/spf13/cobra"

var retainerDeductionsCmd = &cobra.Command{
	Use:     "retainer-deductions",
	Aliases: []string{"retainer-deduction"},
	Short:   "Browse retainer deductions",
	Long: `Browse retainer deductions on the XBE platform.

Retainer deductions record amounts deducted from a retainer and any
supporting notes.

Commands:
  list    List retainer deductions with filtering and pagination
  show    Show retainer deduction details`,
	Example: `  # List retainer deductions
  xbe view retainer-deductions list

  # Filter by retainer
  xbe view retainer-deductions list --retainer 123

  # Show a retainer deduction
  xbe view retainer-deductions show 456

  # Output JSON
  xbe view retainer-deductions list --json`,
}

func init() {
	viewCmd.AddCommand(retainerDeductionsCmd)
}
