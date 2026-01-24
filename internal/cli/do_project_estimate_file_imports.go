package cli

import "github.com/spf13/cobra"

var doProjectEstimateFileImportsCmd = &cobra.Command{
	Use:     "project-estimate-file-imports",
	Aliases: []string{"project-estimate-file-import"},
	Short:   "Import project estimate files",
	Long: `Import project estimate files on the XBE platform.

Commands:
  create    Import a project estimate file`,
	Example: `  # Import a project estimate file
  xbe do project-estimate-file-imports create --project 123 --file-import 456 --file-import-type Bid2Win

  # Dry run a project estimate import
  xbe do project-estimate-file-imports create --project 123 --file-import 456 --file-import-type Bid2Win --is-dry-run`,
}

func init() {
	doCmd.AddCommand(doProjectEstimateFileImportsCmd)
}
