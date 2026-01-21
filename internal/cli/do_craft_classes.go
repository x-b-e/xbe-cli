package cli

import "github.com/spf13/cobra"

var doCraftClassesCmd = &cobra.Command{
	Use:   "craft-classes",
	Short: "Manage craft classes",
	Long: `Create, update, and delete craft classes.

Craft classes are sub-classifications within a craft, used to categorize laborers.

Commands:
  create  Create a new craft class
  update  Update an existing craft class
  delete  Delete a craft class`,
	Example: `  # Create a craft class
  xbe do craft-classes create --name "Journeyman" --code "JRN" --craft 123

  # Update a craft class
  xbe do craft-classes update 456 --name "Updated Name"

  # Delete a craft class
  xbe do craft-classes delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doCraftClassesCmd)
}
