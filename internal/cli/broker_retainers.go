package cli

import "github.com/spf13/cobra"

var brokerRetainersCmd = &cobra.Command{
	Use:     "broker-retainers",
	Aliases: []string{"broker-retainer"},
	Short:   "Browse broker retainers",
	Long: `Browse broker retainers on the XBE platform.

Broker retainers define retainer agreements between brokers and truckers.

Commands:
  list    List broker retainers with filtering and pagination
  show    Show broker retainer details`,
	Example: `  # List broker retainers
  xbe view broker-retainers list

  # Show a broker retainer
  xbe view broker-retainers show 123

  # Output as JSON
  xbe view broker-retainers list --json`,
}

func init() {
	viewCmd.AddCommand(brokerRetainersCmd)
}
