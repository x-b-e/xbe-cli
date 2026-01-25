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

type rawTransportExportsListOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	NoAuth              bool
	Limit               int
	Offset              int
	Sort                string
	Broker              string
	TransportOrder      string
	ExternalOrderNumber string
	ExportType          string
	TargetTable         string
	IssueType           string
	IsExportable        string
	IsExported          string
	Checksum            string
	Sequence            string
	CreatedAtMin        string
	CreatedAtMax        string
	IsCreatedAt         string
	ExportedAtMin       string
	ExportedAtMax       string
	IsExportedAt        string
	FirstSeenAtMin      string
	FirstSeenAtMax      string
	IsFirstSeenAt       string
	Recent              string
}

type rawTransportExportRow struct {
	ID                   string `json:"id"`
	ExternalOrderNumber  string `json:"external_order_number,omitempty"`
	ExportType           string `json:"export_type,omitempty"`
	TargetDatabase       string `json:"target_database,omitempty"`
	TargetTable          string `json:"target_table,omitempty"`
	IsExportable         bool   `json:"is_exportable,omitempty"`
	IsExported           bool   `json:"is_exported,omitempty"`
	IssueType            string `json:"issue_type,omitempty"`
	BrokerID             string `json:"broker_id,omitempty"`
	BrokerName           string `json:"broker_name,omitempty"`
	TransportOrderID     string `json:"transport_order_id,omitempty"`
	TransportOrderNumber string `json:"transport_order_external_number,omitempty"`
}

func newRawTransportExportsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List raw transport exports",
		Long: `List raw transport exports with filtering and pagination.

Output Columns:
  ID          Raw transport export identifier
  ORDER       External order number
  TYPE        Export type
  TARGET      Target database/table
  EXPORTABLE  Exportable flag
  EXPORTED    Exported flag
  ISSUE       Issue type
  BROKER      Broker name or ID

Filters:
  --broker                Filter by broker ID
  --transport-order       Filter by transport order ID
  --external-order-number Filter by external order number
  --export-type           Filter by export type
  --target-table          Filter by target table
  --issue-type            Filter by issue type
  --is-exportable         Filter by exportable flag (true/false)
  --is-exported           Filter by exported flag (true/false)
  --checksum              Filter by checksum
  --sequence              Filter by sequence
  --created-at-min        Filter by created-at on/after (ISO 8601)
  --created-at-max        Filter by created-at on/before (ISO 8601)
  --is-created-at         Filter by has created-at (true/false)
  --exported-at-min       Filter by exported-at on/after (ISO 8601)
  --exported-at-max       Filter by exported-at on/before (ISO 8601)
  --is-exported-at        Filter by has exported-at (true/false)
  --first-seen-at-min     Filter by first-seen-at on/after (ISO 8601)
  --first-seen-at-max     Filter by first-seen-at on/before (ISO 8601)
  --is-first-seen-at      Filter by has first-seen-at (true/false)
  --recent                Filter by records created in the last N hours

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List raw transport exports
  xbe view raw-transport-exports list

  # Filter by broker
  xbe view raw-transport-exports list --broker 123

  # Filter by external order number
  xbe view raw-transport-exports list --external-order-number ORD-456

  # Filter by export status
  xbe view raw-transport-exports list --is-exportable true --is-exported false

  # Output as JSON
  xbe view raw-transport-exports list --json`,
		Args: cobra.NoArgs,
		RunE: runRawTransportExportsList,
	}
	initRawTransportExportsListFlags(cmd)
	return cmd
}

func init() {
	rawTransportExportsCmd.AddCommand(newRawTransportExportsListCmd())
}

func initRawTransportExportsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("transport-order", "", "Filter by transport order ID")
	cmd.Flags().String("external-order-number", "", "Filter by external order number")
	cmd.Flags().String("export-type", "", "Filter by export type")
	cmd.Flags().String("target-table", "", "Filter by target table")
	cmd.Flags().String("issue-type", "", "Filter by issue type")
	cmd.Flags().String("is-exportable", "", "Filter by exportable flag (true/false)")
	cmd.Flags().String("is-exported", "", "Filter by exported flag (true/false)")
	cmd.Flags().String("checksum", "", "Filter by checksum")
	cmd.Flags().String("sequence", "", "Filter by sequence")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by has created-at (true/false)")
	cmd.Flags().String("exported-at-min", "", "Filter by exported-at on/after (ISO 8601)")
	cmd.Flags().String("exported-at-max", "", "Filter by exported-at on/before (ISO 8601)")
	cmd.Flags().String("is-exported-at", "", "Filter by has exported-at (true/false)")
	cmd.Flags().String("first-seen-at-min", "", "Filter by first-seen-at on/after (ISO 8601)")
	cmd.Flags().String("first-seen-at-max", "", "Filter by first-seen-at on/before (ISO 8601)")
	cmd.Flags().String("is-first-seen-at", "", "Filter by has first-seen-at (true/false)")
	cmd.Flags().String("recent", "", "Filter by records created in the last N hours")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runRawTransportExportsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseRawTransportExportsListOptions(cmd)
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
	query.Set("fields[raw-transport-exports]", strings.Join([]string{
		"external-order-number",
		"export-type",
		"target-database",
		"target-table",
		"is-exportable",
		"is-exported",
		"issue-type",
		"broker",
		"transport-order",
	}, ","))
	query.Set("include", "broker,transport-order")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[transport-orders]", "external-order-number")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[transport-order]", opts.TransportOrder)
	setFilterIfPresent(query, "filter[external-order-number]", opts.ExternalOrderNumber)
	setFilterIfPresent(query, "filter[export-type]", opts.ExportType)
	setFilterIfPresent(query, "filter[target-table]", opts.TargetTable)
	setFilterIfPresent(query, "filter[issue-type]", opts.IssueType)
	setFilterIfPresent(query, "filter[is-exportable]", opts.IsExportable)
	setFilterIfPresent(query, "filter[is-exported]", opts.IsExported)
	setFilterIfPresent(query, "filter[checksum]", opts.Checksum)
	setFilterIfPresent(query, "filter[sequence]", opts.Sequence)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[is-created-at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[exported-at-min]", opts.ExportedAtMin)
	setFilterIfPresent(query, "filter[exported-at-max]", opts.ExportedAtMax)
	setFilterIfPresent(query, "filter[is-exported-at]", opts.IsExportedAt)
	setFilterIfPresent(query, "filter[first-seen-at-min]", opts.FirstSeenAtMin)
	setFilterIfPresent(query, "filter[first-seen-at-max]", opts.FirstSeenAtMax)
	setFilterIfPresent(query, "filter[is-first-seen-at]", opts.IsFirstSeenAt)
	setFilterIfPresent(query, "filter[recent]", opts.Recent)

	body, _, err := client.Get(cmd.Context(), "/v1/raw-transport-exports", query)
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

	rows := buildRawTransportExportRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderRawTransportExportsTable(cmd, rows)
}

