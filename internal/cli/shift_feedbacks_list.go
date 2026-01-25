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

type shiftFeedbacksListOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	NoAuth                 bool
	Limit                  int
	Offset                 int
	TenderJobScheduleShift string
	Reason                 string
	Rating                 string
	CreatedBy              string
	Kind                   string
	Automated              string
	Driver                 string
	TruckerID              string
	Trucker                string
	Broker                 string
	Customer               string
	ShiftDateMin           string
	ShiftDateMax           string
}

type shiftFeedbackRow struct {
	ID                       string `json:"id"`
	Rating                   int    `json:"rating,omitempty"`
	Note                     string `json:"note,omitempty"`
	CreatedByBot             bool   `json:"created_by_bot"`
	TenderJobScheduleShiftID string `json:"tender_job_schedule_shift_id,omitempty"`
	ReasonID                 string `json:"reason_id,omitempty"`
	CreatedByID              string `json:"created_by_id,omitempty"`
	TruckerID                string `json:"trucker_id,omitempty"`
}

func newShiftFeedbacksListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List shift feedbacks",
		Long: `List shift feedbacks (trucker/driver performance feedback).

Output Columns:
  ID          Shift feedback identifier
  RATING      Feedback rating
  NOTE        Feedback note
  AUTO        Whether auto-generated
  SHIFT ID    Tender job schedule shift ID
  REASON ID   Shift feedback reason ID
  CREATED BY  User who created the feedback

Filters:
  --tender-job-schedule-shift   Filter by tender job schedule shift ID
  --reason                      Filter by shift feedback reason ID
  --rating                      Filter by rating
  --created-by                  Filter by created-by user ID
  --kind                        Filter by kind
  --automated                   Filter by automated status (true/false)
  --driver                      Filter by driver user ID
  --trucker-id                  Filter by trucker ID
  --trucker                     Filter by trucker ID
  --broker                      Filter by broker ID
  --customer                    Filter by customer ID
  --shift-date-min              Filter by minimum shift date
  --shift-date-max              Filter by maximum shift date`,
		Example: `  # List all shift feedbacks
  xbe view shift-feedbacks list

  # Filter by broker
  xbe view shift-feedbacks list --broker 123

  # Filter by rating
  xbe view shift-feedbacks list --rating 5

  # Output as JSON
  xbe view shift-feedbacks list --json`,
		RunE: runShiftFeedbacksList,
	}
	initShiftFeedbacksListFlags(cmd)
	return cmd
}

func init() {
	shiftFeedbacksCmd.AddCommand(newShiftFeedbacksListCmd())
}

func initShiftFeedbacksListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("tender-job-schedule-shift", "", "Filter by tender job schedule shift ID")
	cmd.Flags().String("reason", "", "Filter by shift feedback reason ID")
	cmd.Flags().String("rating", "", "Filter by rating")
	cmd.Flags().String("created-by", "", "Filter by created-by user ID")
	cmd.Flags().String("kind", "", "Filter by kind")
	cmd.Flags().String("automated", "", "Filter by automated status (true/false)")
	cmd.Flags().String("driver", "", "Filter by driver user ID")
	cmd.Flags().String("trucker-id", "", "Filter by trucker ID")
	cmd.Flags().String("trucker", "", "Filter by trucker ID")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("customer", "", "Filter by customer ID")
	cmd.Flags().String("shift-date-min", "", "Filter by minimum shift date")
	cmd.Flags().String("shift-date-max", "", "Filter by maximum shift date")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runShiftFeedbacksList(cmd *cobra.Command, _ []string) error {
	opts, err := parseShiftFeedbacksListOptions(cmd)
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

	setFilterIfPresent(query, "filter[tender_job_schedule_shift]", opts.TenderJobScheduleShift)
	setFilterIfPresent(query, "filter[reason]", opts.Reason)
	setFilterIfPresent(query, "filter[rating]", opts.Rating)
	setFilterIfPresent(query, "filter[created_by]", opts.CreatedBy)
	setFilterIfPresent(query, "filter[kind]", opts.Kind)
	setFilterIfPresent(query, "filter[automated]", opts.Automated)
	setFilterIfPresent(query, "filter[driver]", opts.Driver)
	setFilterIfPresent(query, "filter[trucker_id]", opts.TruckerID)
	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[customer]", opts.Customer)
	setFilterIfPresent(query, "filter[shift_date_min]", opts.ShiftDateMin)
	setFilterIfPresent(query, "filter[shift_date_max]", opts.ShiftDateMax)

	body, _, err := client.Get(cmd.Context(), "/v1/shift-feedbacks", query)
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

	rows := buildShiftFeedbackRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderShiftFeedbacksTable(cmd, rows)
}

func parseShiftFeedbacksListOptions(cmd *cobra.Command) (shiftFeedbacksListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	tenderJobScheduleShift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	reason, _ := cmd.Flags().GetString("reason")
	rating, _ := cmd.Flags().GetString("rating")
	createdBy, _ := cmd.Flags().GetString("created-by")
	kind, _ := cmd.Flags().GetString("kind")
	automated, _ := cmd.Flags().GetString("automated")
	driver, _ := cmd.Flags().GetString("driver")
	truckerID, _ := cmd.Flags().GetString("trucker-id")
	trucker, _ := cmd.Flags().GetString("trucker")
	broker, _ := cmd.Flags().GetString("broker")
	customer, _ := cmd.Flags().GetString("customer")
	shiftDateMin, _ := cmd.Flags().GetString("shift-date-min")
	shiftDateMax, _ := cmd.Flags().GetString("shift-date-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return shiftFeedbacksListOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		NoAuth:                 noAuth,
		Limit:                  limit,
		Offset:                 offset,
		TenderJobScheduleShift: tenderJobScheduleShift,
		Reason:                 reason,
		Rating:                 rating,
		CreatedBy:              createdBy,
		Kind:                   kind,
		Automated:              automated,
		Driver:                 driver,
		TruckerID:              truckerID,
		Trucker:                trucker,
		Broker:                 broker,
		Customer:               customer,
		ShiftDateMin:           shiftDateMin,
		ShiftDateMax:           shiftDateMax,
	}, nil
}

func buildShiftFeedbackRows(resp jsonAPIResponse) []shiftFeedbackRow {
	rows := make([]shiftFeedbackRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := shiftFeedbackRow{
			ID:           resource.ID,
			Rating:       intAttr(resource.Attributes, "rating"),
			Note:         stringAttr(resource.Attributes, "note"),
			CreatedByBot: boolAttr(resource.Attributes, "created-by-bot"),
		}

		if rel, ok := resource.Relationships["tender-job-schedule-shift"]; ok && rel.Data != nil {
			row.TenderJobScheduleShiftID = rel.Data.ID
		}

		if rel, ok := resource.Relationships["reason"]; ok && rel.Data != nil {
			row.ReasonID = rel.Data.ID
		}

		if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
			row.CreatedByID = rel.Data.ID
		}

		if rel, ok := resource.Relationships["trucker"]; ok && rel.Data != nil {
			row.TruckerID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderShiftFeedbacksTable(cmd *cobra.Command, rows []shiftFeedbackRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No shift feedbacks found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tRATING\tNOTE\tAUTO\tSHIFT ID\tREASON ID\tCREATED BY")
	for _, row := range rows {
		auto := "no"
		if row.CreatedByBot {
			auto = "yes"
		}
		fmt.Fprintf(writer, "%s\t%d\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Rating,
			truncateString(row.Note, 30),
			auto,
			row.TenderJobScheduleShiftID,
			row.ReasonID,
			row.CreatedByID,
		)
	}
	return writer.Flush()
}
