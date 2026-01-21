package cli

import "github.com/spf13/cobra"

var doCertificationTypesCmd = &cobra.Command{
	Use:   "certification-types",
	Short: "Manage certification types",
	Long: `Manage certification types on the XBE platform.

Certification types define categories of certifications that can be assigned
to truckers, users, or other entities (e.g., CDL, OSHA certifications).

Commands:
  create    Create a new certification type
  update    Update an existing certification type
  delete    Delete a certification type`,
	Example: `  # Create a certification type
  xbe do certification-types create --name "CDL Class A" --can-apply-to Trucker --broker 123

  # Update a certification type
  xbe do certification-types update 456 --name "CDL Class A - Updated"

  # Delete a certification type (requires --confirm)
  xbe do certification-types delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doCertificationTypesCmd)
}
