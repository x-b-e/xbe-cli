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

type haskellLemonOutboundMaterialTransactionExportsListOptions struct {
	BaseURL            string
	Token              string
	JSON               bool
	NoAuth             bool
	Limit              int
	Offset             int
	Sort               string
	CreatedBy          string
	TransactionDate    string
	TransactionDateMin string
	TransactionDateMax string
	HasTransactionDate string
}

type haskellLemonOutboundMaterialTransactionExportRow struct {
	ID              string `json:"id"`
	TransactionDate string `json:"transaction_date,omitempty"`
	IsTest          bool   `json:"is_test"`
	CreatedByID     string `json:"created_by_id,omitempty"`
}

func newHaskellLemonOutboundMaterialTransactionExportsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Haskell Lemon outbound material transaction exports",
		Long: `List Haskell Lemon outbound material transaction exports.

Output Columns:
  ID                 Export identifier
  TRANSACTION DATE   Transaction date for the export
  TEST               Whether this is a test export
  CREATED BY         Creator user ID

Filters:
  --created-by            Filter by created-by user ID
  --transaction-date      Filter by transaction date (YYYY-MM-DD)
  --transaction-date-min  Filter by transaction date on/after (YYYY-MM-DD)
  --transaction-date-max  Filter by transaction date on/before (YYYY-MM-DD)
  --has-transaction-date  Filter by presence of transaction date (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List exports
  xbe view haskell-lemon-outbound-material-transaction-exports list

  # Filter by date
  xbe view haskell-lemon-outbound-material-transaction-exports list --transaction-date-min 2025-01-01

  # Output as JSON
  xbe view haskell-lemon-outbound-material-transaction-exports list --json`,
		Args: cobra.NoArgs,
		RunE: runHaskellLemonOutboundMaterialTransactionExportsList,
	}
	initHaskellLemonOutboundMaterialTransactionExportsListFlags(cmd)
	return cmd
}

func init() {
	haskellLemonOutboundMaterialTransactionExportsCmd.AddCommand(newHaskellLemonOutboundMaterialTransactionExportsListCmd())
}

func initHaskellLemonOutboundMaterialTransactionExportsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("created-by", "", "Filter by created-by user ID")
	cmd.Flags().String("transaction-date", "", "Filter by transaction date (YYYY-MM-DD)")
	cmd.Flags().String("transaction-date-min", "", "Filter by transaction date on/after (YYYY-MM-DD)")
	cmd.Flags().String("transaction-date-max", "", "Filter by transaction date on/before (YYYY-MM-DD)")
	cmd.Flags().String("has-transaction-date", "", "Filter by presence of transaction date (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runHaskellLemonOutboundMaterialTransactionExportsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseHaskellLemonOutboundMaterialTransactionExportsListOptions(cmd)
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
	query.Set("fields[haskell-lemon-outbound-material-transaction-exports]", "transaction-date,is-test,created-by")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[created-by]", opts.CreatedBy)
	setFilterIfPresent(query, "filter[transaction-date]", opts.TransactionDate)
	setFilterIfPresent(query, "filter[transaction-date-min]", opts.TransactionDateMin)
	setFilterIfPresent(query, "filter[transaction-date-max]", opts.TransactionDateMax)
	setFilterIfPresent(query, "filter[has-transaction-date]", opts.HasTransactionDate)

	body, _, err := client.Get(cmd.Context(), "/v1/haskell-lemon-outbound-material-transaction-exports", query)
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

	rows := buildHaskellLemonOutboundMaterialTransactionExportRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderHaskellLemonOutboundMaterialTransactionExportsTable(cmd, rows)
}

func parseHaskellLemonOutboundMaterialTransactionExportsListOptions(cmd *cobra.Command) (haskellLemonOutboundMaterialTransactionExportsListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return haskellLemonOutboundMaterialTransactionExportsListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return haskellLemonOutboundMaterialTransactionExportsListOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return haskellLemonOutboundMaterialTransactionExportsListOptions{}, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return haskellLemonOutboundMaterialTransactionExportsListOptions{}, err
	}
	sort, err := cmd.Flags().GetString("sort")
	if err != nil {
		return haskellLemonOutboundMaterialTransactionExportsListOptions{}, err
	}
	createdBy, err := cmd.Flags().GetString("created-by")
	if err != nil {
		return haskellLemonOutboundMaterialTransactionExportsListOptions{}, err
	}
	transactionDate, err := cmd.Flags().GetString("transaction-date")
	if err != nil {
		return haskellLemonOutboundMaterialTransactionExportsListOptions{}, err
	}
	transactionDateMin, err := cmd.Flags().GetString("transaction-date-min")
	if err != nil {
		return haskellLemonOutboundMaterialTransactionExportsListOptions{}, err
	}
	transactionDateMax, err := cmd.Flags().GetString("transaction-date-max")
	if err != nil {
		return haskellLemonOutboundMaterialTransactionExportsListOptions{}, err
	}
	hasTransactionDate, err := cmd.Flags().GetString("has-transaction-date")
	if err != nil {
		return haskellLemonOutboundMaterialTransactionExportsListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return haskellLemonOutboundMaterialTransactionExportsListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return haskellLemonOutboundMaterialTransactionExportsListOptions{}, err
	}

	return haskellLemonOutboundMaterialTransactionExportsListOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		NoAuth:             noAuth,
		Limit:              limit,
		Offset:             offset,
		Sort:               sort,
		CreatedBy:          createdBy,
		TransactionDate:    transactionDate,
		TransactionDateMin: transactionDateMin,
		TransactionDateMax: transactionDateMax,
		HasTransactionDate: hasTransactionDate,
	}, nil
}

func buildHaskellLemonOutboundMaterialTransactionExportRows(resp jsonAPIResponse) []haskellLemonOutboundMaterialTransactionExportRow {
	rows := make([]haskellLemonOutboundMaterialTransactionExportRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, haskellLemonOutboundMaterialTransactionExportRowFromResource(resource))
	}
	return rows
}

func haskellLemonOutboundMaterialTransactionExportRowFromResource(resource jsonAPIResource) haskellLemonOutboundMaterialTransactionExportRow {
	attrs := resource.Attributes
	return haskellLemonOutboundMaterialTransactionExportRow{
		ID:              resource.ID,
		TransactionDate: formatDate(stringAttr(attrs, "transaction-date")),
		IsTest:          boolAttr(attrs, "is-test"),
		CreatedByID:     relationshipIDFromMap(resource.Relationships, "created-by"),
	}
}

func haskellLemonOutboundMaterialTransactionExportRowFromSingle(resp jsonAPISingleResponse) haskellLemonOutboundMaterialTransactionExportRow {
	return haskellLemonOutboundMaterialTransactionExportRowFromResource(resp.Data)
}

func renderHaskellLemonOutboundMaterialTransactionExportsTable(cmd *cobra.Command, rows []haskellLemonOutboundMaterialTransactionExportRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No Haskell Lemon outbound material transaction exports found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, 32, 0)
	fmt.Fprintln(writer, "ID\tTRANSACTION DATE\tTEST\tCREATED BY")

	for _, row := range rows {
		isTest := "no"
		if row.IsTest {
			isTest = "yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			row.TransactionDate,
			isTest,
			row.CreatedByID,
		)
	}

	return writer.Flush()
}
