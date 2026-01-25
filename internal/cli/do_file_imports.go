package cli

import "github.com/spf13/cobra"

var doFileImportsCmd = &cobra.Command{
	Use:   "file-imports",
	Short: "Manage file imports",
	Long: `Create, update, and delete file imports.

File imports track uploaded files and their processing status.

Commands:
  create    Create a new file import
  update    Update an existing file import
  delete    Delete a file import`,
}

func init() {
	doCmd.AddCommand(doFileImportsCmd)
}
