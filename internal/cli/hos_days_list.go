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

type hosDaysListOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	NoAuth            bool
	Limit             int
	Offset            int
	Sort              string
	Broker            string
	Driver            string
	User              string
	ServiceDate       string
	ServiceDateMin    string
	ServiceDateMax    string
	HasServiceDate    string
	RegulationSetCode string
	CreatedAtMin      string
	CreatedAtMax      string
	IsCreatedAt       string
	UpdatedAtMin      string
	UpdatedAtMax      string
	IsUpdatedAt       string
}

type hosDayRow struct {
	ID                string `json:"id"`
	DriverID          string `json:"driver_id,omitempty"`
	BrokerID          string `json:"broker_id,omitempty"`
	ServiceDate       string `json:"service_date,omitempty"`
	RegulationSetCode string `json:"regulation_set_code,omitempty"`
	TimeZoneID        string `json:"time_zone_id,omitempty"`
}

func newHosDaysListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List HOS days",
		Long: `List HOS days.

Output Columns:
  ID            HOS day identifier
  DRIVER        Driver user ID
  SERVICE DATE  Service date (YYYY-MM-DD)
  REG SET       Regulation set code
  TIME ZONE     Time zone ID
  BROKER        Broker ID

Filters:
  --broker               Filter by broker ID
  --driver               Filter by driver user ID
  --user                 Filter by user ID (alias of driver)
  --service-date         Filter by service date (YYYY-MM-DD)
  --service-date-min     Filter by minimum service date (YYYY-MM-DD)
  --service-date-max     Filter by maximum service date (YYYY-MM-DD)
  --has-service-date     Filter by presence of service date (true/false)
  --regulation-set-code  Filter by regulation set code
  --created-at-min       Filter by created-at on/after (ISO 8601)
  --created-at-max       Filter by created-at on/before (ISO 8601)
  --is-created-at        Filter by has created-at (true/false)
  --updated-at-min       Filter by updated-at on/after (ISO 8601)
  --updated-at-max       Filter by updated-at on/before (ISO 8601)
  --is-updated-at        Filter by has updated-at (true/false)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List HOS days
  xbe view hos-days list

  # Filter by driver
  xbe view hos-days list --driver 123

  # Filter by service date range
  xbe view hos-days list --service-date-min 2024-01-01 --service-date-max 2024-01-31

  # Filter by regulation set code
  xbe view hos-days list --regulation-set-code US-70

  # Output as JSON
  xbe view hos-days list --json`,
		Args: cobra.NoArgs,
		RunE: runHosDaysList,
	}
	initHosDaysListFlags(cmd)
	return cmd
}

func init() {
	hosDaysCmd.AddCommand(newHosDaysListCmd())
}

func initHosDaysListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("driver", "", "Filter by driver user ID")
	cmd.Flags().String("user", "", "Filter by user ID (alias of driver)")
	cmd.Flags().String("service-date", "", "Filter by service date (YYYY-MM-DD)")
	cmd.Flags().String("service-date-min", "", "Filter by minimum service date (YYYY-MM-DD)")
	cmd.Flags().String("service-date-max", "", "Filter by maximum service date (YYYY-MM-DD)")
	cmd.Flags().String("has-service-date", "", "Filter by presence of service date (true/false)")
	cmd.Flags().String("regulation-set-code", "", "Filter by regulation set code")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by has created-at (true/false)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("is-updated-at", "", "Filter by has updated-at (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runHosDaysList(cmd *cobra.Command, _ []string) error {
	opts, err := parseHosDaysListOptions(cmd)
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
	query.Set("fields[hos-days]", "service-date,regulation-set-code,time-zone-id,driver,broker")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[driver]", opts.Driver)
	setFilterIfPresent(query, "filter[user]", opts.User)
	setFilterIfPresent(query, "filter[service-date]", opts.ServiceDate)
	setFilterIfPresent(query, "filter[service-date-min]", opts.ServiceDateMin)
	setFilterIfPresent(query, "filter[service-date-max]", opts.ServiceDateMax)
	setFilterIfPresent(query, "filter[has-service-date]", opts.HasServiceDate)
	setFilterIfPresent(query, "filter[regulation-set-code]", opts.RegulationSetCode)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[is-created-at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[is-updated-at]", opts.IsUpdatedAt)

	body, _, err := client.Get(cmd.Context(), "/v1/hos-days", query)
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

	rows := buildHosDayRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderHosDaysTable(cmd, rows)
}

func parseHosDaysListOptions(cmd *cobra.Command) (hosDaysListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	broker, _ := cmd.Flags().GetString("broker")
	driver, _ := cmd.Flags().GetString("driver")
	user, _ := cmd.Flags().GetString("user")
	serviceDate, _ := cmd.Flags().GetString("service-date")
	serviceDateMin, _ := cmd.Flags().GetString("service-date-min")
	serviceDateMax, _ := cmd.Flags().GetString("service-date-max")
	hasServiceDate, _ := cmd.Flags().GetString("has-service-date")
	regulationSetCode, _ := cmd.Flags().GetString("regulation-set-code")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	isCreatedAt, _ := cmd.Flags().GetString("is-created-at")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	isUpdatedAt, _ := cmd.Flags().GetString("is-updated-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return hosDaysListOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		NoAuth:            noAuth,
		Limit:             limit,
		Offset:            offset,
		Sort:              sort,
		Broker:            broker,
		Driver:            driver,
		User:              user,
		ServiceDate:       serviceDate,
		ServiceDateMin:    serviceDateMin,
		ServiceDateMax:    serviceDateMax,
		HasServiceDate:    hasServiceDate,
		RegulationSetCode: regulationSetCode,
		CreatedAtMin:      createdAtMin,
		CreatedAtMax:      createdAtMax,
		IsCreatedAt:       isCreatedAt,
		UpdatedAtMin:      updatedAtMin,
		UpdatedAtMax:      updatedAtMax,
		IsUpdatedAt:       isUpdatedAt,
	}, nil
}

func buildHosDayRows(resp jsonAPIResponse) []hosDayRow {
	rows := make([]hosDayRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildHosDayRow(resource))
	}
	return rows
}

func buildHosDayRow(resource jsonAPIResource) hosDayRow {
	row := hosDayRow{
		ID:                resource.ID,
		ServiceDate:       formatDate(stringAttr(resource.Attributes, "service-date")),
		RegulationSetCode: stringAttr(resource.Attributes, "regulation-set-code"),
		TimeZoneID:        stringAttr(resource.Attributes, "time-zone-id"),
	}

	if rel, ok := resource.Relationships["driver"]; ok && rel.Data != nil {
		row.DriverID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}

	return row
}

func renderHosDaysTable(cmd *cobra.Command, rows []hosDayRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No HOS days found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tDRIVER\tSERVICE DATE\tREG SET\tTIME ZONE\tBROKER")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.DriverID,
			row.ServiceDate,
			row.RegulationSetCode,
			row.TimeZoneID,
			row.BrokerID,
		)
	}
	return writer.Flush()
}
