package cli

import "github.com/spf13/cobra"

var followsCmd = &cobra.Command{
	Use:     "follows",
	Aliases: []string{"follow"},
	Short:   "View follows",
	Long: `View follow relationships between users and creators.

Follows represent users following creators such as users, projects, brokers,
and other entities that can publish posts.

Commands:
  list    List follows
  show    Show follow details`,
}

func init() {
	viewCmd.AddCommand(followsCmd)
}
