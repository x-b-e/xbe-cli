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

type lineupDispatchesListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	Lineup       string
	CreatedAtMin string
	CreatedAtMax string
	UpdatedAtMin string
	UpdatedAtMax string
}

type lineupDispatchRow struct {
	ID               string `json:"id"`
	LineupID         string `json:"lineup_id,omitempty"`
	CreatedByID      string `json:"created_by_id,omitempty"`
	IsFulfilled      bool   `json:"is_fulfilled"`
	IsFulfilling     bool   `json:"is_fulfilling"`
	FulfillmentCount int    `json:"fulfillment_count"`
}

func newLineupDispatchesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List lineup dispatches",
		Long: `List lineup dispatches.

Output Columns:
  ID          Lineup dispatch identifier
  LINEUP      Lineup ID
  CREATED BY  User ID of creator
  FULFILLED   Fulfillment status
  FULFILLING  Fulfillment in progress
  COUNT       Fulfillment attempts count

Filters:
  --lineup          Filter by lineup ID
  --created-at-min  Filter by created-at on/after (ISO 8601)
  --created-at-max  Filter by created-at on/before (ISO 8601)
  --updated-at-min  Filter by updated-at on/after (ISO 8601)
  --updated-at-max  Filter by updated-at on/before (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List lineup dispatches
  xbe view lineup-dispatches list

  # Filter by lineup
  xbe view lineup-dispatches list --lineup 123

  # Output as JSON
  xbe view lineup-dispatches list --json`,
		Args: cobra.NoArgs,
		RunE: runLineupDispatchesList,
	}
	initLineupDispatchesListFlags(cmd)
	return cmd
}

func init() {
	lineupDispatchesCmd.AddCommand(newLineupDispatchesListCmd())
}

func initLineupDispatchesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("lineup", "", "Filter by lineup ID")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runLineupDispatchesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseLineupDispatchesListOptions(cmd)
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
	query.Set("fields[lineup-dispatches]", "is-fulfilled,is-fulfilling,fulfillment-count")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[lineup]", opts.Lineup)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/lineup-dispatches", query)
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

	rows := buildLineupDispatchRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderLineupDispatchesTable(cmd, rows)
}

func parseLineupDispatchesListOptions(cmd *cobra.Command) (lineupDispatchesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	lineup, _ := cmd.Flags().GetString("lineup")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return lineupDispatchesListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		Lineup:       lineup,
		CreatedAtMin: createdAtMin,
		CreatedAtMax: createdAtMax,
		UpdatedAtMin: updatedAtMin,
		UpdatedAtMax: updatedAtMax,
	}, nil
}

func buildLineupDispatchRows(resp jsonAPIResponse) []lineupDispatchRow {
	rows := make([]lineupDispatchRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := lineupDispatchRow{
			ID:               resource.ID,
			IsFulfilled:      boolAttr(resource.Attributes, "is-fulfilled"),
			IsFulfilling:     boolAttr(resource.Attributes, "is-fulfilling"),
			FulfillmentCount: intAttr(resource.Attributes, "fulfillment-count"),
		}

		if rel, ok := resource.Relationships["lineup"]; ok && rel.Data != nil {
			row.LineupID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
			row.CreatedByID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderLineupDispatchesTable(cmd *cobra.Command, rows []lineupDispatchRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No lineup dispatches found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tLINEUP\tCREATED BY\tFULFILLED\tFULFILLING\tCOUNT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%d\n",
			row.ID,
			row.LineupID,
			row.CreatedByID,
			formatYesNo(row.IsFulfilled),
			formatYesNo(row.IsFulfilling),
			row.FulfillmentCount,
		)
	}
	return writer.Flush()
}
