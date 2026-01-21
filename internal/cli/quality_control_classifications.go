package cli

import "github.com/spf13/cobra"

var qualityControlClassificationsCmd = &cobra.Command{
	Use:   "quality-control-classifications",
	Short: "View quality control classifications",
	Long: `View quality control classifications on the XBE platform.

Quality control classifications define types of quality inspections
and checks that can be performed, scoped to a broker organization.

Commands:
  list    List quality control classifications`,
	Example: `  # List quality control classifications
  xbe view quality-control-classifications list

  # Filter by broker
  xbe view quality-control-classifications list --broker 123

  # Filter by name
  xbe view quality-control-classifications list --name "temperature"

  # Output as JSON
  xbe view quality-control-classifications list --json`,
}

func init() {
	viewCmd.AddCommand(qualityControlClassificationsCmd)
}
