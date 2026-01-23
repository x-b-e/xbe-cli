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

type timeCardCostCodeAllocationsListOptions struct {
	BaseURL  string
	Token    string
	JSON     bool
	NoAuth   bool
	Limit    int
	Offset   int
	Sort     string
	TimeCard string
}

type timeCardCostCodeAllocationRow struct {
	ID              string   `json:"id"`
	TimeCardID      string   `json:"time_card_id,omitempty"`
	CostCodeIDs     []string `json:"cost_code_ids,omitempty"`
	AllocationCount int      `json:"allocation_count,omitempty"`
}

func newTimeCardCostCodeAllocationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List time card cost code allocations",
		Long: `List time card cost code allocations.

Output Columns:
  ID          Allocation ID
  TIME CARD   Time card ID
  COST CODES  Cost code IDs (truncated)
  ALLOCS      Number of allocation entries

Filters:
  --time-card  Filter by time card ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List time card cost code allocations
  xbe view time-card-cost-code-allocations list

  # Filter by time card
  xbe view time-card-cost-code-allocations list --time-card 123

  # Output as JSON
  xbe view time-card-cost-code-allocations list --json`,
		Args: cobra.NoArgs,
		RunE: runTimeCardCostCodeAllocationsList,
	}
	initTimeCardCostCodeAllocationsListFlags(cmd)
	return cmd
}

func init() {
	timeCardCostCodeAllocationsCmd.AddCommand(newTimeCardCostCodeAllocationsListCmd())
}

func initTimeCardCostCodeAllocationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("time-card", "", "Filter by time card ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTimeCardCostCodeAllocationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTimeCardCostCodeAllocationsListOptions(cmd)
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
	query.Set("fields[time-card-cost-code-allocations]", "details,time-card,cost-codes")

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

	body, _, err := client.Get(cmd.Context(), "/v1/time-card-cost-code-allocations", query)
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

	rows := buildTimeCardCostCodeAllocationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTimeCardCostCodeAllocationsTable(cmd, rows)
}

func parseTimeCardCostCodeAllocationsListOptions(cmd *cobra.Command) (timeCardCostCodeAllocationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	timeCard, _ := cmd.Flags().GetString("time-card")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return timeCardCostCodeAllocationsListOptions{
		BaseURL:  baseURL,
		Token:    token,
		JSON:     jsonOut,
		NoAuth:   noAuth,
		Limit:    limit,
		Offset:   offset,
		Sort:     sort,
		TimeCard: timeCard,
	}, nil
}

func buildTimeCardCostCodeAllocationRows(resp jsonAPIResponse) []timeCardCostCodeAllocationRow {
	rows := make([]timeCardCostCodeAllocationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildTimeCardCostCodeAllocationRow(resource))
	}
	return rows
}

func buildTimeCardCostCodeAllocationRow(resource jsonAPIResource) timeCardCostCodeAllocationRow {
	details := allocationDetailsValue(resource.Attributes)
	allocationCount := allocationDetailsCount(details)
	costCodeIDs := costCodeIDsFromResource(resource, details)
	if allocationCount == 0 && len(costCodeIDs) > 0 {
		allocationCount = len(costCodeIDs)
	}

	row := timeCardCostCodeAllocationRow{
		ID:              resource.ID,
		CostCodeIDs:     costCodeIDs,
		AllocationCount: allocationCount,
	}
	if rel, ok := resource.Relationships["time-card"]; ok && rel.Data != nil {
		row.TimeCardID = rel.Data.ID
	}

	return row
}

func costCodeIDsFromResource(resource jsonAPIResource, details any) []string {
	if rel, ok := resource.Relationships["cost-codes"]; ok {
		if ids := relationshipIDList(rel); len(ids) > 0 {
			return ids
		}
	}
	return allocationCostCodeIDsFromDetails(details)
}

func renderTimeCardCostCodeAllocationsTable(cmd *cobra.Command, rows []timeCardCostCodeAllocationRow) error {
	out := cmd.OutOrStdout()
	writer := tabwriter.NewWriter(out, 0, 8, 2, ' ', 0)

	fmt.Fprintln(writer, "ID\tTIME CARD\tCOST CODES\tALLOCS")
	for _, row := range rows {
		costCodes := truncateString(strings.Join(row.CostCodeIDs, ", "), 32)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%d\n",
			row.ID,
			row.TimeCardID,
			costCodes,
			row.AllocationCount,
		)
	}

	return writer.Flush()
}
