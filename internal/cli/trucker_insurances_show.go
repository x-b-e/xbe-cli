package cli

import "github.com/spf13/cobra"

func newTruckerInsurancesShowCmd() *cobra.Command {
	return newGenericShowCmd("trucker-insurances")
}

func init() {
	truckerInsurancesCmd.AddCommand(newTruckerInsurancesShowCmd())
}
