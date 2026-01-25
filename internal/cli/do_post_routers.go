package cli

import "github.com/spf13/cobra"

var doPostRoutersCmd = &cobra.Command{
	Use:     "post-routers",
	Aliases: []string{"post-router"},
	Short:   "Manage post routers",
	Long:    "Create post routers to analyze posts and enqueue routing jobs.",
}

func init() {
	doCmd.AddCommand(doPostRoutersCmd)
}
