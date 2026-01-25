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

type tenderRejectionsListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Sort    string
}

type tenderRejectionRow struct {
	ID         string `json:"id"`
	TenderType string `json:"tender_type,omitempty"`
	TenderID   string `json:"tender_id,omitempty"`
	Comment    string `json:"comment,omitempty"`
}

func newTenderRejectionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tender rejections",
		Long: `List tender rejections.

Output Columns:
  ID       Rejection identifier
  TENDER   Tender (type/id)
  COMMENT  Rejection comment (if present)

Filters:
  None

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List rejections
  xbe view tender-rejections list

  # Paginate results
  xbe view tender-rejections list --limit 25 --offset 50

  # Output as JSON
  xbe view tender-rejections list --json`,
		Args: cobra.NoArgs,
		RunE: runTenderRejectionsList,
	}
	initTenderRejectionsListFlags(cmd)
	return cmd
}

func init() {
	tenderRejectionsCmd.AddCommand(newTenderRejectionsListCmd())
}

func initTenderRejectionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTenderRejectionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTenderRejectionsListOptions(cmd)
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
	query.Set("fields[tender-rejections]", "comment,tender")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	body, _, err := client.Get(cmd.Context(), "/v1/tender-rejections", query)
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

	rows := buildTenderRejectionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTenderRejectionsTable(cmd, rows)
}

func parseTenderRejectionsListOptions(cmd *cobra.Command) (tenderRejectionsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return tenderRejectionsListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
	}, nil
}

func buildTenderRejectionRows(resp jsonAPIResponse) []tenderRejectionRow {
	rows := make([]tenderRejectionRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildTenderRejectionRow(resource))
	}
	return rows
}

func buildTenderRejectionRow(resource jsonAPIResource) tenderRejectionRow {
	attrs := resource.Attributes
	row := tenderRejectionRow{
		ID:      resource.ID,
		Comment: stringAttr(attrs, "comment"),
	}

	if rel, ok := resource.Relationships["tender"]; ok && rel.Data != nil {
		row.TenderType = rel.Data.Type
		row.TenderID = rel.Data.ID
	}

	return row
}

func buildTenderRejectionRowFromSingle(resp jsonAPISingleResponse) tenderRejectionRow {
	return buildTenderRejectionRow(resp.Data)
}

func renderTenderRejectionsTable(cmd *cobra.Command, rows []tenderRejectionRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No tender rejections found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTENDER\tCOMMENT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\n",
			row.ID,
			truncateString(formatTypeID(row.TenderType, row.TenderID), 32),
			truncateString(row.Comment, 50),
		)
	}
	return writer.Flush()
}
