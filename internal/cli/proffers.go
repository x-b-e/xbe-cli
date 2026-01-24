package cli

import "github.com/spf13/cobra"

var proffersCmd = &cobra.Command{
	Use:     "proffers",
	Aliases: []string{"proffer"},
	Short:   "View proffers",
	Long:    "Commands for viewing proffers (feature suggestions).",
}

func init() {
	viewCmd.AddCommand(proffersCmd)
}
