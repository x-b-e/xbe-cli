package cli

import "github.com/spf13/cobra"

var doCultureValuesCmd = &cobra.Command{
	Use:   "culture-values",
	Short: "Manage culture values",
	Long: `Manage culture values on the XBE platform.

Culture values define organizational values used for public praise and
recognition. They help reinforce company culture.

Commands:
  create    Create a new culture value
  update    Update an existing culture value
  delete    Delete a culture value`,
	Example: `  # Create a culture value
  xbe do culture-values create --name "Integrity" --organization Broker|123

  # Update a culture value's position
  xbe do culture-values update 456 --position 2

  # Delete a culture value (requires --confirm)
  xbe do culture-values delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doCultureValuesCmd)
}
