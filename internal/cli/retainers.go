package cli

import "github.com/spf13/cobra"

var retainersCmd = &cobra.Command{
	Use:   "retainers",
	Short: "Browse retainers",
	Long: `Browse retainers on the XBE platform.

Retainers define ongoing agreements between buyers and sellers,
including expected earnings and travel limits.

Commands:
  list    List retainers
  show    Show retainer details`,
	Example: `  # List retainers
  xbe view retainers list

  # Filter by status
  xbe view retainers list --status active

  # Show retainer details
  xbe view retainers show 123

  # Output as JSON
  xbe view retainers list --json`,
}

func init() {
	viewCmd.AddCommand(retainersCmd)
}
