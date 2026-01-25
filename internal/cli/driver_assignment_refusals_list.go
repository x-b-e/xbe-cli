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

type driverAssignmentRefusalsListOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	NoAuth                 bool
	Limit                  int
	Offset                 int
	Sort                   string
	TenderJobScheduleShift string
	Driver                 string
	CreatedAtMin           string
	CreatedAtMax           string
	UpdatedAtMin           string
	UpdatedAtMax           string
}

type driverAssignmentRefusalRow struct {
	ID                       string `json:"id"`
	Comment                  string `json:"comment,omitempty"`
	TenderJobScheduleShiftID string `json:"tender_job_schedule_shift_id,omitempty"`
	DriverID                 string `json:"driver_id,omitempty"`
	CreatedByID              string `json:"created_by_id,omitempty"`
	CreatedAt                string `json:"created_at,omitempty"`
}

func newDriverAssignmentRefusalsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List driver assignment refusals",
		Long: `List driver assignment refusals.

Output Columns:
  ID          Refusal identifier
  SHIFT ID    Tender job schedule shift ID
  DRIVER      Driver user ID
  CREATED BY  User who created the refusal
  COMMENT     Refusal comment (truncated)
  CREATED AT  When the refusal was created

Filters:
  --tender-job-schedule-shift  Filter by tender job schedule shift ID
  --driver                     Filter by driver user ID
  --created-at-min             Filter by created-at on/after (ISO 8601)
  --created-at-max             Filter by created-at on/before (ISO 8601)
  --updated-at-min             Filter by updated-at on/after (ISO 8601)
  --updated-at-max             Filter by updated-at on/before (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List refusals
  xbe view driver-assignment-refusals list

  # Filter by tender job schedule shift
  xbe view driver-assignment-refusals list --tender-job-schedule-shift 123

  # Filter by driver
  xbe view driver-assignment-refusals list --driver 456

  # Output as JSON
  xbe view driver-assignment-refusals list --json`,
		Args: cobra.NoArgs,
		RunE: runDriverAssignmentRefusalsList,
	}
	initDriverAssignmentRefusalsListFlags(cmd)
	return cmd
}

func init() {
	driverAssignmentRefusalsCmd.AddCommand(newDriverAssignmentRefusalsListCmd())
}

func initDriverAssignmentRefusalsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("tender-job-schedule-shift", "", "Filter by tender job schedule shift ID")
	cmd.Flags().String("driver", "", "Filter by driver user ID")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDriverAssignmentRefusalsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseDriverAssignmentRefusalsListOptions(cmd)
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

	setFilterIfPresent(query, "filter[tender_job_schedule_shift]", opts.TenderJobScheduleShift)
	setFilterIfPresent(query, "filter[driver]", opts.Driver)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/driver-assignment-refusals", query)
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

	rows := buildDriverAssignmentRefusalRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderDriverAssignmentRefusalsTable(cmd, rows)
}

func parseDriverAssignmentRefusalsListOptions(cmd *cobra.Command) (driverAssignmentRefusalsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	tenderJobScheduleShift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	driver, _ := cmd.Flags().GetString("driver")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return driverAssignmentRefusalsListOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		NoAuth:                 noAuth,
		Limit:                  limit,
		Offset:                 offset,
		Sort:                   sort,
		TenderJobScheduleShift: tenderJobScheduleShift,
		Driver:                 driver,
		CreatedAtMin:           createdAtMin,
		CreatedAtMax:           createdAtMax,
		UpdatedAtMin:           updatedAtMin,
		UpdatedAtMax:           updatedAtMax,
	}, nil
}

func buildDriverAssignmentRefusalRows(resp jsonAPIResponse) []driverAssignmentRefusalRow {
	rows := make([]driverAssignmentRefusalRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := driverAssignmentRefusalRow{
			ID:        resource.ID,
			Comment:   stringAttr(attrs, "comment"),
			CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
		}

		if rel, ok := resource.Relationships["tender-job-schedule-shift"]; ok && rel.Data != nil {
			row.TenderJobScheduleShiftID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["driver"]; ok && rel.Data != nil {
			row.DriverID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
			row.CreatedByID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderDriverAssignmentRefusalsTable(cmd *cobra.Command, rows []driverAssignmentRefusalRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No driver assignment refusals found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSHIFT ID\tDRIVER\tCREATED BY\tCOMMENT\tCREATED AT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.TenderJobScheduleShiftID,
			row.DriverID,
			row.CreatedByID,
			truncateString(row.Comment, 40),
			row.CreatedAt,
		)
	}
	return writer.Flush()
}

func buildDriverAssignmentRefusalRowFromSingle(resp jsonAPISingleResponse) driverAssignmentRefusalRow {
	attrs := resp.Data.Attributes
	row := driverAssignmentRefusalRow{
		ID:        resp.Data.ID,
		Comment:   stringAttr(attrs, "comment"),
		CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
	}

	if rel, ok := resp.Data.Relationships["tender-job-schedule-shift"]; ok && rel.Data != nil {
		row.TenderJobScheduleShiftID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["driver"]; ok && rel.Data != nil {
		row.DriverID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		row.CreatedByID = rel.Data.ID
	}

	return row
}
