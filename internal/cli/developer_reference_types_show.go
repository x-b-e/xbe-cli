package cli

import "github.com/spf13/cobra"

func newDeveloperReferenceTypesShowCmd() *cobra.Command {
	return newGenericShowCmd("developer-reference-types")
}

func init() {
	developerReferenceTypesCmd.AddCommand(newDeveloperReferenceTypesShowCmd())
}
