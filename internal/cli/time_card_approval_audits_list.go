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

type timeCardApprovalAuditsListOptions struct {
	BaseURL    string
	Token      string
	JSON       bool
	NoAuth     bool
	Limit      int
	Offset     int
	Sort       string
	TimeCardID string
	UserID     string
}

type timeCardApprovalAuditRow struct {
	ID         string `json:"id"`
	TimeCardID string `json:"time_card_id,omitempty"`
	UserID     string `json:"user_id,omitempty"`
	IsBot      bool   `json:"is_bot"`
	Note       string `json:"note,omitempty"`
	CreatedAt  string `json:"created_at,omitempty"`
}

func newTimeCardApprovalAuditsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List time card approval audits",
		Long: `List time card approval audits.

Output Columns:
  ID         Audit identifier
  TIME CARD  Time card ID
  USER       User ID (blank when bot)
  BOT        Whether audit was created by a bot
  NOTE       Audit note
  CREATED AT Audit creation timestamp

Filters:
  --time-card  Filter by time card ID
  --user       Filter by user ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List time card approval audits
  xbe view time-card-approval-audits list

  # Filter by time card
  xbe view time-card-approval-audits list --time-card 123

  # Filter by user
  xbe view time-card-approval-audits list --user 456

  # Output as JSON
  xbe view time-card-approval-audits list --json`,
		Args: cobra.NoArgs,
		RunE: runTimeCardApprovalAuditsList,
	}
	initTimeCardApprovalAuditsListFlags(cmd)
	return cmd
}

func init() {
	timeCardApprovalAuditsCmd.AddCommand(newTimeCardApprovalAuditsListCmd())
}

func initTimeCardApprovalAuditsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("time-card", "", "Filter by time card ID")
	cmd.Flags().String("user", "", "Filter by user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTimeCardApprovalAuditsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTimeCardApprovalAuditsListOptions(cmd)
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
	query.Set("fields[time-card-approval-audits]", "note,is-bot,created-at,time-card,user")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[time_card]", opts.TimeCardID)
	setFilterIfPresent(query, "filter[user]", opts.UserID)

	body, _, err := client.Get(cmd.Context(), "/v1/time-card-approval-audits", query)
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

	rows := buildTimeCardApprovalAuditRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTimeCardApprovalAuditsTable(cmd, rows)
}

func parseTimeCardApprovalAuditsListOptions(cmd *cobra.Command) (timeCardApprovalAuditsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	timeCardID, _ := cmd.Flags().GetString("time-card")
	userID, _ := cmd.Flags().GetString("user")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return timeCardApprovalAuditsListOptions{
		BaseURL:    baseURL,
		Token:      token,
		JSON:       jsonOut,
		NoAuth:     noAuth,
		Limit:      limit,
		Offset:     offset,
		Sort:       sort,
		TimeCardID: timeCardID,
		UserID:     userID,
	}, nil
}

func buildTimeCardApprovalAuditRows(resp jsonAPIResponse) []timeCardApprovalAuditRow {
	rows := make([]timeCardApprovalAuditRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := timeCardApprovalAuditRow{
			ID:        resource.ID,
			IsBot:     boolAttr(attrs, "is-bot"),
			Note:      stringAttr(attrs, "note"),
			CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
		}

		if rel, ok := resource.Relationships["time-card"]; ok && rel.Data != nil {
			row.TimeCardID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
			row.UserID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func buildTimeCardApprovalAuditRowFromSingle(resp jsonAPISingleResponse) timeCardApprovalAuditRow {
	resource := resp.Data
	attrs := resource.Attributes
	row := timeCardApprovalAuditRow{
		ID:        resource.ID,
		IsBot:     boolAttr(attrs, "is-bot"),
		Note:      stringAttr(attrs, "note"),
		CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
	}

	if rel, ok := resource.Relationships["time-card"]; ok && rel.Data != nil {
		row.TimeCardID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
		row.UserID = rel.Data.ID
	}

	return row
}

func renderTimeCardApprovalAuditsTable(cmd *cobra.Command, rows []timeCardApprovalAuditRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No time card approval audits found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTIME CARD\tUSER\tBOT\tNOTE\tCREATED AT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.TimeCardID,
			row.UserID,
			boolToYesNo(row.IsBot),
			truncateString(row.Note, 30),
			row.CreatedAt,
		)
	}
	return writer.Flush()
}
