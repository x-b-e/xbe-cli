package cli

import "github.com/spf13/cobra"

var doContractorsCmd = &cobra.Command{
	Use:   "contractors",
	Short: "Manage contractors",
	Long: `Create, update, and delete contractors.

Contractors are broker-associated organizations used in job production plans
and incident tracking.

Commands:
  create    Create a new contractor
  update    Update an existing contractor
  delete    Delete a contractor`,
}

func init() {
	doCmd.AddCommand(doContractorsCmd)
}
