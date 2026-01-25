package cli

import "github.com/spf13/cobra"

var projectAbandonmentsCmd = &cobra.Command{
	Use:     "project-abandonments",
	Aliases: []string{"project-abandonment"},
	Short:   "Browse project abandonments",
	Long: `Browse project abandonments.

Project abandonments record when a project is transitioned to abandoned status.

Commands:
  list    List project abandonments
  show    Show project abandonment details`,
	Example: `  # List project abandonments
  xbe view project-abandonments list

  # Show a project abandonment
  xbe view project-abandonments show 123`,
}

func init() {
	viewCmd.AddCommand(projectAbandonmentsCmd)
}
