package cli

import "github.com/spf13/cobra"

var placePredictionsCmd = &cobra.Command{
	Use:   "place-predictions",
	Short: "View place predictions",
	Long: `View place predictions from the XBE platform.

Place predictions provide location autocomplete suggestions based on a
query string. Use these results to populate forms that require a Google
Place ID.

Commands:
  list    List place predictions by query`,
	Example: `  # List predictions for a query
  xbe view place-predictions list --q "Austin"

  # Output as JSON
  xbe view place-predictions list --q "Austin" --json`,
}

func init() {
	viewCmd.AddCommand(placePredictionsCmd)
}
