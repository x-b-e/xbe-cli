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

type rateAdjustmentsListOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	NoAuth                 bool
	Limit                  int
	Offset                 int
	CostIndex              string
	Rate                   string
	ParentRateAdjustment   string
	IsParentRateAdjustment string
}

type rateAdjustmentRow struct {
	ID                                 string `json:"id"`
	RateID                             string `json:"rate_id,omitempty"`
	CostIndexID                        string `json:"cost_index_id,omitempty"`
	ParentRateAdjustmentID             string `json:"parent_rate_adjustment_id,omitempty"`
	ZeroInterceptValue                 string `json:"zero_intercept_value,omitempty"`
	ZeroInterceptRatio                 string `json:"zero_intercept_ratio,omitempty"`
	AdjustmentMin                      string `json:"adjustment_min,omitempty"`
	AdjustmentMax                      string `json:"adjustment_max,omitempty"`
	PreventRatingWhenIndexValueMissing bool   `json:"prevent_rating_when_index_value_missing"`
}

func newRateAdjustmentsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List rate adjustments",
		Long: `List rate adjustments.

Output Columns:
  ID           Rate adjustment identifier
  RATE         Rate ID
  COST INDEX   Cost index ID
  ZERO VALUE   Zero intercept value
  ZERO RATIO   Zero intercept ratio
  MIN          Minimum adjustment
  MAX          Maximum adjustment

Filters:
  --rate                      Filter by rate ID
  --cost-index                Filter by cost index ID
  --parent-rate-adjustment    Filter by parent rate adjustment ID
  --is-parent-rate-adjustment Filter by parent status (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --base-url, --token, --no-auth`,
		Example: `  # List rate adjustments
  xbe view rate-adjustments list

  # Filter by rate and cost index
  xbe view rate-adjustments list --rate 123 --cost-index 456

  # Only parent adjustments
  xbe view rate-adjustments list --is-parent-rate-adjustment true

  # Output as JSON
  xbe view rate-adjustments list --json`,
		RunE: runRateAdjustmentsList,
	}
	initRateAdjustmentsListFlags(cmd)
	return cmd
}

func init() {
	rateAdjustmentsCmd.AddCommand(newRateAdjustmentsListCmd())
}

func initRateAdjustmentsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("rate", "", "Filter by rate ID")
	cmd.Flags().String("cost-index", "", "Filter by cost index ID")
	cmd.Flags().String("parent-rate-adjustment", "", "Filter by parent rate adjustment ID")
	cmd.Flags().String("is-parent-rate-adjustment", "", "Filter by parent status (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runRateAdjustmentsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseRateAdjustmentsListOptions(cmd)
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
	query.Set("fields[rate-adjustments]", "zero-intercept-value,zero-intercept-ratio,adjustment-min,adjustment-max,prevent-rating-when-index-value-missing,rate,cost-index,parent-rate-adjustment")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	setFilterIfPresent(query, "filter[rate]", opts.Rate)
	setFilterIfPresent(query, "filter[cost-index]", opts.CostIndex)
	setFilterIfPresent(query, "filter[parent-rate-adjustment]", opts.ParentRateAdjustment)
	setFilterIfPresent(query, "filter[is-parent-rate-adjustment]", opts.IsParentRateAdjustment)

	body, _, err := client.Get(cmd.Context(), "/v1/rate-adjustments", query)
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

	rows := buildRateAdjustmentRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderRateAdjustmentsTable(cmd, rows)
}

func parseRateAdjustmentsListOptions(cmd *cobra.Command) (rateAdjustmentsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	costIndex, _ := cmd.Flags().GetString("cost-index")
	rate, _ := cmd.Flags().GetString("rate")
	parentRateAdjustment, _ := cmd.Flags().GetString("parent-rate-adjustment")
	isParentRateAdjustment, _ := cmd.Flags().GetString("is-parent-rate-adjustment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return rateAdjustmentsListOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		NoAuth:                 noAuth,
		Limit:                  limit,
		Offset:                 offset,
		CostIndex:              costIndex,
		Rate:                   rate,
		ParentRateAdjustment:   parentRateAdjustment,
		IsParentRateAdjustment: isParentRateAdjustment,
	}, nil
}

func buildRateAdjustmentRow(resource jsonAPIResource) rateAdjustmentRow {
	row := rateAdjustmentRow{
		ID:                                 resource.ID,
		ZeroInterceptValue:                 stringAttr(resource.Attributes, "zero-intercept-value"),
		ZeroInterceptRatio:                 stringAttr(resource.Attributes, "zero-intercept-ratio"),
		AdjustmentMin:                      stringAttr(resource.Attributes, "adjustment-min"),
		AdjustmentMax:                      stringAttr(resource.Attributes, "adjustment-max"),
		PreventRatingWhenIndexValueMissing: boolAttr(resource.Attributes, "prevent-rating-when-index-value-missing"),
	}

	if rel, ok := resource.Relationships["rate"]; ok && rel.Data != nil {
		row.RateID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["cost-index"]; ok && rel.Data != nil {
		row.CostIndexID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["parent-rate-adjustment"]; ok && rel.Data != nil {
		row.ParentRateAdjustmentID = rel.Data.ID
	}

	return row
}

func buildRateAdjustmentRows(resp jsonAPIResponse) []rateAdjustmentRow {
	rows := make([]rateAdjustmentRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildRateAdjustmentRow(resource))
	}
	return rows
}

func buildRateAdjustmentRowFromSingle(resp jsonAPISingleResponse) rateAdjustmentRow {
	return buildRateAdjustmentRow(resp.Data)
}

func renderRateAdjustmentsTable(cmd *cobra.Command, rows []rateAdjustmentRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No rate adjustments found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tRATE\tCOST INDEX\tZERO VALUE\tZERO RATIO\tMIN\tMAX")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.RateID,
			row.CostIndexID,
			row.ZeroInterceptValue,
			row.ZeroInterceptRatio,
			row.AdjustmentMin,
			row.AdjustmentMax,
		)
	}
	return writer.Flush()
}
