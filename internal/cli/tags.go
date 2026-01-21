package cli

import "github.com/spf13/cobra"

var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "View tags",
	Long: `View tags on the XBE platform.

Tags are labels that can be applied to various entities (predictions, comments, etc.)
and are organized into tag categories.

Commands:
  list    List tags`,
	Example: `  # List tags
  xbe view tags list

  # Filter by name
  xbe view tags list --name "important"

  # Filter by tag category
  xbe view tags list --tag-category-id 123

  # Output as JSON
  xbe view tags list --json`,
}

func init() {
	viewCmd.AddCommand(tagsCmd)
}
