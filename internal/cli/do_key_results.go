package cli

import "github.com/spf13/cobra"

var doKeyResultsCmd = &cobra.Command{
	Use:   "key-results",
	Short: "Manage key results",
	Long: `Manage key results on the XBE platform.

Commands:
  create    Create a key result
  update    Update a key result
  delete    Delete a key result`,
	Example: `  # Create a key result
  xbe do key-results create --title "Launch beta" --objective 123

  # Update a key result
  xbe do key-results update 456 --status completed

  # Delete a key result (requires --confirm)
  xbe do key-results delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doKeyResultsCmd)
}
