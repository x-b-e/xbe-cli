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

type materialTransactionStatusChangesListOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	NoAuth              bool
	Limit               int
	Offset              int
	Sort                string
	MaterialTransaction string
	Status              string
	CreatedAtMin        string
	CreatedAtMax        string
	IsCreatedAt         string
	UpdatedAtMin        string
	UpdatedAtMax        string
	IsUpdatedAt         string
}

type materialTransactionStatusChangeRow struct {
	ID                    string `json:"id"`
	MaterialTransactionID string `json:"material_transaction_id,omitempty"`
	Status                string `json:"status,omitempty"`
	ChangedAt             string `json:"changed_at,omitempty"`
	Comment               string `json:"comment,omitempty"`
	ChangedByID           string `json:"changed_by_id,omitempty"`
	ChangedByName         string `json:"changed_by_name,omitempty"`
	CreatedAt             string `json:"created_at,omitempty"`
	UpdatedAt             string `json:"updated_at,omitempty"`
}

func newMaterialTransactionStatusChangesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List material transaction status changes",
		Long: `List material transaction status changes.

Output Columns:
  ID          Status change identifier
  MTXN        Material transaction ID
  STATUS      Status value
  CHANGED AT  When the status change occurred
  CHANGED BY  User who made the change (when available)
  COMMENT     Status change comment

Filters:
  --material-transaction  Filter by material transaction ID
  --status                Filter by status value
  --created-at-min         Filter by created-at on/after (ISO 8601)
  --created-at-max         Filter by created-at on/before (ISO 8601)
  --is-created-at          Filter by presence of created-at (true/false)
  --updated-at-min         Filter by updated-at on/after (ISO 8601)
  --updated-at-max         Filter by updated-at on/before (ISO 8601)
  --is-updated-at          Filter by presence of updated-at (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List status changes
  xbe view material-transaction-status-changes list

  # Filter by material transaction
  xbe view material-transaction-status-changes list --material-transaction 123

  # Filter by status
  xbe view material-transaction-status-changes list --status accepted

  # Filter by created-at range
  xbe view material-transaction-status-changes list \
    --created-at-min 2026-01-23T00:00:00Z \
    --created-at-max 2026-01-24T00:00:00Z

  # Output as JSON
  xbe view material-transaction-status-changes list --json`,
		Args: cobra.NoArgs,
		RunE: runMaterialTransactionStatusChangesList,
	}
	initMaterialTransactionStatusChangesListFlags(cmd)
	return cmd
}

func init() {
	materialTransactionStatusChangesCmd.AddCommand(newMaterialTransactionStatusChangesListCmd())
}

func initMaterialTransactionStatusChangesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("material-transaction", "", "Filter by material transaction ID")
	cmd.Flags().String("status", "", "Filter by status value")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by presence of created-at (true/false)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("is-updated-at", "", "Filter by presence of updated-at (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialTransactionStatusChangesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMaterialTransactionStatusChangesListOptions(cmd)
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
	query.Set("fields[material-transaction-status-changes]", "status,changed-at,comment,created-at,updated-at,material-transaction,changed-by")
	query.Set("fields[material-transactions]", "ticket-number,transaction-at")
	query.Set("fields[users]", "name")
	query.Set("include", "changed-by,material-transaction")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[material_transaction]", opts.MaterialTransaction)
	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[created_at_min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created_at_max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[is_created_at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[updated_at_min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated_at_max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[is_updated_at]", opts.IsUpdatedAt)

	body, _, err := client.Get(cmd.Context(), "/v1/material-transaction-status-changes", query)
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

	rows := buildMaterialTransactionStatusChangeRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderMaterialTransactionStatusChangesTable(cmd, rows)
}

func parseMaterialTransactionStatusChangesListOptions(cmd *cobra.Command) (materialTransactionStatusChangesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	materialTransaction, _ := cmd.Flags().GetString("material-transaction")
	status, _ := cmd.Flags().GetString("status")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	isCreatedAt, _ := cmd.Flags().GetString("is-created-at")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	isUpdatedAt, _ := cmd.Flags().GetString("is-updated-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialTransactionStatusChangesListOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		NoAuth:              noAuth,
		Limit:               limit,
		Offset:              offset,
		Sort:                sort,
		MaterialTransaction: materialTransaction,
		Status:              status,
		CreatedAtMin:        createdAtMin,
		CreatedAtMax:        createdAtMax,
		IsCreatedAt:         isCreatedAt,
		UpdatedAtMin:        updatedAtMin,
		UpdatedAtMax:        updatedAtMax,
		IsUpdatedAt:         isUpdatedAt,
	}, nil
}

func buildMaterialTransactionStatusChangeRows(resp jsonAPIResponse) []materialTransactionStatusChangeRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]materialTransactionStatusChangeRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildMaterialTransactionStatusChangeRow(resource, included))
	}

	return rows
}

func buildMaterialTransactionStatusChangeRow(resource jsonAPIResource, included map[string]jsonAPIResource) materialTransactionStatusChangeRow {
	attrs := resource.Attributes
	row := materialTransactionStatusChangeRow{
		ID:        resource.ID,
		Status:    strings.TrimSpace(stringAttr(attrs, "status")),
		ChangedAt: formatDateTime(stringAttr(attrs, "changed-at")),
		Comment:   stringAttr(attrs, "comment"),
		CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt: formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resource.Relationships["material-transaction"]; ok && rel.Data != nil {
		row.MaterialTransactionID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["changed-by"]; ok && rel.Data != nil {
		row.ChangedByID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.ChangedByName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
		}
	}

	return row
}

func renderMaterialTransactionStatusChangesTable(cmd *cobra.Command, rows []materialTransactionStatusChangeRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No material transaction status changes found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tMTXN\tSTATUS\tCHANGED AT\tCHANGED BY\tCOMMENT")
	for _, row := range rows {
		changedBy := firstNonEmpty(row.ChangedByName, row.ChangedByID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.MaterialTransactionID,
			row.Status,
			row.ChangedAt,
			truncateString(changedBy, 18),
			truncateString(row.Comment, 32),
		)
	}

	return writer.Flush()
}
