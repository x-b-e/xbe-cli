package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type searchCatalogEntriesListOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	NoAuth      bool
	Limit       int
	Offset      int
	Sort        string
	EntityType  string
	EntityID    string
	Broker      string
	Customer    string
	Trucker     string
	Search      string
	FuzzySearch string
}

type searchCatalogEntryRow struct {
	ID           string `json:"id"`
	EntityType   string `json:"entity_type,omitempty"`
	EntityID     string `json:"entity_id,omitempty"`
	DisplayText  string `json:"display_text,omitempty"`
	BrokerID     string `json:"broker_id,omitempty"`
	BrokerName   string `json:"broker_name,omitempty"`
	CustomerID   string `json:"customer_id,omitempty"`
	CustomerName string `json:"customer_name,omitempty"`
	TruckerID    string `json:"trucker_id,omitempty"`
	TruckerName  string `json:"trucker_name,omitempty"`
}

func newSearchCatalogEntriesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List search catalog entries",
		Long: `List search catalog entries with filtering and pagination.

Output Columns:
  ID           Search catalog entry identifier
  ENTITY TYPE  Entity type for the indexed record
  ENTITY ID    Entity ID for the indexed record
  DISPLAY      Display text used for search results
  BROKER       Broker name (falls back to ID)
  CUSTOMER     Customer name (falls back to ID)
  TRUCKER      Trucker name (falls back to ID)

Filters:
  --entity-type    Filter by entity type (e.g., Customer, Trucker)
  --entity-id      Filter by entity ID
  --broker         Filter by broker ID
  --customer       Filter by customer ID
  --trucker        Filter by trucker ID
  --search         Full-text search for complete words
  --fuzzy-search   Fuzzy search for partial words and typos

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List search catalog entries
  xbe view search-catalog-entries list

  # Full-text search
  xbe view search-catalog-entries list --search "john smith"

  # Fuzzy search
  xbe view search-catalog-entries list --fuzzy-search "joh"

  # Filter by entity type and broker
  xbe view search-catalog-entries list --entity-type Customer --broker 123

  # JSON output
  xbe view search-catalog-entries list --json`,
		Args: cobra.NoArgs,
		RunE: runSearchCatalogEntriesList,
	}
	initSearchCatalogEntriesListFlags(cmd)
	return cmd
}

func init() {
	searchCatalogEntriesCmd.AddCommand(newSearchCatalogEntriesListCmd())
}

func initSearchCatalogEntriesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("entity-type", "", "Filter by entity type")
	cmd.Flags().String("entity-id", "", "Filter by entity ID")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("customer", "", "Filter by customer ID")
	cmd.Flags().String("trucker", "", "Filter by trucker ID")
	cmd.Flags().String("search", "", "Full-text search (complete words)")
	cmd.Flags().String("fuzzy-search", "", "Fuzzy search (partial words, typos)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runSearchCatalogEntriesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseSearchCatalogEntriesListOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[search-catalog-entries]", "entity-id,entity-type,display-text,broker,customer,trucker")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[truckers]", "company-name")
	query.Set("include", "broker,customer,trucker")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[entity-type]", opts.EntityType)
	setFilterIfPresent(query, "filter[entity-id]", opts.EntityID)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[customer]", opts.Customer)
	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)
	setFilterIfPresent(query, "filter[search]", opts.Search)
	setFilterIfPresent(query, "filter[fuzzy-search]", opts.FuzzySearch)

	body, _, err := client.Get(cmd.Context(), "/v1/search-catalog-entries", query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	handled, err := renderSparseListIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	rows := buildSearchCatalogEntryRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderSearchCatalogEntriesTable(cmd, rows)
}

func parseSearchCatalogEntriesListOptions(cmd *cobra.Command) (searchCatalogEntriesListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return searchCatalogEntriesListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return searchCatalogEntriesListOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return searchCatalogEntriesListOptions{}, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return searchCatalogEntriesListOptions{}, err
	}
	sort, err := cmd.Flags().GetString("sort")
	if err != nil {
		return searchCatalogEntriesListOptions{}, err
	}
	entityType, err := cmd.Flags().GetString("entity-type")
	if err != nil {
		return searchCatalogEntriesListOptions{}, err
	}
	entityID, err := cmd.Flags().GetString("entity-id")
	if err != nil {
		return searchCatalogEntriesListOptions{}, err
	}
	broker, err := cmd.Flags().GetString("broker")
	if err != nil {
		return searchCatalogEntriesListOptions{}, err
	}
	customer, err := cmd.Flags().GetString("customer")
	if err != nil {
		return searchCatalogEntriesListOptions{}, err
	}
	trucker, err := cmd.Flags().GetString("trucker")
	if err != nil {
		return searchCatalogEntriesListOptions{}, err
	}
	search, err := cmd.Flags().GetString("search")
	if err != nil {
		return searchCatalogEntriesListOptions{}, err
	}
	fuzzySearch, err := cmd.Flags().GetString("fuzzy-search")
	if err != nil {
		return searchCatalogEntriesListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return searchCatalogEntriesListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return searchCatalogEntriesListOptions{}, err
	}

	return searchCatalogEntriesListOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		NoAuth:      noAuth,
		Limit:       limit,
		Offset:      offset,
		Sort:        sort,
		EntityType:  entityType,
		EntityID:    entityID,
		Broker:      broker,
		Customer:    customer,
		Trucker:     trucker,
		Search:      search,
		FuzzySearch: fuzzySearch,
	}, nil
}

func buildSearchCatalogEntryRows(resp jsonAPIResponse) []searchCatalogEntryRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]searchCatalogEntryRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := searchCatalogEntryRow{
			ID:          resource.ID,
			EntityType:  strings.TrimSpace(stringAttr(resource.Attributes, "entity-type")),
			EntityID:    strings.TrimSpace(stringAttr(resource.Attributes, "entity-id")),
			DisplayText: strings.TrimSpace(stringAttr(resource.Attributes, "display-text")),
		}

		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			row.BrokerID = rel.Data.ID
			if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.BrokerName = strings.TrimSpace(stringAttr(broker.Attributes, "company-name"))
			}
		}
		if rel, ok := resource.Relationships["customer"]; ok && rel.Data != nil {
			row.CustomerID = rel.Data.ID
			if customer, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.CustomerName = strings.TrimSpace(stringAttr(customer.Attributes, "company-name"))
			}
		}
		if rel, ok := resource.Relationships["trucker"]; ok && rel.Data != nil {
			row.TruckerID = rel.Data.ID
			if trucker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.TruckerName = strings.TrimSpace(stringAttr(trucker.Attributes, "company-name"))
			}
		}

		rows = append(rows, row)
	}

	return rows
}

func renderSearchCatalogEntriesTable(cmd *cobra.Command, rows []searchCatalogEntryRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No search catalog entries found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tENTITY TYPE\tENTITY ID\tDISPLAY\tBROKER\tCUSTOMER\tTRUCKER")
	for _, row := range rows {
		brokerDisplay := firstNonEmpty(row.BrokerName, row.BrokerID)
		customerDisplay := firstNonEmpty(row.CustomerName, row.CustomerID)
		truckerDisplay := firstNonEmpty(row.TruckerName, row.TruckerID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.EntityType,
			row.EntityID,
			truncateString(row.DisplayText, 50),
			truncateString(brokerDisplay, 30),
			truncateString(customerDisplay, 30),
			truncateString(truckerDisplay, 30),
		)
	}
	return writer.Flush()
}
