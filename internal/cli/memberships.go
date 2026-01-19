package cli

import "github.com/spf13/cobra"

var membershipsCmd = &cobra.Command{
	Use:   "memberships",
	Short: "Browse and view memberships",
	Long: `Browse and view memberships on the XBE platform.

Memberships represent the relationship between users and organizations.
A user can have memberships in multiple organizations (brokers, customers,
truckers, material suppliers, developers).

Each membership defines:
  - The user's role (operations or manager)
  - Admin privileges within the organization
  - Various notification and access settings

Commands:
  list    List memberships with filtering and pagination
  show    Show full details of a specific membership`,
	Example: `  # List memberships
  xbe view memberships list

  # Filter by broker
  xbe view memberships list --broker 123

  # Filter by user
  xbe view memberships list --user 456

  # Search by user name
  xbe view memberships list --q "John"

  # Filter by kind (operations or manager)
  xbe view memberships list --kind manager

  # Get results as JSON
  xbe view memberships list --json`,
}

func init() {
	viewCmd.AddCommand(membershipsCmd)
}
