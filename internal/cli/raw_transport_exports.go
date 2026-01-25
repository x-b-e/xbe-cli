package cli

import "github.com/spf13/cobra"

var rawTransportExportsCmd = &cobra.Command{
	Use:     "raw-transport-exports",
	Aliases: []string{"raw-transport-export"},
	Short:   "Browse raw transport exports",
	Long: `Browse raw transport exports on the XBE platform.

Raw transport exports capture outbound export payloads, status, and issues for
transport order integrations.

Commands:
  list    List raw transport exports with filtering
  show    Show raw transport export details`,
	Example: `  # List raw transport exports
  xbe view raw-transport-exports list

  # Show a raw transport export
  xbe view raw-transport-exports show 123`,
}

func init() {
	viewCmd.AddCommand(rawTransportExportsCmd)
}
