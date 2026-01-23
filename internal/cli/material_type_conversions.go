package cli

import "github.com/spf13/cobra"

var materialTypeConversionsCmd = &cobra.Command{
	Use:   "material-type-conversions",
	Short: "Browse material type conversions",
	Long: `Browse material type conversions on the XBE platform.

Material type conversions map a material supplier's material type (and optional
material site) to a foreign supplier/material type mapping.

Commands:
  list    List material type conversions
  show    Show material type conversion details`,
	Example: `  # List conversions
  xbe view material-type-conversions list

  # Show conversion details
  xbe view material-type-conversions show 123`,
}

func init() {
	viewCmd.AddCommand(materialTypeConversionsCmd)
}
