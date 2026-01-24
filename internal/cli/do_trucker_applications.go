package cli

import "github.com/spf13/cobra"

var doTruckerApplicationsCmd = &cobra.Command{
	Use:     "trucker-applications",
	Aliases: []string{"trucker-application"},
	Short:   "Manage trucker applications",
	Long:    "Commands for creating, updating, and deleting trucker applications.",
}

func init() {
	doCmd.AddCommand(doTruckerApplicationsCmd)
}
