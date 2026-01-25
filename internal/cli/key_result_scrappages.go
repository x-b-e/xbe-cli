package cli

import "github.com/spf13/cobra"

var keyResultScrappagesCmd = &cobra.Command{
	Use:     "key-result-scrappages",
	Aliases: []string{"key-result-scrappage"},
	Short:   "View key result scrappages",
	Long: `View key result scrappages.

Key result scrappages record when a key result is marked as scrapped.

Commands:
  list    List key result scrappages
  show    Show key result scrappage details`,
	Example: `  # List key result scrappages
  xbe view key-result-scrappages list

  # Show a key result scrappage
  xbe view key-result-scrappages show 123

  # Output JSON
  xbe view key-result-scrappages list --json`,
}

func init() {
	viewCmd.AddCommand(keyResultScrappagesCmd)
}
