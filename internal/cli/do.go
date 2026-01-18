package cli

import "github.com/spf13/cobra"

var doCmd = &cobra.Command{
	Use:   "do",
	Short: "Create, update, and delete XBE resources",
	Long: `Create, update, and delete XBE resources.

The do command provides write access to XBE platform data. Unlike view commands,
these operations modify data and require authentication.

Resources:
  material-transactions  Manage material transaction status (submit, accept, reject)
  glossary-terms         Manage glossary term definitions
  lane-summary           Generate lane (cycle) summaries`,
	Example: `  # Submit a material transaction
  xbe do material-transactions submit 123

  # Accept a material transaction
  xbe do material-transactions accept 123

  # Reject with a comment
  xbe do material-transactions reject 123 --comment "Missing data"

  # Update a glossary term
  xbe do glossary-terms update 123 --definition "New definition"

  # Delete a glossary term
  xbe do glossary-terms delete 123 --confirm

  # Generate a lane summary by origin/destination
  xbe do lane-summary create --group-by origin,destination --filter broker=123 --filter transaction_at_min=2025-01-17T00:00:00Z --filter transaction_at_max=2025-01-17T23:59:59Z`,
	Annotations: map[string]string{"group": GroupCore},
}

func init() {
	rootCmd.AddCommand(doCmd)
}
