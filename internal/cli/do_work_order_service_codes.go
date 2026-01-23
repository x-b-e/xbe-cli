package cli

import "github.com/spf13/cobra"

var doWorkOrderServiceCodesCmd = &cobra.Command{
	Use:   "work-order-service-codes",
	Short: "Manage work order service codes",
	Long: `Create, update, and delete work order service codes.

Work order service codes describe the service categories used on work orders
and are scoped to brokers.

Commands:
  create    Create a new work order service code
  update    Update an existing work order service code
  delete    Delete a work order service code`,
	Example: `  # Create a work order service code
  xbe do work-order-service-codes create --code "HAUL" --broker 123

  # Update a work order service code
  xbe do work-order-service-codes update 456 --description "Hauling service"

  # Delete a work order service code
  xbe do work-order-service-codes delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doWorkOrderServiceCodesCmd)
}
