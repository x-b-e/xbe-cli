package cli

import "github.com/spf13/cobra"

var materialMixDesignsCmd = &cobra.Command{
	Use:     "material-mix-designs",
	Aliases: []string{"material-mix-design"},
	Short:   "View material mix designs",
	Long:    "Commands for viewing material mix designs.",
}

func init() {
	viewCmd.AddCommand(materialMixDesignsCmd)
}
