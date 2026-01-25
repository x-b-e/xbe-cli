package cli

import "github.com/spf13/cobra"

var rawRecordsCmd = &cobra.Command{
	Use:     "raw-records",
	Aliases: []string{"raw-record"},
	Short:   "Browse ingest raw records",
	Long: `Browse ingest raw records.

Raw records capture inbound integration payloads, processing status, and
linkages to internal records.

Commands:
  list    List raw records with filtering and pagination
  show    View the full details of a specific raw record`,
}

func init() {
	viewCmd.AddCommand(rawRecordsCmd)
}
