package cli

import "github.com/spf13/cobra"

var doProjectApprovalsCmd = &cobra.Command{
	Use:     "project-approvals",
	Aliases: []string{"project-approval"},
	Short:   "Approve projects",
	Long: `Approve projects.

Project approvals transition projects from submitted to approved.

Commands:
  create    Approve a project`,
}

func init() {
	doCmd.AddCommand(doProjectApprovalsCmd)
}
