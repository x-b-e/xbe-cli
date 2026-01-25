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

type rawTransportTrailersListOptions struct {
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

type rawTransportTrailerRow struct {
	ID                string `json:"id"`
	ExternalTrailerID string `json:"external_trailer_id,omitempty"`
	Importer          string `json:"importer,omitempty"`
	ImportStatus      string `json:"import_status,omitempty"`
	BrokerID          string `json:"broker_id,omitempty"`
	TrailerID         string `json:"trailer_id,omitempty"`
}

func newRawTransportTrailersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List raw transport trailers",
		Long: `List raw transport trailers with filtering and pagination.

Output Columns:
  ID          Raw transport trailer ID
  EXTERNAL ID External trailer identifier
  IMPORTER    Importer key
  STATUS      Import status
  BROKER      Broker ID
  TRAILER     Trailer ID

Filters:
  --broker         Filter by broker ID
  --importer       Filter by importer key
  --import-status  Filter by import status (pending, success, failed)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List raw transport trailers
  xbe view raw-transport-trailers list

  # Filter by broker
  xbe view raw-transport-trailers list --broker 123

  # Filter by importer
  xbe view raw-transport-trailers list --importer quantix_tmw

  # Filter by import status
  xbe view raw-transport-trailers list --import-status pending

  # Output as JSON
  xbe view raw-transport-trailers list --json`,
		Args: cobra.NoArgs,
		RunE: runRawTransportTrailersList,
	}
	initRawTransportTrailersListFlags(cmd)
	return cmd
}

func init() {
	rawTransportTrailersCmd.AddCommand(newRawTransportTrailersListCmd())
}

func initRawTransportTrailersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("importer", "", "Filter by importer key")
	cmd.Flags().String("import-status", "", "Filter by import status (pending, success, failed)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runRawTransportTrailersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseRawTransportTrailersListOptions(cmd)
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
	query.Set("fields[raw-transport-trailers]", "external-trailer-id,importer,import-status")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[importer]", opts.Importer)
	setFilterIfPresent(query, "filter[import-status]", opts.ImportStatus)

	body, _, err := client.Get(cmd.Context(), "/v1/raw-transport-trailers", query)
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

	rows := buildRawTransportTrailerRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderRawTransportTrailersTable(cmd, rows)
}

func parseRawTransportTrailersListOptions(cmd *cobra.Command) (rawTransportTrailersListOptions, error) {
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

	return rawTransportTrailersListOptions{
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

func buildRawTransportTrailerRows(resp jsonAPIResponse) []rawTransportTrailerRow {
	rows := make([]rawTransportTrailerRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := rawTransportTrailerRow{
			ID:                resource.ID,
			ExternalTrailerID: stringAttr(attrs, "external-trailer-id"),
			Importer:          stringAttr(attrs, "importer"),
			ImportStatus:      stringAttr(attrs, "import-status"),
			BrokerID:          relationshipIDFromMap(resource.Relationships, "broker"),
			TrailerID:         relationshipIDFromMap(resource.Relationships, "trailer"),
		}
		rows = append(rows, row)
	}
	return rows
}

func renderRawTransportTrailersTable(cmd *cobra.Command, rows []rawTransportTrailerRow) error {
	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tEXTERNAL ID\tIMPORTER\tSTATUS\tBROKER\tTRAILER")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.ExternalTrailerID, 24),
			truncateString(row.Importer, 24),
			row.ImportStatus,
			row.BrokerID,
			row.TrailerID,
		)
	}
	return writer.Flush()
}
