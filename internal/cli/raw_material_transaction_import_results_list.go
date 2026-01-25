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

type rawMaterialTransactionImportResultsListOptions struct {
	BaseURL                         string
	Token                           string
	JSON                            bool
	NoAuth                          bool
	Limit                           int
	Offset                          int
	Sort                            string
	Source                          string
	SourceType                      string
	SourceID                        string
	Broker                          string
	BatchID                         string
	LocationID                      string
	HasErrors                       string
	EarliestCreatedTransactionAtMin string
	EarliestCreatedTransactionAtMax string
	IsEarliestCreatedTransactionAt  string
	LatestCreatedTransactionAtMin   string
	LatestCreatedTransactionAtMax   string
	IsLatestCreatedTransactionAt    string
	DisconnectedAtMin               string
	DisconnectedAtMax               string
	IsDisconnectedAt                string
}

type rawMaterialTransactionImportResultRow struct {
	ID                           string `json:"id"`
	Importer                     string `json:"importer,omitempty"`
	ConfigurationID              string `json:"configuration_id,omitempty"`
	LocationID                   string `json:"location_id,omitempty"`
	BatchID                      string `json:"batch_id,omitempty"`
	HasErrors                    bool   `json:"has_errors"`
	IsConnected                  bool   `json:"is_connected"`
	EarliestCreatedTransactionAt string `json:"earliest_created_transaction_at,omitempty"`
	LatestCreatedTransactionAt   string `json:"latest_created_transaction_at,omitempty"`
	DisconnectedAt               string `json:"disconnected_at,omitempty"`
	LastConnectedAt              string `json:"last_connected_at,omitempty"`
	SourceType                   string `json:"source_type,omitempty"`
	SourceID                     string `json:"source_id,omitempty"`
	BrokerID                     string `json:"broker_id,omitempty"`
}

func newRawMaterialTransactionImportResultsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List raw material transaction import results",
		Long: `List raw material transaction import results with filtering and pagination.

Output Columns:
  ID             Import result ID
  IMPORTER       Importer class name
  CONFIG         Configuration identifier
  BROKER         Broker ID
  SOURCE         Source type/id
  ERRORS         Whether the import has errors
  CONNECTED      Whether the importer is currently connected
  LATEST CREATED Latest created transaction timestamp

Filters:
  --source                          Filter by source (Type|ID, comma-separated for multiple)
  --source-type                     Filter by source type (e.g., MaterialSite or material-sites)
  --source-id                       Filter by source ID (used with --source-type)
  --broker                          Filter by broker ID
  --batch-id                        Filter by batch ID
  --location-id                     Filter by location ID
  --has-errors                      Filter by error status (true/false)
  --earliest-created-transaction-at-min  Filter by earliest created transaction on/after (ISO 8601)
  --earliest-created-transaction-at-max  Filter by earliest created transaction on/before (ISO 8601)
  --is-earliest-created-transaction-at   Filter by presence of earliest created transaction (true/false)
  --latest-created-transaction-at-min    Filter by latest created transaction on/after (ISO 8601)
  --latest-created-transaction-at-max    Filter by latest created transaction on/before (ISO 8601)
  --is-latest-created-transaction-at     Filter by presence of latest created transaction (true/false)
  --disconnected-at-min              Filter by disconnected-at on/after (ISO 8601)
  --disconnected-at-max              Filter by disconnected-at on/before (ISO 8601)
  --is-disconnected-at               Filter by presence of disconnected-at (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List import results
  xbe view raw-material-transaction-import-results list

  # Filter by broker
  xbe view raw-material-transaction-import-results list --broker 123

  # Filter by source
  xbe view raw-material-transaction-import-results list --source-type material-sites --source-id 456

  # Filter by error status
  xbe view raw-material-transaction-import-results list --has-errors true

  # Output as JSON
  xbe view raw-material-transaction-import-results list --json`,
		Args: cobra.NoArgs,
		RunE: runRawMaterialTransactionImportResultsList,
	}
	initRawMaterialTransactionImportResultsListFlags(cmd)
	return cmd
}

func init() {
	rawMaterialTransactionImportResultsCmd.AddCommand(newRawMaterialTransactionImportResultsListCmd())
}

func initRawMaterialTransactionImportResultsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("source", "", "Filter by source (Type|ID, comma-separated for multiple)")
	cmd.Flags().String("source-type", "", "Filter by source type (e.g., MaterialSite or material-sites)")
	cmd.Flags().String("source-id", "", "Filter by source ID (used with --source-type)")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("batch-id", "", "Filter by batch ID")
	cmd.Flags().String("location-id", "", "Filter by location ID")
	cmd.Flags().String("has-errors", "", "Filter by error status (true/false)")
	cmd.Flags().String("earliest-created-transaction-at-min", "", "Filter by earliest created transaction on/after (ISO 8601)")
	cmd.Flags().String("earliest-created-transaction-at-max", "", "Filter by earliest created transaction on/before (ISO 8601)")
	cmd.Flags().String("is-earliest-created-transaction-at", "", "Filter by presence of earliest created transaction (true/false)")
	cmd.Flags().String("latest-created-transaction-at-min", "", "Filter by latest created transaction on/after (ISO 8601)")
	cmd.Flags().String("latest-created-transaction-at-max", "", "Filter by latest created transaction on/before (ISO 8601)")
	cmd.Flags().String("is-latest-created-transaction-at", "", "Filter by presence of latest created transaction (true/false)")
	cmd.Flags().String("disconnected-at-min", "", "Filter by disconnected-at on/after (ISO 8601)")
	cmd.Flags().String("disconnected-at-max", "", "Filter by disconnected-at on/before (ISO 8601)")
	cmd.Flags().String("is-disconnected-at", "", "Filter by presence of disconnected-at (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runRawMaterialTransactionImportResultsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseRawMaterialTransactionImportResultsListOptions(cmd)
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
	query.Set("fields[raw-material-transaction-import-results]", "importer,configurationid,locationid,has-errors,is-connected,last-connected-at,earliest-created-transaction-at,latest-created-transaction-at,batch-id,disconnected-at")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	if opts.Source != "" {
		query.Set("filter[source]", normalizePolymorphicFilterValue(opts.Source))
	} else if opts.SourceType != "" && opts.SourceID != "" {
		normalizedType := normalizeResourceTypeForFilter(opts.SourceType)
		if normalizedType == "" {
			normalizedType = strings.TrimSpace(opts.SourceType)
		}
		query.Set("filter[source]", normalizedType+"|"+strings.TrimSpace(opts.SourceID))
	} else if opts.SourceType != "" {
		normalizedType := normalizeResourceTypeForFilter(opts.SourceType)
		if normalizedType == "" {
			normalizedType = strings.TrimSpace(opts.SourceType)
		}
		query.Set("filter[source-type]", normalizedType)
	}

	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[batch-id]", opts.BatchID)
	setFilterIfPresent(query, "filter[locationid]", opts.LocationID)
	setFilterIfPresent(query, "filter[has-errors]", opts.HasErrors)
	setFilterIfPresent(query, "filter[earliest-created-transaction-at-min]", opts.EarliestCreatedTransactionAtMin)
	setFilterIfPresent(query, "filter[earliest-created-transaction-at-max]", opts.EarliestCreatedTransactionAtMax)
	setFilterIfPresent(query, "filter[is-earliest-created-transaction-at]", opts.IsEarliestCreatedTransactionAt)
	setFilterIfPresent(query, "filter[latest-created-transaction-at-min]", opts.LatestCreatedTransactionAtMin)
	setFilterIfPresent(query, "filter[latest-created-transaction-at-max]", opts.LatestCreatedTransactionAtMax)
	setFilterIfPresent(query, "filter[is-latest-created-transaction-at]", opts.IsLatestCreatedTransactionAt)
	setFilterIfPresent(query, "filter[disconnected-at-min]", opts.DisconnectedAtMin)
	setFilterIfPresent(query, "filter[disconnected-at-max]", opts.DisconnectedAtMax)
	setFilterIfPresent(query, "filter[is-disconnected-at]", opts.IsDisconnectedAt)

	body, _, err := client.Get(cmd.Context(), "/v1/raw-material-transaction-import-results", query)
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

	rows := buildRawMaterialTransactionImportResultRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderRawMaterialTransactionImportResultsTable(cmd, rows)
}

