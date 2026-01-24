package cli

import "github.com/spf13/cobra"

var doKeyResultScrappagesCmd = &cobra.Command{
	Use:     "key-result-scrappages",
	Aliases: []string{"key-result-scrappage"},
	Short:   "Manage key result scrappages",
	Long: `Create key result scrappages.

Commands:
  create    Create a key result scrappage`,
}

func init() {
	doCmd.AddCommand(doKeyResultScrappagesCmd)
}
