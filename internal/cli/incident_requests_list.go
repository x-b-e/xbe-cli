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

type incidentRequestsListOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	NoAuth                 bool
	Limit                  int
	Offset                 int
	Sort                   string
	TenderJobScheduleShift string
	Broker                 string
	Customer               string
	Assignee               string
	CreatedBy              string
	StartAtMin             string
	StartAtMax             string
	EndAtMin               string
	EndAtMax               string
	Status                 string
	TimeValueType          string
}

type incidentRequestRow struct {
	ID                       string `json:"id"`
	Status                   string `json:"status,omitempty"`
	StartAt                  string `json:"start_at,omitempty"`
	EndAt                    string `json:"end_at,omitempty"`
	TimeValueType            string `json:"time_value_type,omitempty"`
	IsDownTime               bool   `json:"is_down_time"`
	Description              string `json:"description,omitempty"`
	TenderJobScheduleShiftID string `json:"tender_job_schedule_shift_id,omitempty"`
	AssigneeID               string `json:"assignee_id,omitempty"`
	CreatedByID              string `json:"created_by_id,omitempty"`
	CustomerID               string `json:"customer_id,omitempty"`
	BrokerID                 string `json:"broker_id,omitempty"`
	IncidentID               string `json:"incident_id,omitempty"`
}

func newIncidentRequestsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List incident requests",
		Long: `List incident requests with filtering and pagination.

Output Columns:
  ID          Incident request identifier
  STATUS      Request status
  START AT    Start timestamp
  END AT      End timestamp
  TIME VALUE  Time value type (credited or deducted)
  DOWN TIME   Indicates whether this is downtime
  SHIFT       Tender job schedule shift ID
  ASSIGNEE    Assignee user ID

Filters:
  --tender-job-schedule-shift  Filter by tender job schedule shift ID
  --broker                     Filter by broker ID
  --customer                   Filter by customer ID
  --assignee                   Filter by assignee user ID
  --created-by                 Filter by created-by user ID
  --start-at-min               Filter by start-at on/after (ISO 8601)
  --start-at-max               Filter by start-at on/before (ISO 8601)
  --end-at-min                 Filter by end-at on/after (ISO 8601)
  --end-at-max                 Filter by end-at on/before (ISO 8601)
  --status                     Filter by status
  --time-value-type            Filter by time value type

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List incident requests
  xbe view incident-requests list

  # Filter by status
  xbe view incident-requests list --status submitted

  # Filter by shift
  xbe view incident-requests list --tender-job-schedule-shift 123

  # Filter by time range
  xbe view incident-requests list --start-at-min 2025-01-01T00:00:00Z --end-at-max 2025-01-31T23:59:59Z

  # Output as JSON
  xbe view incident-requests list --json`,
		Args: cobra.NoArgs,
		RunE: runIncidentRequestsList,
	}
	initIncidentRequestsListFlags(cmd)
	return cmd
}

func init() {
	incidentRequestsCmd.AddCommand(newIncidentRequestsListCmd())
}

func initIncidentRequestsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("tender-job-schedule-shift", "", "Filter by tender job schedule shift ID")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("customer", "", "Filter by customer ID")
	cmd.Flags().String("assignee", "", "Filter by assignee user ID")
	cmd.Flags().String("created-by", "", "Filter by created-by user ID")
	cmd.Flags().String("start-at-min", "", "Filter by start-at on/after (ISO 8601)")
	cmd.Flags().String("start-at-max", "", "Filter by start-at on/before (ISO 8601)")
	cmd.Flags().String("end-at-min", "", "Filter by end-at on/after (ISO 8601)")
	cmd.Flags().String("end-at-max", "", "Filter by end-at on/before (ISO 8601)")
	cmd.Flags().String("status", "", "Filter by status")
	cmd.Flags().String("time-value-type", "", "Filter by time value type")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runIncidentRequestsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseIncidentRequestsListOptions(cmd)
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
	query.Set("fields[incident-requests]", "status,start-at,end-at,description,is-down-time,time-value-type,tender-job-schedule-shift,assignee,created-by,customer,broker,incident")

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
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[customer]", opts.Customer)
	setFilterIfPresent(query, "filter[assignee]", opts.Assignee)
	setFilterIfPresent(query, "filter[created_by]", opts.CreatedBy)
	setFilterIfPresent(query, "filter[start_at_min]", opts.StartAtMin)
	setFilterIfPresent(query, "filter[start_at_max]", opts.StartAtMax)
	setFilterIfPresent(query, "filter[end_at_min]", opts.EndAtMin)
	setFilterIfPresent(query, "filter[end_at_max]", opts.EndAtMax)
	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[time_value_type]", opts.TimeValueType)

	body, _, err := client.Get(cmd.Context(), "/v1/incident-requests", query)
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

	rows := buildIncidentRequestRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderIncidentRequestsTable(cmd, rows)
}

