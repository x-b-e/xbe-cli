package cli

import "github.com/spf13/cobra"

var tendersCmd = &cobra.Command{
	Use:     "tenders",
	Aliases: []string{"tender"},
	Short:   "Browse and view tenders",
	Long: `Browse and view tenders on the XBE platform.

Tenders represent offers between buyers and sellers for job work.

Commands:
  list    List tenders with filtering and pagination
  show    Show tender details`,
	Example: `  # List tenders
  xbe view tenders list

  # Show a tender
  xbe view tenders show 123`,
}

func init() {
	viewCmd.AddCommand(tendersCmd)
}
