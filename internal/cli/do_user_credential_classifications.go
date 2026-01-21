package cli

import "github.com/spf13/cobra"

var doUserCredentialClassificationsCmd = &cobra.Command{
	Use:   "user-credential-classifications",
	Short: "Manage user credential classifications",
	Long: `Create, update, and delete user credential classifications.

These classifications define types of credentials that can be assigned to users.

Commands:
  create  Create a new user credential classification
  update  Update an existing user credential classification
  delete  Delete a user credential classification`,
	Example: `  # Create a user credential classification
  xbe do user-credential-classifications create --name "Driver License"

  # Update a user credential classification
  xbe do user-credential-classifications update 456 --name "Updated Name"

  # Delete a user credential classification
  xbe do user-credential-classifications delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doUserCredentialClassificationsCmd)
}
