package cli

import "github.com/spf13/cobra"

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "Browse and view XBE content",
	Long: `Browse and view XBE content.

The view command provides read-only access to XBE platform data including
newsletters, posts, and broker information. All view commands support:

  --json       Output in JSON format for programmatic use
  --no-auth    Access public content without authentication
  --limit      Control the number of results returned
  --offset     Paginate through large result sets

Content Types:
  action-items           Work items (tasks, bugs, features) with status tracking
  material-transactions  Material movement records (loads, weights, timing)
  job-production-plans   Job production plans (daily work schedules)
  memberships            User-organization relationships and roles
  newsletters            Published market newsletters with analysis and insights
  posts                  Status updates, announcements, and various post types
  brokers                Broker/branch information and metadata
  users                  Platform users (for looking up creator IDs)
  transport-orders       Transport orders (basic list)
  material-suppliers     Material supplier companies
  customers              Customer companies
  truckers               Trucking companies
  features               Product features and capabilities
  release-notes          Product release notes and updates
  press-releases         Official press releases and announcements
  glossary-terms         Industry and product terminology definitions`,
	Example: `  # Browse action items
  xbe view action-items list
  xbe view action-items list --status in_progress --kind feature

  # Browse material transactions
  xbe view material-transactions list --date 2025-01-18
  xbe view material-transactions show 12345

  # Browse job production plans
  xbe view job-production-plans list --start-on 2025-01-18

  # Browse memberships
  xbe view memberships list --broker 123
  xbe view memberships list --q "John"

  # Browse newsletters
  xbe view newsletters list
  xbe view newsletters show 123

  # Browse posts
  xbe view posts list
  xbe view posts show 456

  # Browse brokers
  xbe view brokers list

  # Look up creator IDs for post filtering
  xbe view users list --name "John"
  xbe view material-suppliers list --name "Acme"
  xbe view customers list --name "Smith"
  xbe view truckers list --name "Express"

  # Browse product information
  xbe view features list
  xbe view release-notes list
  xbe view press-releases list
  xbe view glossary-terms list`,
	Annotations: map[string]string{"group": GroupCore},
}

func init() {
	rootCmd.AddCommand(viewCmd)
}
