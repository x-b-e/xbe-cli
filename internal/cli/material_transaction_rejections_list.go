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

type materialTransactionRejectionsListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Sort    string
}

type materialTransactionRejectionRow struct {
	ID                    string `json:"id"`
	MaterialTransactionID string `json:"material_transaction_id,omitempty"`
	Comment               string `json:"comment,omitempty"`
}

func newMaterialTransactionRejectionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List material transaction rejections",
		Long: `List material transaction rejections.

Output Columns:
  ID                   Rejection identifier
  MATERIAL TRANSACTION Material transaction ID
  COMMENT              Rejection comment (if present)

Filters:
  None

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List rejections
  xbe view material-transaction-rejections list

  # Paginate results
  xbe view material-transaction-rejections list --limit 25 --offset 50

  # Output as JSON
  xbe view material-transaction-rejections list --json`,
		Args: cobra.NoArgs,
		RunE: runMaterialTransactionRejectionsList,
	}
	initMaterialTransactionRejectionsListFlags(cmd)
	return cmd
}

func init() {
	materialTransactionRejectionsCmd.AddCommand(newMaterialTransactionRejectionsListCmd())
}

func initMaterialTransactionRejectionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialTransactionRejectionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMaterialTransactionRejectionsListOptions(cmd)
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
	query.Set("fields[material-transaction-rejections]", "comment,material-transaction")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	body, _, err := client.Get(cmd.Context(), "/v1/material-transaction-rejections", query)
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

	rows := buildMaterialTransactionRejectionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderMaterialTransactionRejectionsTable(cmd, rows)
}

func parseMaterialTransactionRejectionsListOptions(cmd *cobra.Command) (materialTransactionRejectionsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialTransactionRejectionsListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
	}, nil
}

func buildMaterialTransactionRejectionRows(resp jsonAPIResponse) []materialTransactionRejectionRow {
	rows := make([]materialTransactionRejectionRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := materialTransactionRejectionRow{
			ID:      resource.ID,
			Comment: stringAttr(attrs, "comment"),
		}

		if rel, ok := resource.Relationships["material-transaction"]; ok && rel.Data != nil {
			row.MaterialTransactionID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func buildMaterialTransactionRejectionRowFromSingle(resp jsonAPISingleResponse) materialTransactionRejectionRow {
	resource := resp.Data
	attrs := resource.Attributes
	row := materialTransactionRejectionRow{
		ID:      resource.ID,
		Comment: stringAttr(attrs, "comment"),
	}

	if rel, ok := resource.Relationships["material-transaction"]; ok && rel.Data != nil {
		row.MaterialTransactionID = rel.Data.ID
	}

	return row
}

func renderMaterialTransactionRejectionsTable(cmd *cobra.Command, rows []materialTransactionRejectionRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No material transaction rejections found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tMATERIAL TRANSACTION\tCOMMENT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\n",
			row.ID,
			row.MaterialTransactionID,
			truncateString(row.Comment, 50),
		)
	}
	return writer.Flush()
}
