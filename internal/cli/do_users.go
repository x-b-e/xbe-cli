package cli

import "github.com/spf13/cobra"

var doUsersCmd = &cobra.Command{
	Use:   "users",
	Short: "Manage users",
	Long: `Manage users on the XBE platform.

Commands:
  create    Create a new user
  update    Update an existing user`,
	Example: `  # Create a user
  xbe do users create --name "John Doe" --email "john@example.com"

  # Update a user's name
  xbe do users update 123 --name "Jane Doe"`,
}

func init() {
	doCmd.AddCommand(doUsersCmd)
}
