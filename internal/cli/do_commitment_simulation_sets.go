package cli

import "github.com/spf13/cobra"

var doCommitmentSimulationSetsCmd = &cobra.Command{
	Use:   "commitment-simulation-sets",
	Short: "Manage commitment simulation sets",
	Long: `Manage commitment simulation sets on the XBE platform.

Commands:
  create    Create a new commitment simulation set
  update    Update a commitment simulation set
  delete    Delete a commitment simulation set`,
	Example: `  # Create a commitment simulation set
  xbe do commitment-simulation-sets create --organization-type brokers --organization-id 123 --start-on 2025-01-01 --end-on 2025-01-07 --iteration-count 10

  # Delete a commitment simulation set (requires --confirm)
  xbe do commitment-simulation-sets delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doCommitmentSimulationSetsCmd)
}
