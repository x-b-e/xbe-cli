package cli

import "github.com/spf13/cobra"

var doTruckerInsurancesCmd = &cobra.Command{
	Use:     "trucker-insurances",
	Aliases: []string{"trucker-insurance"},
	Short:   "Manage trucker insurances",
	Long:    "Commands for creating, updating, and deleting trucker insurances.",
}

func init() {
	doCmd.AddCommand(doTruckerInsurancesCmd)
}
