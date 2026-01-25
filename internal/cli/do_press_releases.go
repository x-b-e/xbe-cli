package cli

import "github.com/spf13/cobra"

var doPressReleasesCmd = &cobra.Command{
	Use:     "press-releases",
	Aliases: []string{"press-release"},
	Short:   "Manage press releases",
	Long:    `Create press releases.`,
}

func init() {
	doCmd.AddCommand(doPressReleasesCmd)
}
