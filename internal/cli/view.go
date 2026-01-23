package cli

import "github.com/spf13/cobra"

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "Browse and view XBE content",
	Long: `Browse and view XBE content.

The view command provides read-only access to XBE platform data. All view
commands support common flags documented in 'xbe --help'.`,
	Example: `  xbe view projects list                     # List all
  xbe view projects list --status active     # Filter
  xbe view projects show 123                 # Show one
  xbe view projects list --json              # JSON output`,
	Annotations: map[string]string{"group": GroupCore},
}

func init() {
	rootCmd.AddCommand(viewCmd)
}
