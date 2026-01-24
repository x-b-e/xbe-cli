package cli

import "github.com/spf13/cobra"

var doProjectAbandonmentsCmd = &cobra.Command{
	Use:   "project-abandonments",
	Short: "Abandon projects",
	Long: `Abandon projects on the XBE platform.

Abandonments transition projects to abandoned status. Only projects in editing,
submitted, or rejected status can be abandoned.

Commands:
  create    Abandon a project`,
	Example: `  # Abandon a project
  xbe do project-abandonments create --project 123 --comment "No longer needed"`,
}

func init() {
	doCmd.AddCommand(doProjectAbandonmentsCmd)
}
