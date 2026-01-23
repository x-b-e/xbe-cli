package cli

import "github.com/spf13/cobra"

var doProjectDivisionsCmd = &cobra.Command{
	Use:     "project-divisions",
	Aliases: []string{"project-division"},
	Short:   "Manage project divisions",
	Long:    `Create project divisions.`,
}

func init() {
	doCmd.AddCommand(doProjectDivisionsCmd)
}
