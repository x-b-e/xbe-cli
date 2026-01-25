package cli

import "github.com/spf13/cobra"

var doExternalIdentificationsCmd = &cobra.Command{
	Use:     "external-identifications",
	Aliases: []string{"external-identification", "ext-ids", "ext-id"},
	Short:   "Manage external identifications",
	Long:    "Commands for creating, updating, and deleting external identifications.",
}

func init() {
	doCmd.AddCommand(doExternalIdentificationsCmd)
}
