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

type driverAssignmentAcknowledgementsListOptions struct {
	BaseURL                  string
	Token                    string
	JSON                     bool
	NoAuth                   bool
	Limit                    int
	Offset                   int
	Sort                     string
	TenderJobScheduleShiftID string
	DriverID                 string
}

type driverAssignmentAcknowledgementRow struct {
	ID                       string `json:"id"`
	TenderJobScheduleShiftID string `json:"tender_job_schedule_shift_id,omitempty"`
	DriverID                 string `json:"driver_id,omitempty"`
	CreatedByID              string `json:"created_by_id,omitempty"`
	CreatedAt                string `json:"created_at,omitempty"`
	UpdatedAt                string `json:"updated_at,omitempty"`
}

func newDriverAssignmentAcknowledgementsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List driver assignment acknowledgements",
		Long: `List driver assignment acknowledgements.

Output Columns:
  ID          Acknowledgement identifier
  SHIFT       Tender job schedule shift ID
  DRIVER      Driver user ID
  CREATED BY  Creator user ID
  CREATED AT  Created timestamp

Filters:
  --tender-job-schedule-shift  Filter by tender job schedule shift ID
  --driver                     Filter by driver user ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List acknowledgements
  xbe view driver-assignment-acknowledgements list

  # Filter by tender job schedule shift
  xbe view driver-assignment-acknowledgements list --tender-job-schedule-shift 123

  # Filter by driver
  xbe view driver-assignment-acknowledgements list --driver 456

  # Output as JSON
  xbe view driver-assignment-acknowledgements list --json`,
		Args: cobra.NoArgs,
		RunE: runDriverAssignmentAcknowledgementsList,
	}
	initDriverAssignmentAcknowledgementsListFlags(cmd)
	return cmd
}

func init() {
	driverAssignmentAcknowledgementsCmd.AddCommand(newDriverAssignmentAcknowledgementsListCmd())
}

func initDriverAssignmentAcknowledgementsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("tender-job-schedule-shift", "", "Filter by tender job schedule shift ID")
	cmd.Flags().String("driver", "", "Filter by driver user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDriverAssignmentAcknowledgementsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseDriverAssignmentAcknowledgementsListOptions(cmd)
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
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
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

	setFilterIfPresent(query, "filter[tender-job-schedule-shift]", opts.TenderJobScheduleShiftID)
	setFilterIfPresent(query, "filter[driver]", opts.DriverID)

	body, _, err := client.Get(cmd.Context(), "/v1/driver-assignment-acknowledgements", query)
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

	rows := buildDriverAssignmentAcknowledgementRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderDriverAssignmentAcknowledgementsTable(cmd, rows)
}

func parseDriverAssignmentAcknowledgementsListOptions(cmd *cobra.Command) (driverAssignmentAcknowledgementsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	tenderJobScheduleShiftID, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	driverID, _ := cmd.Flags().GetString("driver")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return driverAssignmentAcknowledgementsListOptions{
		BaseURL:                  baseURL,
		Token:                    token,
		JSON:                     jsonOut,
		NoAuth:                   noAuth,
		Limit:                    limit,
		Offset:                   offset,
		Sort:                     sort,
		TenderJobScheduleShiftID: tenderJobScheduleShiftID,
		DriverID:                 driverID,
	}, nil
}

func buildDriverAssignmentAcknowledgementRows(resp jsonAPIResponse) []driverAssignmentAcknowledgementRow {
	rows := make([]driverAssignmentAcknowledgementRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := driverAssignmentAcknowledgementRow{
			ID:        resource.ID,
			CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
			UpdatedAt: formatDateTime(stringAttr(attrs, "updated-at")),
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

func buildDriverAssignmentAcknowledgementRowFromSingle(resp jsonAPISingleResponse) driverAssignmentAcknowledgementRow {
	resource := resp.Data
	attrs := resource.Attributes
	row := driverAssignmentAcknowledgementRow{
		ID:        resource.ID,
		CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt: formatDateTime(stringAttr(attrs, "updated-at")),
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

	return row
}

func renderDriverAssignmentAcknowledgementsTable(cmd *cobra.Command, rows []driverAssignmentAcknowledgementRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No driver assignment acknowledgements found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSHIFT\tDRIVER\tCREATED BY\tCREATED AT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.TenderJobScheduleShiftID,
			row.DriverID,
			row.CreatedByID,
			row.CreatedAt,
		)
	}
	return writer.Flush()
}
