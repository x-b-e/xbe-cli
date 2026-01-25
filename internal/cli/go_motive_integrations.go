package cli

import "github.com/spf13/cobra"

var goMotiveIntegrationsCmd = &cobra.Command{
	Use:     "go-motive-integrations",
	Aliases: []string{"go-motive-integration"},
	Short:   "Browse GoMotive integrations",
	Long: `Browse GoMotive integrations.

GoMotive integrations connect XBE brokers to GoMotive accounts and track OAuth
connection status for data sync.

Commands:
  list    List GoMotive integrations with filtering
  show    Show GoMotive integration details`,
	Example: `  # List GoMotive integrations
  xbe view go-motive-integrations list

  # Show GoMotive integration details
  xbe view go-motive-integrations show 123`,
}

func init() {
	viewCmd.AddCommand(goMotiveIntegrationsCmd)
}
