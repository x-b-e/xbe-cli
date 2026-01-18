package cli

import "github.com/spf13/cobra"

var usersCmd = &cobra.Command{
	Use:   "users",
	Short: "Browse and view users",
	Long: `Browse and view users on the XBE platform.

Users are individuals who can log in and interact with the platform.
Use the list command to find user IDs for filtering posts by creator.

Commands:
  list    List users with filtering and pagination`,
	Example: `  # List users
  xbe view users list

  # Search by name
  xbe view users list --name "John"

  # Filter by admin status
  xbe view users list --is-admin

  # Get results as JSON
  xbe view users list --json --limit 10`,
}

func init() {
	viewCmd.AddCommand(usersCmd)
}
