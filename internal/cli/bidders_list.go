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

type biddersListOptions struct {
	BaseURL         string
	Token           string
	JSON            bool
	NoAuth          bool
	Limit           int
	Offset          int
	Broker          string
	Name            string
	NameLike        string
	IsSelfForBroker string
}

func newBiddersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List bidders",
		Long: `List bidders with filtering and pagination.

Bidders represent entities that participate in broker bidding workflows.

Output Columns:
  ID      Bidder identifier
  NAME    Bidder name
  SELF    Whether the bidder is the broker's self bidder
  BROKER  Broker organization

Filters:
  --broker              Filter by broker ID
  --name                Filter by name (exact match)
  --name-like           Filter by name (partial match)
  --is-self-for-broker  Filter by self bidder status (true/false)`,
		Example: `  # List all bidders
  xbe view bidders list

  # Filter by broker
  xbe view bidders list --broker 123

  # Search by name
  xbe view bidders list --name "Acme"

  # Search by partial name
  xbe view bidders list --name-like "Ac"

  # Filter by self bidder status
  xbe view bidders list --is-self-for-broker true

  # Output as JSON
  xbe view bidders list --json`,
		RunE: runBiddersList,
	}
	initBiddersListFlags(cmd)
	return cmd
}

func init() {
	biddersCmd.AddCommand(newBiddersListCmd())
}

func initBiddersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("name", "", "Filter by name")
	cmd.Flags().String("name-like", "", "Filter by name (partial match)")
	cmd.Flags().String("is-self-for-broker", "", "Filter by self bidder status (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runBiddersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseBiddersListOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run 'xbe auth login' first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("sort", "name")
	query.Set("fields[bidders]", "name,is-self-for-broker,broker")
	query.Set("include", "broker")
	query.Set("fields[brokers]", "company-name")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[name]", opts.Name)
	setFilterIfPresent(query, "filter[name-like]", opts.NameLike)
	setFilterIfPresent(query, "filter[is-self-for-broker]", opts.IsSelfForBroker)

	body, _, err := client.Get(cmd.Context(), "/v1/bidders", query)
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

	rows := buildBidderRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderBiddersTable(cmd, rows)
}

func parseBiddersListOptions(cmd *cobra.Command) (biddersListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	broker, _ := cmd.Flags().GetString("broker")
	name, _ := cmd.Flags().GetString("name")
	nameLike, _ := cmd.Flags().GetString("name-like")
	isSelfForBroker, _ := cmd.Flags().GetString("is-self-for-broker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return biddersListOptions{
		BaseURL:         baseURL,
		Token:           token,
		JSON:            jsonOut,
		NoAuth:          noAuth,
		Limit:           limit,
		Offset:          offset,
		Broker:          broker,
		Name:            name,
		NameLike:        nameLike,
		IsSelfForBroker: isSelfForBroker,
	}, nil
}

type bidderRow struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	IsSelfForBroker bool   `json:"is_self_for_broker"`
	BrokerID        string `json:"broker_id,omitempty"`
	BrokerName      string `json:"broker_name,omitempty"`
}

func buildBidderRows(resp jsonAPIResponse) []bidderRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	rows := make([]bidderRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := bidderRow{
			ID:              resource.ID,
			Name:            stringAttr(resource.Attributes, "name"),
			IsSelfForBroker: boolAttr(resource.Attributes, "is-self-for-broker"),
		}

		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			row.BrokerID = rel.Data.ID
			if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.BrokerName = stringAttr(broker.Attributes, "company-name")
			}
		}

		rows = append(rows, row)
	}
	return rows
}

func renderBiddersTable(cmd *cobra.Command, rows []bidderRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No bidders found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tSELF\tBROKER")
	for _, row := range rows {
		self := "no"
		if row.IsSelfForBroker {
			self = "yes"
		}

		broker := row.BrokerName
		if broker == "" && row.BrokerID != "" {
			broker = row.BrokerID
		}

		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 30),
			self,
			truncateString(broker, 30),
		)
	}
	return writer.Flush()
}
