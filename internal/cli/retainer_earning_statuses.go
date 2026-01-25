package cli

import "github.com/spf13/cobra"

var retainerEarningStatusesCmd = &cobra.Command{
	Use:     "retainer-earning-statuses",
	Aliases: []string{"retainer-earning-status"},
	Short:   "Browse retainer earning statuses",
	Long: `Browse retainer earning statuses.

Retainer earning statuses capture expected and actual earnings for a retainer on a calculated date.

Commands:
  list    List retainer earning statuses with filtering and pagination
  show    Show full details of a retainer earning status`,
	Example: `  # List retainer earning statuses
  xbe view retainer-earning-statuses list

  # Filter by retainer
  xbe view retainer-earning-statuses list --retainer 123

  # Filter by calculated date
  xbe view retainer-earning-statuses list --calculated-on 2025-01-15

  # Show retainer earning status details
  xbe view retainer-earning-statuses show 456`,
}

func init() {
	viewCmd.AddCommand(retainerEarningStatusesCmd)
}
