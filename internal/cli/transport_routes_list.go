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

type transportRoutesListOptions struct {
	BaseURL                 string
	Token                   string
	JSON                    bool
	NoAuth                  bool
	Limit                   int
	Offset                  int
	Sort                    string
	NearOriginLocation      string
	NearDestinationLocation string
}

type transportRouteRow struct {
	ID                   string  `json:"id"`
	OriginLatitude       string  `json:"origin_latitude,omitempty"`
	OriginLongitude      string  `json:"origin_longitude,omitempty"`
	DestinationLatitude  string  `json:"destination_latitude,omitempty"`
	DestinationLongitude string  `json:"destination_longitude,omitempty"`
	Miles                float64 `json:"miles,omitempty"`
	Minutes              float64 `json:"minutes,omitempty"`
}

func newTransportRoutesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List transport routes",
		Long: `List transport routes with filtering and pagination.

Output Columns:
  ID          Transport route identifier
  ORIGIN      Origin coordinates (lat,lng)
  DESTINATION Destination coordinates (lat,lng)
  MILES       Route distance in miles
  MINUTES     Estimated duration in minutes

Filters:
  --near-origin-location       Filter routes near origin (lat|lng|miles)
  --near-destination-location  Filter routes near destination (lat|lng|miles)`,
		Example: `  # List routes near an origin location
  xbe view transport-routes list --near-origin-location "40.7128|-74.0060|10"

  # List routes near a destination location
  xbe view transport-routes list --near-destination-location "34.0522|-118.2437|25"

  # Paginate results
  xbe view transport-routes list --limit 20 --offset 40

  # Output as JSON
  xbe view transport-routes list --json`,
		RunE: runTransportRoutesList,
	}
	initTransportRoutesListFlags(cmd)
	return cmd
}

func init() {
	transportRoutesCmd.AddCommand(newTransportRoutesListCmd())
}

func initTransportRoutesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("near-origin-location", "", "Filter by origin proximity (lat|lng|miles)")
	cmd.Flags().String("near-destination-location", "", "Filter by destination proximity (lat|lng|miles)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTransportRoutesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTransportRoutesListOptions(cmd)
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
	query.Set("fields[transport-routes]", "origin-latitude,origin-longitude,destination-latitude,destination-longitude,miles,minutes")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[near-origin-location]", opts.NearOriginLocation)
	setFilterIfPresent(query, "filter[near-destination-location]", opts.NearDestinationLocation)

	body, _, err := client.Get(cmd.Context(), "/v1/transport-routes", query)
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

	rows := buildTransportRouteRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTransportRoutesTable(cmd, rows)
}

func parseTransportRoutesListOptions(cmd *cobra.Command) (transportRoutesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	nearOriginLocation, _ := cmd.Flags().GetString("near-origin-location")
	nearDestinationLocation, _ := cmd.Flags().GetString("near-destination-location")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return transportRoutesListOptions{
		BaseURL:                 baseURL,
		Token:                   token,
		JSON:                    jsonOut,
		NoAuth:                  noAuth,
		Limit:                   limit,
		Offset:                  offset,
		Sort:                    sort,
		NearOriginLocation:      nearOriginLocation,
		NearDestinationLocation: nearDestinationLocation,
	}, nil
}

func buildTransportRouteRows(resp jsonAPIResponse) []transportRouteRow {
	rows := make([]transportRouteRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		rows = append(rows, transportRouteRow{
			ID:                   resource.ID,
			OriginLatitude:       stringAttr(attrs, "origin-latitude"),
			OriginLongitude:      stringAttr(attrs, "origin-longitude"),
			DestinationLatitude:  stringAttr(attrs, "destination-latitude"),
			DestinationLongitude: stringAttr(attrs, "destination-longitude"),
			Miles:                floatAttr(attrs, "miles"),
			Minutes:              floatAttr(attrs, "minutes"),
		})
	}
	return rows
}

func renderTransportRoutesTable(cmd *cobra.Command, rows []transportRouteRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No transport routes found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tORIGIN\tDESTINATION\tMILES\tMINUTES")
	for _, row := range rows {
		origin := formatTransportRouteCoordinatePair(row.OriginLatitude, row.OriginLongitude)
		destination := formatTransportRouteCoordinatePair(row.DestinationLatitude, row.DestinationLongitude)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			origin,
			destination,
			formatMiles(row.Miles),
			formatMinutes(row.Minutes),
		)
	}
	return writer.Flush()
}

func formatTransportRouteCoordinatePair(lat, lng string) string {
	lat = strings.TrimSpace(lat)
	lng = strings.TrimSpace(lng)
	switch {
	case lat == "" && lng == "":
		return ""
	case lat == "":
		return lng
	case lng == "":
		return lat
	default:
		return fmt.Sprintf("%s,%s", lat, lng)
	}
}

func formatMinutes(minutes float64) string {
	if minutes == 0 {
		return ""
	}
	return fmt.Sprintf("%.0f", minutes)
}
