package cli

import "github.com/spf13/cobra"

var doIncidentRequestCancellationsCmd = &cobra.Command{
	Use:     "incident-request-cancellations",
	Aliases: []string{"incident-request-cancellation"},
	Short:   "Cancel incident requests",
	Long: `Cancel incident requests.

Cancellations move submitted incident requests to cancelled status.

Commands:
  create    Cancel an incident request`,
	Example: `  # Cancel an incident request
  xbe do incident-request-cancellations create --incident-request 123 --comment "No longer needed"`,
}

func init() {
	doCmd.AddCommand(doIncidentRequestCancellationsCmd)
}
