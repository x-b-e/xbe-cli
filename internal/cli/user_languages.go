package cli

import "github.com/spf13/cobra"

var userLanguagesCmd = &cobra.Command{
	Use:     "user-languages",
	Aliases: []string{"user-language"},
	Short:   "View user languages",
	Long:    "Commands for viewing user language preferences.",
}

func init() {
	viewCmd.AddCommand(userLanguagesCmd)
}
