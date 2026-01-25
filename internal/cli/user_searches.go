package cli

import "github.com/spf13/cobra"

var userSearchesCmd = &cobra.Command{
	Use:     "user-searches",
	Aliases: []string{"user-search"},
	Short:   "Browse user searches",
	Long: `Browse user searches.

User searches look up users by contact method and value. Use the do command
when you need to run a new search.

Commands:
  list    List user searches`,
	Example: `  # List user searches
  xbe view user-searches list

  # Output as JSON
  xbe view user-searches list --json`,
}

func init() {
	viewCmd.AddCommand(userSearchesCmd)
}
