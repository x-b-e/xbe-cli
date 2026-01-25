package cli

import "github.com/spf13/cobra"

var doLoginCodeRequestsCmd = &cobra.Command{
	Use:   "login-code-requests",
	Short: "Request login codes",
	Long: `Request login codes.

Login code requests send a one-time login code to an email address or mobile
number. The API accepts a contact method and device identifier and returns
whether the request was processed.

Commands:
  create    Request a login code`,
	Example: `  # Request a login code for an email
  xbe do login-code-requests create --contact-method user@example.com --device-id "xbe-cli"

  # Request a login code for a mobile number
  xbe do login-code-requests create --contact-method "+18155551234" --device-id "xbe-cli" --json`,
}

func init() {
	doCmd.AddCommand(doLoginCodeRequestsCmd)
}
