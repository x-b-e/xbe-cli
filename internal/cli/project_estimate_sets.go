package cli

import "github.com/spf13/cobra"

var projectEstimateSetsCmd = &cobra.Command{
	Use:     "project-estimate-sets",
	Aliases: []string{"project-estimate-set"},
	Short:   "View project estimate sets",
	Long: `View project estimate sets on the XBE platform.

Project estimate sets group revenue and cost estimates for a project. They
include bid, actual, possible, and custom estimate sets.

Commands:
  list    List project estimate sets
  show    Show project estimate set details`,
	Example: `  # List project estimate sets
  xbe view project-estimate-sets list

  # Show a project estimate set
  xbe view project-estimate-sets show 123`,
}

func init() {
	viewCmd.AddCommand(projectEstimateSetsCmd)
}
