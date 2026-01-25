package cli

import "github.com/spf13/cobra"

func newProjectImportFileVerificationsShowCmd() *cobra.Command {
	return newGenericShowCmd("project-import-file-verifications")
}

func init() {
	projectImportFileVerificationsCmd.AddCommand(newProjectImportFileVerificationsShowCmd())
}
