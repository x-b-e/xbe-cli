package cli

import "github.com/spf13/cobra"

var doCertificationsCmd = &cobra.Command{
	Use:     "certifications",
	Aliases: []string{"certification"},
	Short:   "Manage certifications",
	Long:    "Commands for creating, updating, and deleting certifications.",
}

func init() {
	doCmd.AddCommand(doCertificationsCmd)
}
