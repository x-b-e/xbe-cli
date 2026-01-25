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

type tenderReturnsListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Sort    string
}

type tenderReturnRow struct {
	ID         string `json:"id"`
	TenderID   string `json:"tender_id,omitempty"`
	TenderType string `json:"tender_type,omitempty"`
	Comment    string `json:"comment,omitempty"`
}

func newTenderReturnsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tender returns",
		Long: `List tender returns.

Output Columns:
  ID       Tender return identifier
  TENDER   Tender type and ID
  COMMENT  Return comment

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List tender returns
  xbe view tender-returns list

  # Output as JSON
  xbe view tender-returns list --json`,
		Args: cobra.NoArgs,
		RunE: runTenderReturnsList,
	}
	initTenderReturnsListFlags(cmd)
	return cmd
}

func init() {
	tenderReturnsCmd.AddCommand(newTenderReturnsListCmd())
}

func initTenderReturnsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 0, "Page size (defaults to server default)")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTenderReturnsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTenderReturnsListOptions(cmd)
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
	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	body, _, err := client.Get(cmd.Context(), "/v1/tender-returns", query)
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

	rows := buildTenderReturnRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTenderReturnsTable(cmd, rows)
}

func parseTenderReturnsListOptions(cmd *cobra.Command) (tenderReturnsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return tenderReturnsListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
	}, nil
}

func buildTenderReturnRows(resp jsonAPIResponse) []tenderReturnRow {
	rows := make([]tenderReturnRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := tenderReturnRow{
			ID:      resource.ID,
			Comment: stringAttr(resource.Attributes, "comment"),
		}

		if rel, ok := resource.Relationships["tender"]; ok && rel.Data != nil {
			row.TenderID = rel.Data.ID
			row.TenderType = rel.Data.Type
		}

		rows = append(rows, row)
	}
	return rows
}

func renderTenderReturnsTable(cmd *cobra.Command, rows []tenderReturnRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No tender returns found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTENDER\tCOMMENT")
	for _, row := range rows {
		tender := ""
		if row.TenderType != "" && row.TenderID != "" {
			tender = row.TenderType + "/" + row.TenderID
		} else if row.TenderID != "" {
			tender = row.TenderID
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\n",
			row.ID,
			truncateString(tender, 40),
			truncateString(row.Comment, 40),
		)
	}
	return writer.Flush()
}
