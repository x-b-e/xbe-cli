package cli

import "github.com/spf13/cobra"

var doDeveloperReferenceTypesCmd = &cobra.Command{
	Use:   "developer-reference-types",
	Short: "Manage developer reference types",
	Long: `Create, update, and delete developer reference types.

Developer reference types define custom reference fields for developers.

Commands:
  create  Create a new developer reference type
  update  Update an existing developer reference type
  delete  Delete a developer reference type`,
	Example: `  # Create a developer reference type
  xbe do developer-reference-types create --name "PO Number" --developer 123

  # Update a developer reference type
  xbe do developer-reference-types update 456 --name "Updated Name"

  # Delete a developer reference type
  xbe do developer-reference-types delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doDeveloperReferenceTypesCmd)
}
