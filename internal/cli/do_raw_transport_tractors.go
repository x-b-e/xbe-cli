package cli

import "github.com/spf13/cobra"

var doRawTransportTractorsCmd = &cobra.Command{
	Use:   "raw-transport-tractors",
	Short: "Manage raw transport tractors",
	Long: `Manage raw transport tractor imports on the XBE platform.

Raw transport tractors store inbound tractor payloads, import status, and errors.

Commands:
  create    Create a raw transport tractor
  update    Update a raw transport tractor (not supported)
  delete    Delete a raw transport tractor`,
	Example: `  # Create a raw transport tractor
  xbe do raw-transport-tractors create --external-tractor-id TRC-123 --broker 456 --importer quantix_tmw

  # Attempt to update a raw transport tractor
  xbe do raw-transport-tractors update 123

  # Delete a raw transport tractor
  xbe do raw-transport-tractors delete 123 --confirm`,
}

func init() {
	doCmd.AddCommand(doRawTransportTractorsCmd)
}
