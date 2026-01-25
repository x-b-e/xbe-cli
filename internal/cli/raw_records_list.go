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

type rawRecordsListOptions struct {
	BaseURL                       string
	Token                         string
	JSON                          bool
	NoAuth                        bool
	Limit                         int
	Offset                        int
	Sort                          string
	Broker                        string
	IntegrationConfig             string
	IntegrationConfigOrganization string
	InternalRecord                string
	ExternalRecordType            string
	ExternalRecordID              string
	IsProcessed                   string
	IsFailed                      string
	IsSkipped                     string
	HasData                       string
}

type rawRecordRow struct {
	ID                    string `json:"id"`
	ExternalRecordType    string `json:"external_record_type,omitempty"`
	ExternalRecordID      string `json:"external_record_id,omitempty"`
	InternalRecordType    string `json:"internal_record_type,omitempty"`
	InternalRecordID      string `json:"internal_record_id,omitempty"`
	BrokerID              string `json:"broker_id,omitempty"`
	BrokerName            string `json:"broker_name,omitempty"`
	IntegrationConfigID   string `json:"integration_config_id,omitempty"`
	IntegrationConfigName string `json:"integration_config_name,omitempty"`
	IsProcessed           bool   `json:"is_processed,omitempty"`
	IsFailed              bool   `json:"is_failed,omitempty"`
	IsSkipped             bool   `json:"is_skipped,omitempty"`
}

func newRawRecordsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List raw records",
		Long: `List ingest raw records with filtering and pagination.

Output Columns:
  ID         Raw record identifier
  EXTERNAL   External record type and ID
  INTERNAL   Internal record type and ID
  BROKER     Broker name or ID
  CONFIG     Integration config name or ID
  PROCESSED  Processed status
  FAILED     Failed status
  SKIPPED    Skipped status

Filters:
  --broker                          Filter by broker ID
  --integration-config              Filter by integration config ID
  --integration-config-organization Filter by integration config organization (Type|ID)
  --internal-record                 Filter by internal record (Type|ID)
  --external-record-type            Filter by external record type
  --external-record-id              Filter by external record ID
  --is-processed                    Filter by processed status (true/false)
  --is-failed                       Filter by failed status (true/false)
  --is-skipped                      Filter by skipped status (true/false)
  --has-data                        Filter by data presence (true/false)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List raw records
  xbe view raw-records list

  # Filter by external record
  xbe view raw-records list --external-record-type Equipment --external-record-id 123

  # Filter by internal record
  xbe view raw-records list --internal-record "Project|456"

  # Filter by processing status
  xbe view raw-records list --is-processed true --is-failed false

  # Output as JSON
  xbe view raw-records list --json`,
		Args: cobra.NoArgs,
		RunE: runRawRecordsList,
	}
	initRawRecordsListFlags(cmd)
	return cmd
}

func init() {
	rawRecordsCmd.AddCommand(newRawRecordsListCmd())
}

func initRawRecordsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("integration-config", "", "Filter by integration config ID")
	cmd.Flags().String("integration-config-organization", "", "Filter by integration config organization (Type|ID)")
	cmd.Flags().String("internal-record", "", "Filter by internal record (Type|ID)")
	cmd.Flags().String("external-record-type", "", "Filter by external record type")
	cmd.Flags().String("external-record-id", "", "Filter by external record ID")
	cmd.Flags().String("is-processed", "", "Filter by processed status (true/false)")
	cmd.Flags().String("is-failed", "", "Filter by failed status (true/false)")
	cmd.Flags().String("is-skipped", "", "Filter by skipped status (true/false)")
	cmd.Flags().String("has-data", "", "Filter by data presence (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runRawRecordsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseRawRecordsListOptions(cmd)
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
	query.Set("fields[raw-records]", "external-record-type,external-record-id,internal-record-type,internal-record-id,is-processed,is-failed,is-skipped,broker,integration-config")
	query.Set("include", "broker,integration-config")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[integration-configs]", "friendly-name")

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
	setFilterIfPresent(query, "filter[integration-config]", opts.IntegrationConfig)
	setFilterIfPresent(query, "filter[integration-config-organization]", opts.IntegrationConfigOrganization)
	setFilterIfPresent(query, "filter[internal-record]", opts.InternalRecord)
	setFilterIfPresent(query, "filter[external-record-type]", opts.ExternalRecordType)
	setFilterIfPresent(query, "filter[external-record-id]", opts.ExternalRecordID)
	setFilterIfPresent(query, "filter[is-processed]", opts.IsProcessed)
	setFilterIfPresent(query, "filter[is-failed]", opts.IsFailed)
	setFilterIfPresent(query, "filter[is-skipped]", opts.IsSkipped)
	setFilterIfPresent(query, "filter[has-data]", opts.HasData)

	body, _, err := client.Get(cmd.Context(), "/v1/ingest/raw-records", query)
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

	rows := buildRawRecordRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderRawRecordsTable(cmd, rows)
}

