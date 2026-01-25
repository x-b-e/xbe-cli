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

type actionItemKeyResultsListOptions struct {
	BaseURL    string
	Token      string
	JSON       bool
	NoAuth     bool
	Limit      int
	Offset     int
	Sort       string
	ActionItem string
	KeyResult  string
}

type actionItemKeyResultRow struct {
	ID              string `json:"id"`
	ActionItemID    string `json:"action_item_id,omitempty"`
	ActionItemTitle string `json:"action_item_title,omitempty"`
	KeyResultID     string `json:"key_result_id,omitempty"`
	KeyResultTitle  string `json:"key_result_title,omitempty"`
}

func newActionItemKeyResultsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List action item key result links",
		Long: `List action item key result links with filtering and pagination.

Output Columns:
  ID           Link identifier
  ACTION ITEM  Action item title
  KEY RESULT   Key result title

Filters:
  --action-item  Filter by action item ID
  --key-result   Filter by key result ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List action item key results
  xbe view action-item-key-results list

  # Filter by action item
  xbe view action-item-key-results list --action-item 123

  # Filter by key result
  xbe view action-item-key-results list --key-result 456

  # JSON output
  xbe view action-item-key-results list --json`,
		Args: cobra.NoArgs,
		RunE: runActionItemKeyResultsList,
	}
	initActionItemKeyResultsListFlags(cmd)
	return cmd
}

func init() {
	actionItemKeyResultsCmd.AddCommand(newActionItemKeyResultsListCmd())
}

func initActionItemKeyResultsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("action-item", "", "Filter by action item ID")
	cmd.Flags().String("key-result", "", "Filter by key result ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runActionItemKeyResultsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseActionItemKeyResultsListOptions(cmd)
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
	query.Set("fields[action-item-key-results]", "action-item,key-result")
	query.Set("fields[action-items]", "title")
	query.Set("fields[key-results]", "title")
	query.Set("include", "action-item,key-result")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[action-item]", opts.ActionItem)
	setFilterIfPresent(query, "filter[key-result]", opts.KeyResult)

	body, _, err := client.Get(cmd.Context(), "/v1/action-item-key-results", query)
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

	rows := buildActionItemKeyResultRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderActionItemKeyResultsTable(cmd, rows)
}

func parseActionItemKeyResultsListOptions(cmd *cobra.Command) (actionItemKeyResultsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	actionItem, _ := cmd.Flags().GetString("action-item")
	keyResult, _ := cmd.Flags().GetString("key-result")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return actionItemKeyResultsListOptions{
		BaseURL:    baseURL,
		Token:      token,
		JSON:       jsonOut,
		NoAuth:     noAuth,
		Limit:      limit,
		Offset:     offset,
		Sort:       sort,
		ActionItem: actionItem,
		KeyResult:  keyResult,
	}, nil
}

func buildActionItemKeyResultRows(resp jsonAPIResponse) []actionItemKeyResultRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]actionItemKeyResultRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := actionItemKeyResultRow{
			ID: resource.ID,
		}

		if rel, ok := resource.Relationships["action-item"]; ok && rel.Data != nil {
			row.ActionItemID = rel.Data.ID
			if actionItem, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.ActionItemTitle = strings.TrimSpace(stringAttr(actionItem.Attributes, "title"))
			}
		}

		if rel, ok := resource.Relationships["key-result"]; ok && rel.Data != nil {
			row.KeyResultID = rel.Data.ID
			if keyResult, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.KeyResultTitle = strings.TrimSpace(stringAttr(keyResult.Attributes, "title"))
			}
		}

		rows = append(rows, row)
	}

	return rows
}

func renderActionItemKeyResultsTable(cmd *cobra.Command, rows []actionItemKeyResultRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No action item key results found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tACTION ITEM\tKEY RESULT")
	for _, row := range rows {
		actionItemDisplay := firstNonEmpty(row.ActionItemTitle, row.ActionItemID)
		keyResultDisplay := firstNonEmpty(row.KeyResultTitle, row.KeyResultID)
		fmt.Fprintf(writer, "%s\t%s\t%s\n",
			row.ID,
			truncateString(actionItemDisplay, 40),
			truncateString(keyResultDisplay, 40),
		)
	}
	return writer.Flush()
}
