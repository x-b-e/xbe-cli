package cli

import "github.com/spf13/cobra"

var jobProductionPlanInspectorsCmd = &cobra.Command{
	Use:     "job-production-plan-inspectors",
	Aliases: []string{"job-production-plan-inspector"},
	Short:   "Browse job production plan inspectors",
	Long: `Browse job production plan inspectors on the XBE platform.

Job production plan inspectors link a job production plan to a user inspector.

Commands:
  list    List job production plan inspectors with filtering and pagination
  show    Show job production plan inspector details`,
	Example: `  # List job production plan inspectors
  xbe view job-production-plan-inspectors list

  # Show a job production plan inspector
  xbe view job-production-plan-inspectors show 123

  # Output as JSON
  xbe view job-production-plan-inspectors list --json`,
}

func init() {
	viewCmd.AddCommand(jobProductionPlanInspectorsCmd)
}
