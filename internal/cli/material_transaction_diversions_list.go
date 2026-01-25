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

type materialTransactionDiversionsListOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	NoAuth              bool
	Limit               int
	Offset              int
	Sort                string
	MaterialTransaction string
	BrokerID            string
	Broker              string
	CreatedBy           string
	NewDeliveryDate     string
	NewDeliveryDateMin  string
	NewDeliveryDateMax  string
	HasNewDeliveryDate  string
}

type materialTransactionDiversionRow struct {
	ID                              string `json:"id"`
	MaterialTransactionID           string `json:"material_transaction_id,omitempty"`
	MaterialTransactionTicketNumber string `json:"material_transaction_ticket_number,omitempty"`
	NewJobSiteID                    string `json:"new_job_site_id,omitempty"`
	NewJobSiteName                  string `json:"new_job_site_name,omitempty"`
	NewDeliveryDate                 string `json:"new_delivery_date,omitempty"`
	DivertedTonsExplicit            string `json:"diverted_tons_explicit,omitempty"`
	DivertedTons                    string `json:"diverted_tons,omitempty"`
	BrokerID                        string `json:"broker_id,omitempty"`
	BrokerName                      string `json:"broker_name,omitempty"`
	CreatedByID                     string `json:"created_by_id,omitempty"`
	CreatedByName                   string `json:"created_by_name,omitempty"`
}

func newMaterialTransactionDiversionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List material transaction diversions",
		Long: `List material transaction diversions.

Output Columns:
  ID         Diversion identifier
  MTXN       Material transaction ID
  TICKET     Material transaction ticket number
  NEW SITE   New job site name
  NEW DATE   New delivery date
  TONS       Diverted tons
  BROKER     Broker name
  CREATED BY Creator name

Filters:
  --material-transaction   Filter by material transaction ID
  --broker-id              Filter by broker ID (joined)
  --broker                 Filter by broker ID
  --created-by             Filter by creator user ID
  --new-delivery-date      Filter by new delivery date (YYYY-MM-DD)
  --new-delivery-date-min  Filter by minimum new delivery date (YYYY-MM-DD)
  --new-delivery-date-max  Filter by maximum new delivery date (YYYY-MM-DD)
  --has-new-delivery-date  Filter by presence of new delivery date (true/false)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List diversions
  xbe view material-transaction-diversions list

  # Filter by material transaction
  xbe view material-transaction-diversions list --material-transaction 123

  # Filter by broker
  xbe view material-transaction-diversions list --broker-id 456

  # Filter by new delivery date range
  xbe view material-transaction-diversions list --new-delivery-date-min 2025-01-01 --new-delivery-date-max 2025-01-31

  # Output as JSON
  xbe view material-transaction-diversions list --json`,
		Args: cobra.NoArgs,
		RunE: runMaterialTransactionDiversionsList,
	}
	initMaterialTransactionDiversionsListFlags(cmd)
	return cmd
}

func init() {
	materialTransactionDiversionsCmd.AddCommand(newMaterialTransactionDiversionsListCmd())
}

func initMaterialTransactionDiversionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("material-transaction", "", "Filter by material transaction ID")
	cmd.Flags().String("broker-id", "", "Filter by broker ID (joined)")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("created-by", "", "Filter by creator user ID")
	cmd.Flags().String("new-delivery-date", "", "Filter by new delivery date (YYYY-MM-DD)")
	cmd.Flags().String("new-delivery-date-min", "", "Filter by minimum new delivery date (YYYY-MM-DD)")
	cmd.Flags().String("new-delivery-date-max", "", "Filter by maximum new delivery date (YYYY-MM-DD)")
	cmd.Flags().String("has-new-delivery-date", "", "Filter by presence of new delivery date (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialTransactionDiversionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMaterialTransactionDiversionsListOptions(cmd)
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
	query.Set("fields[material-transaction-diversions]", "material-transaction,new-job-site,new-delivery-date,diverted-tons-explicit,diverted-tons,created-by,broker")
	query.Set("include", "material-transaction,new-job-site,created-by,broker")
	query.Set("fields[material-transactions]", "ticket-number")
	query.Set("fields[job-sites]", "name")
	query.Set("fields[users]", "name")
	query.Set("fields[brokers]", "company-name")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[material-transaction]", opts.MaterialTransaction)
	setFilterIfPresent(query, "filter[broker-id]", opts.BrokerID)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[created-by]", opts.CreatedBy)
	setFilterIfPresent(query, "filter[new-delivery-date]", opts.NewDeliveryDate)
	setFilterIfPresent(query, "filter[new-delivery-date-min]", opts.NewDeliveryDateMin)
	setFilterIfPresent(query, "filter[new-delivery-date-max]", opts.NewDeliveryDateMax)
	setFilterIfPresent(query, "filter[has-new-delivery-date]", opts.HasNewDeliveryDate)

	body, _, err := client.Get(cmd.Context(), "/v1/material-transaction-diversions", query)
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

	rows := buildMaterialTransactionDiversionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderMaterialTransactionDiversionsTable(cmd, rows)
}

