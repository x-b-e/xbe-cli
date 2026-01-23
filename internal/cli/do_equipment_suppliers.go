package cli

import "github.com/spf13/cobra"

var doEquipmentSuppliersCmd = &cobra.Command{
	Use:   "equipment-suppliers",
	Short: "Manage equipment suppliers",
	Long: `Create, update, and delete equipment suppliers.

Equipment suppliers provide rental equipment and related services for projects.

Commands:
  create    Create a new equipment supplier
  update    Update an existing equipment supplier
  delete    Delete an equipment supplier`,
}

func init() {
	doCmd.AddCommand(doEquipmentSuppliersCmd)
}
