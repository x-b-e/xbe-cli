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

type materialTransactionCostCodeAllocationsListOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	NoAuth              bool
	Limit               int
	Offset              int
	Sort                string
	MaterialTransaction string
	CreatedAtMin        string
	CreatedAtMax        string
	UpdatedAtMin        string
	UpdatedAtMax        string
}

type materialTransactionCostCodeAllocationRow struct {
	ID                    string   `json:"id"`
	MaterialTransactionID string   `json:"material_transaction_id,omitempty"`
	CostCodeIDs           []string `json:"cost_code_ids,omitempty"`
	CostCodes             []string `json:"cost_codes,omitempty"`
	AllocationCount       int      `json:"allocations,omitempty"`
	CreatedAt             string   `json:"created_at,omitempty"`
	UpdatedAt             string   `json:"updated_at,omitempty"`
}

func newMaterialTransactionCostCodeAllocationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List material transaction cost code allocations",
		Long: `List material transaction cost code allocations with filtering and pagination.

Output Columns:
  ID           Allocation identifier
  TRANSACTION  Material transaction ID
  ALLOCATIONS  Number of cost code allocations
  COST_CODES   Cost code values (when available)
  CREATED_AT   Allocation creation timestamp

Filters:
  --material-transaction  Filter by material transaction ID (comma-separated for multiple)
  --created-at-min        Filter by created-at on/after (ISO 8601)
  --created-at-max        Filter by created-at on/before (ISO 8601)
  --updated-at-min        Filter by updated-at on/after (ISO 8601)
  --updated-at-max        Filter by updated-at on/before (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List allocations
  xbe view material-transaction-cost-code-allocations list

  # Filter by material transaction
  xbe view material-transaction-cost-code-allocations list --material-transaction 123

  # Filter by created-at range
  xbe view material-transaction-cost-code-allocations list \
    --created-at-min 2025-01-01T00:00:00Z \
    --created-at-max 2025-12-31T23:59:59Z

  # Output as JSON
  xbe view material-transaction-cost-code-allocations list --json`,
		Args: cobra.NoArgs,
		RunE: runMaterialTransactionCostCodeAllocationsList,
	}
	initMaterialTransactionCostCodeAllocationsListFlags(cmd)
	return cmd
}

func init() {
	materialTransactionCostCodeAllocationsCmd.AddCommand(newMaterialTransactionCostCodeAllocationsListCmd())
}

func initMaterialTransactionCostCodeAllocationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("material-transaction", "", "Filter by material transaction ID (comma-separated for multiple)")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialTransactionCostCodeAllocationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMaterialTransactionCostCodeAllocationsListOptions(cmd)
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
	query.Set("fields[material-transaction-cost-code-allocations]", "details,created-at,updated-at,material-transaction,cost-codes")
	query.Set("fields[cost-codes]", "code")
	query.Set("include", "cost-codes")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[material-transaction]", opts.MaterialTransaction)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/material-transaction-cost-code-allocations", query)
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

	rows := buildMaterialTransactionCostCodeAllocationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderMaterialTransactionCostCodeAllocationsTable(cmd, rows)
}

func parseMaterialTransactionCostCodeAllocationsListOptions(cmd *cobra.Command) (materialTransactionCostCodeAllocationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	materialTransaction, _ := cmd.Flags().GetString("material-transaction")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialTransactionCostCodeAllocationsListOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		NoAuth:              noAuth,
		Limit:               limit,
		Offset:              offset,
		Sort:                sort,
		MaterialTransaction: materialTransaction,
		CreatedAtMin:        createdAtMin,
		CreatedAtMax:        createdAtMax,
		UpdatedAtMin:        updatedAtMin,
		UpdatedAtMax:        updatedAtMax,
	}, nil
}

func buildMaterialTransactionCostCodeAllocationRows(resp jsonAPIResponse) []materialTransactionCostCodeAllocationRow {
	included := map[string]map[string]any{}
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc.Attributes
	}

	rows := make([]materialTransactionCostCodeAllocationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		costCodeIDs := relationshipIDsFromMap(resource.Relationships, "cost-codes")
		row := materialTransactionCostCodeAllocationRow{
			ID:                    resource.ID,
			MaterialTransactionID: relationshipIDFromMap(resource.Relationships, "material-transaction"),
			CostCodeIDs:           costCodeIDs,
			CostCodes:             resolveCostCodeCodes(costCodeIDs, included),
			AllocationCount:       allocationDetailsCount(attrs),
			CreatedAt:             formatDateTime(stringAttr(attrs, "created-at")),
			UpdatedAt:             formatDateTime(stringAttr(attrs, "updated-at")),
		}
		rows = append(rows, row)
	}

	return rows
}

func renderMaterialTransactionCostCodeAllocationsTable(cmd *cobra.Command, rows []materialTransactionCostCodeAllocationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No material transaction cost code allocations found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTRANSACTION\tALLOCATIONS\tCOST_CODES\tCREATED_AT")
	for _, row := range rows {
		codes := strings.Join(row.CostCodes, ",")
		if codes == "" {
			codes = strings.Join(row.CostCodeIDs, ",")
		}
		fmt.Fprintf(writer, "%s\t%s\t%d\t%s\t%s\n",
			row.ID,
			row.MaterialTransactionID,
			row.AllocationCount,
			truncateString(codes, 40),
			row.CreatedAt,
		)
	}
	return writer.Flush()
}

func allocationDetailsCount(attrs map[string]any) int {
	if attrs == nil {
		return 0
	}
	value, ok := attrs["details"]
	if !ok || value == nil {
		return 0
	}
	switch typed := value.(type) {
	case []any:
		return len(typed)
	case []map[string]any:
		return len(typed)
	case string:
		var decoded []any
		if err := json.Unmarshal([]byte(typed), &decoded); err == nil {
			return len(decoded)
		}
	}
	return 0
}

func resolveCostCodeCodes(costCodeIDs []string, included map[string]map[string]any) []string {
	if len(costCodeIDs) == 0 {
		return nil
	}
	codes := make([]string, 0, len(costCodeIDs))
	for _, id := range costCodeIDs {
		if attrs, ok := included[resourceKey("cost-codes", id)]; ok {
			if code := stringAttr(attrs, "code"); code != "" {
				codes = append(codes, code)
				continue
			}
		}
		codes = append(codes, id)
	}
	return codes
}
