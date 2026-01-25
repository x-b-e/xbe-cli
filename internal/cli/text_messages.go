package cli

import "github.com/spf13/cobra"

var textMessagesCmd = &cobra.Command{
	Use:     "text-messages",
	Aliases: []string{"text-message"},
	Short:   "Browse text messages",
	Long: `Browse text messages sent through the XBE platform.

Text messages are sourced from Twilio. Listing text messages is restricted
by policy to admin users and defaults to messages sent today.

Commands:
  list    List text messages with filtering
  show    Show text message details`,
	Example: `  # List today's text messages
  xbe view text-messages list

  # Filter by recipient and date
  xbe view text-messages list --to +15551234567 --date-sent 2025-01-20

  # Show text message details
  xbe view text-messages show SM123`,
}

func init() {
	viewCmd.AddCommand(textMessagesCmd)
}
