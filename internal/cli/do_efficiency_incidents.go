package cli

import "github.com/spf13/cobra"

var doEfficiencyIncidentsCmd = &cobra.Command{
	Use:   "efficiency-incidents",
	Short: "Manage efficiency incidents",
	Long: `Manage efficiency incidents on the XBE platform.

Commands:
  create    Create a new efficiency incident
  update    Update an existing efficiency incident
  delete    Delete an efficiency incident`,
	Example: `  # Create an efficiency incident
  xbe do efficiency-incidents create --subject Broker|123 --start-at 2025-01-01T08:00:00Z --status open --kind over_trucking

  # Update an efficiency incident
  xbe do efficiency-incidents update 123 --status closed --end-at 2025-01-01T10:00:00Z

  # Delete an efficiency incident
  xbe do efficiency-incidents delete 123 --confirm`,
}

func init() {
	doCmd.AddCommand(doEfficiencyIncidentsCmd)
}
