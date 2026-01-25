package cli

import "github.com/spf13/cobra"

var doUserSearchesCmd = &cobra.Command{
	Use:   "user-searches",
	Short: "Search for users by contact method",
	Long: `Search for users by contact method and value.

User searches look up users by email address or mobile number and return the
matching user if one exists.

Commands:
  create    Run a user search`,
	Example: `  # Search by email address
  xbe do user-searches create --contact-method email_address --contact-value "user@example.com"

  # Search by mobile number
  xbe do user-searches create --contact-method mobile_number --contact-value "+15551234567"

  # Restrict matches to admins or members
  xbe do user-searches create --contact-method email_address --contact-value "user@example.com" --only-admin-or-member true

  # Output as JSON
  xbe do user-searches create --contact-method email_address --contact-value "user@example.com" --json`,
}

func init() {
	doCmd.AddCommand(doUserSearchesCmd)
}
