package cli

import "github.com/spf13/cobra"

var doImporterConfigurationsCmd = &cobra.Command{
	Use:     "importer-configurations",
	Aliases: []string{"importer-configuration"},
	Short:   "Manage importer configurations",
	Long: `Create, update, and delete importer configurations.

Importer configurations define inbound integrations for free ticketing data.
Access is restricted to admin users.`,
}

func init() {
	doCmd.AddCommand(doImporterConfigurationsCmd)
}
