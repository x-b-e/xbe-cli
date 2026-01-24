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

type rawTransportTractorsListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	Broker       string
	Importer     string
	ImportStatus string
}

type rawTransportTractorRow struct {
	ID                string `json:"id"`
	ExternalTractorID string `json:"external_tractor_id,omitempty"`
	Importer          string `json:"importer,omitempty"`
	ImportStatus      string `json:"import_status,omitempty"`
	BrokerID          string `json:"broker_id,omitempty"`
	BrokerName        string `json:"broker_name,omitempty"`
	TractorID         string `json:"tractor_id,omitempty"`
	TractorNumber     string `json:"tractor_number,omitempty"`
}

func newRawTransportTractorsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List raw transport tractors",
		Long: `List raw transport tractors with filtering and pagination.

Output Columns:
  ID          Raw transport tractor identifier
  EXTERNAL ID External tractor identifier
  IMPORTER    Importer name
  STATUS      Import status
  TRACTOR     Tractor number or ID
  BROKER      Broker name or ID

Filters:
  --broker        Filter by broker ID
  --importer      Filter by importer (e.g., quantix_tmw)
  --import-status Filter by import status (pending, success, failed)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List raw transport tractors
  xbe view raw-transport-tractors list

  # Filter by broker
  xbe view raw-transport-tractors list --broker 123

  # Filter by importer
  xbe view raw-transport-tractors list --importer quantix_tmw

  # Filter by import status
  xbe view raw-transport-tractors list --import-status pending

  # Output as JSON
  xbe view raw-transport-tractors list --json`,
		Args: cobra.NoArgs,
		RunE: runRawTransportTractorsList,
	}
	initRawTransportTractorsListFlags(cmd)
	return cmd
}

func init() {
	rawTransportTractorsCmd.AddCommand(newRawTransportTractorsListCmd())
}

func initRawTransportTractorsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("importer", "", "Filter by importer (e.g., quantix_tmw)")
	cmd.Flags().String("import-status", "", "Filter by import status (pending, success, failed)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runRawTransportTractorsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseRawTransportTractorsListOptions(cmd)
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
	query.Set("fields[raw-transport-tractors]", "external-tractor-id,importer,import-status,broker,tractor")
	query.Set("include", "broker,tractor")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[tractors]", "number")

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
	setFilterIfPresent(query, "filter[importer]", opts.Importer)
	setFilterIfPresent(query, "filter[import_status]", opts.ImportStatus)

	body, _, err := client.Get(cmd.Context(), "/v1/raw-transport-tractors", query)
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

	rows := buildRawTransportTractorRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderRawTransportTractorsTable(cmd, rows)
}

func parseRawTransportTractorsListOptions(cmd *cobra.Command) (rawTransportTractorsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	broker, _ := cmd.Flags().GetString("broker")
	importer, _ := cmd.Flags().GetString("importer")
	importStatus, _ := cmd.Flags().GetString("import-status")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return rawTransportTractorsListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		Broker:       broker,
		Importer:     importer,
		ImportStatus: importStatus,
	}, nil
}

func buildRawTransportTractorRows(resp jsonAPIResponse) []rawTransportTractorRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]rawTransportTractorRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildRawTransportTractorRow(resource, included))
	}
	return rows
}

func rawTransportTractorRowFromSingle(resp jsonAPISingleResponse) rawTransportTractorRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}
	return buildRawTransportTractorRow(resp.Data, included)
}

func buildRawTransportTractorRow(resource jsonAPIResource, included map[string]jsonAPIResource) rawTransportTractorRow {
	row := rawTransportTractorRow{
		ID:                resource.ID,
		ExternalTractorID: stringAttr(resource.Attributes, "external-tractor-id"),
		Importer:          stringAttr(resource.Attributes, "importer"),
		ImportStatus:      stringAttr(resource.Attributes, "import-status"),
	}

	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
		if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.BrokerName = stringAttr(broker.Attributes, "company-name")
		}
	}

	if rel, ok := resource.Relationships["tractor"]; ok && rel.Data != nil {
		row.TractorID = rel.Data.ID
		if tractor, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.TractorNumber = stringAttr(tractor.Attributes, "number")
		}
	}

	return row
}

func renderRawTransportTractorsTable(cmd *cobra.Command, rows []rawTransportTractorRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No raw transport tractors found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tEXTERNAL ID\tIMPORTER\tSTATUS\tTRACTOR\tBROKER")
	for _, row := range rows {
		tractor := firstNonEmpty(row.TractorNumber, row.TractorID)
		broker := firstNonEmpty(row.BrokerName, row.BrokerID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.ExternalTractorID, 22),
			truncateString(row.Importer, 12),
			truncateString(row.ImportStatus, 10),
			truncateString(tractor, 20),
			truncateString(broker, 24),
		)
	}
	return writer.Flush()
}
