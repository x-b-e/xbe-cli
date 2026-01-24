package cli

import "github.com/spf13/cobra"

var incidentRequestApprovalsCmd = &cobra.Command{
	Use:     "incident-request-approvals",
	Aliases: []string{"incident-request-approval"},
	Short:   "View incident request approvals",
	Long: `View incident request approvals.

Incident request approvals transition submitted incident requests to approved and create incidents.

Commands:
  list    List incident request approvals
  show    Show incident request approval details`,
	Example: `  # List incident request approvals
  xbe view incident-request-approvals list

  # Show an incident request approval
  xbe view incident-request-approvals show 123

  # Output JSON
  xbe view incident-request-approvals list --json`,
}

func init() {
	viewCmd.AddCommand(incidentRequestApprovalsCmd)
}
