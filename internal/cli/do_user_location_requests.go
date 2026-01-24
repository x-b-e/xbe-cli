package cli

import "github.com/spf13/cobra"

var doUserLocationRequestsCmd = &cobra.Command{
	Use:     "user-location-requests",
	Aliases: []string{"user-location-request"},
	Short:   "Manage user location requests",
	Long:    "Commands for creating user location requests.",
	Example: `  # Create a user location request
  xbe do user-location-requests create --user 123
`,
}

func init() {
	doCmd.AddCommand(doUserLocationRequestsCmd)
}
