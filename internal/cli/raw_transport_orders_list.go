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

type rawTransportOrdersListOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	NoAuth              bool
	Limit               int
	Offset              int
	Broker              string
	TablesRowversionMin string
	TablesRowversionMax string
}

type rawTransportOrderRow struct {
	ID                  string `json:"id"`
	ExternalOrderNumber string `json:"external_order_number,omitempty"`
	Importer            string `json:"importer,omitempty"`
	ImportStatus        string `json:"import_status,omitempty"`
	IsManaged           bool   `json:"is_managed,omitempty"`
	TablesRowversionMin string `json:"tables_rowversion_min,omitempty"`
	TablesRowversionMax string `json:"tables_rowversion_max,omitempty"`
	BrokerID            string `json:"broker_id,omitempty"`
	TransportOrderID    string `json:"transport_order_id,omitempty"`
}

func newRawTransportOrdersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List raw transport orders",
		Long: `List raw transport orders with filtering and pagination.

Raw transport orders hold imported order payloads before they are normalized
into transport orders. Use filters to narrow by broker or rowversion range.

Output Columns:
  ID       Raw transport order ID
  ORDER    External order number
  STATUS   Import status
  IMPORTER Importer key
  MANAGED  Managed flag
  BROKER   Broker ID
  TO       Linked transport order ID

Filters:
  --broker                Filter by broker ID (comma-separated for multiple)
  --tables-rowversion-min Filter by minimum tables rowversion
  --tables-rowversion-max Filter by maximum tables rowversion

Global flags (see xbe --help): --json, --limit, --offset, --base-url, --token, --no-auth`,
		Example: `  # List raw transport orders
  xbe view raw-transport-orders list

  # Filter by broker
  xbe view raw-transport-orders list --broker 123

  # Filter by rowversion range
  xbe view raw-transport-orders list --tables-rowversion-min 100 --tables-rowversion-max 200

  # JSON output
  xbe view raw-transport-orders list --json`,
		RunE: runRawTransportOrdersList,
	}
	initRawTransportOrdersListFlags(cmd)
	return cmd
}

func init() {
	rawTransportOrdersCmd.AddCommand(newRawTransportOrdersListCmd())
}

func initRawTransportOrdersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 0, "Page size (defaults to server default)")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("broker", "", "Filter by broker ID (comma-separated for multiple)")
	cmd.Flags().String("tables-rowversion-min", "", "Filter by minimum tables rowversion")
	cmd.Flags().String("tables-rowversion-max", "", "Filter by maximum tables rowversion")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runRawTransportOrdersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseRawTransportOrdersListOptions(cmd)
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
	query.Set("fields[raw-transport-orders]", "external-order-number,importer,import-status,is-managed,tables-rowversion-min,tables-rowversion-max,broker,transport-order")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[tables_rowversion_min]", opts.TablesRowversionMin)
	setFilterIfPresent(query, "filter[tables_rowversion_max]", opts.TablesRowversionMax)

	body, _, err := client.Get(cmd.Context(), "/v1/raw-transport-orders", query)
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

	rows := buildRawTransportOrderRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderRawTransportOrdersTable(cmd, rows)
}

func parseRawTransportOrdersListOptions(cmd *cobra.Command) (rawTransportOrdersListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return rawTransportOrdersListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return rawTransportOrdersListOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return rawTransportOrdersListOptions{}, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return rawTransportOrdersListOptions{}, err
	}
	broker, err := cmd.Flags().GetString("broker")
	if err != nil {
		return rawTransportOrdersListOptions{}, err
	}
	rowversionMin, err := cmd.Flags().GetString("tables-rowversion-min")
	if err != nil {
		return rawTransportOrdersListOptions{}, err
	}
	rowversionMax, err := cmd.Flags().GetString("tables-rowversion-max")
	if err != nil {
		return rawTransportOrdersListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return rawTransportOrdersListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return rawTransportOrdersListOptions{}, err
	}

	return rawTransportOrdersListOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		NoAuth:              noAuth,
		Limit:               limit,
		Offset:              offset,
		Broker:              broker,
		TablesRowversionMin: rowversionMin,
		TablesRowversionMax: rowversionMax,
	}, nil
}

func buildRawTransportOrderRows(resp jsonAPIResponse) []rawTransportOrderRow {
	rows := make([]rawTransportOrderRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildRawTransportOrderRow(resource))
	}
	return rows
}

func buildRawTransportOrderRow(resource jsonAPIResource) rawTransportOrderRow {
	attrs := resource.Attributes
	row := rawTransportOrderRow{
		ID:                  resource.ID,
		ExternalOrderNumber: stringAttr(attrs, "external-order-number"),
		Importer:            stringAttr(attrs, "importer"),
		ImportStatus:        stringAttr(attrs, "import-status"),
		IsManaged:           boolAttr(attrs, "is-managed"),
		TablesRowversionMin: stringAttr(attrs, "tables-rowversion-min"),
		TablesRowversionMax: stringAttr(attrs, "tables-rowversion-max"),
	}

	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["transport-order"]; ok && rel.Data != nil {
		row.TransportOrderID = rel.Data.ID
	}

	return row
}

func renderRawTransportOrdersTable(cmd *cobra.Command, rows []rawTransportOrderRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No raw transport orders found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tORDER\tSTATUS\tIMPORTER\tMANAGED\tBROKER\tTO")
	for _, row := range rows {
		managed := ""
		if row.IsManaged {
			managed = "Y"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.ExternalOrderNumber, 18),
			truncateString(row.ImportStatus, 10),
			truncateString(row.Importer, 12),
			managed,
			truncateString(row.BrokerID, 10),
			truncateString(row.TransportOrderID, 10),
		)
	}
	return writer.Flush()
}
