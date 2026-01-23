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

type timeSheetsListOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	NoAuth            bool
	Limit             int
	Offset            int
	Sort              string
	SubjectType       string
	SubjectID         string
	StartAtMin        string
	StartAtMax        string
	EndAtMin          string
	EndAtMax          string
	StartOnMin        string
	StartOnMax        string
	Broker            string
	Trucker           string
	Driver            string
	Laborer           string
	LaborerUser       string
	Equipment         string
	HasExports        string
	Status            string
	MissingCraftClass string
}

type timeSheetRow struct {
	ID              string `json:"id"`
	Status          string `json:"status,omitempty"`
	StartAt         string `json:"start_at,omitempty"`
	EndAt           string `json:"end_at,omitempty"`
	DurationMinutes string `json:"duration_minutes,omitempty"`
	SubjectType     string `json:"subject_type,omitempty"`
	SubjectID       string `json:"subject_id,omitempty"`
	DriverID        string `json:"driver_id,omitempty"`
}

func newTimeSheetsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List time sheets",
		Long: `List time sheets with filtering and pagination.

Output Columns:
  ID        Time sheet identifier
  STATUS    Current status
  SUBJECT   Subject type and ID
  START AT  Start timestamp
  END AT    End timestamp
  DURATION  Duration in minutes
  DRIVER    Driver user ID

Filters:
  --subject-type       Filter by subject type (WorkOrder, CrewRequirement, TruckerShiftSet; also accepts work-orders, crew-requirements, trucker-shift-sets)
  --subject-id         Filter by subject ID (use with --subject-type)
  --start-at-min       Filter by minimum start time (ISO 8601)
  --start-at-max       Filter by maximum start time (ISO 8601)
  --end-at-min         Filter by minimum end time (ISO 8601)
  --end-at-max         Filter by maximum end time (ISO 8601)
  --start-on-min       Filter by minimum start date (YYYY-MM-DD)
  --start-on-max       Filter by maximum start date (YYYY-MM-DD)
  --broker             Filter by broker ID
  --trucker            Filter by trucker ID
  --driver             Filter by driver user ID
  --laborer            Filter by laborer ID
  --laborer-user       Filter by laborer user ID
  --equipment          Filter by equipment ID
  --has-exports         Filter by time sheets with exports (true/false)
  --status             Filter by status
  --missing-craft-class Filter by missing craft class (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List time sheets
  xbe view time-sheets list

  # Filter by subject
  xbe view time-sheets list --subject-type WorkOrder --subject-id 123

  # Filter by date range
  xbe view time-sheets list --start-at-min 2026-01-01T00:00:00Z --end-at-max 2026-01-31T23:59:59Z

  # Filter by status
  xbe view time-sheets list --status submitted

  # Output as JSON
  xbe view time-sheets list --json`,
		Args: cobra.NoArgs,
		RunE: runTimeSheetsList,
	}
	initTimeSheetsListFlags(cmd)
	return cmd
}

func init() {
	timeSheetsCmd.AddCommand(newTimeSheetsListCmd())
}

func initTimeSheetsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("subject-type", "", "Filter by subject type (WorkOrder, CrewRequirement, TruckerShiftSet; also accepts work-orders, crew-requirements, trucker-shift-sets)")
	cmd.Flags().String("subject-id", "", "Filter by subject ID (use with --subject-type)")
	cmd.Flags().String("start-at-min", "", "Filter by minimum start time (ISO 8601)")
	cmd.Flags().String("start-at-max", "", "Filter by maximum start time (ISO 8601)")
	cmd.Flags().String("end-at-min", "", "Filter by minimum end time (ISO 8601)")
	cmd.Flags().String("end-at-max", "", "Filter by maximum end time (ISO 8601)")
	cmd.Flags().String("start-on-min", "", "Filter by minimum start date (YYYY-MM-DD)")
	cmd.Flags().String("start-on-max", "", "Filter by maximum start date (YYYY-MM-DD)")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("trucker", "", "Filter by trucker ID")
	cmd.Flags().String("driver", "", "Filter by driver user ID")
	cmd.Flags().String("laborer", "", "Filter by laborer ID")
	cmd.Flags().String("laborer-user", "", "Filter by laborer user ID")
	cmd.Flags().String("equipment", "", "Filter by equipment ID")
	cmd.Flags().String("has-exports", "", "Filter by has exports (true/false)")
	cmd.Flags().String("status", "", "Filter by status")
	cmd.Flags().String("missing-craft-class", "", "Filter by missing craft class (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTimeSheetsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTimeSheetsListOptions(cmd)
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
	query.Set("fields[time-sheets]", "status,start-at,end-at,duration-minutes,subject,driver")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	if opts.SubjectType != "" && opts.SubjectID != "" {
		normalizedType := normalizeTimeSheetSubjectFilter(opts.SubjectType)
		query.Set("filter[subject]", normalizedType+"|"+opts.SubjectID)
	}
	setFilterIfPresent(query, "filter[start_at_min]", opts.StartAtMin)
	setFilterIfPresent(query, "filter[start_at_max]", opts.StartAtMax)
	setFilterIfPresent(query, "filter[end_at_min]", opts.EndAtMin)
	setFilterIfPresent(query, "filter[end_at_max]", opts.EndAtMax)
	setFilterIfPresent(query, "filter[start_on_min]", opts.StartOnMin)
	setFilterIfPresent(query, "filter[start_on_max]", opts.StartOnMax)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)
	setFilterIfPresent(query, "filter[driver]", opts.Driver)
	setFilterIfPresent(query, "filter[laborer]", opts.Laborer)
	setFilterIfPresent(query, "filter[laborer_user]", opts.LaborerUser)
	setFilterIfPresent(query, "filter[equipment]", opts.Equipment)
	setFilterIfPresent(query, "filter[has_exports]", opts.HasExports)
	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[missing_craft_class]", opts.MissingCraftClass)

	body, _, err := client.Get(cmd.Context(), "/v1/time-sheets", query)
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

	rows := buildTimeSheetRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTimeSheetsTable(cmd, rows)
}

