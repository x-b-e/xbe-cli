package cli

import "github.com/spf13/cobra"

var serviceTypesCmd = &cobra.Command{
	Use:   "service-types",
	Short: "View service types",
	Long: `View service types on the XBE platform.

Service types define the types of services that can be performed on jobs
(e.g., hauling, spreading, compaction). They are used to categorize work
and track costs.

Commands:
  list    List service types`,
	Example: `  # List service types
  xbe view service-types list

  # Search by name
  xbe view service-types list --name "haul"`,
}

func init() {
	viewCmd.AddCommand(serviceTypesCmd)
}
