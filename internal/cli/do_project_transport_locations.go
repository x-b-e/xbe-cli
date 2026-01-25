package cli

import "github.com/spf13/cobra"

var doProjectTransportLocationsCmd = &cobra.Command{
	Use:   "project-transport-locations",
	Short: "Manage project transport locations",
	Long: `Create, update, and delete project transport locations.

Project transport locations represent pickup, delivery, and staging locations
used in transport planning.

Commands:
  create  Create a new project transport location
  update  Update an existing project transport location
  delete  Delete a project transport location`,
	Example: `  # Create a project transport location
  xbe do project-transport-locations create --name "North Yard" --geocoding-method explicit --broker 123

  # Update a project transport location
  xbe do project-transport-locations update 456 --is-active=false

  # Delete a project transport location
  xbe do project-transport-locations delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doProjectTransportLocationsCmd)
}
