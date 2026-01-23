package cli

import "github.com/spf13/cobra"

var externalIdentificationsCmd = &cobra.Command{
	Use:     "external-identifications",
	Aliases: []string{"external-identification", "ext-ids", "ext-id"},
	Short:   "View external identifications",
	Long:    "Commands for viewing external identifications.",
}

func init() {
	viewCmd.AddCommand(externalIdentificationsCmd)
}
