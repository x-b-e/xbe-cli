package cli

import "github.com/spf13/cobra"

var samsaraIntegrationsCmd = &cobra.Command{
	Use:     "samsara-integrations",
	Aliases: []string{"samsara-integration"},
	Short:   "Browse Samsara integrations",
	Long: `Browse Samsara integrations.

Samsara integrations connect Samsara accounts to XBE for telematics data.

Commands:
  list    List Samsara integrations with filtering
  show    Show Samsara integration details`,
	Example: `  # List Samsara integrations
  xbe view samsara-integrations list

  # Show a Samsara integration
  xbe view samsara-integrations show 123`,
}

func init() {
	viewCmd.AddCommand(samsaraIntegrationsCmd)
}
