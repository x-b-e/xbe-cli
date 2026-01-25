package cli

import "github.com/spf13/cobra"

func newProjectsShowCmd() *cobra.Command {
	return newGenericShowCmd("projects")
}

func init() {
	projectsCmd.AddCommand(newProjectsShowCmd())
}
