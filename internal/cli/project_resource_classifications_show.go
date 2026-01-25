package cli

import "github.com/spf13/cobra"

func newProjectResourceClassificationsShowCmd() *cobra.Command {
	return newGenericShowCmd("project-resource-classifications")
}

func init() {
	projectResourceClassificationsCmd.AddCommand(newProjectResourceClassificationsShowCmd())
}
