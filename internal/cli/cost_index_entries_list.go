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

type costIndexEntriesListOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	NoAuth    bool
	Limit     int
	Offset    int
	CostIndex string
	StartOn   string
	EndOn     string
	Value     string
}

func newCostIndexEntriesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List cost index entries",
		Long: `List cost index entries with filtering and pagination.

Cost index entries are time-series values for cost indexes, used in rate adjustments.

Output Columns:
  ID           Entry identifier
  COST INDEX   Parent cost index ID
  START ON     Entry start date
  END ON       Entry end date
  VALUE        Entry value

Filters:
  --cost-index  Filter by cost index ID (required for most queries)
  --start-on    Filter by start date
  --end-on      Filter by end date
  --value       Filter by value`,
		Example: `  # List entries for a cost index
  xbe view cost-index-entries list --cost-index 123

  # Filter by date range
  xbe view cost-index-entries list --cost-index 123 --start-on "2024-01-01"

  # Output as JSON
  xbe view cost-index-entries list --cost-index 123 --json`,
		RunE: runCostIndexEntriesList,
	}
	initCostIndexEntriesListFlags(cmd)
	return cmd
}

func init() {
	costIndexEntriesCmd.AddCommand(newCostIndexEntriesListCmd())
}

func initCostIndexEntriesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("cost-index", "", "Filter by cost index ID")
	cmd.Flags().String("start-on", "", "Filter by start date")
	cmd.Flags().String("end-on", "", "Filter by end date")
	cmd.Flags().String("value", "", "Filter by value")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCostIndexEntriesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseCostIndexEntriesListOptions(cmd)
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
	query.Set("sort", "-start-on")
	query.Set("fields[cost-index-entries]", "start-on,end-on,value,cost-index")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[cost-index]", opts.CostIndex)
	setFilterIfPresent(query, "filter[start-on]", opts.StartOn)
	setFilterIfPresent(query, "filter[end-on]", opts.EndOn)
	setFilterIfPresent(query, "filter[value]", opts.Value)

	body, _, err := client.Get(cmd.Context(), "/v1/cost-index-entries", query)
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

	rows := buildCostIndexEntryRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderCostIndexEntriesTable(cmd, rows)
}

func parseCostIndexEntriesListOptions(cmd *cobra.Command) (costIndexEntriesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	costIndex, _ := cmd.Flags().GetString("cost-index")
	startOn, _ := cmd.Flags().GetString("start-on")
	endOn, _ := cmd.Flags().GetString("end-on")
	value, _ := cmd.Flags().GetString("value")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return costIndexEntriesListOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		NoAuth:    noAuth,
		Limit:     limit,
		Offset:    offset,
		CostIndex: costIndex,
		StartOn:   startOn,
		EndOn:     endOn,
		Value:     value,
	}, nil
}

type costIndexEntryRow struct {
	ID          string `json:"id"`
	CostIndexID string `json:"cost_index_id,omitempty"`
	StartOn     string `json:"start_on,omitempty"`
	EndOn       string `json:"end_on,omitempty"`
	Value       any    `json:"value,omitempty"`
}

func buildCostIndexEntryRows(resp jsonAPIResponse) []costIndexEntryRow {
	rows := make([]costIndexEntryRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := costIndexEntryRow{
			ID:      resource.ID,
			StartOn: stringAttr(resource.Attributes, "start-on"),
			EndOn:   stringAttr(resource.Attributes, "end-on"),
			Value:   resource.Attributes["value"],
		}

		if rel, ok := resource.Relationships["cost-index"]; ok && rel.Data != nil {
			row.CostIndexID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderCostIndexEntriesTable(cmd *cobra.Command, rows []costIndexEntryRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No cost index entries found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tCOST INDEX\tSTART ON\tEND ON\tVALUE")
	for _, row := range rows {
		value := ""
		if row.Value != nil {
			value = fmt.Sprintf("%v", row.Value)
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.CostIndexID,
			row.StartOn,
			row.EndOn,
			value,
		)
	}
	return writer.Flush()
}
