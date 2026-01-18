package cli

import "github.com/spf13/cobra"

var featuresCmd = &cobra.Command{
	Use:   "features",
	Short: "Browse and view features",
	Long: `Browse and view features on the XBE platform.

Features are product capabilities and enhancements tracked in the system.
Each feature includes a name, description, release date, and categorization.

Commands:
  list    List features with filtering and pagination
  show    View the full details of a specific feature`,
	Example: `  # List recent features
  xbe view features list

  # Filter by PDCA stage
  xbe view features list --pdca-stage plan

  # View a specific feature
  xbe view features show 123`,
}

func init() {
	viewCmd.AddCommand(featuresCmd)
}
