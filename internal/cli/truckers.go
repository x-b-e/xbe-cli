package cli

import "github.com/spf13/cobra"

var truckersCmd = &cobra.Command{
	Use:   "truckers",
	Short: "Browse and view truckers",
	Long: `Browse and view truckers on the XBE platform.

Truckers are trucking companies that transport materials.
Use the list command to find trucker IDs for filtering posts by creator.

Commands:
  list    List truckers with filtering and pagination`,
	Example: `  # List truckers
  xbe view truckers list

  # Search by company name
  xbe view truckers list --name "Acme"

  # Filter by active status
  xbe view truckers list --active

  # Get results as JSON
  xbe view truckers list --json --limit 10`,
}

func init() {
	viewCmd.AddCommand(truckersCmd)
}
