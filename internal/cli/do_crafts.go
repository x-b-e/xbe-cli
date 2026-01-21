package cli

import "github.com/spf13/cobra"

var doCraftsCmd = &cobra.Command{
	Use:   "crafts",
	Short: "Manage crafts",
	Long: `Create, update, and delete crafts.

Crafts define trade classifications for workers (e.g., carpenter, electrician)
and are scoped to a broker organization.

Commands:
  create    Create a new craft
  update    Update an existing craft
  delete    Delete a craft`,
	Example: `  # Create a craft
  xbe do crafts create --name "Carpenter" --code "CARP" --broker 123

  # Update a craft
  xbe do crafts update 456 --name "Senior Carpenter"

  # Delete a craft
  xbe do crafts delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doCraftsCmd)
}
