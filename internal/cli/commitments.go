package cli

import "github.com/spf13/cobra"

var commitmentsCmd = &cobra.Command{
	Use:     "commitments",
	Aliases: []string{"commitment"},
	Short:   "Browse commitments",
	Long: `Browse commitments.

Commitments capture agreements between buyers and sellers.

Commands:
  list    List commitments with filtering and pagination
  show    Show full details of a commitment

Note: Commitments are read-only in the CLI.`,
	Example: `  # List commitments
  xbe view commitments list

  # Filter by status
  xbe view commitments list --status active

  # Filter by broker
  xbe view commitments list --broker 123

  # Show a commitment
  xbe view commitments show 456`,
}

func init() {
	viewCmd.AddCommand(commitmentsCmd)
}
