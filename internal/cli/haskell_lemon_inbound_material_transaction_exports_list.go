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

type haskellLemonInboundMaterialTransactionExportsListOptions struct {
	BaseURL            string
	Token              string
	JSON               bool
	NoAuth             bool
	Limit              int
	Offset             int
	Sort               string
	TransactionDate    string
	TransactionDateMin string
	TransactionDateMax string
	HasTransactionDate string
	CreatedBy          string
	CreatedAtMin       string
	CreatedAtMax       string
	IsCreatedAt        string
	UpdatedAtMin       string
	UpdatedAtMax       string
	IsUpdatedAt        string
}

type haskellLemonInboundMaterialTransactionExportRow struct {
	ID              string `json:"id"`
	TransactionDate string `json:"transaction_date,omitempty"`
	IsTest          bool   `json:"is_test,omitempty"`
	CreatedByID     string `json:"created_by_id,omitempty"`
	CreatedByName   string `json:"created_by_name,omitempty"`
	CreatedAt       string `json:"created_at,omitempty"`
}

func newHaskellLemonInboundMaterialTransactionExportsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Haskell Lemon inbound material transaction exports",
		Long: `List Haskell Lemon inbound material transaction exports with filtering and pagination.

Output Columns:
  ID          Export identifier
  DATE        Transaction date
  TEST        Test export flag
  CREATED BY  Creator
  CREATED AT  Creation timestamp

Filters:
  --transaction-date       Filter by transaction date (YYYY-MM-DD)
  --transaction-date-min   Filter by minimum transaction date (YYYY-MM-DD)
  --transaction-date-max   Filter by maximum transaction date (YYYY-MM-DD)
  --has-transaction-date   Filter by presence of transaction date (true/false)
  --created-by             Filter by creator user ID
  --created-at-min         Filter by created-at on/after (ISO 8601)
  --created-at-max         Filter by created-at on/before (ISO 8601)
  --is-created-at          Filter by has created-at (true/false)
  --updated-at-min         Filter by updated-at on/after (ISO 8601)
  --updated-at-max         Filter by updated-at on/before (ISO 8601)
  --is-updated-at          Filter by has updated-at (true/false)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List exports
  xbe view haskell-lemon-inbound-material-transaction-exports list

  # Filter by transaction date
  xbe view haskell-lemon-inbound-material-transaction-exports list --transaction-date 2025-01-15

  # Filter by creator
  xbe view haskell-lemon-inbound-material-transaction-exports list --created-by 456

  # Output as JSON
  xbe view haskell-lemon-inbound-material-transaction-exports list --json`,
		Args: cobra.NoArgs,
		RunE: runHaskellLemonInboundMaterialTransactionExportsList,
	}
	initHaskellLemonInboundMaterialTransactionExportsListFlags(cmd)
	return cmd
}

func init() {
	haskellLemonInboundMaterialTransactionExportsCmd.AddCommand(newHaskellLemonInboundMaterialTransactionExportsListCmd())
}

func initHaskellLemonInboundMaterialTransactionExportsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("transaction-date", "", "Filter by transaction date (YYYY-MM-DD)")
	cmd.Flags().String("transaction-date-min", "", "Filter by minimum transaction date (YYYY-MM-DD)")
	cmd.Flags().String("transaction-date-max", "", "Filter by maximum transaction date (YYYY-MM-DD)")
	cmd.Flags().String("has-transaction-date", "", "Filter by presence of transaction date (true/false)")
	cmd.Flags().String("created-by", "", "Filter by creator user ID")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by has created-at (true/false)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("is-updated-at", "", "Filter by has updated-at (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runHaskellLemonInboundMaterialTransactionExportsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseHaskellLemonInboundMaterialTransactionExportsListOptions(cmd)
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
	query.Set("fields[haskell-lemon-inbound-material-transaction-exports]", strings.Join([]string{
		"transaction-date",
		"is-test",
		"created-by",
		"created-at",
	}, ","))
	query.Set("include", "created-by")
	query.Set("fields[users]", "name")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[transaction-date]", opts.TransactionDate)
	setFilterIfPresent(query, "filter[transaction-date-min]", opts.TransactionDateMin)
	setFilterIfPresent(query, "filter[transaction-date-max]", opts.TransactionDateMax)
	setFilterIfPresent(query, "filter[has-transaction-date]", opts.HasTransactionDate)
	setFilterIfPresent(query, "filter[created-by]", opts.CreatedBy)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[is-created-at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[is-updated-at]", opts.IsUpdatedAt)

	body, _, err := client.Get(cmd.Context(), "/v1/haskell-lemon-inbound-material-transaction-exports", query)
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

	rows := buildHaskellLemonInboundMaterialTransactionExportRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderHaskellLemonInboundMaterialTransactionExportsTable(cmd, rows)
}

func parseHaskellLemonInboundMaterialTransactionExportsListOptions(cmd *cobra.Command) (haskellLemonInboundMaterialTransactionExportsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	transactionDate, _ := cmd.Flags().GetString("transaction-date")
	transactionDateMin, _ := cmd.Flags().GetString("transaction-date-min")
	transactionDateMax, _ := cmd.Flags().GetString("transaction-date-max")
	hasTransactionDate, _ := cmd.Flags().GetString("has-transaction-date")
	createdBy, _ := cmd.Flags().GetString("created-by")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	isCreatedAt, _ := cmd.Flags().GetString("is-created-at")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	isUpdatedAt, _ := cmd.Flags().GetString("is-updated-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return haskellLemonInboundMaterialTransactionExportsListOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		NoAuth:             noAuth,
		Limit:              limit,
		Offset:             offset,
		Sort:               sort,
		TransactionDate:    transactionDate,
		TransactionDateMin: transactionDateMin,
		TransactionDateMax: transactionDateMax,
		HasTransactionDate: hasTransactionDate,
		CreatedBy:          createdBy,
		CreatedAtMin:       createdAtMin,
		CreatedAtMax:       createdAtMax,
		IsCreatedAt:        isCreatedAt,
		UpdatedAtMin:       updatedAtMin,
		UpdatedAtMax:       updatedAtMax,
		IsUpdatedAt:        isUpdatedAt,
	}, nil
}

func buildHaskellLemonInboundMaterialTransactionExportRows(resp jsonAPIResponse) []haskellLemonInboundMaterialTransactionExportRow {
	rows := make([]haskellLemonInboundMaterialTransactionExportRow, 0, len(resp.Data))
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := haskellLemonInboundMaterialTransactionExportRow{
			ID:              resource.ID,
			TransactionDate: formatDate(stringAttr(attrs, "transaction-date")),
			IsTest:          boolAttr(attrs, "is-test"),
			CreatedAt:       formatDateTime(stringAttr(attrs, "created-at")),
		}

		if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
			row.CreatedByID = rel.Data.ID
			if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.CreatedByName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
			}
		}

		rows = append(rows, row)
	}

	return rows
}

func renderHaskellLemonInboundMaterialTransactionExportsTable(cmd *cobra.Command, rows []haskellLemonInboundMaterialTransactionExportRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No Haskell Lemon inbound material transaction exports found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tDATE\tTEST\tCREATED BY\tCREATED AT")
	for _, row := range rows {
		createdBy := formatRelated(row.CreatedByName, row.CreatedByID)
		fmt.Fprintf(writer, "%s\t%s\t%t\t%s\t%s\n",
			row.ID,
			row.TransactionDate,
			row.IsTest,
			truncateString(createdBy, 32),
			row.CreatedAt,
		)
	}
	return writer.Flush()
}
