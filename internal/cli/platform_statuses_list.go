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

type platformStatusesListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
}

func newPlatformStatusesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List platform statuses",
		Long: `List platform status updates.

Platform statuses communicate incidents, maintenance windows, and other
service updates.

Output Columns:
  ID         Platform status identifier
  TITLE      Status title
  PUBLISHED  Published timestamp
  START      Start timestamp
  END        End timestamp`,
		Example: `  # List platform statuses
  xbe view platform-statuses list

  # Paginate results
  xbe view platform-statuses list --limit 10 --offset 10

  # Output as JSON
  xbe view platform-statuses list --json`,
		RunE: runPlatformStatusesList,
	}
	initPlatformStatusesListFlags(cmd)
	return cmd
}

func init() {
	platformStatusesCmd.AddCommand(newPlatformStatusesListCmd())
}

func initPlatformStatusesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPlatformStatusesList(cmd *cobra.Command, _ []string) error {
	opts, err := parsePlatformStatusesListOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("sort", "-published-at")
	query.Set("fields[platform-statuses]", "title,description,published-at,start-at,end-at")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	body, _, err := client.Get(cmd.Context(), "/v1/platform-statuses", query)
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

	rows := buildPlatformStatusRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderPlatformStatusesTable(cmd, rows)
}

func parsePlatformStatusesListOptions(cmd *cobra.Command) (platformStatusesListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return platformStatusesListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return platformStatusesListOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return platformStatusesListOptions{}, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return platformStatusesListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return platformStatusesListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return platformStatusesListOptions{}, err
	}

	return platformStatusesListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
	}, nil
}

func renderPlatformStatusesTable(cmd *cobra.Command, rows []platformStatusRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No platform statuses found.")
		return nil
	}

	const titleMax = 48

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, 32, 0)
	fmt.Fprintln(writer, "ID\tTITLE\tPUBLISHED\tSTART\tEND")

	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Title, titleMax),
			row.PublishedAt,
			row.StartAt,
			row.EndAt,
		)
	}

	return writer.Flush()
}

type platformStatusRow struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	PublishedAt string `json:"published_at,omitempty"`
	StartAt     string `json:"start_at,omitempty"`
	EndAt       string `json:"end_at,omitempty"`
}

func buildPlatformStatusRows(resp jsonAPIResponse) []platformStatusRow {
	rows := make([]platformStatusRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, platformStatusRowFromResource(resource))
	}
	return rows
}

func platformStatusRowFromSingle(resp jsonAPISingleResponse) platformStatusRow {
	return platformStatusRowFromResource(resp.Data)
}

func platformStatusRowFromResource(resource jsonAPIResource) platformStatusRow {
	attrs := resource.Attributes

	return platformStatusRow{
		ID:          resource.ID,
		Title:       strings.TrimSpace(stringAttr(attrs, "title")),
		Description: strings.TrimSpace(stringAttr(attrs, "description")),
		PublishedAt: formatDateTime(stringAttr(attrs, "published-at")),
		StartAt:     formatDateTime(stringAttr(attrs, "start-at")),
		EndAt:       formatDateTime(stringAttr(attrs, "end-at")),
	}
}
