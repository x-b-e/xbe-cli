package cli

import "github.com/spf13/cobra"

var doTimeCardScrappagesCmd = &cobra.Command{
	Use:     "time-card-scrappages",
	Aliases: []string{"time-card-scrappage"},
	Short:   "Scrap time cards",
	Long:    "Commands for scrapping time cards.",
}

func init() {
	doCmd.AddCommand(doTimeCardScrappagesCmd)
}
