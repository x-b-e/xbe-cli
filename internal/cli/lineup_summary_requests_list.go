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

type lineupSummaryRequestsListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	CreatedAtMin string
	CreatedAtMax string
	UpdatedAtMin string
	UpdatedAtMax string
}

type lineupSummaryRequestRow struct {
	ID             string   `json:"id"`
	LevelType      string   `json:"level_type,omitempty"`
	LevelID        string   `json:"level_id,omitempty"`
	StartAtMin     string   `json:"start_at_min,omitempty"`
	StartAtMax     string   `json:"start_at_max,omitempty"`
	EmailTo        []string `json:"email_to,omitempty"`
	SendIfNoShifts bool     `json:"send_if_no_shifts"`
	CreatedByID    string   `json:"created_by_id,omitempty"`
}

func newLineupSummaryRequestsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List lineup summary requests",
		Long: `List lineup summary requests.

Output Columns:
  ID            Lineup summary request ID
  LEVEL         Level type and ID (brokers/123 or customers/456)
  START MIN     Minimum shift start time
  START MAX     Maximum shift start time
  EMAIL TO      Email recipients
  SEND IF NONE  Send summary even if no shifts
  CREATED BY    User ID of the requester (if present)

Filters:
  --created-at-min   Filter by created-at on/after (ISO 8601)
  --created-at-max   Filter by created-at on/before (ISO 8601)
  --updated-at-min   Filter by updated-at on/after (ISO 8601)
  --updated-at-max   Filter by updated-at on/before (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List lineup summary requests
  xbe view lineup-summary-requests list

  # Paginate results
  xbe view lineup-summary-requests list --limit 25 --offset 50

  # Output as JSON
  xbe view lineup-summary-requests list --json`,
		Args: cobra.NoArgs,
		RunE: runLineupSummaryRequestsList,
	}
	initLineupSummaryRequestsListFlags(cmd)
	return cmd
}

func init() {
	lineupSummaryRequestsCmd.AddCommand(newLineupSummaryRequestsListCmd())
}

func initLineupSummaryRequestsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runLineupSummaryRequestsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseLineupSummaryRequestsListOptions(cmd)
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
	query.Set("fields[lineup-summary-requests]", "start-at-min,start-at-max,email-to,send-if-no-shifts")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/lineup-summary-requests", query)
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

	rows := buildLineupSummaryRequestRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderLineupSummaryRequestsTable(cmd, rows)
}

func parseLineupSummaryRequestsListOptions(cmd *cobra.Command) (lineupSummaryRequestsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return lineupSummaryRequestsListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		CreatedAtMin: createdAtMin,
		CreatedAtMax: createdAtMax,
		UpdatedAtMin: updatedAtMin,
		UpdatedAtMax: updatedAtMax,
	}, nil
}

func buildLineupSummaryRequestRows(resp jsonAPIResponse) []lineupSummaryRequestRow {
	rows := make([]lineupSummaryRequestRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildLineupSummaryRequestRow(resource))
	}
	return rows
}

func buildLineupSummaryRequestRow(resource jsonAPIResource) lineupSummaryRequestRow {
	attrs := resource.Attributes
	row := lineupSummaryRequestRow{
		ID:             resource.ID,
		StartAtMin:     formatDateTime(stringAttr(attrs, "start-at-min")),
		StartAtMax:     formatDateTime(stringAttr(attrs, "start-at-max")),
		EmailTo:        stringSliceAttr(attrs, "email-to"),
		SendIfNoShifts: boolAttr(attrs, "send-if-no-shifts"),
	}

	if rel, ok := resource.Relationships["level"]; ok && rel.Data != nil {
		row.LevelType = rel.Data.Type
		row.LevelID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		row.CreatedByID = rel.Data.ID
	}

	return row
}

func buildLineupSummaryRequestRowFromSingle(resp jsonAPISingleResponse) lineupSummaryRequestRow {
	return buildLineupSummaryRequestRow(resp.Data)
}

func renderLineupSummaryRequestsTable(cmd *cobra.Command, rows []lineupSummaryRequestRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No lineup summary requests found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tLEVEL\tSTART MIN\tSTART MAX\tEMAIL TO\tSEND IF NONE\tCREATED BY")
	for _, row := range rows {
		level := ""
		if row.LevelType != "" && row.LevelID != "" {
			level = row.LevelType + "/" + row.LevelID
		}
		emailTo := truncateString(strings.Join(row.EmailTo, ", "), 40)

		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%t\t%s\n",
			row.ID,
			truncateString(level, 25),
			truncateString(row.StartAtMin, 20),
			truncateString(row.StartAtMax, 20),
			emailTo,
			row.SendIfNoShifts,
			row.CreatedByID,
		)
	}
	return writer.Flush()
}
