package cli

import "github.com/spf13/cobra"

var doIncidentTagsCmd = &cobra.Command{
	Use:     "incident-tags",
	Aliases: []string{"incident-tag"},
	Short:   "Manage incident tags",
	Long:    `Create incident tags.`,
}

func init() {
	doCmd.AddCommand(doIncidentTagsCmd)
}
