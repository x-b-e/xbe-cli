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

type jobScheduleShiftStartSiteChangesListOptions struct {
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

type jobScheduleShiftStartSiteChangeRow struct {
	ID               string `json:"id"`
	JobScheduleShift string `json:"job_schedule_shift_id,omitempty"`
	OldStartSiteType string `json:"old_start_site_type,omitempty"`
	OldStartSiteID   string `json:"old_start_site_id,omitempty"`
	NewStartSiteType string `json:"new_start_site_type,omitempty"`
	NewStartSiteID   string `json:"new_start_site_id,omitempty"`
	CreatedByID      string `json:"created_by_id,omitempty"`
	CreatedAt        string `json:"created_at,omitempty"`
}

func newJobScheduleShiftStartSiteChangesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List job schedule shift start site changes",
		Long: `List job schedule shift start site changes.

Output Columns:
  ID            Start site change identifier
  SHIFT         Job schedule shift ID
  OLD START     Previous start site (type/id)
  NEW START     New start site (type/id)
  CREATED BY    User who created the change
  CREATED AT    When the change was created

Filters:
  --created-at-min  Filter by created-at on/after (ISO 8601)
  --created-at-max  Filter by created-at on/before (ISO 8601)
  --updated-at-min  Filter by updated-at on/after (ISO 8601)
  --updated-at-max  Filter by updated-at on/before (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List start site changes
  xbe view job-schedule-shift-start-site-changes list

  # Filter by date range
  xbe view job-schedule-shift-start-site-changes list --created-at-min 2025-01-01T00:00:00Z

  # Output as JSON
  xbe view job-schedule-shift-start-site-changes list --json`,
		Args: cobra.NoArgs,
		RunE: runJobScheduleShiftStartSiteChangesList,
	}
	initJobScheduleShiftStartSiteChangesListFlags(cmd)
	return cmd
}

func init() {
	jobScheduleShiftStartSiteChangesCmd.AddCommand(newJobScheduleShiftStartSiteChangesListCmd())
}

func initJobScheduleShiftStartSiteChangesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
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

func runJobScheduleShiftStartSiteChangesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseJobScheduleShiftStartSiteChangesListOptions(cmd)
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
	query.Set("fields[job-schedule-shift-start-site-changes]", "job-schedule-shift,old-start-site,new-start-site,created-by,created-at")

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

	body, _, err := client.Get(cmd.Context(), "/v1/job-schedule-shift-start-site-changes", query)
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

	rows := buildJobScheduleShiftStartSiteChangeRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderJobScheduleShiftStartSiteChangesTable(cmd, rows)
}

func parseJobScheduleShiftStartSiteChangesListOptions(cmd *cobra.Command) (jobScheduleShiftStartSiteChangesListOptions, error) {
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

	return jobScheduleShiftStartSiteChangesListOptions{
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

func buildJobScheduleShiftStartSiteChangeRows(resp jsonAPIResponse) []jobScheduleShiftStartSiteChangeRow {
	rows := make([]jobScheduleShiftStartSiteChangeRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := buildJobScheduleShiftStartSiteChangeRow(resource)
		rows = append(rows, row)
	}
	return rows
}

func buildJobScheduleShiftStartSiteChangeRow(resource jsonAPIResource) jobScheduleShiftStartSiteChangeRow {
	row := jobScheduleShiftStartSiteChangeRow{
		ID:        resource.ID,
		CreatedAt: formatDateTime(stringAttr(resource.Attributes, "created-at")),
	}

	if rel, ok := resource.Relationships["job-schedule-shift"]; ok && rel.Data != nil {
		row.JobScheduleShift = rel.Data.ID
	}
	if rel, ok := resource.Relationships["old-start-site"]; ok && rel.Data != nil {
		row.OldStartSiteType = rel.Data.Type
		row.OldStartSiteID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["new-start-site"]; ok && rel.Data != nil {
		row.NewStartSiteType = rel.Data.Type
		row.NewStartSiteID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		row.CreatedByID = rel.Data.ID
	}

	return row
}

func renderJobScheduleShiftStartSiteChangesTable(cmd *cobra.Command, rows []jobScheduleShiftStartSiteChangeRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No job schedule shift start site changes found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSHIFT\tOLD START\tNEW START\tCREATED BY\tCREATED AT")
	for _, row := range rows {
		oldStart := formatResourceRef(row.OldStartSiteType, row.OldStartSiteID)
		newStart := formatResourceRef(row.NewStartSiteType, row.NewStartSiteID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.JobScheduleShift,
			truncateString(oldStart, 28),
			truncateString(newStart, 28),
			row.CreatedByID,
			row.CreatedAt,
		)
	}
	return writer.Flush()
}

func formatResourceRef(resourceType, resourceID string) string {
	if resourceType == "" && resourceID == "" {
		return ""
	}
	if resourceType != "" && resourceID != "" {
		return resourceType + "/" + resourceID
	}
	if resourceType != "" {
		return resourceType
	}
	return resourceID
}
