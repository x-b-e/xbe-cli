package cli

import "github.com/spf13/cobra"

var searchCatalogEntriesCmd = &cobra.Command{
	Use:   "search-catalog-entries",
	Short: "Browse search catalog entries",
	Long: `Browse search catalog entries used by the XBE search catalog.

Search catalog entries index brokers, customers, truckers, and other entities
for search and autocomplete. Each entry stores an entity type/id and display
text used for search results.

Commands:
  list    List search catalog entries with filtering and pagination
  show    View the full details of a search catalog entry`,
	Example: `  # Search by full text
  xbe view search-catalog-entries list --search "john smith"

  # Fuzzy search for partial text
  xbe view search-catalog-entries list --fuzzy-search "john"

  # Filter by entity type
  xbe view search-catalog-entries list --entity-type Customer

  # View a specific entry
  xbe view search-catalog-entries show 123`,
}

func init() {
	viewCmd.AddCommand(searchCatalogEntriesCmd)
}
