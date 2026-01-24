package cli

import "github.com/spf13/cobra"

var doProjectEstimateSetsCmd = &cobra.Command{
	Use:     "project-estimate-sets",
	Aliases: []string{"project-estimate-set"},
	Short:   "Manage project estimate sets",
	Long:    "Create, update, and delete project estimate sets.",
}

func init() {
	doCmd.AddCommand(doProjectEstimateSetsCmd)
}
