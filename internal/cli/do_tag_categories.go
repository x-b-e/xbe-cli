package cli

import "github.com/spf13/cobra"

var doTagCategoriesCmd = &cobra.Command{
	Use:   "tag-categories",
	Short: "Manage tag categories",
	Long: `Manage tag categories on the XBE platform.

Tag categories organize tags into groups based on what they can be applied to
(e.g., predictions, comments, posts).

Commands:
  create    Create a new tag category
  update    Update an existing tag category
  delete    Delete a tag category`,
	Example: `  # Create a tag category
  xbe do tag-categories create --name "Market Area" --slug "market-area" --can-apply-to PredictionSubject

  # Update a tag category
  xbe do tag-categories update 123 --description "Updated description"

  # Delete a tag category (requires --confirm)
  xbe do tag-categories delete 123 --confirm`,
}

func init() {
	doCmd.AddCommand(doTagCategoriesCmd)
}
