package cli

import "github.com/spf13/cobra"

var doProjectRejectionsCmd = &cobra.Command{
	Use:     "project-rejections",
	Aliases: []string{"project-rejection"},
	Short:   "Reject projects",
	Long: `Reject projects.

Rejections move submitted projects to rejected status.

Commands:
  create    Reject a project`,
	Example: `  # Reject a project
  xbe do project-rejections create --project 123 --comment "Not ready"`,
}

func init() {
	doCmd.AddCommand(doProjectRejectionsCmd)
}
