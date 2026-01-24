package cli

import "github.com/spf13/cobra"

var rawTransportDriversCmd = &cobra.Command{
	Use:     "raw-transport-drivers",
	Aliases: []string{"raw-transport-driver"},
	Short:   "Browse raw transport drivers",
	Long: `Browse raw transport drivers on the XBE platform.

Raw transport drivers represent upstream driver records imported from
transport systems for validation and processing.

Commands:
  list    List raw transport drivers
  show    Show raw transport driver details`,
	Example: `  # List raw transport drivers
  xbe view raw-transport-drivers list

  # Show raw transport driver details
  xbe view raw-transport-drivers show 123`,
}

func init() {
	viewCmd.AddCommand(rawTransportDriversCmd)
}
