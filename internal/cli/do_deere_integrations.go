package cli

import "github.com/spf13/cobra"

var doDeereIntegrationsCmd = &cobra.Command{
	Use:     "deere-integrations",
	Aliases: []string{"deere-integration"},
	Short:   "Manage Deere integrations",
	Long: `Create, update, and delete Deere integrations.

Deere integrations connect John Deere accounts to XBE and track OAuth status.
Access is restricted to admin users.`,
}

func init() {
	doCmd.AddCommand(doDeereIntegrationsCmd)
}
