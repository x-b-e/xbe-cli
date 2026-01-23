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

type jobScheduleShiftStartAtChangesListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	CreatedAtMin string
	CreatedAtMax string
	UpdatedAtMin string
	UpdatedAtMax string
}

type jobScheduleShiftStartAtChangeRow struct {
	ID                 string `json:"id"`
	JobScheduleShiftID string `json:"job_schedule_shift_id,omitempty"`
	OldStartAt         string `json:"old_start_at,omitempty"`
	NewStartAt         string `json:"new_start_at,omitempty"`
	CreatedAt          string `json:"created_at,omitempty"`
}

func newJobScheduleShiftStartAtChangesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List job schedule shift start-at changes",
		Long: `List job schedule shift start-at changes with filtering and pagination.

Output Columns:
  ID          Change identifier
  SHIFT       Job schedule shift ID
  NEW_START   New shift start time
  OLD_START   Previous shift start time
  CREATED_AT  When the change was created

Filters:
  --created-at-min  Filter by created-at on/after (ISO 8601)
  --created-at-max  Filter by created-at on/before (ISO 8601)
  --updated-at-min  Filter by updated-at on/after (ISO 8601)
  --updated-at-max  Filter by updated-at on/before (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List start-at changes
  xbe view job-schedule-shift-start-at-changes list

  # Filter by created-at range
  xbe view job-schedule-shift-start-at-changes list \
    --created-at-min 2025-01-01T00:00:00Z \
    --created-at-max 2025-12-31T23:59:59Z

  # Output as JSON
  xbe view job-schedule-shift-start-at-changes list --json`,
		Args: cobra.NoArgs,
		RunE: runJobScheduleShiftStartAtChangesList,
	}
	initJobScheduleShiftStartAtChangesListFlags(cmd)
	return cmd
}

func init() {
	jobScheduleShiftStartAtChangesCmd.AddCommand(newJobScheduleShiftStartAtChangesListCmd())
}

func initJobScheduleShiftStartAtChangesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobScheduleShiftStartAtChangesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseJobScheduleShiftStartAtChangesListOptions(cmd)
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
	query.Set("fields[job-schedule-shift-start-at-changes]", "new-start-at,old-start-at,created-at,job-schedule-shift")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/job-schedule-shift-start-at-changes", query)
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

	rows := buildJobScheduleShiftStartAtChangeRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderJobScheduleShiftStartAtChangesTable(cmd, rows)
}

func parseJobScheduleShiftStartAtChangesListOptions(cmd *cobra.Command) (jobScheduleShiftStartAtChangesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobScheduleShiftStartAtChangesListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		CreatedAtMin: createdAtMin,
		CreatedAtMax: createdAtMax,
		UpdatedAtMin: updatedAtMin,
		UpdatedAtMax: updatedAtMax,
	}, nil
}

func buildJobScheduleShiftStartAtChangeRows(resp jsonAPIResponse) []jobScheduleShiftStartAtChangeRow {
	rows := make([]jobScheduleShiftStartAtChangeRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := jobScheduleShiftStartAtChangeRow{
			ID:         resource.ID,
			OldStartAt: formatDateTime(stringAttr(attrs, "old-start-at")),
			NewStartAt: formatDateTime(stringAttr(attrs, "new-start-at")),
			CreatedAt:  formatDateTime(stringAttr(attrs, "created-at")),
		}

		if rel, ok := resource.Relationships["job-schedule-shift"]; ok && rel.Data != nil {
			row.JobScheduleShiftID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderJobScheduleShiftStartAtChangesTable(cmd *cobra.Command, rows []jobScheduleShiftStartAtChangeRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No job schedule shift start-at changes found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSHIFT\tNEW_START\tOLD_START\tCREATED_AT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.JobScheduleShiftID,
			row.NewStartAt,
			row.OldStartAt,
			row.CreatedAt,
		)
	}
	return writer.Flush()
}
