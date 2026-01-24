package cli

import "github.com/spf13/cobra"

var doModelFilterInfosCmd = &cobra.Command{
	Use:     "model-filter-infos",
	Aliases: []string{"model-filter-info"},
	Short:   "Fetch filter options for resources",
	Long: `Fetch filter options for resources.

Model filter infos return available option values for resource filters. Results
are generated on demand and are not persisted.

Commands:
  create    Fetch filter options for a resource`,
	Example: `  # Fetch filter options for projects
  xbe do model-filter-infos create --resource-type projects

  # Limit to selected filter keys
  xbe do model-filter-infos create --resource-type projects --filter-keys customer,project_manager

  # Scope options to a broker
  xbe do model-filter-infos create --resource-type projects --scope-filter broker=123`,
}

func init() {
	doCmd.AddCommand(doModelFilterInfosCmd)
}
