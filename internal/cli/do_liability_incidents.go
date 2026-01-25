package cli

import "github.com/spf13/cobra"

var doLiabilityIncidentsCmd = &cobra.Command{
	Use:     "liability-incidents",
	Aliases: []string{"liability-incident"},
	Short:   "Manage liability incidents",
	Long:    "Create, update, and delete liability incidents.",
}

func init() {
	doCmd.AddCommand(doLiabilityIncidentsCmd)
}
