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

type driverDayAdjustmentsListOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	NoAuth    bool
	Limit     int
	Offset    int
	Sort      string
	DriverDay string
	Trucker   string
	TruckerID string
	Driver    string
	DriverID  string
}

type driverDayAdjustmentRow struct {
	ID              string `json:"id"`
	DriverDayID     string `json:"driver_day_id,omitempty"`
	TruckerID       string `json:"trucker_id,omitempty"`
	DriverID        string `json:"driver_id,omitempty"`
	Amount          string `json:"amount,omitempty"`
	AmountExplicit  string `json:"amount_explicit,omitempty"`
	AmountGenerated string `json:"amount_generated,omitempty"`
}

func newDriverDayAdjustmentsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List driver day adjustments",
		Long: `List driver day adjustments.

Output Columns:
  ID         Adjustment identifier
  DRIVER DAY Driver day ID
  DRIVER     Driver user ID
  TRUCKER    Trucker ID
  AMOUNT     Final adjustment amount
  EXPLICIT   Explicit adjustment amount
  GENERATED  Generated adjustment amount

Filters:
  --driver-day  Filter by driver day ID
  --trucker     Filter by trucker ID
  --trucker-id  Filter by driver day trucker ID
  --driver      Filter by driver user ID
  --driver-id   Filter by driver day driver ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List adjustments
  xbe view driver-day-adjustments list

  # Filter by driver day
  xbe view driver-day-adjustments list --driver-day 123

  # Filter by driver
  xbe view driver-day-adjustments list --driver 456

  # Output as JSON
  xbe view driver-day-adjustments list --json`,
		Args: cobra.NoArgs,
		RunE: runDriverDayAdjustmentsList,
	}
	initDriverDayAdjustmentsListFlags(cmd)
	return cmd
}

func init() {
	driverDayAdjustmentsCmd.AddCommand(newDriverDayAdjustmentsListCmd())
}

func initDriverDayAdjustmentsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("driver-day", "", "Filter by driver day ID")
	cmd.Flags().String("trucker", "", "Filter by trucker ID")
	cmd.Flags().String("trucker-id", "", "Filter by driver day trucker ID")
	cmd.Flags().String("driver", "", "Filter by driver user ID")
	cmd.Flags().String("driver-id", "", "Filter by driver day driver ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDriverDayAdjustmentsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseDriverDayAdjustmentsListOptions(cmd)
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

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[driver-day]", opts.DriverDay)
	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)
	setFilterIfPresent(query, "filter[trucker-id]", opts.TruckerID)
	setFilterIfPresent(query, "filter[driver]", opts.Driver)
	setFilterIfPresent(query, "filter[driver-id]", opts.DriverID)

	body, _, err := client.Get(cmd.Context(), "/v1/driver-day-adjustments", query)
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

	rows := buildDriverDayAdjustmentRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderDriverDayAdjustmentsTable(cmd, rows)
}

func parseDriverDayAdjustmentsListOptions(cmd *cobra.Command) (driverDayAdjustmentsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	driverDay, _ := cmd.Flags().GetString("driver-day")
	trucker, _ := cmd.Flags().GetString("trucker")
	truckerID, _ := cmd.Flags().GetString("trucker-id")
	driver, _ := cmd.Flags().GetString("driver")
	driverID, _ := cmd.Flags().GetString("driver-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return driverDayAdjustmentsListOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		NoAuth:    noAuth,
		Limit:     limit,
		Offset:    offset,
		Sort:      sort,
		DriverDay: driverDay,
		Trucker:   trucker,
		TruckerID: truckerID,
		Driver:    driver,
		DriverID:  driverID,
	}, nil
}

func buildDriverDayAdjustmentRows(resp jsonAPIResponse) []driverDayAdjustmentRow {
	rows := make([]driverDayAdjustmentRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := driverDayAdjustmentRow{
			ID:              resource.ID,
			Amount:          stringAttr(attrs, "amount"),
			AmountExplicit:  stringAttr(attrs, "amount-explicit"),
			AmountGenerated: stringAttr(attrs, "amount-generated"),
		}

		if rel, ok := resource.Relationships["driver-day"]; ok && rel.Data != nil {
			row.DriverDayID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["trucker"]; ok && rel.Data != nil {
			row.TruckerID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["driver"]; ok && rel.Data != nil {
			row.DriverID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func buildDriverDayAdjustmentRowFromSingle(resp jsonAPISingleResponse) driverDayAdjustmentRow {
	resource := resp.Data
	attrs := resource.Attributes
	row := driverDayAdjustmentRow{
		ID:              resource.ID,
		Amount:          stringAttr(attrs, "amount"),
		AmountExplicit:  stringAttr(attrs, "amount-explicit"),
		AmountGenerated: stringAttr(attrs, "amount-generated"),
	}

	if rel, ok := resource.Relationships["driver-day"]; ok && rel.Data != nil {
		row.DriverDayID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["trucker"]; ok && rel.Data != nil {
		row.TruckerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["driver"]; ok && rel.Data != nil {
		row.DriverID = rel.Data.ID
	}

	return row
}

func renderDriverDayAdjustmentsTable(cmd *cobra.Command, rows []driverDayAdjustmentRow) error {
	out := cmd.OutOrStdout()

	w := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tDRIVER DAY\tDRIVER\tTRUCKER\tAMOUNT\tEXPLICIT\tGENERATED")
	for _, row := range rows {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.DriverDayID,
			row.DriverID,
			row.TruckerID,
			row.Amount,
			row.AmountExplicit,
			row.AmountGenerated,
		)
	}
	return w.Flush()
}
