package cli

import "github.com/spf13/cobra"

var profitImprovementsCmd = &cobra.Command{
	Use:   "profit-improvements",
	Short: "Browse and view profit improvements",
	Long: `Browse and view profit improvements on the XBE platform.

Profit improvements capture initiatives to improve profitability, with estimated
and validated impact, ownership, and gain-share details.

Commands:
  list    List profit improvements with filtering and pagination
  show    Show profit improvement details`,
	Example: `  # List profit improvements
  xbe view profit-improvements list

  # Filter by status
  xbe view profit-improvements list --status submitted

  # Show a profit improvement
  xbe view profit-improvements show 123

  # Output JSON
  xbe view profit-improvements list --json`,
}

func init() {
	viewCmd.AddCommand(profitImprovementsCmd)
}
