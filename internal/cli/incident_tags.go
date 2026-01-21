package cli

import "github.com/spf13/cobra"

var incidentTagsCmd = &cobra.Command{
	Use:   "incident-tags",
	Short: "View incident tags",
	Long: `View incident tags on the XBE platform.

Incident tags are used to categorize and label safety incidents for
reporting and analysis purposes.

Commands:
  list    List incident tags`,
	Example: `  # List incident tags
  xbe view incident-tags list

  # Filter by slug
  xbe view incident-tags list --slug "property-damage"`,
}

func init() {
	viewCmd.AddCommand(incidentTagsCmd)
}
