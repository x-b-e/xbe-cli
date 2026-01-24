package cli

import "github.com/spf13/cobra"

var doEmailAddressStatusesCmd = &cobra.Command{
	Use:     "email-address-statuses",
	Aliases: []string{"email-address-status"},
	Short:   "Check email address status",
	Long: `Check email address rejection status.

Email address status lookups query the rejection list and return whether the
address is rejected. Results are generated on demand and are not persisted.

Commands:
  create    Check the status of an email address`,
	Example: `  # Check an email address
  xbe do email-address-statuses create --email-address "user@example.com"

  # JSON output
  xbe do email-address-statuses create --email-address "user@example.com" --json`,
}

func init() {
	doCmd.AddCommand(doEmailAddressStatusesCmd)
}
