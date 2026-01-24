package cli

import "github.com/spf13/cobra"

var tenderReRatesCmd = &cobra.Command{
	Use:     "tender-re-rates",
	Aliases: []string{"tender-re-rate"},
	Short:   "View tender re-rates",
	Long: `View tender re-rates.

Tender re-rates re-run pricing for one or more tenders and can optionally
re-constrain rates. These entries are write-only on the API.

Commands:
  list    List tender re-rates
  show    Show tender re-rate details`,
	Example: `  # List tender re-rates
  xbe view tender-re-rates list

  # Show a tender re-rate
  xbe view tender-re-rates show 123

  # JSON output
  xbe view tender-re-rates list --json`,
}

func init() {
	viewCmd.AddCommand(tenderReRatesCmd)
}
