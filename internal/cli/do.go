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
  glossary-terms         Manage glossary term definitions`,
	Example: `  # Submit a material transaction
  xbe do material-transactions submit 123

  # Accept a material transaction
  xbe do material-transactions accept 123

  # Reject with a comment
  xbe do material-transactions reject 123 --comment "Missing data"

  # Update a glossary term
  xbe do glossary-terms update 123 --definition "New definition"

  # Delete a glossary term
  xbe do glossary-terms delete 123 --confirm`,
	Annotations: map[string]string{"group": GroupCore},
}

func init() {
	rootCmd.AddCommand(doCmd)
}
