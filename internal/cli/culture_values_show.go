package cli

import "github.com/spf13/cobra"

func newCultureValuesShowCmd() *cobra.Command {
	return newGenericShowCmd("culture-values")
}

func init() {
	cultureValuesCmd.AddCommand(newCultureValuesShowCmd())
}
