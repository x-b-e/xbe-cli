package cli

import "github.com/spf13/cobra"

var communicationsCmd = &cobra.Command{
	Use:     "communications",
	Aliases: []string{"communication"},
	Short:   "Browse communications",
	Long: `Browse communications.

Communications capture inbound and outbound messages, along with delivery status.

Commands:
  list    List communications with filtering and pagination
  show    Show full details of a communication`,
	Example: `  # List communications
  xbe view communications list

  # Show communication details
  xbe view communications show 123`,
}

func init() {
	viewCmd.AddCommand(communicationsCmd)
}
