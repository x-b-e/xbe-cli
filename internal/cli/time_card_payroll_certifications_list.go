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

type timeCardPayrollCertificationsListOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	NoAuth      bool
	Limit       int
	Offset      int
	Sort        string
	TimeCardID  string
	CreatedByID string
}

type timeCardPayrollCertificationRow struct {
	ID          string `json:"id"`
	TimeCardID  string `json:"time_card_id,omitempty"`
	CreatedByID string `json:"created_by_id,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
}

func newTimeCardPayrollCertificationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List time card payroll certifications",
		Long: `List time card payroll certifications.

Output Columns:
  ID         Payroll certification identifier
  TIME CARD  Time card ID
  CREATED BY User who certified the time card
  CREATED AT Certification timestamp

Filters:
  --time-card   Filter by time card ID
  --created-by  Filter by created-by user ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List time card payroll certifications
  xbe view time-card-payroll-certifications list

  # Filter by time card
  xbe view time-card-payroll-certifications list --time-card 123

  # Filter by creator
  xbe view time-card-payroll-certifications list --created-by 456

  # Output as JSON
  xbe view time-card-payroll-certifications list --json`,
		Args: cobra.NoArgs,
		RunE: runTimeCardPayrollCertificationsList,
	}
	initTimeCardPayrollCertificationsListFlags(cmd)
	return cmd
}

func init() {
	timeCardPayrollCertificationsCmd.AddCommand(newTimeCardPayrollCertificationsListCmd())
}

func initTimeCardPayrollCertificationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("time-card", "", "Filter by time card ID")
	cmd.Flags().String("created-by", "", "Filter by created-by user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTimeCardPayrollCertificationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTimeCardPayrollCertificationsListOptions(cmd)
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
	query.Set("fields[time-card-payroll-certifications]", "created-at,time-card,created-by")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[time_card]", opts.TimeCardID)
	setFilterIfPresent(query, "filter[created_by]", opts.CreatedByID)

	body, _, err := client.Get(cmd.Context(), "/v1/time-card-payroll-certifications", query)
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

	rows := buildTimeCardPayrollCertificationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTimeCardPayrollCertificationsTable(cmd, rows)
}

func parseTimeCardPayrollCertificationsListOptions(cmd *cobra.Command) (timeCardPayrollCertificationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	timeCardID, _ := cmd.Flags().GetString("time-card")
	createdByID, _ := cmd.Flags().GetString("created-by")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return timeCardPayrollCertificationsListOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		NoAuth:      noAuth,
		Limit:       limit,
		Offset:      offset,
		Sort:        sort,
		TimeCardID:  timeCardID,
		CreatedByID: createdByID,
	}, nil
}

func buildTimeCardPayrollCertificationRows(resp jsonAPIResponse) []timeCardPayrollCertificationRow {
	rows := make([]timeCardPayrollCertificationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := timeCardPayrollCertificationRow{
			ID:        resource.ID,
			CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
		}

		if rel, ok := resource.Relationships["time-card"]; ok && rel.Data != nil {
			row.TimeCardID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
			row.CreatedByID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func buildTimeCardPayrollCertificationRowFromSingle(resp jsonAPISingleResponse) timeCardPayrollCertificationRow {
	resource := resp.Data
	attrs := resource.Attributes
	row := timeCardPayrollCertificationRow{
		ID:        resource.ID,
		CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
	}

	if rel, ok := resource.Relationships["time-card"]; ok && rel.Data != nil {
		row.TimeCardID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		row.CreatedByID = rel.Data.ID
	}

	return row
}

func renderTimeCardPayrollCertificationsTable(cmd *cobra.Command, rows []timeCardPayrollCertificationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No time card payroll certifications found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTIME CARD\tCREATED BY\tCREATED AT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			row.TimeCardID,
			row.CreatedByID,
			row.CreatedAt,
		)
	}
	return writer.Flush()
}
