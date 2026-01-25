package cli

import "github.com/spf13/cobra"

var deereEquipmentsCmd = &cobra.Command{
	Use:     "deere-equipments",
	Aliases: []string{"deere-equipment"},
	Short:   "Browse Deere equipment",
	Long: `Browse Deere equipment integrations.

Deere equipment records capture equipment metadata from John Deere integrations
and show assignment status to internal equipment records.

Commands:
  list    List Deere equipment with filtering
  show    Show Deere equipment details`,
	Example: `  # List Deere equipment
  xbe view deere-equipments list

  # Show Deere equipment details
  xbe view deere-equipments show 123`,
}

func init() {
	viewCmd.AddCommand(deereEquipmentsCmd)
}
