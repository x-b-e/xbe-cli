package cli

import "github.com/spf13/cobra"

var projectUnabandonmentsCmd = &cobra.Command{
	Use:     "project-unabandonments",
	Aliases: []string{"project-unabandonment"},
	Short:   "View project unabandonments",
	Long: `View project unabandonments.

Project unabandonments restore abandoned projects to their previous status.

Commands:
  list    List project unabandonments
  show    Show project unabandonment details`,
	Example: `  # List project unabandonments
  xbe view project-unabandonments list

  # Show a project unabandonment
  xbe view project-unabandonments show 123

  # Output JSON
  xbe view project-unabandonments list --json`,
}

func init() {
	viewCmd.AddCommand(projectUnabandonmentsCmd)
}
