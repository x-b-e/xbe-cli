package cli

import "github.com/spf13/cobra"

var doRawTransportExportsCmd = &cobra.Command{
	Use:   "raw-transport-exports",
	Short: "Manage raw transport exports",
	Long: `Manage raw transport exports on the XBE platform.

Raw transport exports store outbound payloads and export metadata for
transport order integrations.

Commands:
  create    Create a raw transport export
  update    Update a raw transport export (not supported)
  delete    Delete a raw transport export (not supported)`,
	Example: `  # Create a raw transport export
  xbe do raw-transport-exports create --external-order-number ORD-123 --target-database tmw

  # Attempt to update a raw transport export
  xbe do raw-transport-exports update 123

  # Attempt to delete a raw transport export
  xbe do raw-transport-exports delete 123`,
}

func init() {
	doCmd.AddCommand(doRawTransportExportsCmd)
}
