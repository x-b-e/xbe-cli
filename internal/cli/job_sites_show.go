package cli

import "github.com/spf13/cobra"

func newJobSitesShowCmd() *cobra.Command {
	return newGenericShowCmd("job-sites")
}

func init() {
	jobSitesCmd.AddCommand(newJobSitesShowCmd())
}
