package cli

import "github.com/spf13/cobra"

var projectTransportOrganizationsCmd = &cobra.Command{
	Use:   "project-transport-organizations",
	Short: "Browse project transport organizations",
	Long: `Browse project transport organizations on the XBE platform.

Project transport organizations represent transport companies tied to a broker.
They are referenced by project transport locations and transport orders.

Commands:
  list    List project transport organizations
  show    Show project transport organization details`,
	Example: `  # List project transport organizations
  xbe view project-transport-organizations list

  # Search by name
  xbe view project-transport-organizations list --q "Acme"

  # Show a project transport organization
  xbe view project-transport-organizations show 123`,
}

func init() {
	viewCmd.AddCommand(projectTransportOrganizationsCmd)
}
