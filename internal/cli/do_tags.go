package cli

import "github.com/spf13/cobra"

var doTagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "Manage tags",
	Long: `Create, update, and delete tags.

Tags are labels that can be applied to various entities (predictions, comments, etc.)
and are organized into tag categories.

Commands:
  create    Create a new tag
  update    Update an existing tag
  delete    Delete a tag`,
	Example: `  # Create a tag
  xbe do tags create --name "Urgent" --tag-category 123

  # Update a tag
  xbe do tags update 456 --name "High Priority"

  # Delete a tag
  xbe do tags delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doTagsCmd)
}
