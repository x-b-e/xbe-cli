package cli

import "github.com/spf13/cobra"

var doProjectCategoriesCmd = &cobra.Command{
	Use:     "project-categories",
	Aliases: []string{"project-category"},
	Short:   "Manage project categories",
	Long:    `Create project categories.`,
}

func init() {
	doCmd.AddCommand(doProjectCategoriesCmd)
}
