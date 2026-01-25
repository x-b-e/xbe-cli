package cli

import "github.com/spf13/cobra"

var doProjectImportFileVerificationsCmd = &cobra.Command{
	Use:     "project-import-file-verifications",
	Aliases: []string{"project-import-file-verification"},
	Short:   "Verify project import files",
	Long:    "Commands for verifying project import files.",
}

func init() {
	doCmd.AddCommand(doProjectImportFileVerificationsCmd)
}
