package cli

import "github.com/spf13/cobra"

var doExternalIdentificationTypesCmd = &cobra.Command{
	Use:   "external-identification-types",
	Short: "Manage external identification types",
	Long: `Manage external identification types on the XBE platform.

External identification types define categories of external identifiers that can
be assigned to various entities (e.g., Employee ID, Tax ID, License Number).

Commands:
  create    Create a new external identification type
  update    Update an existing external identification type
  delete    Delete an external identification type`,
	Example: `  # Create an external identification type
  xbe do external-identification-types create --name "Employee ID" --can-apply-to User

  # Update an external identification type
  xbe do external-identification-types update 123 --name "Updated Name"

  # Delete an external identification type (requires --confirm)
  xbe do external-identification-types delete 123 --confirm`,
}

func init() {
	doCmd.AddCommand(doExternalIdentificationTypesCmd)
}
