package cli

import "github.com/spf13/cobra"

var doProjectBidLocationsCmd = &cobra.Command{
	Use:     "project-bid-locations",
	Aliases: []string{"project-bid-location"},
	Short:   "Manage project bid locations",
	Long: `Create, update, and delete project bid locations.

Project bid locations define geometry-based locations associated with
projects and their prediction subjects.

Commands:
  create    Create a project bid location
  update    Update a project bid location
  delete    Delete a project bid location`,
	Example: `  # Create a project bid location
  xbe do project-bid-locations create --project 123 --geometry "POINT(-77.0365 38.8977)"

  # Update a project bid location
  xbe do project-bid-locations update 456 --name "Updated Location"

  # Delete a project bid location
  xbe do project-bid-locations delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doProjectBidLocationsCmd)
}
