package cli

import "github.com/spf13/cobra"

var incidentRequestsCmd = &cobra.Command{
	Use:     "incident-requests",
	Aliases: []string{"incident-request"},
	Short:   "View incident requests",
	Long: `View incident requests submitted for job schedule shifts.

Incident requests capture time-impacting incidents with start/end times,
status, and supporting details.

Commands:
  list    List incident requests
  show    Show incident request details`,
	Example: `  # List incident requests
  xbe view incident-requests list

  # Filter by status
  xbe view incident-requests list --status submitted

  # Show an incident request
  xbe view incident-requests show 123`,
}

func init() {
	viewCmd.AddCommand(incidentRequestsCmd)
}
