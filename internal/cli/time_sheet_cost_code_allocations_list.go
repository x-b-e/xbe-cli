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

type timeSheetCostCodeAllocationsListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	TimeSheetID  string
	CreatedAtMin string
	CreatedAtMax string
	IsCreatedAt  string
	UpdatedAtMin string
	UpdatedAtMax string
	IsUpdatedAt  string
}

type timeSheetCostCodeAllocationRow struct {
	ID          string                              `json:"id"`
	TimeSheetID string                              `json:"time_sheet_id,omitempty"`
	Details     []timeSheetCostCodeAllocationDetail `json:"details,omitempty"`
	CreatedAt   string                              `json:"created_at,omitempty"`
	UpdatedAt   string                              `json:"updated_at,omitempty"`
}

func newTimeSheetCostCodeAllocationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List time sheet cost code allocations",
		Long: `List time sheet cost code allocations.

Output Columns:
  ID           Allocation identifier
  TIME SHEET   Time sheet ID
  ALLOCATIONS  Cost code allocations (cost_code_id:percentage)

Filters:
  --time-sheet       Filter by time sheet ID
  --created-at-min   Filter by created-at on/after (ISO 8601)
  --created-at-max   Filter by created-at on/before (ISO 8601)
  --is-created-at    Filter by presence of created-at (true/false)
  --updated-at-min   Filter by updated-at on/after (ISO 8601)
  --updated-at-max   Filter by updated-at on/before (ISO 8601)
  --is-updated-at    Filter by presence of updated-at (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List time sheet cost code allocations
  xbe view time-sheet-cost-code-allocations list

  # Filter by time sheet
  xbe view time-sheet-cost-code-allocations list --time-sheet 123

  # Filter by created-at range
  xbe view time-sheet-cost-code-allocations list \
    --created-at-min 2026-01-01T00:00:00Z \
    --created-at-max 2026-01-31T23:59:59Z

  # Output as JSON
  xbe view time-sheet-cost-code-allocations list --json`,
		Args: cobra.NoArgs,
		RunE: runTimeSheetCostCodeAllocationsList,
	}
	initTimeSheetCostCodeAllocationsListFlags(cmd)
	return cmd
}

func init() {
	timeSheetCostCodeAllocationsCmd.AddCommand(newTimeSheetCostCodeAllocationsListCmd())
}

func initTimeSheetCostCodeAllocationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("time-sheet", "", "Filter by time sheet ID")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by presence of created-at (true/false)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("is-updated-at", "", "Filter by presence of updated-at (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTimeSheetCostCodeAllocationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTimeSheetCostCodeAllocationsListOptions(cmd)
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
	query.Set("fields[time-sheet-cost-code-allocations]", "details,time-sheet,created-at,updated-at")
	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[time_sheet]", opts.TimeSheetID)
	setFilterIfPresent(query, "filter[created_at_min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created_at_max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[is_created_at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[updated_at_min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated_at_max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[is_updated_at]", opts.IsUpdatedAt)

	body, _, err := client.Get(cmd.Context(), "/v1/time-sheet-cost-code-allocations", query)
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

	rows := buildTimeSheetCostCodeAllocationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTimeSheetCostCodeAllocationsTable(cmd, rows)
}

func parseTimeSheetCostCodeAllocationsListOptions(cmd *cobra.Command) (timeSheetCostCodeAllocationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	timeSheetID, _ := cmd.Flags().GetString("time-sheet")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	isCreatedAt, _ := cmd.Flags().GetString("is-created-at")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	isUpdatedAt, _ := cmd.Flags().GetString("is-updated-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return timeSheetCostCodeAllocationsListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		TimeSheetID:  timeSheetID,
		CreatedAtMin: createdAtMin,
		CreatedAtMax: createdAtMax,
		IsCreatedAt:  isCreatedAt,
		UpdatedAtMin: updatedAtMin,
		UpdatedAtMax: updatedAtMax,
		IsUpdatedAt:  isUpdatedAt,
	}, nil
}

func buildTimeSheetCostCodeAllocationRows(resp jsonAPIResponse) []timeSheetCostCodeAllocationRow {
	rows := make([]timeSheetCostCodeAllocationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := timeSheetCostCodeAllocationRow{
			ID:        resource.ID,
			Details:   parseTimeSheetCostCodeAllocationDetails(attrs),
			CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
			UpdatedAt: formatDateTime(stringAttr(attrs, "updated-at")),
		}

		if rel, ok := resource.Relationships["time-sheet"]; ok && rel.Data != nil {
			row.TimeSheetID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func buildTimeSheetCostCodeAllocationRowFromSingle(resp jsonAPISingleResponse) timeSheetCostCodeAllocationRow {
	resource := resp.Data
	attrs := resource.Attributes
	row := timeSheetCostCodeAllocationRow{
		ID:        resource.ID,
		Details:   parseTimeSheetCostCodeAllocationDetails(attrs),
		CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt: formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resource.Relationships["time-sheet"]; ok && rel.Data != nil {
		row.TimeSheetID = rel.Data.ID
	}

	return row
}

func renderTimeSheetCostCodeAllocationsTable(cmd *cobra.Command, rows []timeSheetCostCodeAllocationRow) error {
	out := cmd.OutOrStdout()
	if len(rows) == 0 {
		fmt.Fprintln(out, "No time sheet cost code allocations found.")
		return nil
	}

	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTIME SHEET\tALLOCATIONS")
	for _, row := range rows {
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			row.ID,
			row.TimeSheetID,
			formatTimeSheetCostCodeAllocationSummary(row.Details),
		)
	}
	return w.Flush()
}
