package cli

import "github.com/spf13/cobra"

var doNewslettersCmd = &cobra.Command{
	Use:     "newsletters",
	Aliases: []string{"newsletter"},
	Short:   "Manage newsletters",
	Long:    `Create newsletters.`,
}

func init() {
	doCmd.AddCommand(doNewslettersCmd)
}
