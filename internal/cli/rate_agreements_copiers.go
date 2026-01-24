package cli

import "github.com/spf13/cobra"

var rateAgreementsCopiersCmd = &cobra.Command{
	Use:     "rate-agreements-copiers",
	Aliases: []string{"rate-agreements-copier"},
	Short:   "View rate agreements copiers",
	Long: `View rate agreements copiers.

Rate agreements copiers copy a template rate agreement to multiple
customers or truckers.

Commands:
  list    List rate agreements copiers
  show    Show rate agreements copier details`,
	Example: `  # List rate agreements copiers
  xbe view rate-agreements-copiers list

  # Show a rate agreements copier
  xbe view rate-agreements-copiers show 123`,
}

func init() {
	viewCmd.AddCommand(rateAgreementsCopiersCmd)
}
