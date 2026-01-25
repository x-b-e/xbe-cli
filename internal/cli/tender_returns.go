package cli

import "github.com/spf13/cobra"

var tenderReturnsCmd = &cobra.Command{
	Use:     "tender-returns",
	Aliases: []string{"tender-return"},
	Short:   "Browse tender returns",
	Long: `Browse tender returns.

Tender returns record when accepted tenders are returned.

Commands:
  list    List tender returns
  show    Show tender return details`,
	Example: `  # List tender returns
  xbe view tender-returns list

  # Show a tender return
  xbe view tender-returns show 123`,
}

func init() {
	viewCmd.AddCommand(tenderReturnsCmd)
}
