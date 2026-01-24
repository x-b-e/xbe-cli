package cli

import "github.com/spf13/cobra"

var incidentRequestRejectionsCmd = &cobra.Command{
	Use:     "incident-request-rejections",
	Aliases: []string{"incident-request-rejection"},
	Short:   "View incident request rejections",
	Long: `View incident request rejections.

Incident request rejections transition submitted incident requests to rejected.

Commands:
  list    List incident request rejections
  show    Show incident request rejection details`,
	Example: `  # List incident request rejections
  xbe view incident-request-rejections list

  # Show an incident request rejection
  xbe view incident-request-rejections show 123

  # Output JSON
  xbe view incident-request-rejections list --json`,
}

func init() {
	viewCmd.AddCommand(incidentRequestRejectionsCmd)
}
