package cli

import "github.com/spf13/cobra"

var doDevelopersCmd = &cobra.Command{
	Use:   "developers",
	Short: "Manage developers",
	Long: `Create, update, and delete developers on the XBE platform.

Developers are companies that develop projects.

Commands:
  create    Create a new developer
  update    Update an existing developer
  delete    Delete a developer`,
	Example: `  # Create a new developer
  xbe do developers create --name "Acme Development" --broker 123

  # Update a developer
  xbe do developers update 456 --name "New Name"

  # Delete a developer
  xbe do developers delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doDevelopersCmd)
}
