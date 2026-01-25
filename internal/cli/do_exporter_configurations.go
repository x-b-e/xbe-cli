package cli

import "github.com/spf13/cobra"

var doExporterConfigurationsCmd = &cobra.Command{
	Use:     "exporter-configurations",
	Aliases: []string{"exporter-configuration"},
	Short:   "Manage exporter configurations",
	Long: `Create, update, and delete exporter configurations.

Exporter configurations define outbound integrations for free ticketing data.
Access is restricted to admin users.`,
}

func init() {
	doCmd.AddCommand(doExporterConfigurationsCmd)
}
