package cli

import "github.com/spf13/cobra"

var projectPhasesCmd = &cobra.Command{
	Use:     "project-phases",
	Aliases: []string{"project-phase"},
	Short:   "View project phases",
	Long:    "Commands for viewing project phases.",
}

func init() {
	viewCmd.AddCommand(projectPhasesCmd)
}
