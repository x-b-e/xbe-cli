package cli

import "github.com/spf13/cobra"

var materialSiteReadingsCmd = &cobra.Command{
	Use:   "material-site-readings",
	Short: "Browse material site readings",
	Long: `Browse material site readings on the XBE platform.

Material site readings capture measured values for a material site measure
(e.g., inventory readings or mixing readings) at a specific timestamp.

Commands:
  list    List material site readings
  show    Show material site reading details`,
	Example: `  # List material site readings
  xbe view material-site-readings list

  # Show a material site reading
  xbe view material-site-readings show 123`,
}

func init() {
	viewCmd.AddCommand(materialSiteReadingsCmd)
}
