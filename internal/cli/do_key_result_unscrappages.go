package cli

import "github.com/spf13/cobra"

var doKeyResultUnscrappagesCmd = &cobra.Command{
	Use:   "key-result-unscrappages",
	Short: "Unscrap key results",
	Long: `Unscrap key results on the XBE platform.

Unscrappages transition key results from scrapped back to the most recent
non-scrapped status (or unknown when none exists).

Commands:
  create    Unscrap a key result`,
	Example: `  # Unscrap a key result
  xbe do key-result-unscrappages create --key-result 123 --comment "Restoring key result"`,
}

func init() {
	doCmd.AddCommand(doKeyResultUnscrappagesCmd)
}
