package cli

import "github.com/spf13/cobra"

var doProjectTransportOrganizationsCmd = &cobra.Command{
	Use:   "project-transport-organizations",
	Short: "Manage project transport organizations",
	Long: `Manage project transport organizations on the XBE platform.

Project transport organizations represent transport companies tied to a broker.
They are referenced by project transport locations and transport orders.

Commands:
  create    Create a project transport organization
  update    Update a project transport organization
  delete    Delete a project transport organization`,
	Example: `  # Create a project transport organization
  xbe do project-transport-organizations create --name "Acme Transport" --broker 123

  # Update a project transport organization
  xbe do project-transport-organizations update 456 --name "Acme Transport West"

  # Delete a project transport organization (requires --confirm)
  xbe do project-transport-organizations delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doProjectTransportOrganizationsCmd)
}
