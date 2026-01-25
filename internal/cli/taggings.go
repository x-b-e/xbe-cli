package cli

import "github.com/spf13/cobra"

var taggingsCmd = &cobra.Command{
	Use:   "taggings",
	Short: "View taggings",
	Long: `View taggings that associate tags with taggable records.

Taggings link tags to resources such as prediction subjects.

Commands:
  list    List taggings
  show    Show tagging details`,
	Example: `  # List taggings
  xbe view taggings list

  # Filter by tag
  xbe view taggings list --tag-id 123

  # Filter by taggable
  xbe view taggings list --taggable-type PredictionSubject --taggable-id 456

  # Show a tagging
  xbe view taggings show 789`,
}

func init() {
	viewCmd.AddCommand(taggingsCmd)
}
