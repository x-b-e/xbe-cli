package cli

import "github.com/spf13/cobra"

var brokersCmd = &cobra.Command{
	Use:     "brokers",
	Aliases: []string{"branches"},
	Short:   "Browse broker/branch information",
	Long: `Browse broker (branch) information.

Brokers (also called branches) are organizations on the XBE platform that
publish newsletters and other content. Use the broker commands to discover
available brokers and get their IDs for filtering other commands.

Commands:
  list    List all brokers with optional filtering`,
	Example: `  # List all active brokers
  xbe view brokers list

  # Search brokers by name
  xbe view brokers list --company-name "Acme"

  # List only active brokers
  xbe view brokers list --is-active true

  # Get broker IDs for use with newsletter filtering
  xbe view brokers list --json | jq '.[].id'`,
}

func init() {
	viewCmd.AddCommand(brokersCmd)
}
