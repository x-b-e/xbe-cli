package cli

import "github.com/spf13/cobra"

var doAdministrativeIncidentsCmd = &cobra.Command{
	Use:   "administrative-incidents",
	Short: "Manage administrative incidents",
	Long: `Manage administrative incidents on the XBE platform.

Commands:
  create    Create a new administrative incident
  update    Update an existing administrative incident
  delete    Delete an administrative incident`,
	Example: `  # Create an administrative incident
  xbe do administrative-incidents create --subject Broker|123 --start-at 2025-01-01T08:00:00Z --status open --kind capacity

  # Update an administrative incident
  xbe do administrative-incidents update 123 --status closed --end-at 2025-01-01T10:00:00Z

  # Delete an administrative incident
  xbe do administrative-incidents delete 123 --confirm`,
}

func init() {
	doCmd.AddCommand(doAdministrativeIncidentsCmd)
}
