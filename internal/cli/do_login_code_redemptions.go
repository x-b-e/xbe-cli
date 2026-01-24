package cli

import "github.com/spf13/cobra"

var doLoginCodeRedemptionsCmd = &cobra.Command{
	Use:   "login-code-redemptions",
	Short: "Redeem login codes for auth tokens",
	Long: `Redeem login codes for auth tokens.

Login code redemptions exchange a one-time code for an auth token.

Commands:
  create    Redeem a login code`,
	Example: `  # Redeem a login code
  xbe do login-code-redemptions create --code 123456

  # Redeem with a device ID
  xbe do login-code-redemptions create --code 123456 --device-id "device-abc"

  # Output as JSON
  xbe do login-code-redemptions create --code 123456 --json`,
}

func init() {
	doCmd.AddCommand(doLoginCodeRedemptionsCmd)
}
