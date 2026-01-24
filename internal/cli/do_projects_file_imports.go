package cli

import "github.com/spf13/cobra"

var doProjectsFileImportsCmd = &cobra.Command{
	Use:     "projects-file-imports",
	Aliases: []string{"projects-file-import"},
	Short:   "Import projects files",
	Long: `Import projects files on the XBE platform.

Commands:
  create    Import projects from a file`,
	Example: `  # Import projects from a file
  xbe do projects-file-imports create --file-import 123 --file-import-type SageProjectsFileImport

  # Dry run a projects file import
  xbe do projects-file-imports create --file-import 123 --file-import-type SageProjectsFileImport --is-dry-run`,
}

func init() {
	doCmd.AddCommand(doProjectsFileImportsCmd)
}
