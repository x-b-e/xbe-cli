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

type textMessagesListOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	NoAuth         bool
	To             string
	From           string
	DateSent       string
	DateSentAfter  string
	DateSentBefore string
	MaxMessages    int
	PageSize       int
	Limit          int
}

type textMessageRow struct {
	ID        string `json:"id"`
	Status    string `json:"status,omitempty"`
	Direction string `json:"direction,omitempty"`
	From      string `json:"from,omitempty"`
	To        string `json:"to,omitempty"`
	DateSent  string `json:"date_sent,omitempty"`
	Body      string `json:"body,omitempty"`
}

func newTextMessagesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List text messages",
		Long: `List text messages with filtering.

Text message listing defaults to messages sent today and requires an admin user.

Output Columns:
  ID         Text message SID
  STATUS     Twilio delivery status
  DIRECTION  Message direction (inbound/outbound)
  FROM       Sender phone number
  TO         Recipient phone number
  SENT AT    When the message was sent
  BODY       Message body

Filters:
  --to                Filter by recipient phone number
  --from              Filter by sender phone number
  --date-sent         Filter by date sent (YYYY-MM-DD)
  --date-sent-after   Filter by date sent on/after (YYYY-MM-DD)
  --date-sent-before  Filter by date sent on/before (YYYY-MM-DD)
  --max-messages      Max number of messages to return
  --page-size         Page size for Twilio requests

Global flags (see xbe --help): --json, --limit, --base-url, --token, --no-auth`,
		Example: `  # List today's text messages
  xbe view text-messages list

  # Filter by recipient and date range
  xbe view text-messages list --to +15551234567 --date-sent-after 2025-01-01

  # Limit results
  xbe view text-messages list --max-messages 10

  # Output as JSON
  xbe view text-messages list --json`,
		Args: cobra.NoArgs,
		RunE: runTextMessagesList,
	}
	initTextMessagesListFlags(cmd)
	return cmd
}

func init() {
	textMessagesCmd.AddCommand(newTextMessagesListCmd())
}

func initTextMessagesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 0, "Max messages to return (alias for --max-messages)")
	cmd.Flags().String("to", "", "Filter by recipient phone number")
	cmd.Flags().String("from", "", "Filter by sender phone number")
	cmd.Flags().String("date-sent", "", "Filter by date sent (YYYY-MM-DD)")
	cmd.Flags().String("date-sent-after", "", "Filter by date sent on/after (YYYY-MM-DD)")
	cmd.Flags().String("date-sent-before", "", "Filter by date sent on/before (YYYY-MM-DD)")
	cmd.Flags().Int("max-messages", 0, "Max number of messages to return")
	cmd.Flags().Int("page-size", 0, "Page size for Twilio requests")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTextMessagesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTextMessagesListOptions(cmd)
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
	query.Set("fields[text-messages]", "status,direction,from,to,date-sent,body")

	setFilterIfPresent(query, "filter[to]", opts.To)
	setFilterIfPresent(query, "filter[from]", opts.From)
	setFilterIfPresent(query, "filter[date-sent]", opts.DateSent)
	setFilterIfPresent(query, "filter[date-sent-after]", opts.DateSentAfter)
	setFilterIfPresent(query, "filter[date-sent-before]", opts.DateSentBefore)

	maxMessages := opts.MaxMessages
	if maxMessages == 0 {
		maxMessages = opts.Limit
	}
	if maxMessages > 0 {
		query.Set("filter[max-messages]", strconv.Itoa(maxMessages))
	}
	if opts.PageSize > 0 {
		query.Set("filter[page-size]", strconv.Itoa(opts.PageSize))
	}

	body, _, err := client.Get(cmd.Context(), "/v1/text-messages", query)
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

	rows := buildTextMessageRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTextMessagesTable(cmd, rows)
}

func parseTextMessagesListOptions(cmd *cobra.Command) (textMessagesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	to, _ := cmd.Flags().GetString("to")
	from, _ := cmd.Flags().GetString("from")
	dateSent, _ := cmd.Flags().GetString("date-sent")
	dateSentAfter, _ := cmd.Flags().GetString("date-sent-after")
	dateSentBefore, _ := cmd.Flags().GetString("date-sent-before")
	maxMessages, _ := cmd.Flags().GetInt("max-messages")
	pageSize, _ := cmd.Flags().GetInt("page-size")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return textMessagesListOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		NoAuth:         noAuth,
		To:             to,
		From:           from,
		DateSent:       dateSent,
		DateSentAfter:  dateSentAfter,
		DateSentBefore: dateSentBefore,
		MaxMessages:    maxMessages,
		PageSize:       pageSize,
		Limit:          limit,
	}, nil
}

func buildTextMessageRows(resp jsonAPIResponse) []textMessageRow {
	rows := make([]textMessageRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := textMessageRow{
			ID:        resource.ID,
			Status:    stringAttr(attrs, "status"),
			Direction: stringAttr(attrs, "direction"),
			From:      stringAttr(attrs, "from"),
			To:        stringAttr(attrs, "to"),
			DateSent:  formatDateTime(stringAttr(attrs, "date-sent")),
			Body:      stringAttr(attrs, "body"),
		}
		rows = append(rows, row)
	}
	return rows
}

func renderTextMessagesTable(cmd *cobra.Command, rows []textMessageRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No text messages found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tDIRECTION\tFROM\tTO\tSENT AT\tBODY")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Status, 18),
			truncateString(row.Direction, 14),
			truncateString(row.From, 18),
			truncateString(row.To, 18),
			truncateString(row.DateSent, 20),
			truncateString(row.Body, 40),
		)
	}
	return writer.Flush()
}
