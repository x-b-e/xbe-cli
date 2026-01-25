package cli

import "github.com/spf13/cobra"

var doNativeAppReleasesCmd = &cobra.Command{
	Use:     "native-app-releases",
	Aliases: []string{"native-app-release"},
	Short:   "Manage native app releases",
	Long:    "Create, update, and delete native app releases.",
}

func init() {
	doCmd.AddCommand(doNativeAppReleasesCmd)
}
