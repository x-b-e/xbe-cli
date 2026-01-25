package cli

import "github.com/spf13/cobra"

func newIncidentTagsShowCmd() *cobra.Command {
	return newGenericShowCmd("incident-tags")
}

func init() {
	incidentTagsCmd.AddCommand(newIncidentTagsShowCmd())
}
