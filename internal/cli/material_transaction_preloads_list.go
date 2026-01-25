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

type materialTransactionPreloadsListOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	NoAuth              bool
	Limit               int
	Offset              int
	Sort                string
	Trailer             string
	MaterialTransaction string
	PreloadedAtMin      string
	PreloadedAtMax      string
}

type materialTransactionPreloadRow struct {
	ID                              string `json:"id"`
	PreloadedAt                     string `json:"preloaded_at,omitempty"`
	PreloadMinutes                  int    `json:"preload_minutes,omitempty"`
	TrailerID                       string `json:"trailer_id,omitempty"`
	TrailerNumber                   string `json:"trailer_number,omitempty"`
	MaterialTransactionID           string `json:"material_transaction_id,omitempty"`
	MaterialTransactionTicketNumber string `json:"material_transaction_ticket_number,omitempty"`
	MaterialTransactionAt           string `json:"material_transaction_at,omitempty"`
}

func newMaterialTransactionPreloadsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List material transaction preloads",
		Long: `List material transaction preloads with filtering and pagination.

Output Columns:
  ID                   Preload identifier
  PRELOADED_AT         Preload timestamp
  PRELOAD_MIN          Estimated preload minutes
  TRAILER              Trailer number or ID
  MATERIAL_TRANSACTION Material transaction ticket or ID

Filters:
  --trailer                Filter by trailer ID (comma-separated for multiple)
  --material-transaction   Filter by material transaction ID (comma-separated for multiple)
  --preloaded-at-min       Filter by preload timestamp on/after (ISO 8601)
  --preloaded-at-max       Filter by preload timestamp on/before (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List preloads
  xbe view material-transaction-preloads list

  # Filter by trailer
  xbe view material-transaction-preloads list --trailer 123

  # Filter by preloaded timestamp range
  xbe view material-transaction-preloads list --preloaded-at-min 2025-01-01T00:00:00Z --preloaded-at-max 2025-01-31T23:59:59Z

  # Output as JSON
  xbe view material-transaction-preloads list --json`,
		Args: cobra.NoArgs,
		RunE: runMaterialTransactionPreloadsList,
	}
	initMaterialTransactionPreloadsListFlags(cmd)
	return cmd
}

func init() {
	materialTransactionPreloadsCmd.AddCommand(newMaterialTransactionPreloadsListCmd())
}

func initMaterialTransactionPreloadsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("trailer", "", "Filter by trailer ID (comma-separated for multiple)")
	cmd.Flags().String("material-transaction", "", "Filter by material transaction ID (comma-separated for multiple)")
	cmd.Flags().String("preloaded-at-min", "", "Filter by preload timestamp on/after (ISO 8601)")
	cmd.Flags().String("preloaded-at-max", "", "Filter by preload timestamp on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialTransactionPreloadsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMaterialTransactionPreloadsListOptions(cmd)
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
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run xbe auth login first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[material-transaction-preloads]", "preloaded-at,preload-minutes,material-transaction,trailer")
	query.Set("fields[material-transactions]", "ticket-number,transaction-at")
	query.Set("fields[trailers]", "number")
	query.Set("include", "material-transaction,trailer")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[trailer]", opts.Trailer)
	setFilterIfPresent(query, "filter[material-transaction]", opts.MaterialTransaction)
	setFilterIfPresent(query, "filter[preloaded-at-min]", opts.PreloadedAtMin)
	setFilterIfPresent(query, "filter[preloaded-at-max]", opts.PreloadedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/material-transaction-preloads", query)
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

	rows := buildMaterialTransactionPreloadRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderMaterialTransactionPreloadsTable(cmd, rows)
}

func parseMaterialTransactionPreloadsListOptions(cmd *cobra.Command) (materialTransactionPreloadsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	trailer, _ := cmd.Flags().GetString("trailer")
	materialTransaction, _ := cmd.Flags().GetString("material-transaction")
	preloadedAtMin, _ := cmd.Flags().GetString("preloaded-at-min")
	preloadedAtMax, _ := cmd.Flags().GetString("preloaded-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialTransactionPreloadsListOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		NoAuth:              noAuth,
		Limit:               limit,
		Offset:              offset,
		Sort:                sort,
		Trailer:             trailer,
		MaterialTransaction: materialTransaction,
		PreloadedAtMin:      preloadedAtMin,
		PreloadedAtMax:      preloadedAtMax,
	}, nil
}

func buildMaterialTransactionPreloadRows(resp jsonAPIResponse) []materialTransactionPreloadRow {
	included := map[string]map[string]any{}
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc.Attributes
	}

	rows := make([]materialTransactionPreloadRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		trailerID := relationshipIDFromMap(resource.Relationships, "trailer")
		materialTransactionID := relationshipIDFromMap(resource.Relationships, "material-transaction")
		row := materialTransactionPreloadRow{
			ID:                              resource.ID,
			PreloadedAt:                     formatDateTime(stringAttr(attrs, "preloaded-at")),
			PreloadMinutes:                  intAttr(attrs, "preload-minutes"),
			TrailerID:                       trailerID,
			TrailerNumber:                   resolveTrailerNumber(trailerID, included),
			MaterialTransactionID:           materialTransactionID,
			MaterialTransactionTicketNumber: resolveMaterialTransactionTicketNumber(materialTransactionID, included),
			MaterialTransactionAt:           formatDateTime(resolveMaterialTransactionAt(materialTransactionID, included)),
		}
		rows = append(rows, row)
	}

	return rows
}

func renderMaterialTransactionPreloadsTable(cmd *cobra.Command, rows []materialTransactionPreloadRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No material transaction preloads found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPRELOADED_AT\tPRELOAD_MIN\tTRAILER\tMATERIAL_TRANSACTION")
	for _, row := range rows {
		preloadMinutes := ""
		if row.PreloadMinutes > 0 {
			preloadMinutes = fmt.Sprintf("%d", row.PreloadMinutes)
		}
		trailer := firstNonEmpty(row.TrailerNumber, row.TrailerID)
		materialTransaction := firstNonEmpty(row.MaterialTransactionTicketNumber, row.MaterialTransactionID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.PreloadedAt,
			preloadMinutes,
			truncateString(trailer, 30),
			truncateString(materialTransaction, 30),
		)
	}
	return writer.Flush()
}

func resolveTrailerNumber(trailerID string, included map[string]map[string]any) string {
	if trailerID == "" {
		return ""
	}
	if attrs, ok := included[resourceKey("trailers", trailerID)]; ok {
		return stringAttr(attrs, "number")
	}
	return ""
}

func resolveMaterialTransactionTicketNumber(materialTransactionID string, included map[string]map[string]any) string {
	if materialTransactionID == "" {
		return ""
	}
	if attrs, ok := included[resourceKey("material-transactions", materialTransactionID)]; ok {
		return stringAttr(attrs, "ticket-number")
	}
	return ""
}

func resolveMaterialTransactionAt(materialTransactionID string, included map[string]map[string]any) string {
	if materialTransactionID == "" {
		return ""
	}
	if attrs, ok := included[resourceKey("material-transactions", materialTransactionID)]; ok {
		return stringAttr(attrs, "transaction-at")
	}
	return ""
}