func parseRawTransportExportsListOptions(cmd *cobra.Command) (rawTransportExportsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	broker, _ := cmd.Flags().GetString("broker")
	transportOrder, _ := cmd.Flags().GetString("transport-order")
	externalOrderNumber, _ := cmd.Flags().GetString("external-order-number")
	exportType, _ := cmd.Flags().GetString("export-type")
	targetTable, _ := cmd.Flags().GetString("target-table")
	issueType, _ := cmd.Flags().GetString("issue-type")
	isExportable, _ := cmd.Flags().GetString("is-exportable")
	isExported, _ := cmd.Flags().GetString("is-exported")
	checksum, _ := cmd.Flags().GetString("checksum")
	sequence, _ := cmd.Flags().GetString("sequence")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	isCreatedAt, _ := cmd.Flags().GetString("is-created-at")
	exportedAtMin, _ := cmd.Flags().GetString("exported-at-min")
	exportedAtMax, _ := cmd.Flags().GetString("exported-at-max")
	isExportedAt, _ := cmd.Flags().GetString("is-exported-at")
	firstSeenAtMin, _ := cmd.Flags().GetString("first-seen-at-min")
	firstSeenAtMax, _ := cmd.Flags().GetString("first-seen-at-max")
	isFirstSeenAt, _ := cmd.Flags().GetString("is-first-seen-at")
	recent, _ := cmd.Flags().GetString("recent")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return rawTransportExportsListOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		NoAuth:              noAuth,
		Limit:               limit,
		Offset:              offset,
		Sort:                sort,
		Broker:              broker,
		TransportOrder:      transportOrder,
		ExternalOrderNumber: externalOrderNumber,
		ExportType:          exportType,
		TargetTable:         targetTable,
		IssueType:           issueType,
		IsExportable:        isExportable,
		IsExported:          isExported,
		Checksum:            checksum,
		Sequence:            sequence,
		CreatedAtMin:        createdAtMin,
		CreatedAtMax:        createdAtMax,
		IsCreatedAt:         isCreatedAt,
		ExportedAtMin:       exportedAtMin,
		ExportedAtMax:       exportedAtMax,
		IsExportedAt:        isExportedAt,
		FirstSeenAtMin:      firstSeenAtMin,
		FirstSeenAtMax:      firstSeenAtMax,
		IsFirstSeenAt:       isFirstSeenAt,
		Recent:              recent,
	}, nil
}

func buildRawTransportExportRows(resp jsonAPIResponse) []rawTransportExportRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]rawTransportExportRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildRawTransportExportRow(resource, included))
	}
	return rows
}

func rawTransportExportRowFromSingle(resp jsonAPISingleResponse) rawTransportExportRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}
	return buildRawTransportExportRow(resp.Data, included)
}

func buildRawTransportExportRow(resource jsonAPIResource, included map[string]jsonAPIResource) rawTransportExportRow {
	row := rawTransportExportRow{
		ID:                  resource.ID,
		ExternalOrderNumber: stringAttr(resource.Attributes, "external-order-number"),
		ExportType:          stringAttr(resource.Attributes, "export-type"),
		TargetDatabase:      stringAttr(resource.Attributes, "target-database"),
		TargetTable:         stringAttr(resource.Attributes, "target-table"),
		IsExportable:        boolAttr(resource.Attributes, "is-exportable"),
		IsExported:          boolAttr(resource.Attributes, "is-exported"),
		IssueType:           stringAttr(resource.Attributes, "issue-type"),
	}

	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
		if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.BrokerName = stringAttr(broker.Attributes, "company-name")
		}
	}

	if rel, ok := resource.Relationships["transport-order"]; ok && rel.Data != nil {
		row.TransportOrderID = rel.Data.ID
		if order, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.TransportOrderNumber = stringAttr(order.Attributes, "external-order-number")
		}
	}

	return row
}

func renderRawTransportExportsTable(cmd *cobra.Command, rows []rawTransportExportRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No raw transport exports found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tORDER\tTYPE\tTARGET\tEXPORTABLE\tEXPORTED\tISSUE\tBROKER")
	for _, row := range rows {
		broker := formatRelated(row.BrokerName, row.BrokerID)
		target := formatRawTransportExportTarget(row.TargetDatabase, row.TargetTable)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%t\t%t\t%s\t%s\n",
			row.ID,
			truncateString(firstNonEmpty(row.ExternalOrderNumber, row.TransportOrderNumber), 22),
			truncateString(row.ExportType, 14),
			truncateString(target, 22),
			row.IsExportable,
			row.IsExported,
			truncateString(row.IssueType, 14),
			truncateString(broker, 24),
		)
	}
	return writer.Flush()
}

func formatRawTransportExportTarget(database, table string) string {
	database = strings.TrimSpace(database)
	table = strings.TrimSpace(table)
	if database == "" {
		return table
	}
	if table == "" {
		return database
	}
	return fmt.Sprintf("%s.%s", database, table)
}
