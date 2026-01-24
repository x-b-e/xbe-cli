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

type tenderJobScheduleShiftDriversListOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	NoAuth                 bool
	Limit                  int
	Offset                 int
	Sort                   string
	TenderJobScheduleShift string
	User                   string
}

type tenderJobScheduleShiftDriverRow struct {
	ID                       string `json:"id"`
	IsPrimary                bool   `json:"is_primary"`
	TenderJobScheduleShiftID string `json:"tender_job_schedule_shift_id,omitempty"`
	UserID                   string `json:"user_id,omitempty"`
	CreatedByID              string `json:"created_by_id,omitempty"`
	CreatedAt                string `json:"created_at,omitempty"`
}

func newTenderJobScheduleShiftDriversListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tender job schedule shift drivers",
		Long: `List tender job schedule shift drivers.

Output Columns:
  ID          Shift driver identifier
  SHIFT ID    Tender job schedule shift ID
  USER        Driver user ID
  PRIMARY     Whether the driver is primary
  CREATED BY  User who created the shift driver
  CREATED AT  When the shift driver was created

Filters:
  --tender-job-schedule-shift  Filter by tender job schedule shift ID
  --user                       Filter by user ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List shift drivers
  xbe view tender-job-schedule-shift-drivers list

  # Filter by tender job schedule shift
  xbe view tender-job-schedule-shift-drivers list --tender-job-schedule-shift 123

  # Filter by user
  xbe view tender-job-schedule-shift-drivers list --user 456

  # Output as JSON
  xbe view tender-job-schedule-shift-drivers list --json`,
		Args: cobra.NoArgs,
		RunE: runTenderJobScheduleShiftDriversList,
	}
	initTenderJobScheduleShiftDriversListFlags(cmd)
	return cmd
}

func init() {
	tenderJobScheduleShiftDriversCmd.AddCommand(newTenderJobScheduleShiftDriversListCmd())
}

func initTenderJobScheduleShiftDriversListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("tender-job-schedule-shift", "", "Filter by tender job schedule shift ID")
	cmd.Flags().String("user", "", "Filter by user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTenderJobScheduleShiftDriversList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTenderJobScheduleShiftDriversListOptions(cmd)
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
	setFilterIfPresent(query, "filter[user]", opts.User)

	body, _, err := client.Get(cmd.Context(), "/v1/tender-job-schedule-shift-drivers", query)
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

	rows := buildTenderJobScheduleShiftDriverRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTenderJobScheduleShiftDriversTable(cmd, rows)
}

func parseTenderJobScheduleShiftDriversListOptions(cmd *cobra.Command) (tenderJobScheduleShiftDriversListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	tenderJobScheduleShift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	user, _ := cmd.Flags().GetString("user")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return tenderJobScheduleShiftDriversListOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		NoAuth:                 noAuth,
		Limit:                  limit,
		Offset:                 offset,
		Sort:                   sort,
		TenderJobScheduleShift: tenderJobScheduleShift,
		User:                   user,
	}, nil
}

func buildTenderJobScheduleShiftDriverRows(resp jsonAPIResponse) []tenderJobScheduleShiftDriverRow {
	rows := make([]tenderJobScheduleShiftDriverRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := tenderJobScheduleShiftDriverRow{
			ID:        resource.ID,
			IsPrimary: boolAttr(attrs, "is-primary"),
			CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
		}

		if rel, ok := resource.Relationships["tender-job-schedule-shift"]; ok && rel.Data != nil {
			row.TenderJobScheduleShiftID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
			row.UserID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
			row.CreatedByID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderTenderJobScheduleShiftDriversTable(cmd *cobra.Command, rows []tenderJobScheduleShiftDriverRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No tender job schedule shift drivers found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSHIFT ID\tUSER\tPRIMARY\tCREATED BY\tCREATED AT")
	for _, row := range rows {
		primary := "no"
		if row.IsPrimary {
			primary = "yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.TenderJobScheduleShiftID,
			row.UserID,
			primary,
			row.CreatedByID,
			row.CreatedAt,
		)
	}
	return writer.Flush()
}

func buildTenderJobScheduleShiftDriverRowFromSingle(resp jsonAPISingleResponse) tenderJobScheduleShiftDriverRow {
	attrs := resp.Data.Attributes
	row := tenderJobScheduleShiftDriverRow{
		ID:        resp.Data.ID,
		IsPrimary: boolAttr(attrs, "is-primary"),
		CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
	}

	if rel, ok := resp.Data.Relationships["tender-job-schedule-shift"]; ok && rel.Data != nil {
		row.TenderJobScheduleShiftID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["user"]; ok && rel.Data != nil {
		row.UserID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		row.CreatedByID = rel.Data.ID
	}

	return row
}
