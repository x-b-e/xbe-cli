package cli

import "github.com/spf13/cobra"

var tenderAcceptancesCmd = &cobra.Command{
	Use:     "tender-acceptances",
	Aliases: []string{"tender-acceptance"},
	Short:   "View tender acceptances",
	Long: `View tender acceptances.

Tender acceptances mark broker or customer tenders as accepted and can optionally
reject specific tender job schedule shifts.

Commands:
  list    List tender acceptances
  show    Show tender acceptance details`,
	Example: `  # List tender acceptances
  xbe view tender-acceptances list

  # Show a tender acceptance
  xbe view tender-acceptances show 123

  # JSON output
  xbe view tender-acceptances list --json`,
}

func init() {
	viewCmd.AddCommand(tenderAcceptancesCmd)
}
