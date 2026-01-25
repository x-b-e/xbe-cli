package cli

import "github.com/spf13/cobra"

var doLineupsCmd = &cobra.Command{
	Use:   "lineups",
	Short: "Manage lineups",
	Long: `Create, update, and delete lineups.

Lineups define scheduling windows for a customer, identified by a time range
and optional name.

Commands:
  create    Create a lineup
  update    Update a lineup
  delete    Delete a lineup`,
}

func init() {
	doCmd.AddCommand(doLineupsCmd)
}
