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

type materialTransactionFieldScopesListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
}

type materialTransactionFieldScopeRow struct {
	ID            string `json:"id"`
	TicketNumber  string `json:"ticket_number,omitempty"`
	TransactionAt string `json:"transaction_at,omitempty"`
}

func newMaterialTransactionFieldScopesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List material transaction field scopes",
		Long: `List material transaction field scopes.

Material transaction field scopes are primarily used for diagnostics and
matching context. Listing is restricted to admins.

Output Columns:
  ID             Field scope identifier (material transaction ID)
  TICKET         Ticket number
  TRANSACTION AT Transaction timestamp

Global flags (see xbe --help): --json, --no-auth, --limit, --offset, --base-url, --token`,
		Example: `  # List material transaction field scopes (admin only)
  xbe view material-transaction-field-scopes list

  # Output as JSON
  xbe view material-transaction-field-scopes list --json`,
		RunE: runMaterialTransactionFieldScopesList,
	}
	initMaterialTransactionFieldScopesListFlags(cmd)
	return cmd
}

func init() {
	materialTransactionFieldScopesCmd.AddCommand(newMaterialTransactionFieldScopesListCmd())
}

func initMaterialTransactionFieldScopesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialTransactionFieldScopesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMaterialTransactionFieldScopesListOptions(cmd)
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
	query.Set("fields[material-transaction-field-scopes]", "ticket-number,transaction-at")
	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	body, _, err := client.Get(cmd.Context(), "/v1/material-transaction-field-scopes", query)
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

	rows := buildMaterialTransactionFieldScopeRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderMaterialTransactionFieldScopeTable(cmd, rows)
}

func parseMaterialTransactionFieldScopesListOptions(cmd *cobra.Command) (materialTransactionFieldScopesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialTransactionFieldScopesListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
	}, nil
}

func buildMaterialTransactionFieldScopeRows(resp jsonAPIResponse) []materialTransactionFieldScopeRow {
	rows := make([]materialTransactionFieldScopeRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, materialTransactionFieldScopeRow{
			ID:            resource.ID,
			TicketNumber:  stringAttr(resource.Attributes, "ticket-number"),
			TransactionAt: formatDateTime(stringAttr(resource.Attributes, "transaction-at")),
		})
	}
	return rows
}

func renderMaterialTransactionFieldScopeTable(cmd *cobra.Command, rows []materialTransactionFieldScopeRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No material transaction field scopes found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTICKET\tTRANSACTION AT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\n", row.ID, row.TicketNumber, row.TransactionAt)
	}
	return writer.Flush()
}
