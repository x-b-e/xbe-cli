package cli

import "github.com/spf13/cobra"

var equipmentSuppliersCmd = &cobra.Command{
	Use:   "equipment-suppliers",
	Short: "Browse equipment suppliers",
	Long: `Browse equipment suppliers on the XBE platform.

Equipment suppliers provide rental equipment and related services for projects.

Commands:
  list    List equipment suppliers with filtering and pagination
  show    Show equipment supplier details`,
	Example: `  # List equipment suppliers
  xbe view equipment-suppliers list

  # Show details for a supplier
  xbe view equipment-suppliers show 123`,
}

func init() {
	viewCmd.AddCommand(equipmentSuppliersCmd)
}
