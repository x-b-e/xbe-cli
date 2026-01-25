package cli

import "github.com/spf13/cobra"

var projectImportFileVerificationsCmd = &cobra.Command{
	Use:     "project-import-file-verifications",
	Aliases: []string{"project-import-file-verification"},
	Short:   "View project import file verifications",
	Long: `View project import file verifications on the XBE platform.

Project import file verifications validate project import files against
supported verification types.

Commands:
  list    List project import file verifications`,
	Example: `  # List project import file verifications for a project
  xbe view project-import-file-verifications list --project 123

  # Output as JSON
  xbe view project-import-file-verifications list --project 123 --json`,
}

func init() {
	viewCmd.AddCommand(projectImportFileVerificationsCmd)
}
