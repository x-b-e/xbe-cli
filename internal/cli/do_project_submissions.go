package cli

import "github.com/spf13/cobra"

var doProjectSubmissionsCmd = &cobra.Command{
	Use:     "project-submissions",
	Aliases: []string{"project-submission"},
	Short:   "Submit projects",
	Long: `Submit projects.

Project submissions transition projects from editing or rejected to submitted.

Commands:
  create    Submit a project`,
}

func init() {
	doCmd.AddCommand(doProjectSubmissionsCmd)
}
