package cli

import "github.com/spf13/cobra"

var developerReferenceTypesCmd = &cobra.Command{
	Use:   "developer-reference-types",
	Short: "View developer reference types",
	Long: `View developer reference types.

Developer reference types define custom reference fields for developers.

Commands:
  list  List developer reference types`,
}

func init() {
	viewCmd.AddCommand(developerReferenceTypesCmd)
}
