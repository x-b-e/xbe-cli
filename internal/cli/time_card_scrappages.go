package cli

import "github.com/spf13/cobra"

var timeCardScrappagesCmd = &cobra.Command{
	Use:     "time-card-scrappages",
	Aliases: []string{"time-card-scrappage"},
	Short:   "View time card scrappages",
	Long: `View time card scrappages.

Scrappages record a status change from editing/submitted to scrapped and
may include a comment.

Commands:
  list    List time card scrappages
  show    Show time card scrappage details`,
	Example: `  # List scrappages
  xbe view time-card-scrappages list

  # Show a scrappage
  xbe view time-card-scrappages show 123`,
}

func init() {
	viewCmd.AddCommand(timeCardScrappagesCmd)
}
