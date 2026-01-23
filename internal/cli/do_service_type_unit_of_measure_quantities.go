package cli

import "github.com/spf13/cobra"

var doServiceTypeUnitOfMeasureQuantitiesCmd = &cobra.Command{
	Use:   "service-type-unit-of-measure-quantities",
	Short: "Manage service type unit of measure quantities",
	Long: `Create, update, and delete service type unit of measure quantities.

Service type unit of measure quantities represent quantified amounts tied to
resources such as time cards, including explicit and calculated quantities.

Commands:
  create    Create a service type unit of measure quantity
  update    Update a service type unit of measure quantity
  delete    Delete a service type unit of measure quantity`,
}

func init() {
	doCmd.AddCommand(doServiceTypeUnitOfMeasureQuantitiesCmd)
}
