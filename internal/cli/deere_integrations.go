package cli

import "github.com/spf13/cobra"

var deereIntegrationsCmd = &cobra.Command{
	Use:     "deere-integrations",
	Aliases: []string{"deere-integration"},
	Short:   "Browse Deere integrations",
	Long: `Browse Deere integrations.

Deere integrations connect John Deere accounts to XBE and track OAuth status.

Commands:
  list    List Deere integrations with filtering
  show    Show Deere integration details`,
	Example: `  # List Deere integrations
  xbe view deere-integrations list

  # Show a Deere integration
  xbe view deere-integrations show 123`,
}

func init() {
	viewCmd.AddCommand(deereIntegrationsCmd)
}