func parseRawRecordsListOptions(cmd *cobra.Command) (rawRecordsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	broker, _ := cmd.Flags().GetString("broker")
	integrationConfig, _ := cmd.Flags().GetString("integration-config")
	integrationConfigOrganization, _ := cmd.Flags().GetString("integration-config-organization")
	internalRecord, _ := cmd.Flags().GetString("internal-record")
	externalRecordType, _ := cmd.Flags().GetString("external-record-type")
	externalRecordID, _ := cmd.Flags().GetString("external-record-id")
	isProcessed, _ := cmd.Flags().GetString("is-processed")
	isFailed, _ := cmd.Flags().GetString("is-failed")
	isSkipped, _ := cmd.Flags().GetString("is-skipped")
	hasData, _ := cmd.Flags().GetString("has-data")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return rawRecordsListOptions{
		BaseURL:                       baseURL,
		Token:                         token,
		JSON:                          jsonOut,
		NoAuth:                        noAuth,
		Limit:                         limit,
		Offset:                        offset,
		Sort:                          sort,
		Broker:                        broker,
		IntegrationConfig:             integrationConfig,
		IntegrationConfigOrganization: integrationConfigOrganization,
		InternalRecord:                internalRecord,
		ExternalRecordType:            externalRecordType,
		ExternalRecordID:              externalRecordID,
		IsProcessed:                   isProcessed,
		IsFailed:                      isFailed,
		IsSkipped:                     isSkipped,
		HasData:                       hasData,
	}, nil
}

func buildRawRecordRows(resp jsonAPIResponse) []rawRecordRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]rawRecordRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildRawRecordRow(resource, included))
	}

	return rows
}

func rawRecordRowFromSingle(resp jsonAPISingleResponse) rawRecordRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	return buildRawRecordRow(resp.Data, included)
}

func buildRawRecordRow(resource jsonAPIResource, included map[string]jsonAPIResource) rawRecordRow {
	attrs := resource.Attributes
	row := rawRecordRow{
		ID:                 resource.ID,
		ExternalRecordType: stringAttr(attrs, "external-record-type"),
		ExternalRecordID:   stringAttr(attrs, "external-record-id"),
		InternalRecordType: stringAttr(attrs, "internal-record-type"),
		InternalRecordID:   stringAttr(attrs, "internal-record-id"),
		IsProcessed:        boolAttr(attrs, "is-processed"),
		IsFailed:           boolAttr(attrs, "is-failed"),
		IsSkipped:          boolAttr(attrs, "is-skipped"),
	}

	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
		if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.BrokerName = stringAttr(broker.Attributes, "company-name")
		}
	}

	if rel, ok := resource.Relationships["integration-config"]; ok && rel.Data != nil {
		row.IntegrationConfigID = rel.Data.ID
		if config, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.IntegrationConfigName = stringAttr(config.Attributes, "friendly-name")
		}
	}

	return row
}

func renderRawRecordsTable(cmd *cobra.Command, rows []rawRecordRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No raw records found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tEXTERNAL\tINTERNAL\tBROKER\tCONFIG\tPROCESSED\tFAILED\tSKIPPED")
	for _, row := range rows {
		external := formatPolymorphic(row.ExternalRecordType, row.ExternalRecordID)
		internal := formatPolymorphic(row.InternalRecordType, row.InternalRecordID)
		brokerLabel := formatRelated(row.BrokerName, row.BrokerID)
		configLabel := formatRelated(row.IntegrationConfigName, row.IntegrationConfigID)

		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%t\t%t\t%t\n",
			row.ID,
			truncateString(external, 32),
			truncateString(internal, 32),
			truncateString(brokerLabel, 24),
			truncateString(configLabel, 24),
			row.IsProcessed,
			row.IsFailed,
			row.IsSkipped,
		)
	}
	return writer.Flush()
}
