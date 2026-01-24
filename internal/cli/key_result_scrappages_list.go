package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type keyResultScrappagesListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Sort    string
}

type keyResultScrappageRow struct {
	ID          string `json:"id"`
	KeyResultID string `json:"key_result_id,omitempty"`
	Comment     string `json:"comment,omitempty"`
}

func newKeyResultScrappagesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List key result scrappages",
		Long: `List key result scrappages.

Output Columns:
  ID         Scrappage identifier
  KEY RESULT Key result ID
  COMMENT    Comment (truncated)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List key result scrappages
  xbe view key-result-scrappages list

  # JSON output
  xbe view key-result-scrappages list --json`,
		Args: cobra.NoArgs,
		RunE: runKeyResultScrappagesList,
	}
	initKeyResultScrappagesListFlags(cmd)
	return cmd
}

func init() {
	keyResultScrappagesCmd.AddCommand(newKeyResultScrappagesListCmd())
}

func initKeyResultScrappagesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runKeyResultScrappagesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseKeyResultScrappagesListOptions(cmd)
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
	query.Set("fields[key-result-scrappages]", "key-result,comment")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	body, status, err := client.Get(cmd.Context(), "/v1/key-result-scrappages", query)
	if err != nil {
		if status == http.StatusNotFound {
			return renderKeyResultScrappagesUnavailable(cmd, opts.JSON)
		}
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

	rows := buildKeyResultScrappageRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderKeyResultScrappagesTable(cmd, rows)
}

func renderKeyResultScrappagesUnavailable(cmd *cobra.Command, jsonOut bool) error {
	if jsonOut {
		return writeJSON(cmd.OutOrStdout(), []keyResultScrappageRow{})
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Key result scrappages are write-only; list is not available.")
	return nil
}

func parseKeyResultScrappagesListOptions(cmd *cobra.Command) (keyResultScrappagesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return keyResultScrappagesListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
	}, nil
}

func buildKeyResultScrappageRows(resp jsonAPIResponse) []keyResultScrappageRow {
	rows := make([]keyResultScrappageRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := buildKeyResultScrappageRow(resource)
		rows = append(rows, row)
	}
	return rows
}

func buildKeyResultScrappageRow(resource jsonAPIResource) keyResultScrappageRow {
	attrs := resource.Attributes
	row := keyResultScrappageRow{
		ID:      resource.ID,
		Comment: strings.TrimSpace(stringAttr(attrs, "comment")),
	}

	if rel, ok := resource.Relationships["key-result"]; ok && rel.Data != nil {
		row.KeyResultID = rel.Data.ID
	}

	return row
}

func renderKeyResultScrappagesTable(cmd *cobra.Command, rows []keyResultScrappageRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No key result scrappages found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tKEY RESULT\tCOMMENT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\n",
			row.ID,
			row.KeyResultID,
			truncateString(row.Comment, 40),
		)
	}
	return writer.Flush()
}
