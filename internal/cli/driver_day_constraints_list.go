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

type driverDayConstraintsListOptions struct {
	BaseURL    string
	Token      string
	JSON       bool
	NoAuth     bool
	Limit      int
	Offset     int
	Sort       string
	DriverDay  string
	Constraint string
}

type driverDayConstraintRow struct {
	ID           string `json:"id"`
	DriverDayID  string `json:"driver_day_id,omitempty"`
	ConstraintID string `json:"constraint_id,omitempty"`
}

func newDriverDayConstraintsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List driver day constraints",
		Long: `List driver day constraints.

Output Columns:
  ID           Driver day constraint identifier
  DRIVER DAY   Driver day ID
  CONSTRAINT   Shift set time card constraint ID

Filters:
  --driver-day  Filter by driver day ID
  --constraint  Filter by constraint ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List driver day constraints
  xbe view driver-day-constraints list

  # Filter by driver day
  xbe view driver-day-constraints list --driver-day 123

  # Filter by constraint
  xbe view driver-day-constraints list --constraint 456

  # Output as JSON
  xbe view driver-day-constraints list --json`,
		Args: cobra.NoArgs,
		RunE: runDriverDayConstraintsList,
	}
	initDriverDayConstraintsListFlags(cmd)
	return cmd
}

func init() {
	driverDayConstraintsCmd.AddCommand(newDriverDayConstraintsListCmd())
}

func initDriverDayConstraintsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("driver-day", "", "Filter by driver day ID")
	cmd.Flags().String("constraint", "", "Filter by constraint ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDriverDayConstraintsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseDriverDayConstraintsListOptions(cmd)
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
	query.Set("fields[driver-day-constraints]", "driver-day,constraint")
	query.Set("include", "driver-day,constraint")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[driver_day]", opts.DriverDay)
	setFilterIfPresent(query, "filter[constraint]", opts.Constraint)

	body, _, err := client.Get(cmd.Context(), "/v1/driver-day-constraints", query)
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

	rows := buildDriverDayConstraintRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderDriverDayConstraintsTable(cmd, rows)
}

func parseDriverDayConstraintsListOptions(cmd *cobra.Command) (driverDayConstraintsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	driverDay, _ := cmd.Flags().GetString("driver-day")
	constraint, _ := cmd.Flags().GetString("constraint")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return driverDayConstraintsListOptions{
		BaseURL:    baseURL,
		Token:      token,
		JSON:       jsonOut,
		NoAuth:     noAuth,
		Limit:      limit,
		Offset:     offset,
		Sort:       sort,
		DriverDay:  driverDay,
		Constraint: constraint,
	}, nil
}

func buildDriverDayConstraintRows(resp jsonAPIResponse) []driverDayConstraintRow {
	rows := make([]driverDayConstraintRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildDriverDayConstraintRow(resource))
	}
	return rows
}

func buildDriverDayConstraintRow(resource jsonAPIResource) driverDayConstraintRow {
	row := driverDayConstraintRow{
		ID: resource.ID,
	}
	if rel, ok := resource.Relationships["driver-day"]; ok && rel.Data != nil {
		row.DriverDayID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["constraint"]; ok && rel.Data != nil {
		row.ConstraintID = rel.Data.ID
	}
	return row
}

func buildDriverDayConstraintRowFromSingle(resp jsonAPISingleResponse) driverDayConstraintRow {
	return buildDriverDayConstraintRow(resp.Data)
}

func renderDriverDayConstraintsTable(cmd *cobra.Command, rows []driverDayConstraintRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No driver day constraints found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tDRIVER DAY\tCONSTRAINT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\n", row.ID, row.DriverDayID, row.ConstraintID)
	}
	writer.Flush()
	return nil
}
