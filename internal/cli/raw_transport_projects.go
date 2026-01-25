package cli

import "github.com/spf13/cobra"

var rawTransportProjectsCmd = &cobra.Command{
	Use:     "raw-transport-projects",
	Aliases: []string{"raw-transport-project"},
	Short:   "Browse raw transport projects",
	Long: `Browse raw transport projects on the XBE platform.

Raw transport projects represent upstream project records imported from
transport systems for validation and processing.

Commands:
  list    List raw transport projects
  show    Show raw transport project details`,
	Example: `  # List raw transport projects
  xbe view raw-transport-projects list

  # Show raw transport project details
  xbe view raw-transport-projects show 123`,
}

func init() {
	viewCmd.AddCommand(rawTransportProjectsCmd)
}
