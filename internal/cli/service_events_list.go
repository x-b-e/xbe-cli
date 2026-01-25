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

type serviceEventsListOptions struct {
	BaseURL                          string
	Token                            string
	JSON                             bool
	NoAuth                           bool
	Limit                            int
	Offset                           int
	Sort                             string
	TenderJobScheduleShiftID         string
	Kind                             string
	OccurredAtMin                    string
	OccurredAtMax                    string
	IsOccurredAt                     string
	CreatedByID                      string
	ViaGPS                           string
	ViaMaterialTransactionAcceptance string
	JobScheduleShiftID               string
}

type serviceEventRow struct {
	ID                               string `json:"id"`
	TenderJobScheduleShiftID         string `json:"tender_job_schedule_shift_id,omitempty"`
	OccurredAt                       string `json:"occurred_at,omitempty"`
	Kind                             string `json:"kind,omitempty"`
	Note                             string `json:"note,omitempty"`
	ViaGPS                           bool   `json:"via_gps"`
	ViaMaterialTransactionAcceptance bool   `json:"via_material_transaction_acceptance"`
	CreatedByID                      string `json:"created_by_id,omitempty"`
	UpdatedByID                      string `json:"updated_by_id,omitempty"`
}

func newServiceEventsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List service events",
		Long: `List service events.

Output Columns:
  ID          Service event identifier
  SHIFT       Tender job schedule shift ID
  KIND        Event kind
  OCCURRED AT Event timestamp
  NOTE        Event note
  GPS         Whether event was via GPS
  MTXN        Whether event was via material transaction acceptance

Filters:
  --tender-job-schedule-shift             Filter by tender job schedule shift ID
  --kind                                 Filter by event kind (ready_to_work, work_start_at)
  --occurred-at-min                       Filter by occurred-at on/after (ISO 8601)
  --occurred-at-max                       Filter by occurred-at on/before (ISO 8601)
  --is-occurred-at                        Filter by presence of occurred-at (true/false)
  --created-by                            Filter by created-by user ID
  --via-gps                               Filter by via GPS (true/false)
  --via-material-transaction-acceptance   Filter by via material transaction acceptance (true/false)
  --job-schedule-shift                    Filter by job schedule shift ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List service events
  xbe view service-events list

  # Filter by tender job schedule shift
  xbe view service-events list --tender-job-schedule-shift 123

  # Filter by kind
  xbe view service-events list --kind ready_to_work

  # Filter by occurred-at range
  xbe view service-events list \
    --occurred-at-min 2026-01-22T00:00:00Z \
    --occurred-at-max 2026-01-23T00:00:00Z

  # Output as JSON
  xbe view service-events list --json`,
		Args: cobra.NoArgs,
		RunE: runServiceEventsList,
	}
	initServiceEventsListFlags(cmd)
	return cmd
}

func init() {
	serviceEventsCmd.AddCommand(newServiceEventsListCmd())
}

func initServiceEventsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("tender-job-schedule-shift", "", "Filter by tender job schedule shift ID")
	cmd.Flags().String("kind", "", "Filter by event kind (ready_to_work, work_start_at)")
	cmd.Flags().String("occurred-at-min", "", "Filter by occurred-at on/after (ISO 8601)")
	cmd.Flags().String("occurred-at-max", "", "Filter by occurred-at on/before (ISO 8601)")
	cmd.Flags().String("is-occurred-at", "", "Filter by presence of occurred-at (true/false)")
	cmd.Flags().String("created-by", "", "Filter by created-by user ID")
	cmd.Flags().String("via-gps", "", "Filter by via GPS (true/false)")
	cmd.Flags().String("via-material-transaction-acceptance", "", "Filter by via material transaction acceptance (true/false)")
	cmd.Flags().String("job-schedule-shift", "", "Filter by job schedule shift ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runServiceEventsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseServiceEventsListOptions(cmd)
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
	query.Set("fields[service-events]", "occurred-at,kind,note,via-gps,via-material-transaction-acceptance,tender-job-schedule-shift,created-by,updated-by")

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
	setFilterIfPresent(query, "filter[kind]", opts.Kind)
	setFilterIfPresent(query, "filter[occurred-at-min]", opts.OccurredAtMin)
	setFilterIfPresent(query, "filter[occurred-at-max]", opts.OccurredAtMax)
	setFilterIfPresent(query, "filter[is-occurred-at]", opts.IsOccurredAt)
	setFilterIfPresent(query, "filter[created-by]", opts.CreatedByID)
	setFilterIfPresent(query, "filter[via-gps]", opts.ViaGPS)
	setFilterIfPresent(query, "filter[via-material-transaction-acceptance]", opts.ViaMaterialTransactionAcceptance)
	setFilterIfPresent(query, "filter[job-schedule-shift]", opts.JobScheduleShiftID)

	body, _, err := client.Get(cmd.Context(), "/v1/service-events", query)
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

	rows := buildServiceEventRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderServiceEventsTable(cmd, rows)
}

