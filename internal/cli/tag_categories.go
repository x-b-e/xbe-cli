package cli

import "github.com/spf13/cobra"

var tagCategoriesCmd = &cobra.Command{
	Use:   "tag-categories",
	Short: "View tag categories",
	Long: `View tag categories on the XBE platform.

Tag categories organize tags into groups based on what they can be applied to
(e.g., predictions, comments, posts).

Commands:
  list    List tag categories`,
	Example: `  # List tag categories
  xbe view tag-categories list

  # Filter by name
  xbe view tag-categories list --name "market"

  # Output as JSON
  xbe view tag-categories list --json`,
}

func init() {
	viewCmd.AddCommand(tagCategoriesCmd)
}