func parseIncidentRequestsListOptions(cmd *cobra.Command) (incidentRequestsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	tenderJobScheduleShift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	broker, _ := cmd.Flags().GetString("broker")
	customer, _ := cmd.Flags().GetString("customer")
	assignee, _ := cmd.Flags().GetString("assignee")
	createdBy, _ := cmd.Flags().GetString("created-by")
	startAtMin, _ := cmd.Flags().GetString("start-at-min")
	startAtMax, _ := cmd.Flags().GetString("start-at-max")
	endAtMin, _ := cmd.Flags().GetString("end-at-min")
	endAtMax, _ := cmd.Flags().GetString("end-at-max")
	status, _ := cmd.Flags().GetString("status")
	timeValueType, _ := cmd.Flags().GetString("time-value-type")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return incidentRequestsListOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		NoAuth:                 noAuth,
		Limit:                  limit,
		Offset:                 offset,
		Sort:                   sort,
		TenderJobScheduleShift: tenderJobScheduleShift,
		Broker:                 broker,
		Customer:               customer,
		Assignee:               assignee,
		CreatedBy:              createdBy,
		StartAtMin:             startAtMin,
		StartAtMax:             startAtMax,
		EndAtMin:               endAtMin,
		EndAtMax:               endAtMax,
		Status:                 status,
		TimeValueType:          timeValueType,
	}, nil
}

func buildIncidentRequestRows(resp jsonAPIResponse) []incidentRequestRow {
	rows := make([]incidentRequestRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := incidentRequestRow{
			ID:                       resource.ID,
			Status:                   stringAttr(attrs, "status"),
			StartAt:                  formatDateTime(stringAttr(attrs, "start-at")),
			EndAt:                    formatDateTime(stringAttr(attrs, "end-at")),
			TimeValueType:            stringAttr(attrs, "time-value-type"),
			IsDownTime:               boolAttr(attrs, "is-down-time"),
			Description:              stringAttr(attrs, "description"),
			TenderJobScheduleShiftID: relationshipIDFromMap(resource.Relationships, "tender-job-schedule-shift"),
			AssigneeID:               relationshipIDFromMap(resource.Relationships, "assignee"),
			CreatedByID:              relationshipIDFromMap(resource.Relationships, "created-by"),
			CustomerID:               relationshipIDFromMap(resource.Relationships, "customer"),
			BrokerID:                 relationshipIDFromMap(resource.Relationships, "broker"),
			IncidentID:               relationshipIDFromMap(resource.Relationships, "incident"),
		}
		rows = append(rows, row)
	}
	return rows
}

func renderIncidentRequestsTable(cmd *cobra.Command, rows []incidentRequestRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No incident requests found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tSTART AT\tEND AT\tTIME VALUE\tDOWN TIME\tSHIFT\tASSIGNEE")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Status, 12),
			truncateString(row.StartAt, 20),
			truncateString(row.EndAt, 20),
			truncateString(row.TimeValueType, 16),
			formatYesNo(row.IsDownTime),
			truncateString(row.TenderJobScheduleShiftID, 20),
			truncateString(row.AssigneeID, 16),
		)
	}
	return writer.Flush()
}
