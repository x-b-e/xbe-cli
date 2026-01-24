package cli

import "github.com/spf13/cobra"

var doRawTransportProjectsCmd = &cobra.Command{
	Use:     "raw-transport-projects",
	Aliases: []string{"raw-transport-project"},
	Short:   "Manage raw transport projects",
	Long: `Create and delete raw transport project records on the XBE platform.

Raw transport projects capture upstream project data imports and are typically
created by integrations. Deleting removes the raw record.

Commands:
  create   Create a raw transport project
  delete   Delete a raw transport project`,
	Example: `  # Create a raw transport project
  xbe do raw-transport-projects create --broker 123 --external-project-number PROJ-0001 --tables '[]'

  # Delete a raw transport project
  xbe do raw-transport-projects delete 123 --confirm`,
}

func init() {
	doCmd.AddCommand(doRawTransportProjectsCmd)
}
