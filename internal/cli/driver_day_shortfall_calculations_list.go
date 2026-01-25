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

type driverDayShortfallCalculationsListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Sort    string
}

type driverDayShortfallCalculationRow struct {
	ID                         string  `json:"id"`
	Quantity                   float64 `json:"quantity,omitempty"`
	ServiceTypeUnitOfMeasure   string  `json:"service_type_unit_of_measure_id,omitempty"`
	TimeCardCount              int     `json:"time_card_count,omitempty"`
	AllocatableTimeCardCount   int     `json:"allocatable_time_card_count,omitempty"`
	UnallocatableTimeCardCount int     `json:"unallocatable_time_card_count,omitempty"`
	ConstraintCount            int     `json:"driver_day_time_card_constraint_count,omitempty"`
}

func newDriverDayShortfallCalculationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List driver day shortfall calculations",
		Long: `List driver day shortfall calculations.

Output Columns:
  ID             Calculation identifier
  QUANTITY       Shortfall quantity
  UOM            Service type unit of measure ID
  TIME CARDS     Total time cards
  ALLOCATABLE    Allocatable time cards
  UNALLOCATABLE  Unallocatable time cards
  CONSTRAINTS    Driver day time card constraints

Filters:
  None

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List calculations
  xbe view driver-day-shortfall-calculations list

  # Paginate results
  xbe view driver-day-shortfall-calculations list --limit 10 --offset 20

  # JSON output
  xbe view driver-day-shortfall-calculations list --json`,
		Args: cobra.NoArgs,
		RunE: runDriverDayShortfallCalculationsList,
	}
	initDriverDayShortfallCalculationsListFlags(cmd)
	return cmd
}

func init() {
	driverDayShortfallCalculationsCmd.AddCommand(newDriverDayShortfallCalculationsListCmd())
}

func initDriverDayShortfallCalculationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 0, "Limit results")
	cmd.Flags().Int("offset", 0, "Offset results")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDriverDayShortfallCalculationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseDriverDayShortfallCalculationsListOptions(cmd)
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
	query.Set("fields[driver-day-shortfall-calculations]", "quantity,allocatable-time-card-ids,unallocatable-time-card-ids,driver-day-time-card-constraint-ids,time-card-ids")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	body, _, err := client.Get(cmd.Context(), "/v1/driver-day-shortfall-calculations", query)
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

	rows := buildDriverDayShortfallCalculationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderDriverDayShortfallCalculationsTable(cmd, rows)
}

func parseDriverDayShortfallCalculationsListOptions(cmd *cobra.Command) (driverDayShortfallCalculationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return driverDayShortfallCalculationsListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
	}, nil
}

func buildDriverDayShortfallCalculationRows(resp jsonAPIResponse) []driverDayShortfallCalculationRow {
	rows := make([]driverDayShortfallCalculationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		timeCardIDs := stringSliceAttr(attrs, "time-card-ids")
		allocatableTimeCardIDs := stringSliceAttr(attrs, "allocatable-time-card-ids")
		unallocatableTimeCardIDs := stringSliceAttr(attrs, "unallocatable-time-card-ids")
		constraintIDs := stringSliceAttr(attrs, "driver-day-time-card-constraint-ids")

		row := driverDayShortfallCalculationRow{
			ID:                         resource.ID,
			Quantity:                   floatAttr(attrs, "quantity"),
			TimeCardCount:              len(timeCardIDs),
			AllocatableTimeCardCount:   len(allocatableTimeCardIDs),
			UnallocatableTimeCardCount: len(unallocatableTimeCardIDs),
			ConstraintCount:            len(constraintIDs),
		}

		if rel, ok := resource.Relationships["service-type-unit-of-measure"]; ok && rel.Data != nil {
			row.ServiceTypeUnitOfMeasure = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderDriverDayShortfallCalculationsTable(cmd *cobra.Command, rows []driverDayShortfallCalculationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No driver day shortfall calculations found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tQUANTITY\tUOM\tTIME CARDS\tALLOCATABLE\tUNALLOCATABLE\tCONSTRAINTS")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%.2f\t%s\t%d\t%d\t%d\t%d\n",
			row.ID,
			row.Quantity,
			row.ServiceTypeUnitOfMeasure,
			row.TimeCardCount,
			row.AllocatableTimeCardCount,
			row.UnallocatableTimeCardCount,
			row.ConstraintCount,
		)
	}
	return writer.Flush()
}
