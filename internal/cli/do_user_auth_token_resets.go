package cli

import "github.com/spf13/cobra"

var doUserAuthTokenResetsCmd = &cobra.Command{
	Use:   "user-auth-token-resets",
	Short: "Reset user auth tokens",
	Long: `Reset user auth tokens.

User auth token resets invalidate a user's current auth token and issue a new one.

Commands:
  create    Reset a user's auth token`,
	Example: `  # Reset a user's auth token
  xbe do user-auth-token-resets create --user-id 123

  # Output as JSON
  xbe do user-auth-token-resets create --user-id 123 --json`,
}

func init() {
	doCmd.AddCommand(doUserAuthTokenResetsCmd)
}
