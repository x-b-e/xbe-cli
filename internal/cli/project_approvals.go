package cli

import "github.com/spf13/cobra"

var projectApprovalsCmd = &cobra.Command{
	Use:     "project-approvals",
	Aliases: []string{"project-approval"},
	Short:   "View project approvals",
	Long: `View project approvals.

Project approvals transition projects from submitted to approved.

Commands:
  list    List project approvals
  show    Show project approval details`,
	Example: `  # List project approvals
  xbe view project-approvals list

  # Show a project approval
  xbe view project-approvals show 123

  # Output JSON
  xbe view project-approvals list --json`,
}

func init() {
	viewCmd.AddCommand(projectApprovalsCmd)
}
