package cli

import "github.com/spf13/cobra"

var materialSuppliersCmd = &cobra.Command{
	Use:   "material-suppliers",
	Short: "Browse and view material suppliers",
	Long: `Browse and view material suppliers on the XBE platform.

Material suppliers are companies that provide materials like asphalt, concrete,
aggregates, etc. Use the list command to find supplier IDs for filtering posts
by creator.

Commands:
  list    List material suppliers with filtering and pagination`,
	Example: `  # List material suppliers
  xbe view material-suppliers list

  # Search by name
  xbe view material-suppliers list --name "Acme"

  # Filter by active status
  xbe view material-suppliers list --active

  # Get results as JSON
  xbe view material-suppliers list --json --limit 10`,
}

func init() {
	viewCmd.AddCommand(materialSuppliersCmd)
}
