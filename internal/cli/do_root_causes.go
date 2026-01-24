package cli

import "github.com/spf13/cobra"

var doRootCausesCmd = &cobra.Command{
	Use:     "root-causes",
	Aliases: []string{"root-cause"},
	Short:   "Manage root causes",
	Long: `Create, update, and delete root causes.

Root causes track underlying issues for incidents and can be linked
hierarchically to group related causes.`,
	Example: `  # Create a root cause
  xbe do root-causes create \
    --incident-type production-incidents --incident-id 123 \
    --title "Mechanical failure" \
    --description "Hydraulic leak caused downtime" \
    --is-triaged

  # Update a root cause
  xbe do root-causes update 456 --title "Updated title"

  # Delete a root cause
  xbe do root-causes delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doRootCausesCmd)
}
