package cli

import "github.com/spf13/cobra"

var doMaterialSupplierMembershipsCmd = &cobra.Command{
	Use:   "material-supplier-memberships",
	Short: "Manage material supplier memberships",
	Long: `Create, update, and delete material supplier memberships.

Material supplier memberships link users to material suppliers and control
role settings, permissions, and notifications.

Commands:
  create    Create a new material supplier membership
  update    Update an existing material supplier membership
  delete    Delete a material supplier membership`,
}

func init() {
	doCmd.AddCommand(doMaterialSupplierMembershipsCmd)
}
