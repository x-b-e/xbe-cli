package cli

import "github.com/spf13/cobra"

var doDeveloperCertifiedWeighersCmd = &cobra.Command{
	Use:   "developer-certified-weighers",
	Short: "Manage developer certified weighers",
	Long: `Manage developer certified weighers on the XBE platform.

Developer certified weighers link developers to users who are certified
for weighing materials.

Commands:
  create    Create a developer certified weigher
  update    Update a developer certified weigher
  delete    Delete a developer certified weigher`,
	Example: `  # Create a developer certified weigher
  xbe do developer-certified-weighers create --developer 123 --user 456 --number CW-001

  # Update a developer certified weigher
  xbe do developer-certified-weighers update 789 --number CW-002 --is-active false

  # Delete a developer certified weigher
  xbe do developer-certified-weighers delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doDeveloperCertifiedWeighersCmd)
}
