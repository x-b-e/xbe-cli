package cli

import "github.com/spf13/cobra"

var materialSupplierMembershipsCmd = &cobra.Command{
	Use:   "material-supplier-memberships",
	Short: "Browse material supplier memberships",
	Long: `Browse and view material supplier memberships on the XBE platform.

Material supplier memberships link users to material suppliers and control
role settings, permissions, and notifications.

Commands:
  list    List material supplier memberships with filtering and pagination
  show    Show material supplier membership details`,
	Example: `  # List material supplier memberships
  xbe view material-supplier-memberships list

  # Show a material supplier membership
  xbe view material-supplier-memberships show 123`,
}

func init() {
	viewCmd.AddCommand(materialSupplierMembershipsCmd)
}
