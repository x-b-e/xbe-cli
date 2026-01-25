package cli

import "github.com/spf13/cobra"

var doTaggingsCmd = &cobra.Command{
	Use:   "taggings",
	Short: "Manage taggings",
	Long: `Create and delete taggings.

Taggings link tags to taggable resources such as prediction subjects.

Commands:
  create    Create a tagging
  delete    Delete a tagging`,
	Example: `  # Create a tagging
  xbe do taggings create --tag 123 --taggable-type prediction-subjects --taggable-id 456

  # Delete a tagging
  xbe do taggings delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doTaggingsCmd)
}
