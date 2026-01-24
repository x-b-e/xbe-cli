package cli

import "github.com/spf13/cobra"

var doProjectUnabandonmentsCmd = &cobra.Command{
	Use:     "project-unabandonments",
	Aliases: []string{"project-unabandonment"},
	Short:   "Unabandon projects",
	Long: `Unabandon projects.

Project unabandonments restore abandoned projects to their previous status.

Commands:
  create    Unabandon a project`,
}

func init() {
	doCmd.AddCommand(doProjectUnabandonmentsCmd)
}
