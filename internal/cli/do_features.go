package cli

import "github.com/spf13/cobra"

var doFeaturesCmd = &cobra.Command{
	Use:     "features",
	Aliases: []string{"feature"},
	Short:   "Manage features",
	Long:    `Create, update, and delete features.`,
}

func init() {
	doCmd.AddCommand(doFeaturesCmd)
}
