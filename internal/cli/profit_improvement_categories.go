package cli

import "github.com/spf13/cobra"

var profitImprovementCategoriesCmd = &cobra.Command{
	Use:   "profit-improvement-categories",
	Short: "View profit improvement categories",
	Long: `View profit improvement categories on the XBE platform.

Profit improvement categories organize profit improvements into groups for
reporting and analysis purposes.

Note: Profit improvement categories are read-only and cannot be created,
updated, or deleted through the API.

Commands:
  list    List profit improvement categories`,
	Example: `  # List profit improvement categories
  xbe view profit-improvement-categories list

  # Filter by name
  xbe view profit-improvement-categories list --name "safety"

  # Output as JSON
  xbe view profit-improvement-categories list --json`,
}

func init() {
	viewCmd.AddCommand(profitImprovementCategoriesCmd)
}
