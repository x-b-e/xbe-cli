package cli

import "github.com/spf13/cobra"

var doProductionIncidentsCmd = &cobra.Command{
	Use:     "production-incidents",
	Aliases: []string{"production-incident"},
	Short:   "Manage production incidents",
	Long:    "Create, update, and delete production incidents.",
}

func init() {
	doCmd.AddCommand(doProductionIncidentsCmd)
}
