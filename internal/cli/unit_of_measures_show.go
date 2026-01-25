package cli

import "github.com/spf13/cobra"

func newUnitOfMeasuresShowCmd() *cobra.Command {
	return newGenericShowCmd("unit-of-measures")
}

func init() {
	unitOfMeasuresCmd.AddCommand(newUnitOfMeasuresShowCmd())
}
