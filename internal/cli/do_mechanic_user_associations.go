package cli

import "github.com/spf13/cobra"

var doMechanicUserAssociationsCmd = &cobra.Command{
	Use:     "mechanic-user-associations",
	Aliases: []string{"mechanic-user-association"},
	Short:   "Manage mechanic user associations",
	Long: `Create, update, and delete mechanic user associations.

Mechanic user associations link users to maintenance requirements.

Commands:
  create    Create a record
  update    Update a record
  delete    Delete a record`,
}

func init() {
	doCmd.AddCommand(doMechanicUserAssociationsCmd)
}
