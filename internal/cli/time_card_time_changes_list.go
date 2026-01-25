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

type timeCardTimeChangesListOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	NoAuth      bool
	Limit       int
	Offset      int
	Sort        string
	TimeCard    string
	CreatedBy   string
	IsProcessed string
	Broker      string
}

type timeCardTimeChangeRow struct {
	ID          string `json:"id"`
	TimeCardID  string `json:"time_card_id,omitempty"`
	CreatedByID string `json:"created_by_id,omitempty"`
	IsProcessed bool   `json:"is_processed"`
	Comment     string `json:"comment,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
}

func newTimeCardTimeChangesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List time card time changes",
		Long: `List time card time changes with filtering and pagination.

Output Columns:
  ID          Time card time change identifier
  TIME_CARD   Time card ID
  CREATED_BY  Created by user ID
  PROCESSED   Whether the change has been processed
  COMMENT     Comment preview
  CREATED_AT  Creation timestamp

Filters:
  --time-card     Filter by time card ID (comma-separated for multiple)
  --created-by    Filter by created by user ID (comma-separated for multiple)
  --is-processed  Filter by processed status (true/false)
  --broker        Filter by broker ID (comma-separated for multiple)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List time card time changes
  xbe view time-card-time-changes list

  # Filter by time card
  xbe view time-card-time-changes list --time-card 123

  # Filter by processed status
  xbe view time-card-time-changes list --is-processed false

  # Output as JSON
  xbe view time-card-time-changes list --json`,
		Args: cobra.NoArgs,
		RunE: runTimeCardTimeChangesList,
	}
	initTimeCardTimeChangesListFlags(cmd)
	return cmd
}

func init() {
	timeCardTimeChangesCmd.AddCommand(newTimeCardTimeChangesListCmd())
}

func initTimeCardTimeChangesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("time-card", "", "Filter by time card ID (comma-separated for multiple)")
	cmd.Flags().String("created-by", "", "Filter by created by user ID (comma-separated for multiple)")
	cmd.Flags().String("is-processed", "", "Filter by processed status (true/false)")
	cmd.Flags().String("broker", "", "Filter by broker ID (comma-separated for multiple)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTimeCardTimeChangesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTimeCardTimeChangesListOptions(cmd)
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
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[time-card-time-changes]", "created-at,comment,is-processed,time-card,created-by")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[time-card]", opts.TimeCard)
	setFilterIfPresent(query, "filter[created-by]", opts.CreatedBy)
	setFilterIfPresent(query, "filter[is-processed]", opts.IsProcessed)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)

	body, _, err := client.Get(cmd.Context(), "/v1/time-card-time-changes", query)
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

	rows := buildTimeCardTimeChangeRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTimeCardTimeChangesTable(cmd, rows)
}

func parseTimeCardTimeChangesListOptions(cmd *cobra.Command) (timeCardTimeChangesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	timeCard, _ := cmd.Flags().GetString("time-card")
	createdBy, _ := cmd.Flags().GetString("created-by")
	isProcessed, _ := cmd.Flags().GetString("is-processed")
	broker, _ := cmd.Flags().GetString("broker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return timeCardTimeChangesListOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		NoAuth:      noAuth,
		Limit:       limit,
		Offset:      offset,
		Sort:        sort,
		TimeCard:    timeCard,
		CreatedBy:   createdBy,
		IsProcessed: isProcessed,
		Broker:      broker,
	}, nil
}

func buildTimeCardTimeChangeRows(resp jsonAPIResponse) []timeCardTimeChangeRow {
	rows := make([]timeCardTimeChangeRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := timeCardTimeChangeRow{
			ID:          resource.ID,
			TimeCardID:  relationshipIDFromMap(resource.Relationships, "time-card"),
			CreatedByID: relationshipIDFromMap(resource.Relationships, "created-by"),
			IsProcessed: boolAttr(attrs, "is-processed"),
			Comment:     stringAttr(attrs, "comment"),
			CreatedAt:   formatDateTime(stringAttr(attrs, "created-at")),
		}
		rows = append(rows, row)
	}
	return rows
}

func renderTimeCardTimeChangesTable(cmd *cobra.Command, rows []timeCardTimeChangeRow) error {
	out := cmd.OutOrStdout()
	if len(rows) == 0 {
		fmt.Fprintln(out, "No time card time changes found.")
		return nil
	}

	writer := tabwriter.NewWriter(out, 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTIME_CARD\tCREATED_BY\tPROCESSED\tCOMMENT\tCREATED_AT")
	for _, row := range rows {
		processed := ""
		if row.IsProcessed {
			processed = "yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.TimeCardID,
			row.CreatedByID,
			processed,
			truncateString(row.Comment, 30),
			row.CreatedAt,
		)
	}

	return writer.Flush()
}
