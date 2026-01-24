package cli

import "github.com/spf13/cobra"

var doDispatchUserMatchersCmd = &cobra.Command{
	Use:   "dispatch-user-matchers",
	Short: "Match dispatch users by phone number",
	Long: `Match dispatch users by phone number.

Dispatch user matchers return the dispatch phone number for a caller.

Commands:
  create    Match a dispatch user by phone number`,
	Example: `  # Match a dispatch user by phone number
  xbe do dispatch-user-matchers create --phone-number "+15551234567"

  # Output as JSON
  xbe do dispatch-user-matchers create --phone-number "+15551234567" --json`,
}

func init() {
	doCmd.AddCommand(doDispatchUserMatchersCmd)
}
