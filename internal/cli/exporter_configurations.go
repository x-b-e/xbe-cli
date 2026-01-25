package cli

import "github.com/spf13/cobra"

var exporterConfigurationsCmd = &cobra.Command{
	Use:     "exporter-configurations",
	Aliases: []string{"exporter-configuration"},
	Short:   "Browse exporter configurations",
	Long: `Browse exporter configurations.

Exporter configurations define outbound integrations for free ticketing data.

Commands:
  list    List exporter configurations with filtering
  show    Show exporter configuration details`,
	Example: `  # List exporter configurations
  xbe view exporter-configurations list

  # Show an exporter configuration
  xbe view exporter-configurations show 123`,
}

func init() {
	viewCmd.AddCommand(exporterConfigurationsCmd)
}
