package cli

import "github.com/spf13/cobra"

var crewAssignmentConfirmationsCmd = &cobra.Command{
	Use:     "crew-assignment-confirmations",
	Aliases: []string{"crew-assignment-confirmation"},
	Short:   "Browse crew assignment confirmations",
	Long: `Browse crew assignment confirmations.

Crew assignment confirmations record when a resource confirms a crew
requirement assignment.

Commands:
  list    List confirmations with filtering and pagination
  show    Show full details of a confirmation`,
	Example: `  # List confirmations
  xbe view crew-assignment-confirmations list

  # Filter by crew requirement
  xbe view crew-assignment-confirmations list --crew-requirement 123

  # Filter by resource
  xbe view crew-assignment-confirmations list --resource-type laborers --resource-id 456

  # Show a confirmation
  xbe view crew-assignment-confirmations show 789`,
}

func init() {
	viewCmd.AddCommand(crewAssignmentConfirmationsCmd)
}
