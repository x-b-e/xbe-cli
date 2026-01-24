package cli

import "github.com/spf13/cobra"

var doRawTransportDriversCmd = &cobra.Command{
	Use:     "raw-transport-drivers",
	Aliases: []string{"raw-transport-driver"},
	Short:   "Manage raw transport drivers",
	Long: `Create and delete raw transport driver records on the XBE platform.

Raw transport drivers capture upstream driver data imports and are typically
created by integrations. Deleting removes the raw record.

Commands:
  create   Create a raw transport driver
  delete   Delete a raw transport driver`,
	Example: `  # Create a raw transport driver
  xbe do raw-transport-drivers create --broker 123 --external-driver-id DRV-0001 --tables '[]'

  # Delete a raw transport driver
  xbe do raw-transport-drivers delete 123 --confirm`,
}

func init() {
	doCmd.AddCommand(doRawTransportDriversCmd)
}
