package cli

import "github.com/spf13/cobra"

var doTractorTrailerCredentialClassificationsCmd = &cobra.Command{
	Use:   "tractor-trailer-credential-classifications",
	Short: "Manage tractor/trailer credential classifications",
	Long: `Create, update, and delete tractor and trailer credential classifications.

These classifications define types of credentials that can be assigned to tractors and trailers.

Commands:
  create  Create a new tractor/trailer credential classification
  update  Update an existing tractor/trailer credential classification
  delete  Delete a tractor/trailer credential classification`,
	Example: `  # Create a tractor/trailer credential classification
  xbe do tractor-trailer-credential-classifications create --name "Insurance" --organization-type brokers --organization-id 123

  # Update a tractor/trailer credential classification
  xbe do tractor-trailer-credential-classifications update 456 --name "Updated Name"

  # Delete a tractor/trailer credential classification
  xbe do tractor-trailer-credential-classifications delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doTractorTrailerCredentialClassificationsCmd)
}
