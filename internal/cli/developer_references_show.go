package cli

import "github.com/spf13/cobra"

func newDeveloperReferencesShowCmd() *cobra.Command {
	return newGenericShowCmd("developer-references")
}

func init() {
	developerReferencesCmd.AddCommand(newDeveloperReferencesShowCmd())
}
