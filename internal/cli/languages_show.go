package cli

import "github.com/spf13/cobra"

func newLanguagesShowCmd() *cobra.Command {
	return newGenericShowCmd("languages")
}

func init() {
	languagesCmd.AddCommand(newLanguagesShowCmd())
}
