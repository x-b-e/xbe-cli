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

type tenderStatusChangesListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Sort    string
	Tender  string
	Status  string
}

type tenderStatusChangeRow struct {
	ID          string `json:"id"`
	TenderID    string `json:"tender_id,omitempty"`
	Status      string `json:"status,omitempty"`
	ChangedAt   string `json:"changed_at,omitempty"`
	ChangedByID string `json:"changed_by_id,omitempty"`
	Comment     string `json:"comment,omitempty"`
}

func newTenderStatusChangesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tender status changes",
		Long: `List tender status changes with filtering and pagination.

Output Columns:
  ID          Status change identifier
  TENDER      Tender ID
  STATUS      Tender status
  CHANGED AT  Status change timestamp
  CHANGED BY  User who changed the status
  COMMENT     Status change comment

Pagination:
  Use --limit and --offset to paginate through large result sets.

Filters:
  --tender   Filter by tender ID
  --status   Filter by status (accepted, cancelled, editing, expired, offered, rejected, returned, sourced)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List tender status changes
  xbe view tender-status-changes list

  # Filter by tender
  xbe view tender-status-changes list --tender 123

  # Filter by status
  xbe view tender-status-changes list --status accepted

  # Output as JSON
  xbe view tender-status-changes list --json`,
		Args: cobra.NoArgs,
		RunE: runTenderStatusChangesList,
	}
	initTenderStatusChangesListFlags(cmd)
	return cmd
}

func init() {
	tenderStatusChangesCmd.AddCommand(newTenderStatusChangesListCmd())
}

func initTenderStatusChangesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("tender", "", "Filter by tender ID")
	cmd.Flags().String("status", "", "Filter by status")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTenderStatusChangesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTenderStatusChangesListOptions(cmd)
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
	query.Set("fields[tender-status-changes]", "tender,status,changed-at,comment,changed-by")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[tender]", opts.Tender)
	setFilterIfPresent(query, "filter[status]", opts.Status)

	body, _, err := client.Get(cmd.Context(), "/v1/tender-status-changes", query)
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

	rows := buildTenderStatusChangeRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTenderStatusChangesTable(cmd, rows)
}

func parseTenderStatusChangesListOptions(cmd *cobra.Command) (tenderStatusChangesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	tender, _ := cmd.Flags().GetString("tender")
	status, _ := cmd.Flags().GetString("status")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return tenderStatusChangesListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
		Tender:  tender,
		Status:  status,
	}, nil
}

func buildTenderStatusChangeRows(resp jsonAPIResponse) []tenderStatusChangeRow {
	rows := make([]tenderStatusChangeRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := tenderStatusChangeRow{
			ID:        resource.ID,
			Status:    stringAttr(resource.Attributes, "status"),
			ChangedAt: formatDateTime(stringAttr(resource.Attributes, "changed-at")),
			Comment:   stringAttr(resource.Attributes, "comment"),
		}
		if rel, ok := resource.Relationships["tender"]; ok && rel.Data != nil {
			row.TenderID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["changed-by"]; ok && rel.Data != nil {
			row.ChangedByID = rel.Data.ID
		}
		rows = append(rows, row)
	}
	return rows
}

func renderTenderStatusChangesTable(cmd *cobra.Command, rows []tenderStatusChangeRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No tender status changes found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTENDER\tSTATUS\tCHANGED AT\tCHANGED BY\tCOMMENT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.TenderID,
			row.Status,
			row.ChangedAt,
			row.ChangedByID,
			truncateString(row.Comment, 40),
		)
	}
	return writer.Flush()
}
