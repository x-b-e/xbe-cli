package cli

import "github.com/spf13/cobra"

var commitmentItemsCmd = &cobra.Command{
	Use:     "commitment-items",
	Aliases: []string{"commitment-item"},
	Short:   "Browse commitment items",
	Long: `Browse commitment items.

Commitment items define scheduling and adjustment rules for commitments.

Commands:
  list    List commitment items with filtering and pagination
  show    Show full details of a commitment item`,
	Example: `  # List commitment items
  xbe view commitment-items list

  # Filter by commitment and status
  xbe view commitment-items list --commitment-type customer-commitments --commitment-id 123 --status active

  # Show a commitment item
  xbe view commitment-items show 456`,
}

func init() {
	viewCmd.AddCommand(commitmentItemsCmd)
}
