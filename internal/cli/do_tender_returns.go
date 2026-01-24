package cli

import "github.com/spf13/cobra"

var doTenderReturnsCmd = &cobra.Command{
	Use:   "tender-returns",
	Short: "Return tenders",
	Long: `Return tenders on the XBE platform.

Tender returns transition accepted tenders to returned status.

Commands:
  create    Return a tender`,
	Example: `  # Return a tender
  xbe do tender-returns create --tender-type broker-tenders --tender-id 123 --comment "Returned"`,
}

func init() {
	doCmd.AddCommand(doTenderReturnsCmd)
}