func parseTimeSheetsListOptions(cmd *cobra.Command) (timeSheetsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	subjectType, _ := cmd.Flags().GetString("subject-type")
	subjectID, _ := cmd.Flags().GetString("subject-id")
	startAtMin, _ := cmd.Flags().GetString("start-at-min")
	startAtMax, _ := cmd.Flags().GetString("start-at-max")
	endAtMin, _ := cmd.Flags().GetString("end-at-min")
	endAtMax, _ := cmd.Flags().GetString("end-at-max")
	startOnMin, _ := cmd.Flags().GetString("start-on-min")
	startOnMax, _ := cmd.Flags().GetString("start-on-max")
	broker, _ := cmd.Flags().GetString("broker")
	trucker, _ := cmd.Flags().GetString("trucker")
	driver, _ := cmd.Flags().GetString("driver")
	laborer, _ := cmd.Flags().GetString("laborer")
	laborerUser, _ := cmd.Flags().GetString("laborer-user")
	equipment, _ := cmd.Flags().GetString("equipment")
	hasExports, _ := cmd.Flags().GetString("has-exports")
	status, _ := cmd.Flags().GetString("status")
	missingCraftClass, _ := cmd.Flags().GetString("missing-craft-class")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return timeSheetsListOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		NoAuth:            noAuth,
		Limit:             limit,
		Offset:            offset,
		Sort:              sort,
		SubjectType:       subjectType,
		SubjectID:         subjectID,
		StartAtMin:        startAtMin,
		StartAtMax:        startAtMax,
		EndAtMin:          endAtMin,
		EndAtMax:          endAtMax,
		StartOnMin:        startOnMin,
		StartOnMax:        startOnMax,
		Broker:            broker,
		Trucker:           trucker,
		Driver:            driver,
		Laborer:           laborer,
		LaborerUser:       laborerUser,
		Equipment:         equipment,
		HasExports:        hasExports,
		Status:            status,
		MissingCraftClass: missingCraftClass,
	}, nil
}

func buildTimeSheetRows(resp jsonAPIResponse) []timeSheetRow {
	rows := make([]timeSheetRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildTimeSheetRow(resource))
	}
	return rows
}

func buildTimeSheetRow(resource jsonAPIResource) timeSheetRow {
	attrs := resource.Attributes
	row := timeSheetRow{
		ID:              resource.ID,
		Status:          stringAttr(attrs, "status"),
		StartAt:         formatDateTime(stringAttr(attrs, "start-at")),
		EndAt:           formatDateTime(stringAttr(attrs, "end-at")),
		DurationMinutes: stringAttr(attrs, "duration-minutes"),
	}

	if rel, ok := resource.Relationships["subject"]; ok && rel.Data != nil {
		row.SubjectType = rel.Data.Type
		row.SubjectID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["driver"]; ok && rel.Data != nil {
		row.DriverID = rel.Data.ID
	}

	return row
}

func renderTimeSheetsTable(cmd *cobra.Command, rows []timeSheetRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No time sheets found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tSUBJECT\tSTART AT\tEND AT\tDURATION\tDRIVER")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Status,
			formatTimeSheetSubject(row.SubjectType, row.SubjectID),
			row.StartAt,
			row.EndAt,
			row.DurationMinutes,
			row.DriverID,
		)
	}
	return writer.Flush()
}
