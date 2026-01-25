package cli

import "github.com/spf13/cobra"

var doLineupScenariosCmd = &cobra.Command{
	Use:     "lineup-scenarios",
	Aliases: []string{"lineup-scenario"},
	Short:   "Manage lineup scenarios",
	Long: `Manage lineup scenarios on the XBE platform.

Commands:
  create    Create a lineup scenario
  update    Update a lineup scenario
  delete    Delete a lineup scenario`,
	Example: `  # Create a lineup scenario
  xbe do lineup-scenarios create --broker 123 --date 2026-01-23 --window day`,
}

func init() {
	doCmd.AddCommand(doLineupScenariosCmd)
}
