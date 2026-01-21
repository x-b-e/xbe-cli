package cli

import "github.com/spf13/cobra"

var externalIdentificationTypesCmd = &cobra.Command{
	Use:   "external-identification-types",
	Short: "View external identification types",
	Long: `View external identification types on the XBE platform.

External identification types define the kinds of external IDs that can be
associated with entities (e.g., license numbers for truckers, tax IDs for brokers).

Commands:
  list    List external identification types`,
	Example: `  # List external identification types
  xbe view external-identification-types list

  # Filter by name
  xbe view external-identification-types list --name "license"

  # Output as JSON
  xbe view external-identification-types list --json`,
}

func init() {
	viewCmd.AddCommand(externalIdentificationTypesCmd)
}
