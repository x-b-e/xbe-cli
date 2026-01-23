package cli

import "github.com/spf13/cobra"

var truckerInsurancesCmd = &cobra.Command{
	Use:     "trucker-insurances",
	Aliases: []string{"trucker-insurance"},
	Short:   "View trucker insurances",
	Long:    "Commands for viewing trucker insurances.",
}

func init() {
	viewCmd.AddCommand(truckerInsurancesCmd)
}
