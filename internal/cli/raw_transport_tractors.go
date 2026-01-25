package cli

import "github.com/spf13/cobra"

var rawTransportTractorsCmd = &cobra.Command{
	Use:     "raw-transport-tractors",
	Aliases: []string{"raw-transport-tractor"},
	Short:   "Browse raw transport tractors",
	Long: `Browse raw transport tractors on the XBE platform.

Raw transport tractors capture inbound tractor data from transport integrations
alongside import status and errors.

Commands:
  list    List raw transport tractors with filtering
  show    Show raw transport tractor details`,
	Example: `  # List raw transport tractors
  xbe view raw-transport-tractors list

  # Show a raw transport tractor
  xbe view raw-transport-tractors show 123`,
}

func init() {
	viewCmd.AddCommand(rawTransportTractorsCmd)
}
