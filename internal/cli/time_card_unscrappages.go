package cli

import "github.com/spf13/cobra"

var timeCardUnscrappagesCmd = &cobra.Command{
	Use:     "time-card-unscrappages",
	Aliases: []string{"time-card-unscrappage"},
	Short:   "View time card unscrappages",
	Long: `View time card unscrappages.

Time card unscrappages record when a scrapped time card is restored
back to the submitted status.

Commands:
  list    List time card unscrappages
  show    Show time card unscrappage details`,
	Example: `  # List time card unscrappages
  xbe view time-card-unscrappages list

  # Show a time card unscrappage
  xbe view time-card-unscrappages show 123

  # Output JSON
  xbe view time-card-unscrappages list --json`,
}

func init() {
	viewCmd.AddCommand(timeCardUnscrappagesCmd)
}
