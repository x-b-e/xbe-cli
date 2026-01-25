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

type invoiceSendsListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Sort    string
}

type invoiceSendRow struct {
	ID        string `json:"id"`
	InvoiceID string `json:"invoice_id,omitempty"`
	Comment   string `json:"comment,omitempty"`
}

func newInvoiceSendsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List invoice sends",
		Long: `List invoice sends.

Output Columns:
  ID       Send identifier
  INVOICE  Invoice ID
  COMMENT  Send comment (if present)

Filters:
  None

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List sends
  xbe view invoice-sends list

  # Paginate results
  xbe view invoice-sends list --limit 25 --offset 50

  # Output as JSON
  xbe view invoice-sends list --json`,
		Args: cobra.NoArgs,
		RunE: runInvoiceSendsList,
	}
	initInvoiceSendsListFlags(cmd)
	return cmd
}

func init() {
	invoiceSendsCmd.AddCommand(newInvoiceSendsListCmd())
}

func initInvoiceSendsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runInvoiceSendsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseInvoiceSendsListOptions(cmd)
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
	query.Set("fields[invoice-sends]", "comment,invoice")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	body, _, err := client.Get(cmd.Context(), "/v1/invoice-sends", query)
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

	rows := buildInvoiceSendRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderInvoiceSendsTable(cmd, rows)
}

func parseInvoiceSendsListOptions(cmd *cobra.Command) (invoiceSendsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return invoiceSendsListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
	}, nil
}

func buildInvoiceSendRows(resp jsonAPIResponse) []invoiceSendRow {
	rows := make([]invoiceSendRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := invoiceSendRow{
			ID:      resource.ID,
			Comment: stringAttr(attrs, "comment"),
		}

		if rel, ok := resource.Relationships["invoice"]; ok && rel.Data != nil {
			row.InvoiceID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func buildInvoiceSendRowFromSingle(resp jsonAPISingleResponse) invoiceSendRow {
	resource := resp.Data
	attrs := resource.Attributes
	row := invoiceSendRow{
		ID:      resource.ID,
		Comment: stringAttr(attrs, "comment"),
	}

	if rel, ok := resource.Relationships["invoice"]; ok && rel.Data != nil {
		row.InvoiceID = rel.Data.ID
	}

	return row
}

func renderInvoiceSendsTable(cmd *cobra.Command, rows []invoiceSendRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No invoice sends found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tINVOICE\tCOMMENT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\n",
			row.ID,
			row.InvoiceID,
			truncateString(row.Comment, 50),
		)
	}
	return writer.Flush()
}
