package cli

import "github.com/spf13/cobra"

var fileImportsCmd = &cobra.Command{
	Use:   "file-imports",
	Short: "Browse file imports",
	Long: `Browse file imports on the XBE platform.

File imports track uploaded files and their processing status.

Commands:
  list    List file imports with filtering
  show    Show file import details`,
	Example: `  # List file imports
  xbe view file-imports list

  # Show a file import
  xbe view file-imports show 123`,
}

func init() {
	viewCmd.AddCommand(fileImportsCmd)
}
