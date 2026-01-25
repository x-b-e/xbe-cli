package cli

import "github.com/spf13/cobra"

var doGoMotiveIntegrationsCmd = &cobra.Command{
	Use:     "go-motive-integrations",
	Aliases: []string{"go-motive-integration"},
	Short:   "Manage GoMotive integrations",
	Long: `Manage GoMotive integrations on the XBE platform.

GoMotive integrations connect brokers to GoMotive accounts and track OAuth
connection status for data sync.

Commands:
  create   Create a GoMotive integration
  update   Update a GoMotive integration
  delete   Delete a GoMotive integration`,
}

func init() {
	doCmd.AddCommand(doGoMotiveIntegrationsCmd)
}
