package cli

import "github.com/spf13/cobra"

var languagesCmd = &cobra.Command{
	Use:   "languages",
	Short: "View languages",
	Long: `View languages on the XBE platform.

Languages represent available language options (e.g., English, Spanish)
that can be associated with users and content.

Commands:
  list    List languages`,
	Example: `  # List languages
  xbe view languages list

  # Filter by code
  xbe view languages list --code en

  # Output as JSON
  xbe view languages list --json`,
}

func init() {
	viewCmd.AddCommand(languagesCmd)
}
