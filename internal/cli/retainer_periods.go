package cli

import "github.com/spf13/cobra"

var retainerPeriodsCmd = &cobra.Command{
	Use:     "retainer-periods",
	Aliases: []string{"retainer-period"},
	Short:   "Browse retainer periods",
	Long: `Browse retainer periods on the XBE platform.

Retainer periods define the start/end range and weekly payment amount for a retainer.

Commands:
  list    List retainer periods with filtering and pagination
  show    Show retainer period details`,
	Example: `  # List retainer periods
  xbe view retainer-periods list

  # Filter by retainer
  xbe view retainer-periods list --retainer 123

  # Show a retainer period
  xbe view retainer-periods show 456

  # Output JSON
  xbe view retainer-periods list --json`,
}

func init() {
	viewCmd.AddCommand(retainerPeriodsCmd)
}