func parseServiceEventsListOptions(cmd *cobra.Command) (serviceEventsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	tenderJobScheduleShiftID, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	kind, _ := cmd.Flags().GetString("kind")
	occurredAtMin, _ := cmd.Flags().GetString("occurred-at-min")
	occurredAtMax, _ := cmd.Flags().GetString("occurred-at-max")
	isOccurredAt, _ := cmd.Flags().GetString("is-occurred-at")
	createdByID, _ := cmd.Flags().GetString("created-by")
	viaGPS, _ := cmd.Flags().GetString("via-gps")
	viaMaterialTransactionAcceptance, _ := cmd.Flags().GetString("via-material-transaction-acceptance")
	jobScheduleShiftID, _ := cmd.Flags().GetString("job-schedule-shift")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return serviceEventsListOptions{
		BaseURL:                          baseURL,
		Token:                            token,
		JSON:                             jsonOut,
		NoAuth:                           noAuth,
		Limit:                            limit,
		Offset:                           offset,
		Sort:                             sort,
		TenderJobScheduleShiftID:         tenderJobScheduleShiftID,
		Kind:                             kind,
		OccurredAtMin:                    occurredAtMin,
		OccurredAtMax:                    occurredAtMax,
		IsOccurredAt:                     isOccurredAt,
		CreatedByID:                      createdByID,
		ViaGPS:                           viaGPS,
		ViaMaterialTransactionAcceptance: viaMaterialTransactionAcceptance,
		JobScheduleShiftID:               jobScheduleShiftID,
	}, nil
}

func buildServiceEventRows(resp jsonAPIResponse) []serviceEventRow {
	rows := make([]serviceEventRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := serviceEventRow{
			ID:                               resource.ID,
			OccurredAt:                       formatDateTime(stringAttr(attrs, "occurred-at")),
			Kind:                             stringAttr(attrs, "kind"),
			Note:                             stringAttr(attrs, "note"),
			ViaGPS:                           boolAttr(attrs, "via-gps"),
			ViaMaterialTransactionAcceptance: boolAttr(attrs, "via-material-transaction-acceptance"),
		}

		if rel, ok := resource.Relationships["tender-job-schedule-shift"]; ok && rel.Data != nil {
			row.TenderJobScheduleShiftID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
			row.CreatedByID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["updated-by"]; ok && rel.Data != nil {
			row.UpdatedByID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func buildServiceEventRowFromSingle(resp jsonAPISingleResponse) serviceEventRow {
	resource := resp.Data
	attrs := resource.Attributes
	row := serviceEventRow{
		ID:                               resource.ID,
		OccurredAt:                       formatDateTime(stringAttr(attrs, "occurred-at")),
		Kind:                             stringAttr(attrs, "kind"),
		Note:                             stringAttr(attrs, "note"),
		ViaGPS:                           boolAttr(attrs, "via-gps"),
		ViaMaterialTransactionAcceptance: boolAttr(attrs, "via-material-transaction-acceptance"),
	}

	if rel, ok := resource.Relationships["tender-job-schedule-shift"]; ok && rel.Data != nil {
		row.TenderJobScheduleShiftID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		row.CreatedByID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["updated-by"]; ok && rel.Data != nil {
		row.UpdatedByID = rel.Data.ID
	}

	return row
}

func renderServiceEventsTable(cmd *cobra.Command, rows []serviceEventRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No service events found.")
		return nil
	}

	const noteMax = 40

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, 32, 0)
	fmt.Fprintln(writer, "ID\tSHIFT\tKIND\tOCCURRED AT\tNOTE\tGPS\tMTXN")

	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%t\t%t\n",
			row.ID,
			row.TenderJobScheduleShiftID,
			row.Kind,
			row.OccurredAt,
			truncateString(row.Note, noteMax),
			row.ViaGPS,
			row.ViaMaterialTransactionAcceptance,
		)
	}

	return writer.Flush()
}
