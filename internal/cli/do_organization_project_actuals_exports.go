package cli

import "github.com/spf13/cobra"

var doOrganizationProjectActualsExportsCmd = &cobra.Command{
	Use:     "organization-project-actuals-exports",
	Aliases: []string{"organization-project-actuals-export"},
	Short:   "Export organization project actuals",
	Long: `Export organization project actuals.

Organization project actuals exports send formatted project actuals export files
through partner integrations.

Commands:
  create    Export an organization project actuals export`,
	Example: `  # Export organization project actuals
  xbe do organization-project-actuals-exports create --project-actuals-export 123

  # Run export as a dry run
  xbe do organization-project-actuals-exports create --project-actuals-export 123 --dry-run`,
}

func init() {
	doCmd.AddCommand(doOrganizationProjectActualsExportsCmd)
}
