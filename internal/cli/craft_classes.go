package cli

import "github.com/spf13/cobra"

var craftClassesCmd = &cobra.Command{
	Use:   "craft-classes",
	Short: "View craft classes",
	Long: `View craft classes.

Craft classes are sub-classifications within a craft, used to categorize laborers.

Commands:
  list  List craft classes`,
}

func init() {
	viewCmd.AddCommand(craftClassesCmd)
}