func parseRawMaterialTransactionImportResultsListOptions(cmd *cobra.Command) (rawMaterialTransactionImportResultsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	source, _ := cmd.Flags().GetString("source")
	sourceType, _ := cmd.Flags().GetString("source-type")
	sourceID, _ := cmd.Flags().GetString("source-id")
	broker, _ := cmd.Flags().GetString("broker")
	batchID, _ := cmd.Flags().GetString("batch-id")
	locationID, _ := cmd.Flags().GetString("location-id")
	hasErrors, _ := cmd.Flags().GetString("has-errors")
	earliestCreatedTransactionAtMin, _ := cmd.Flags().GetString("earliest-created-transaction-at-min")
	earliestCreatedTransactionAtMax, _ := cmd.Flags().GetString("earliest-created-transaction-at-max")
	isEarliestCreatedTransactionAt, _ := cmd.Flags().GetString("is-earliest-created-transaction-at")
	latestCreatedTransactionAtMin, _ := cmd.Flags().GetString("latest-created-transaction-at-min")
	latestCreatedTransactionAtMax, _ := cmd.Flags().GetString("latest-created-transaction-at-max")
	isLatestCreatedTransactionAt, _ := cmd.Flags().GetString("is-latest-created-transaction-at")
	disconnectedAtMin, _ := cmd.Flags().GetString("disconnected-at-min")
	disconnectedAtMax, _ := cmd.Flags().GetString("disconnected-at-max")
	isDisconnectedAt, _ := cmd.Flags().GetString("is-disconnected-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return rawMaterialTransactionImportResultsListOptions{
		BaseURL:                         baseURL,
		Token:                           token,
		JSON:                            jsonOut,
		NoAuth:                          noAuth,
		Limit:                           limit,
		Offset:                          offset,
		Sort:                            sort,
		Source:                          source,
		SourceType:                      sourceType,
		SourceID:                        sourceID,
		Broker:                          broker,
		BatchID:                         batchID,
		LocationID:                      locationID,
		HasErrors:                       hasErrors,
		EarliestCreatedTransactionAtMin: earliestCreatedTransactionAtMin,
		EarliestCreatedTransactionAtMax: earliestCreatedTransactionAtMax,
		IsEarliestCreatedTransactionAt:  isEarliestCreatedTransactionAt,
		LatestCreatedTransactionAtMin:   latestCreatedTransactionAtMin,
		LatestCreatedTransactionAtMax:   latestCreatedTransactionAtMax,
		IsLatestCreatedTransactionAt:    isLatestCreatedTransactionAt,
		DisconnectedAtMin:               disconnectedAtMin,
		DisconnectedAtMax:               disconnectedAtMax,
		IsDisconnectedAt:                isDisconnectedAt,
	}, nil
}

func buildRawMaterialTransactionImportResultRows(resp jsonAPIResponse) []rawMaterialTransactionImportResultRow {
	rows := make([]rawMaterialTransactionImportResultRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := rawMaterialTransactionImportResultRow{
			ID:                           resource.ID,
			Importer:                     stringAttr(attrs, "importer"),
			ConfigurationID:              stringAttr(attrs, "configurationid"),
			LocationID:                   stringAttr(attrs, "locationid"),
			BatchID:                      stringAttr(attrs, "batch-id"),
			HasErrors:                    boolAttr(attrs, "has-errors"),
			IsConnected:                  boolAttr(attrs, "is-connected"),
			EarliestCreatedTransactionAt: formatDateTime(stringAttr(attrs, "earliest-created-transaction-at")),
			LatestCreatedTransactionAt:   formatDateTime(stringAttr(attrs, "latest-created-transaction-at")),
			DisconnectedAt:               formatDateTime(stringAttr(attrs, "disconnected-at")),
			LastConnectedAt:              formatDateTime(stringAttr(attrs, "last-connected-at")),
		}

		if rel, ok := resource.Relationships["source"]; ok && rel.Data != nil {
			row.SourceType = rel.Data.Type
			row.SourceID = rel.Data.ID
		}
		row.BrokerID = relationshipIDFromMap(resource.Relationships, "broker")

		rows = append(rows, row)
	}
	return rows
}

func renderRawMaterialTransactionImportResultsTable(cmd *cobra.Command, rows []rawMaterialTransactionImportResultRow) error {
	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tIMPORTER\tCONFIG\tBROKER\tSOURCE\tERRORS\tCONNECTED\tLATEST CREATED")
	for _, row := range rows {
		source := ""
		if row.SourceType != "" && row.SourceID != "" {
			source = row.SourceType + "/" + row.SourceID
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%t\t%t\t%s\n",
			row.ID,
			truncateString(row.Importer, 24),
			row.ConfigurationID,
			row.BrokerID,
			truncateString(source, 32),
			row.HasErrors,
			row.IsConnected,
			row.LatestCreatedTransactionAt,
		)
	}
	return writer.Flush()
}
