package cli

import "github.com/spf13/cobra"

var doProjectDuplicationsCmd = &cobra.Command{
	Use:   "project-duplications",
	Short: "Duplicate projects",
	Long: `Duplicate projects on the XBE platform.

Project duplications create a new project based on a template project, with
optional overrides and the ability to skip copying specific relationships.

Commands:
  create    Duplicate a project`,
	Example: `  # Duplicate a project from a template
  xbe do project-duplications create --project-template 123

  # Duplicate with overrides and skip relations
  xbe do project-duplications create \
    --project-template 123 \
    --derived-project-template-name "Template Copy" \
    --derived-project-number "TMP-001" \
    --derived-due-on 2026-02-01 \
    --derived-is-prevailing-wage-applicable \
    --skip-project-material-types \
    --skip-project-revenue-items`,
}

func init() {
	doCmd.AddCommand(doProjectDuplicationsCmd)
}
