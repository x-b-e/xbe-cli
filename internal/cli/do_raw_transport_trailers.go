package cli

import "github.com/spf13/cobra"

var doRawTransportTrailersCmd = &cobra.Command{
	Use:     "raw-transport-trailers",
	Aliases: []string{"raw-transport-trailer"},
	Short:   "Manage raw transport trailers",
	Long: `Create and delete raw transport trailer records on the XBE platform.

Raw transport trailers capture upstream trailer data imports and are typically
created by integrations. Deleting removes the raw record.

Commands:
  create   Create a raw transport trailer
  delete   Delete a raw transport trailer`,
	Example: `  # Create a raw transport trailer
  xbe do raw-transport-trailers create --broker 123 --external-trailer-id TRL-0001 --tables '[]'

  # Delete a raw transport trailer
  xbe do raw-transport-trailers delete 123 --confirm`,
}

func init() {
	doCmd.AddCommand(doRawTransportTrailersCmd)
}
