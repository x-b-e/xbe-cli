package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type searchCatalogEntriesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type searchCatalogEntryDetails struct {
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

func newSearchCatalogEntriesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show search catalog entry details",
		Long: `Show the full details of a search catalog entry.

Output Fields:
  ID           Search catalog entry identifier
  Entity Type  Entity type for the indexed record
  Entity ID    Entity ID for the indexed record
  Display Text Display text used for search results
  Broker       Broker name (falls back to ID)
  Customer     Customer name (falls back to ID)
  Trucker      Trucker name (falls back to ID)

Arguments:
  <id>    Search catalog entry ID (required)

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a search catalog entry
  xbe view search-catalog-entries show 123

  # JSON output
  xbe view search-catalog-entries show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runSearchCatalogEntriesShow,
	}
	initSearchCatalogEntriesShowFlags(cmd)
	return cmd
}

func init() {
	searchCatalogEntriesCmd.AddCommand(newSearchCatalogEntriesShowCmd())
}

func initSearchCatalogEntriesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runSearchCatalogEntriesShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseSearchCatalogEntriesShowOptions(cmd)
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("search catalog entry id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[search-catalog-entries]", "entity-id,entity-type,display-text,broker,customer,trucker")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[truckers]", "company-name")
	query.Set("include", "broker,customer,trucker")

	body, _, err := client.Get(cmd.Context(), "/v1/search-catalog-entries/"+id, query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPISingleResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildSearchCatalogEntryDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderSearchCatalogEntryDetails(cmd, details)
}

func parseSearchCatalogEntriesShowOptions(cmd *cobra.Command) (searchCatalogEntriesShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return searchCatalogEntriesShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return searchCatalogEntriesShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return searchCatalogEntriesShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return searchCatalogEntriesShowOptions{}, err
	}

	return searchCatalogEntriesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildSearchCatalogEntryDetails(resp jsonAPISingleResponse) searchCatalogEntryDetails {
	details := searchCatalogEntryDetails{
		ID:          resp.Data.ID,
		EntityType:  strings.TrimSpace(stringAttr(resp.Data.Attributes, "entity-type")),
		EntityID:    strings.TrimSpace(stringAttr(resp.Data.Attributes, "entity-id")),
		DisplayText: strings.TrimSpace(stringAttr(resp.Data.Attributes, "display-text")),
	}

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
		if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.BrokerName = strings.TrimSpace(stringAttr(broker.Attributes, "company-name"))
		}
	}
	if rel, ok := resp.Data.Relationships["customer"]; ok && rel.Data != nil {
		details.CustomerID = rel.Data.ID
		if customer, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.CustomerName = strings.TrimSpace(stringAttr(customer.Attributes, "company-name"))
		}
	}
	if rel, ok := resp.Data.Relationships["trucker"]; ok && rel.Data != nil {
		details.TruckerID = rel.Data.ID
		if trucker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.TruckerName = strings.TrimSpace(stringAttr(trucker.Attributes, "company-name"))
		}
	}

	return details
}

func renderSearchCatalogEntryDetails(cmd *cobra.Command, details searchCatalogEntryDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	fmt.Fprintf(out, "Entity Type: %s\n", formatOptional(details.EntityType))
	fmt.Fprintf(out, "Entity ID: %s\n", formatOptional(details.EntityID))
	fmt.Fprintf(out, "Display Text: %s\n", formatOptional(details.DisplayText))

	if details.BrokerID != "" || details.BrokerName != "" {
		label := details.BrokerID
		if details.BrokerName != "" {
			label = fmt.Sprintf("%s (%s)", details.BrokerName, details.BrokerID)
		}
		fmt.Fprintf(out, "Broker: %s\n", formatOptional(label))
	}
	if details.CustomerID != "" || details.CustomerName != "" {
		label := details.CustomerID
		if details.CustomerName != "" {
			label = fmt.Sprintf("%s (%s)", details.CustomerName, details.CustomerID)
		}
		fmt.Fprintf(out, "Customer: %s\n", formatOptional(label))
	}
	if details.TruckerID != "" || details.TruckerName != "" {
		label := details.TruckerID
		if details.TruckerName != "" {
			label = fmt.Sprintf("%s (%s)", details.TruckerName, details.TruckerID)
		}
		fmt.Fprintf(out, "Trucker: %s\n", formatOptional(label))
	}

	return nil
}
