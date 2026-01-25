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

type paveFrameActualHoursListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Sort    string
}

type paveFrameActualHourRow struct {
	ID          string `json:"id"`
	Date        string `json:"date,omitempty"`
	Hour        string `json:"hour,omitempty"`
	Window      string `json:"window,omitempty"`
	Latitude    string `json:"latitude,omitempty"`
	Longitude   string `json:"longitude,omitempty"`
	TempMinF    string `json:"temp_min_f,omitempty"`
	Precip1hrIn string `json:"precip_1hr_in,omitempty"`
}

func newPaveFrameActualHoursListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List pave frame actual hours",
		Long: `List pave frame actual hours with pagination.

Pave frame actual hours capture hourly paving conditions for specific
coordinates, including temperature and precipitation.

Output Columns:
  ID             Record identifier
  DATE           Date of the hour record
  HOUR           Hour of day (0-23)
  WINDOW         Window (day/night)
  LAT            Latitude
  LNG            Longitude
  TEMP_MIN_F     Minimum temperature (F)
  PRECIP_1HR_IN  Precipitation in the last hour (inches)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List pave frame actual hours
  xbe view pave-frame-actual-hours list

  # Output as JSON
  xbe view pave-frame-actual-hours list --json`,
		RunE: runPaveFrameActualHoursList,
	}
	initPaveFrameActualHoursListFlags(cmd)
	return cmd
}

func init() {
	paveFrameActualHoursCmd.AddCommand(newPaveFrameActualHoursListCmd())
}

func initPaveFrameActualHoursListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPaveFrameActualHoursList(cmd *cobra.Command, _ []string) error {
	opts, err := parsePaveFrameActualHoursListOptions(cmd)
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
	query.Set("fields[pave-frame-actual-hours]", "date,latitude,longitude,hour,window,temp-min-f,precip-1hr-in")

	sortValue := opts.Sort
	if sortValue == "" {
		sortValue = "date,hour"
	}
	query.Set("sort", sortValue)

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	body, _, err := client.Get(cmd.Context(), "/v1/pave-frame-actual-hours", query)
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

	rows := buildPaveFrameActualHourRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderPaveFrameActualHoursTable(cmd, rows)
}

func parsePaveFrameActualHoursListOptions(cmd *cobra.Command) (paveFrameActualHoursListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return paveFrameActualHoursListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
	}, nil
}

func buildPaveFrameActualHourRows(resp jsonAPIResponse) []paveFrameActualHourRow {
	rows := make([]paveFrameActualHourRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, paveFrameActualHourRow{
			ID:          resource.ID,
			Date:        stringAttr(resource.Attributes, "date"),
			Hour:        stringAttr(resource.Attributes, "hour"),
			Window:      stringAttr(resource.Attributes, "window"),
			Latitude:    stringAttr(resource.Attributes, "latitude"),
			Longitude:   stringAttr(resource.Attributes, "longitude"),
			TempMinF:    stringAttr(resource.Attributes, "temp-min-f"),
			Precip1hrIn: stringAttr(resource.Attributes, "precip-1hr-in"),
		})
	}
	return rows
}

func buildPaveFrameActualHourRowFromSingle(resp jsonAPISingleResponse) paveFrameActualHourRow {
	return paveFrameActualHourRow{
		ID:          resp.Data.ID,
		Date:        stringAttr(resp.Data.Attributes, "date"),
		Hour:        stringAttr(resp.Data.Attributes, "hour"),
		Window:      stringAttr(resp.Data.Attributes, "window"),
		Latitude:    stringAttr(resp.Data.Attributes, "latitude"),
		Longitude:   stringAttr(resp.Data.Attributes, "longitude"),
		TempMinF:    stringAttr(resp.Data.Attributes, "temp-min-f"),
		Precip1hrIn: stringAttr(resp.Data.Attributes, "precip-1hr-in"),
	}
}

func renderPaveFrameActualHoursTable(cmd *cobra.Command, rows []paveFrameActualHourRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No pave frame actual hours found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tDATE\tHOUR\tWINDOW\tLAT\tLNG\tTEMP_MIN_F\tPRECIP_1HR_IN")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Date,
			row.Hour,
			row.Window,
			row.Latitude,
			row.Longitude,
			row.TempMinF,
			row.Precip1hrIn,
		)
	}
	return writer.Flush()
}
