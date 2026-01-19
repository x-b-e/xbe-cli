package cli

import "github.com/spf13/cobra"

var doCmd = &cobra.Command{
	Use:   "do",
	Short: "Create, update, and delete XBE resources",
	Long: `Create, update, and delete XBE resources.

The do command provides write access to XBE platform data. Unlike view commands,
these operations modify data and require authentication.

Resources:
  glossary-terms                   Manage glossary term definitions
  lane-summary                     Generate lane (cycle) summaries
  material-transaction-summary     Generate material transaction summaries
  memberships                      Manage user-organization memberships`,
	Example: `  # Update a glossary term
  xbe do glossary-terms update 123 --definition "New definition"

  # Delete a glossary term
  xbe do glossary-terms delete 123 --confirm

  # Generate a lane summary by origin/destination
  xbe do lane-summary create --group-by origin,destination --filter broker=123

  # Generate a material transaction summary by material site
  xbe do material-transaction-summary create --group-by material_site --filter broker=123

  # Create a membership
  xbe do memberships create --user 123 --organization Broker|4 --kind manager

  # Update a membership
  xbe do memberships update 686 --kind operations

  # Delete a membership
  xbe do memberships delete 686 --confirm`,
	Annotations: map[string]string{"group": GroupCore},
}

func init() {
	rootCmd.AddCommand(doCmd)
}
