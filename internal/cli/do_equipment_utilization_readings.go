package cli

import "github.com/spf13/cobra"

var doEquipmentUtilizationReadingsCmd = &cobra.Command{
	Use:     "equipment-utilization-readings",
	Aliases: []string{"equipment-utilization-reading"},
	Short:   "Manage equipment utilization readings",
	Long:    "Commands for creating, updating, and deleting equipment utilization readings.",
}

func init() {
	doCmd.AddCommand(doEquipmentUtilizationReadingsCmd)
}
