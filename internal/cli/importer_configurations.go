package cli

import "github.com/spf13/cobra"

var importerConfigurationsCmd = &cobra.Command{
	Use:     "importer-configurations",
	Aliases: []string{"importer-configuration"},
	Short:   "Browse importer configurations",
	Long: `Browse importer configurations.

Importer configurations define inbound integrations for free ticketing data.

Commands:
  list    List importer configurations with filtering
  show    Show importer configuration details`,
	Example: `  # List importer configurations
  xbe view importer-configurations list

  # Show an importer configuration
  xbe view importer-configurations show 123`,
}

func init() {
	viewCmd.AddCommand(importerConfigurationsCmd)
}
