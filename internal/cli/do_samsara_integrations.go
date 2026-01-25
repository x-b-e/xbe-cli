package cli

import "github.com/spf13/cobra"

var doSamsaraIntegrationsCmd = &cobra.Command{
	Use:     "samsara-integrations",
	Aliases: []string{"samsara-integration"},
	Short:   "Manage Samsara integrations",
	Long: `Create, update, and delete Samsara integrations.

Samsara integrations connect Samsara accounts to XBE for telematics data.
Access is restricted to admin users.`,
}

func init() {
	doCmd.AddCommand(doSamsaraIntegrationsCmd)
}