func parseMaterialTransactionDiversionsListOptions(cmd *cobra.Command) (materialTransactionDiversionsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	materialTransaction, _ := cmd.Flags().GetString("material-transaction")
	brokerID, _ := cmd.Flags().GetString("broker-id")
	broker, _ := cmd.Flags().GetString("broker")
	createdBy, _ := cmd.Flags().GetString("created-by")
	newDeliveryDate, _ := cmd.Flags().GetString("new-delivery-date")
	newDeliveryDateMin, _ := cmd.Flags().GetString("new-delivery-date-min")
	newDeliveryDateMax, _ := cmd.Flags().GetString("new-delivery-date-max")
	hasNewDeliveryDate, _ := cmd.Flags().GetString("has-new-delivery-date")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialTransactionDiversionsListOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		NoAuth:              noAuth,
		Limit:               limit,
		Offset:              offset,
		Sort:                sort,
		MaterialTransaction: materialTransaction,
		BrokerID:            brokerID,
		Broker:              broker,
		CreatedBy:           createdBy,
		NewDeliveryDate:     newDeliveryDate,
		NewDeliveryDateMin:  newDeliveryDateMin,
		NewDeliveryDateMax:  newDeliveryDateMax,
		HasNewDeliveryDate:  hasNewDeliveryDate,
	}, nil
}

func buildMaterialTransactionDiversionRows(resp jsonAPIResponse) []materialTransactionDiversionRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]materialTransactionDiversionRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildMaterialTransactionDiversionRow(resource, included))
	}
	return rows
}

func materialTransactionDiversionRowFromSingle(resp jsonAPISingleResponse) materialTransactionDiversionRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}
	return buildMaterialTransactionDiversionRow(resp.Data, included)
}

func buildMaterialTransactionDiversionRow(resource jsonAPIResource, included map[string]jsonAPIResource) materialTransactionDiversionRow {
	attrs := resource.Attributes
	row := materialTransactionDiversionRow{
		ID:                   resource.ID,
		NewDeliveryDate:      formatDate(stringAttr(attrs, "new-delivery-date")),
		DivertedTonsExplicit: stringAttr(attrs, "diverted-tons-explicit"),
		DivertedTons:         stringAttr(attrs, "diverted-tons"),
	}

	if rel, ok := resource.Relationships["material-transaction"]; ok && rel.Data != nil {
		row.MaterialTransactionID = rel.Data.ID
		if mtxn, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.MaterialTransactionTicketNumber = stringAttr(mtxn.Attributes, "ticket-number")
		}
	}

	if rel, ok := resource.Relationships["new-job-site"]; ok && rel.Data != nil {
		row.NewJobSiteID = rel.Data.ID
		if site, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.NewJobSiteName = stringAttr(site.Attributes, "name")
		}
	}

	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
		if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.BrokerName = firstNonEmpty(
				stringAttr(broker.Attributes, "company-name"),
				stringAttr(broker.Attributes, "name"),
			)
		}
	}

	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		row.CreatedByID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.CreatedByName = stringAttr(user.Attributes, "name")
		}
	}

	return row
}

func renderMaterialTransactionDiversionsTable(cmd *cobra.Command, rows []materialTransactionDiversionRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No material transaction diversions found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tMTXN\tTICKET\tNEW SITE\tNEW DATE\tTONS\tBROKER\tCREATED BY")
	for _, row := range rows {
		newSite := firstNonEmpty(row.NewJobSiteName, row.NewJobSiteID)
		broker := firstNonEmpty(row.BrokerName, row.BrokerID)
		createdBy := firstNonEmpty(row.CreatedByName, row.CreatedByID)
		tons := firstNonEmpty(row.DivertedTons, row.DivertedTonsExplicit)

		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.MaterialTransactionID,
			truncateString(row.MaterialTransactionTicketNumber, 15),
			truncateString(newSite, 25),
			row.NewDeliveryDate,
			tons,
			truncateString(broker, 20),
			truncateString(createdBy, 20),
		)
	}
	return writer.Flush()
}
